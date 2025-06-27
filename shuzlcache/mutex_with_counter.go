package shulzcache

import (
	"sync"
	"sync/atomic"
)

type MutexWithCounter struct {
	sync.Mutex
	counter int32
}

func NewMutexWithCounter() *MutexWithCounter {
	return &MutexWithCounter{
		Mutex:   sync.Mutex{},
		counter: 0,
	}
}

func (m *MutexWithCounter) Lock() {
	m.Mutex.Lock()
}

func (m *MutexWithCounter) Unlock() {
	m.Mutex.Unlock()
}

func (m *MutexWithCounter) Inc() int32 {
	return atomic.AddInt32(&m.counter, 1)
}

func (m *MutexWithCounter) Dec() int32 {
	return atomic.AddInt32(&m.counter, -1)
}
