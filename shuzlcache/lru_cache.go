package shulzcache

import (
	"sync"
	"time"
)

type CacheEntry struct {
	value     string
	createdAt time.Time
}

type LRUCache struct {
	cache      map[int]CacheEntry
	lock       sync.RWMutex
	lru        LRU
	maxEntries int
	ttl        time.Duration
}

// NewLRUCache creates a new LRUCache with the specified maximum number of entries and time-to-live (TTL) for cache entries.
func NewLRUCache(maxEntries int, ttl time.Duration) *LRUCache {
	return &LRUCache{
		cache:      make(map[int]CacheEntry),
		lru:        NewLRU(),
		maxEntries: maxEntries,
		ttl:        ttl,
	}
}

// GetIfFresh checks if the cache contains a fresh entry for the given parameter.
// It returns the value and a boolean indicating whether the entry exist and fresh (not expired).
// If the entry is not found or expired, it returns an empty string and false.
func (c *LRUCache) GetIfFresh(param int) (string, bool) {
	// First, check the cache with a read lock for fast access
	c.lock.RLock()
	defer c.lock.RUnlock()

	entry, ok := c.cache[param]
	if !ok || time.Since(entry.createdAt) >= c.ttl {
		return "", false
	}

	// Cache hit and not expired, update LRU
	c.lru.Hit(param)
	return entry.value, true
}

// Put adds or updates the cache entry for the given parameter with the specified value.
// It also deletes old entries if the cache exceeds the maximum size.
func (c *LRUCache) Put(param int, value string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	// Add or update the cache entry
	c.cache[param] = CacheEntry{
		value:     value,
		createdAt: time.Now(),
	}

	// Update the LRU cache
	c.lru.HitOrAdd(param)

	// Check LRU list size and remove old entries if necessary
	keysToDelete := c.lru.SizeTo(c.maxEntries)
	for _, key := range keysToDelete {
		delete(c.cache, key)
	}
}
