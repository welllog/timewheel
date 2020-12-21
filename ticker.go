package timewheel

import "github.com/welllog/timewheel/timing"

type Ticker struct {
	C     <-chan struct{}
	stop  chan struct{}
	timer timing.Timer
}

func (t *Ticker) Stop() {
	close(t.stop)
	t.timer.Stop()
}
