package timewheel

import (
	"sync/atomic"
	"time"
)

type TimeWheelPool struct {
	incr uint64
	size uint64
	pool []*TimeWheel
}

func NewTimeWheelPool(size int, interval time.Duration, slotNum int) *TimeWheelPool {
	twp := &TimeWheelPool{
		pool: make([]*TimeWheel, size),
		size: uint64(size),
	}

	for i := range twp.pool {
		twp.pool[i] = NewTimeWheel(interval, slotNum)
	}
	return twp
}

func (twp *TimeWheelPool) Start() {
	for _, tw := range twp.pool {
		tw.Start()
	}
}

func (twp *TimeWheelPool) Stop() {
	for _, tw := range twp.pool {
		tw.Stop()
	}
}

func (twp *TimeWheelPool) Get() *TimeWheel {
	return twp.pool[twp.getIndex()]
}

func (twp *TimeWheelPool) getIndex() (index uint64) {
	return atomic.AddUint64(&twp.incr, 1) % twp.size
}
