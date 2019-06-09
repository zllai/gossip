package ringbuffer

import (
	"errors"
	"fmt"
	"sync"
)

type RingBuffer struct {
	capacity int
	head     int
	len      int
	buf      [][]byte
	lock     *sync.Mutex
	cond     *sync.Cond
}

func New(capacity int) *RingBuffer {
	lock := &sync.Mutex{}
	cond := sync.NewCond(lock)
	return &RingBuffer{capacity, 0, 0, make([][]byte, capacity), lock, cond}
}

func (ring *RingBuffer) Push(data []byte) error {
	ring.lock.Lock()
	defer ring.lock.Unlock()
	index := (ring.head + ring.len) % ring.capacity
	ring.buf[index] = data
	ring.len++
	ring.cond.Signal()
	if ring.len > ring.capacity {
		ring.len = ring.capacity
		ring.head++
		return errors.New("Buffer full, truncate last message")
	} else {
		return nil
	}
}

func (ring *RingBuffer) Pop() []byte {
	ring.lock.Lock()
	defer ring.lock.Unlock()
	for ring.len == 0 {
		ring.cond.Wait()
	}
	index := (ring.head + ring.len - 1) % ring.capacity
	ret := ring.buf[index]
	ring.len--
	return ret
}

func (ring *RingBuffer) print() {
	for i := ring.head; i < ring.head+ring.len; i++ {
		fmt.Println(string(ring.buf[i%ring.capacity]))
	}
}
