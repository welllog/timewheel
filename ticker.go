package timewheel

type Ticker struct {
	C      <-chan struct{}
	stop   chan struct{}
	taskid TaskId
	tw     *TimeWheel
}

func (t *Ticker) Stop() {
	close(t.stop)
	t.tw.Remove(t.taskid)
}
