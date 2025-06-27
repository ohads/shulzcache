package shulzcache

import (
	"container/list"
	"sync"
)

type LRU interface {
	Hit(key int) bool
	HitOrAdd(key int)
	SizeTo(max int) []int
}

// LinkedListLRU implements a simple LRU cache using a doubly linked list and a map.
// It allows for O(1) time complexity for adding, deleting, and hitting keys.
type LinkedListLRU struct {
	aList *list.List
	aMap  map[int]*list.Element
	lock  sync.Mutex
}

// NewLRU creates a new instance of LinkedListLRU.
// It initializes the linked list and the map that will hold the keys and their corresponding list elements
func NewLRU() LRU {
	return &LinkedListLRU{
		aList: list.New(),
		aMap:  make(map[int]*list.Element),
	}
}

// Hit checks if the key exists in the LRU cache.
// If it exists, it moves the element to the front of the list to mark it as recently used.
// It locks the cache to ensure thread safety while checking the key
func (lru *LinkedListLRU) Hit(key int) bool {
	lru.lock.Lock()
	defer lru.lock.Unlock()

	return lru.hit(key)
}

// HitOrAdd checks if the key exists in the LRU cache.
// If it exists, it marks it as recently used.
// If it does not exist, it adds the key to the cache.
// It locks the cache to ensure thread safety while checking or adding the key.
func (lru *LinkedListLRU) HitOrAdd(key int) {
	lru.lock.Lock()
	defer lru.lock.Unlock()

	if !(lru.hit(key)) {
		lru.add(key)
	}
}

// SizeTo checks the current size of the LRU cache and removes the oldest entries
// until the size is less than or equal to the specified maximum size.
// It returns a slice of keys that were removed from the cache.
// It locks the cache to ensure thread safety while checking the size and removing elements.
// If the current size is less than or equal to the maximum size, it returns an empty slice.
func (lru *LinkedListLRU) SizeTo(max int) []int {
	lru.lock.Lock()
	defer lru.lock.Unlock()

	// get num of elements to remove to fit max size.
	if lru.aList.Len() <= max {
		return []int{}
	}

	toRemove := lru.aList.Len() - max
	keys := make([]int, 0, toRemove)

	for i := 0; i < toRemove; i++ {
		elem := lru.aList.Back()
		if elem == nil {
			break
		}

		key := elem.Value.(int)
		keys = append(keys, key)
		lru.delete(key)
	}

	return keys
}

func (lru *LinkedListLRU) add(key int) {
	element := lru.aList.PushFront(key)
	lru.aMap[key] = element
}

func (lru *LinkedListLRU) delete(key int) {
	if elem, ok := lru.aMap[key]; ok {
		lru.aList.Remove(elem)
		delete(lru.aMap, key)
	}
}

func (lru *LinkedListLRU) hit(key int) bool {
	element, ok := lru.aMap[key]
	if ok {
		lru.aList.MoveToFront(element)
	}
	return ok
}
