package timewheel

import (
	"sync"
	"time"

	"github.com/welllog/timewheel/timing"
	"github.com/welllog/timewheel/timing/dqdriver"
)

type DRIVER string

type defscheduler struct {
	delay time.Duration
}

func (s *defscheduler) Next(now time.Time) time.Time {
	return now.Add(s.delay)
}

const (
	DELAY_QUEUE_DRV DRIVER = "delay_queue"
)

var Timing timing.Timing

func InitTiming(tick time.Duration, slotNum int, drive DRIVER) {
	Timing = dqdriver.NewTimingWheel(tick, slotNum)
	Timing.Start()
}

func StopTiming() {
	Timing.Stop()
}

func NewTimer(d time.Duration) *Timer {
	c := make(chan struct{}, 1)
	t := Timing.AddTask(d, func() {
		c <- struct{}{}
	})
	return &Timer{
		C:      c,
		recv:   c,
		timing: Timing,
		timer:  t,
	}
}

func NewTicker(d time.Duration) *Ticker {
	c := make(chan struct{})
	stop := make(chan struct{})

	t := Timing.ScheduleTask(&defscheduler{delay: d}, func() {
		select {
		case c <- struct{}{}:
		case <-stop:
		}
	})
	return &Ticker{
		C:     c,
		stop:  stop,
		timer: t,
	}
}

func Sleep(d time.Duration) {
	w := &sync.WaitGroup{}
	w.Add(1)
	Timing.AddTask(d, func() { w.Done() })
	w.Wait()
}

func After(d time.Duration) <-chan time.Time {
	c := make(chan time.Time, 1)
	Timing.AddTask(d, func() { c <- time.Now() })
	return c
}

func AfterFunc(d time.Duration, f func()) *Timer {
	c := make(chan struct{}, 1)
	t := Timing.AddTask(d, func() {
		f()
		c <- struct{}{}
	})
	return &Timer{
		C:      c,
		recv:   c,
		fn:     f,
		timing: Timing,
		timer:  t,
	}
}
