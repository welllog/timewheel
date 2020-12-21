package timing

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type DelayQueue interface {
	Offer(elem interface{}, expiration time.Time)
	Poll(ctx context.Context, nowF func() time.Time)
	Chan() <-chan interface{}
	Size() int
}

func NewDelayQueue(capacity int, precision time.Duration) DelayQueue {
	return &delayQueue{
		C:         make(chan interface{}, capacity),
		pq:        NewPriorityQueue(capacity),
		precision: precision,
		wakeupC:   make(chan struct{}),
	}
}

type delayQueue struct {
	C         chan interface{}
	mu        sync.Mutex
	pq        PriorityQueue
	sleeping  int32
	precision time.Duration
	wakeupC   chan struct{}
}

func (q *delayQueue) Offer(elem interface{}, expiration time.Time) {
	q.mu.Lock()
	index := q.pq.Add(elem, expiration.UnixNano()/int64(q.precision))
	q.mu.Unlock()

	if index == 0 {
		if atomic.CompareAndSwapInt32(&q.sleeping, 1, 0) {
			q.wakeupC <- struct{}{}
		}
	}
}

func (q *delayQueue) Poll(ctx context.Context, nowF func() time.Time) {
	for {
		now := nowF().UnixNano() / int64(q.precision)

		q.mu.Lock()
		elem, exp := q.pq.PriorityShift(now)
		if elem == nil {
			atomic.StoreInt32(&q.sleeping, 1)
			// StoreInt32放在锁中，防止队列为空时，同时发生Offer操作,而Offer中的CAS先执行而未向wakeupC发信号,
			// 导致下面判断exp为0而长时间阻塞
		}
		q.mu.Unlock()

		if elem == nil {
			if exp == 0 {
				select {
				case <-q.wakeupC:
					continue
				case <-ctx.Done():
					if atomic.SwapInt32(&q.sleeping, 0) == 0 {
						<-q.wakeupC
					}
					return
				}
			} else {
				select {
				case <-q.wakeupC:
					continue
				case <-time.After(time.Duration(exp-now) * q.precision):
					if atomic.SwapInt32(&q.sleeping, 0) == 0 {
						<-q.wakeupC
					}
					continue
				case <-ctx.Done():
					if atomic.SwapInt32(&q.sleeping, 0) == 0 {
						<-q.wakeupC
					}
					return
				}
			}
		}

		select {
		case q.C <- elem:
		case <-ctx.Done():
			return
		}
	}
}

func (q *delayQueue) Chan() <-chan interface{} {
	return q.C
}

func (q *delayQueue) Size() int {
	q.mu.Lock()
	s := q.pq.Size()
	q.mu.Unlock()
	return s
}
