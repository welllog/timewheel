package dqdriver

import (
	"sync"
	"sync/atomic"
)

type bucket struct {
	expiration int64
	mu         sync.Mutex
	root       elem
	last       *elem
}

type elem struct {
	value *timer
	next  *elem
}

func (e *elem) Next() *elem {
	return e.next
}

func newBucket() *bucket {
	b := &bucket{
		expiration: -1,
	}
	b.last = &b.root
	return b
}

func (b *bucket) push(e *elem) {
	b.last.next = e
	b.last = e
}

func (b *bucket) removeFirst() {
	rm := b.root.next
	if rm == nil {
		return
	}
	b.root.next = rm.next
	rm.next = nil
	if b.root.next == nil { // 移除的是最后一个元素
		b.last = &b.root
	}
}

func (b *bucket) Front() *elem {
	return b.root.next
}

func (b *bucket) Expiration() int64 {
	return atomic.LoadInt64(&b.expiration)
}

func (b *bucket) SetExpiration(expiration int64) bool {
	return atomic.SwapInt64(&b.expiration, expiration) != expiration
}

func (b *bucket) Add(t *timer) {
	b.mu.Lock()

	if t.elem == nil {
		t.elem = &elem{
			value: t,
		}
	}

	b.push(t.elem)

	b.mu.Unlock()
}

func (b *bucket) Flush(reinsert func(*timer)) {
	var ts []*timer

	b.mu.Lock()
	for e := b.Front(); e != nil; {
		next := e.Next()

		b.removeFirst()

		if !e.value.isStop() {
			ts = append(ts, e.value)
		}

		e = next
	}
	b.mu.Unlock()

	b.SetExpiration(-1)
	for _, t := range ts {
		reinsert(t)
	}
}
