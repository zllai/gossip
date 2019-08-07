package gossip

import (
	"container/list"
	"sync"
)

type RingBuffer struct {
	capacity int
	buf      *list.List
	lock     *sync.Mutex
	cond     *sync.Cond
}

func NewRingBuffer(cap int) *RingBuffer {
	lock := &sync.Mutex{}
	cond := sync.NewCond(lock)
	return &RingBuffer{
		capacity: cap,
		buf:      list.New(),
		lock:     lock,
		cond:     cond,
	}
}

func (r *RingBuffer) Len() int {
	return r.buf.Len()
}

func (r *RingBuffer) Push(value interface{}) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.buf.PushBack(value)
	r.cond.Signal()
	if r.buf.Len() > r.capacity {
		r.buf.Remove(r.buf.Front())
	}
}

func (r *RingBuffer) Pop() interface{} {
	r.lock.Lock()
	defer r.lock.Unlock()
	for r.buf.Len() == 0 {
		r.cond.Wait()
	}
	ret := r.buf.Front().Value
	r.buf.Remove(r.buf.Front())
	return ret
}

func (r *RingBuffer) RemoveIf(filter func(interface{}) bool) {
	r.lock.Lock()
	defer r.lock.Unlock()
	for e := r.buf.Front(); e != nil; {
		next := e.Next()
		if !filter(e.Value) {
			r.buf.Remove(e)
		}
		e = next
	}
}

func (r *RingBuffer) Do(do func(interface{})) {
	r.lock.Lock()
	defer r.lock.Unlock()
	for e := r.buf.Front(); e != nil; e = e.Next() {
		do(e.Value)
	}
}

func (r *RingBuffer) DoRemove(do func(interface{})) {
	r.lock.Lock()
	defer r.lock.Unlock()
	for e := r.buf.Front(); e != nil; {
		next := e.Next()
		do(e.Value)
		r.buf.Remove(e)
		e = next
	}
}
