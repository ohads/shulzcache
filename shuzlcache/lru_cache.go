package shulzcache

import (
	"sync"
	"time"
)

type CacheEntry struct {
	value     string
	timestamp time.Time
}

type LRUCache struct {
	cache      map[int]CacheEntry
	lock       sync.RWMutex
	lru        *LRU
	maxEntries int
	ttl        time.Duration
}

func NewLRUCache(maxEntries int, ttl time.Duration) *LRUCache {
	return &LRUCache{
		cache:      make(map[int]CacheEntry),
		lru:        NewLRU(),
		maxEntries: maxEntries,
		ttl:        ttl,
	}
}

func (c *LRUCache) GetIfFresh(param int) (string, bool) {
	// First, check the cache with a read lock for fast access
	c.lock.RLock()
	defer c.lock.RUnlock()

	entry, ok := c.cache[param]
	expired := ok && time.Since(entry.timestamp) >= c.ttl
	if ok && !expired {
		// Cache hit and not expired, release read lock first, then update LRU
		c.lru.Hit(param)
		return entry.value, true
	}

	// If expired or not found, release the read lock and acquire write lock
	return "", false
}

func (c *LRUCache) Put(param int, value string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	// Add or update the cache entry
	c.cache[param] = CacheEntry{
		value:     value,
		timestamp: time.Now(),
	}

	// Update the LRU cache
	c.lru.HitOrAdd(param)

	// Check LRU list size and remove old entries if necessary
	keysToDelete := c.lru.SizeTo(c.maxEntries)
	for _, key := range keysToDelete {
		delete(c.cache, key)
	}
}
