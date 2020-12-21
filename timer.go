package timewheel

import (
	"github.com/welllog/timewheel/timing"
	"time"
)

type Timer struct {
	C      <-chan struct{}
	recv   chan<- struct{}
	fn     func()
	timing timing.Timing
	timer  timing.Timer
}

func (t *Timer) Stop() bool {
	return t.timer.Stop()
}

// 	if !t.Stop() {
// 		<-t.C
// 	}
// 	t.Reset(d)
func (t *Timer) Reset(d time.Duration) {
	if t.fn != nil {
		t.timer = t.timing.AddTask(d, func() {
			t.fn()
			t.recv <- struct{}{}
		})
		return
	}

	t.timer = t.timing.AddTask(d, func() {
		t.recv <- struct{}{}
	})
}
