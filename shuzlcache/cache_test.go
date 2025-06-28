package shulzcache

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	var counter int32 = 0

	f := func(param int) (string, error) {
		atomic.AddInt32(&counter, 1)
		time.Sleep(2 * time.Millisecond)
		// Simulate a long-running operation
		return fmt.Sprintf("Result for ID %d", param), nil
	}
	cachedFunc := NewCachedFunction(f)
	cachedFunc(1)
	if atomic.LoadInt32(&counter) != 1 {
		t.Errorf("Expected counter to be 1, got %d", atomic.LoadInt32(&counter))
	}

	cachedFunc(1)
	if atomic.LoadInt32(&counter) != 1 {
		t.Errorf("Expected counter to be 1, got %d", atomic.LoadInt32(&counter))
	}
}

func TestExpire(t *testing.T) {
	var counter int32 = 0

	f := func(param int) (string, error) {
		atomic.AddInt32(&counter, 1)
		time.Sleep(2 * time.Millisecond)
		// Simulate a long-running operation
		return fmt.Sprintf("Result for ID %d", param), nil
	}
	cachedFunc := NewCachedFunctionWithOptions(f, 10, 10*time.Millisecond)

	cachedFunc(1)
	if atomic.LoadInt32(&counter) != 1 {
		t.Errorf("Expected counter to be 1, got %d", atomic.LoadInt32(&counter))
	}

	cachedFunc(1)
	if atomic.LoadInt32(&counter) != 1 {
		t.Errorf("Expected counter to be 1, got %d", atomic.LoadInt32(&counter))
	}

	time.Sleep(11 * time.Millisecond) // Wait for the cache to expire

	cachedFunc(1)
	if atomic.LoadInt32(&counter) != 2 {
		t.Errorf("Expected counter to be 2 after expiration, got %d", atomic.LoadInt32(&counter))
	}
}

func TestConcurrentCalls(t *testing.T) {
	var counter int32 = 0

	f := func(param int) (string, error) {
		atomic.AddInt32(&counter, 1)
		time.Sleep(2 * time.Millisecond)
		// Simulate a long-running operation
		return fmt.Sprintf("Result for ID %d", param), nil
	}
	cachedFunc := NewCachedFunctionWithOptions(f, 10, 10*time.Millisecond)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			cachedFunc(i % 2)
		}(i)
	}
	wg.Wait()

	if atomic.LoadInt32(&counter) != 2 {
		t.Errorf("Expected counter to be 2 after concurrent calls, got %d", atomic.LoadInt32(&counter))
	}
}

func TestCacheExceeds(t *testing.T) {
	var counter int32 = 0

	f := func(param int) (string, error) {
		atomic.AddInt32(&counter, 1)
		time.Sleep(2 * time.Millisecond)
		// Simulate a long-running operation
		return fmt.Sprintf("Result for ID %d", param), nil
	}

	cache := NewLRUCache(10, 10*time.Millisecond)
	cachedFunc := NewCachedFunctionWithCache(f, cache)

	for i := 0; i < 20; i++ {
		cachedFunc(i)
	}

	// check cache size
	t.Log("Cache size after adding 20 items:", len(cache.cache))
	if len(cache.cache) > 10 {
		t.Errorf("Expected cache size to be at most 10, got %d", len(cache.cache))
	}

	//wait for cache to expire
	time.Sleep(11 * time.Millisecond)

	for i := 0; i < 20; i++ {
		cachedFunc(100 + i)
	}

	// check cache size
	t.Log("Cache size after adding another 20 items:", len(cache.cache))
	if len(cache.cache) > 10 {
		t.Errorf("Expected cache size to be at most 10, got %d", len(cache.cache))
	}

}

func TestOldestEntriesEvicted(t *testing.T) {
	var counter int32 = 0

	f := func(param int) (string, error) {
		atomic.AddInt32(&counter, 1)
		time.Sleep(2 * time.Millisecond)
		// Simulate a long-running operation
		return fmt.Sprintf("Result for ID %d", param), nil
	}

	cache := NewLRUCache(10, 10*time.Millisecond)
	cachedFunc := NewCachedFunctionWithCache(f, cache)

	for i := 0; i < 10; i++ {
		cachedFunc(i)
	}

	cachedFunc(11)

	if _, ok := cache.cache[0]; ok {
		t.Errorf("Expected entry for key 0 to be evicted, but it still exists")
	}

	if _, ok := cache.cache[1]; !ok {
		t.Errorf("Expected entry for key 1 to not be evicted, but it not exists")
	}

}

const (
	benchmarkSleep = 10 * time.Millisecond
)

func BenchmarkDirectFunction(b *testing.B) {
	var counter int32 = 0

	f := func(param int) (string, error) {
		atomic.AddInt32(&counter, 1)
		time.Sleep(benchmarkSleep)
		// Simulate a long-running operation
		return fmt.Sprintf("Result for ID %d", param), nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f(i % 10)
	}
}

func BenchmarkCachedFunctionCold(b *testing.B) {
	var counter int32 = 0

	f := func(param int) (string, error) {
		atomic.AddInt32(&counter, 1)
		time.Sleep(benchmarkSleep)
		// Simulate a long-running operation
		return fmt.Sprintf("Result for ID %d", param), nil
	}
	cachedFunc := NewCachedFunction(f)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cachedFunc(i % 10)
	}
}

func BenchmarkCachedFunctionWarm(b *testing.B) {
	var counter int32 = 0

	f := func(param int) (string, error) {
		atomic.AddInt32(&counter, 1)
		time.Sleep(benchmarkSleep)
		// Simulate a long-running operation
		return fmt.Sprintf("Result for ID %d", param), nil
	}
	cachedFunc := NewCachedFunction(f)

	for i := 0; i < b.N; i++ {
		cachedFunc(i % 10)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cachedFunc(i % 10)
	}
}

func BenchmarkHighConcurrency(b *testing.B) {
	var counter int32 = 0

	f := func(param int) (string, error) {
		atomic.AddInt32(&counter, 1)
		time.Sleep(benchmarkSleep)
		// Simulate a long-running operation
		return fmt.Sprintf("Result for ID %d", param), nil
	}
	cachedFunc := NewCachedFunction(f)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cachedFunc(1)
		}
	})
}
