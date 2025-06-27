package shulzcache

import (
	"container/list"
	"sync"
)

type LRU struct {
	aList *list.List
	aMap  map[int]*list.Element
	lock  sync.Mutex
}

func NewLRU() *LRU {
	return &LRU{
		aList: list.New(),
		aMap:  make(map[int]*list.Element),
	}
}

func (lru *LRU) Add(key int) {
	lru.lock.Lock()
	defer lru.lock.Unlock()

	lru.add(key)
}

func (lru *LRU) add(key int) {
	element := lru.aList.PushFront(key)
	lru.aMap[key] = element
}

func (lru *LRU) Delete(key int) {
	lru.lock.Lock()
	defer lru.lock.Unlock()

	if elem, ok := lru.aMap[key]; ok {
		lru.aList.Remove(elem)
		delete(lru.aMap, key)
	}
}

func (lru *LRU) Hit(key int) bool {
	lru.lock.Lock()
	defer lru.lock.Unlock()

	return lru.hit(key)
}

func (lru *LRU) hit(key int) bool {
	element, ok := lru.aMap[key]
	if ok {
		lru.aList.MoveToFront(element)
	}
	return ok
}

func (lru *LRU) HitOrAdd(key int) {
	lru.lock.Lock()
	defer lru.lock.Unlock()

	if !(lru.hit(key)) {
		lru.add(key)
	}
}

func (lru *LRU) SizeTo(max int) []int {
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
		lru.aList.Remove(elem)
		delete(lru.aMap, key)
	}

	return keys
}
