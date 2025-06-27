package shulzcache

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

const (
	maxEntries = 1000
	ttl        = 5 * time.Minute
)

type cachedFunction func(param int) (string, error)

type Cache interface {
	GetIfFresh(param int) (string, bool)
	Put(param int, value string)
}

type MutexWithCounter struct {
	sync.Mutex
	counter int32
}

func NewCachedFunction(aFunc cachedFunction) cachedFunction {
	return NewCachedFunctionWithOptions(aFunc, maxEntries, ttl)
}

func NewCachedFunctionWithOptions(aFunc cachedFunction, maxEntries int, ttl time.Duration) cachedFunction {
	return NewCachedFunctionWithCache(aFunc, NewLRUCache(maxEntries, ttl))
}

func NewCachedFunctionWithCache(aFunc cachedFunction, cache Cache) cachedFunction {
	var running sync.Map

	newFunc := func(param int) (string, error) {
		// try to read the value from cache.
		val, ok := cache.GetIfFresh(param)
		if ok {
			return val, nil
		}

		// acquire per key lock to ensure only one goroutine runs the function for this key at a time.
		lockAsAny, _ := running.LoadOrStore(param, &MutexWithCounter{
			Mutex:   sync.Mutex{},
			counter: 0,
		})

		lock, ok := lockAsAny.(*MutexWithCounter)
		if !ok {
			return "", fmt.Errorf("failed to cast to *MutexWithCounter")
		}

		atomic.AddInt32(&lock.counter, 1)
		lock.Lock()
		defer func() {
			lock.Unlock()
			if atomic.AddInt32(&lock.counter, -1) == 0 {
				running.Delete(param)
			}
		}()

		// try to read the value from cache again, after we have the lock.
		val, ok = cache.GetIfFresh(param)
		if ok {
			return val, nil
		}

		// not in cache, we need to run the function and store the result.
		newVal, err := aFunc(param)
		cache.Put(param, newVal)
		return newVal, err

	}

	return newFunc

}
