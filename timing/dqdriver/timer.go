package dqdriver

import (
	"github.com/welllog/timewheel/timing"
	"sync/atomic"
)

type timer struct {
	expiration int64
	state      int32
	task       func()
	next       *timer
}

func (t *timer) Stop() bool {
	return atomic.SwapInt32(&t.state, 1) == 0
}

func (t *timer) Next() *timer {
	return t.next
}

func (t *timer) run(wrapper *timing.WaitGroupWrapper) bool {
	if atomic.CompareAndSwapInt32(&t.state, 0, 2) {
		wrapper.Wrap(t.task)
		return true
	}
	return false
}

func (t *timer) resetState() bool {
	return atomic.CompareAndSwapInt32(&t.state, 2, 0)
}

func (t *timer) isStop() bool {
	return atomic.LoadInt32(&t.state) == 1
}
