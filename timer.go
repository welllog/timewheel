package timewheel

import "time"

type Timer struct {
	C      <-chan struct{}
	recv   chan<- struct{}
	taskid TaskId
	fn     func()
	tw     *TimeWheel
}

func (t *Timer) Stop() bool {
	return !t.tw.RemoveAndHasRun(t.taskid)
}

// 	if !t.Stop() {
// 		<-t.C
// 	}
// 	t.Reset(d)
func (t *Timer) Reset(d time.Duration) {
	if t.fn != nil {
		t.taskid = t.tw.addTask(d, func() {
			t.fn()
			t.recv <- struct{}{}
		}, 1, true)
		return
	}

	t.taskid = t.tw.addTask(d, func() {
		t.recv <- struct{}{}
	}, 1, true)
}
