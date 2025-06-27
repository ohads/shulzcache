package shulzcache

import (
	"fmt"
	"sync"
	"time"
)

const (
	maxEntries = 1000
	ttl        = 5 * time.Minute
)

type CachedFunction func(param int) (string, error)

type Cache interface {
	GetIfFresh(param int) (string, bool)
	Put(param int, value string)
}

// NewCachedFunction creates a new CachedFunction that caches the results of the provided function.
// It uses an LRU cache with a default maximum number of entries (1000) and TTL (5 Minutes) for cache entries.
// The function is designed to handle concurrent calls efficiently, ensuring that only one goroutine runs the function for a given key at a time.
func NewCachedFunction(aFunc CachedFunction) CachedFunction {
	return NewCachedFunctionWithOptions(aFunc, maxEntries, ttl)
}

// NewCachedFunctionWithOptions creates a new CachedFunction with specified options for maximum entries and TTL.
func NewCachedFunctionWithOptions(aFunc CachedFunction, maxEntries int, ttl time.Duration) CachedFunction {
	return NewCachedFunctionWithCache(aFunc, NewLRUCache(maxEntries, ttl))
}

// NewCachedFunctionWithCache creates a new CachedFunction that uses the provided cache.
func NewCachedFunctionWithCache(aFunc CachedFunction, cache Cache) CachedFunction {
	var running sync.Map

	newFunc := func(param int) (string, error) {
		// fast path - try to read the value from cache.
		val, ok := cache.GetIfFresh(param)
		if ok {
			return val, nil
		}

		// acquire per key lock to ensure only one goroutine runs the function for this key at a time.
		lockAsAny, _ := running.LoadOrStore(param, NewMutexWithCounter())

		lock, ok := lockAsAny.(*MutexWithCounter)
		if !ok {
			return "", fmt.Errorf("failed to cast to *MutexWithCounter")
		}

		lock.Inc() // increment the counter to track how many goroutines are waiting for this lock.
		lock.Lock()
		defer func() {
			lock.Unlock()
			if lock.Dec() == 0 {
				running.Delete(param)
			}
		}()

		// try to read the value from cache again, after we acquired the lock.
		val, ok = cache.GetIfFresh(param)
		if ok {
			return val, nil
		}

		// not in cache, we need to run the function and store the result.
		newVal, err := aFunc(param)
		if err != nil {
			return "", err
		}

		cache.Put(param, newVal)
		return newVal, nil

	}

	return newFunc

}
