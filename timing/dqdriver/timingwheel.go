package dqdriver

import (
	"context"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/welllog/timewheel/timing"
)

type timingWheel struct {
	tick          int64
	slotNum       int64
	interval      int64
	curTime       int64
	slots         []*bucket
	queue         timing.DelayQueue
	overflowWheel unsafe.Pointer
	exitC         chan struct{}
	waitGroup     timing.WaitGroupWrapper
}

func NewTimingWheel(tick time.Duration, slotNum int) timing.Timing {
	if tick < time.Millisecond {
		panic("tick must be greater than or equal to 1ms")
	}
	return newTimingWheel(int64(tick), int64(slotNum), truncate(time.Now().UnixNano(), int64(tick)),
		timing.NewDelayQueue(slotNum, tick))
}

func newTimingWheel(tick, slotNum, curTime int64, dq timing.DelayQueue) *timingWheel {
	buckets := make([]*bucket, slotNum)
	for i := range buckets {
		buckets[i] = newBucket()
	}
	return &timingWheel{
		tick:     tick,
		slotNum:  slotNum,
		interval: tick * slotNum,
		curTime:  curTime,
		slots:    buckets,
		queue:    dq,
		exitC:    make(chan struct{}),
	}
}

func (tw *timingWheel) add(t *timer) bool {
	curTime := atomic.LoadInt64(&tw.curTime)
	if t.expiration < curTime+tw.tick {
		return false
	} else if t.expiration < curTime+tw.interval {
		virtualID := t.expiration / tw.tick
		b := tw.slots[virtualID%tw.slotNum]
		b.Add(t)

		if b.SetExpiration(virtualID * tw.tick) {
			tw.queue.Offer(b, time.Unix(0, b.Expiration()))
		}

		return true
	} else {
		overflowWheel := atomic.LoadPointer(&tw.overflowWheel)
		if overflowWheel == nil {
			atomic.CompareAndSwapPointer(
				&tw.overflowWheel,
				nil,
				unsafe.Pointer(newTimingWheel(tw.interval, tw.slotNum, curTime, tw.queue)),
			)
			overflowWheel = atomic.LoadPointer(&tw.overflowWheel)
		}
		return (*timingWheel)(overflowWheel).add(t)
	}
}

func (tw *timingWheel) addOrRun(t *timer) {
	if !tw.add(t) {
		t.run(&tw.waitGroup)
	}
}

func (tw *timingWheel) advanceClock(expiration int64) {
	curTime := atomic.LoadInt64(&tw.curTime)
	if expiration >= curTime+tw.tick {
		curTime = truncate(expiration, tw.tick)
		atomic.StoreInt64(&tw.curTime, curTime)

		overflowWheel := atomic.LoadPointer(&tw.overflowWheel)
		if overflowWheel != nil {
			(*timingWheel)(overflowWheel).advanceClock(curTime)
		}
	}
}

func (tw *timingWheel) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	tw.waitGroup.Wrap(func() {
		tw.queue.Poll(ctx, time.Now)
	})

	tw.waitGroup.Wrap(func() {
		ch := tw.queue.Chan()
		for {
			select {
			case elem := <-ch:
				b := elem.(*bucket)
				tw.advanceClock(b.Expiration())
				b.Flush(tw.addOrRun)
			case <-tw.exitC:
				cancel()
				return
			}
		}
	})
}

func (tw *timingWheel) Stop() {
	close(tw.exitC)

	complete := make(chan struct{})
	go func() {
		tw.waitGroup.Wait()
		complete <- struct{}{}
	}()

	select {
	case <-complete:
	case <-time.After(8 * time.Second):
	}
}

func (tw *timingWheel) AddTask(delay time.Duration, task func()) timing.Timer {
	t := &timer{task: task, expiration: time.Now().Add(delay).UnixNano()}
	tw.addOrRun(t)
	return t
}

func (tw *timingWheel) ScheduleTask(s timing.Scheduler, task func()) timing.Timer {
	t := &timer{}
	expiration := s.Next(time.Now())
	if expiration.IsZero() {
		return t
	}

	t.expiration = expiration.UnixNano()
	t.task = func() {
		task()

		nexpiration := s.Next(time.Now())
		if !nexpiration.IsZero() && t.resetState() {
			t.expiration = nexpiration.UnixNano()
			tw.addOrRun(t)
		}
	}
	tw.addOrRun(t)

	return t
}

func truncate(x, m int64) int64 {
	if m <= 0 {
		return x
	}
	return x - x%m
}
