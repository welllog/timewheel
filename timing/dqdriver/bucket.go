package dqdriver

import (
	"sync"
	"sync/atomic"
)

type bucket struct {
	expiration int64
	mu         sync.Mutex
	root       timer
	last       *timer
}

func newBucket() *bucket {
	b := &bucket{
		expiration: -1,
	}
	b.last = &b.root
	return b
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

func (b *bucket) Front() *timer {
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

	b.last.next = t
	b.last = t

	b.mu.Unlock()
}

func (b *bucket) Flush(reinsert func(*timer)) {
	var ts []*timer

	b.mu.Lock()
	for t := b.Front(); t != nil; {
		next := t.Next()

		b.removeFirst()

		if !t.isStop() {
			ts = append(ts, t)
		}

		t = next
	}
	b.SetExpiration(-1)
	b.mu.Unlock()

	for _, t := range ts {
		reinsert(t)
	}
}
