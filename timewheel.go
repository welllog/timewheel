package timewheel

import (
	"container/list"
	"sync"
	"sync/atomic"
	"time"
)

type TimeWheel struct {
	randomID uint64
	interval time.Duration
	ticker   *time.Ticker

	slotNum    int
	slots      []*list.List
	currentPos int

	onceStart  sync.Once
	addTaskC   chan *task
	stopC      chan struct{}
	taskRecord sync.Map
}

func NewTimeWheel(interval time.Duration, slotNum int) *TimeWheel {
	if interval.Seconds() < 0.1 {
		panic("invalid params, must tick >= 100 ms")
	}
	if slotNum <= 0 {
		panic("invalid params, must bucketsNum > 0")
	}

	tw := &TimeWheel{
		interval: interval,
		slotNum:  slotNum,
		slots:    make([]*list.List, slotNum),
		addTaskC: make(chan *task, 1024),
		stopC:    make(chan struct{}),
	}

	for i := range tw.slots {
		tw.slots[i] = list.New()
	}

	return tw
}

func (tw *TimeWheel) Start() {
	tw.onceStart.Do(func() {
		tw.ticker = time.NewTicker(tw.interval)
		go tw.run()
	})
}

func (tw *TimeWheel) Stop() {
	tw.stopC <- struct{}{}
}

func (tw *TimeWheel) AddOnce(delay time.Duration, callback func()) TaskId {
	return tw.addTask(delay, callback, 1, true)
}

func (tw *TimeWheel) AddWithTimes(delay time.Duration, times int32, callback func()) TaskId {
	if times <= 0 {
		times = 1
	}
	return tw.addTask(delay, callback, times, true)
}

func (tw *TimeWheel) AddCron(delay time.Duration, callback func()) TaskId {
	return tw.addTask(delay, callback, timesUnlimit, true)
}

func (tw *TimeWheel) Remove(id TaskId) {
	val, ok := tw.taskRecord.Load(id)
	if ok {
		obj := val.(*task)
		obj.Stop(id)
	}
}

// not idempotent, the second call may get the wrong return value
// so it is better to remove it once if you need use the return value.
func (tw *TimeWheel) RemoveAndHasRun(id TaskId) bool {
	val, ok := tw.taskRecord.Load(id)
	if ok {
		obj := val.(*task)
		return obj.Stop(id)
	}
	return true
}

func (tw *TimeWheel) NewTimer(d time.Duration) *Timer {
	c := make(chan struct{}, 1)
	id := tw.addTask(d, func() { c <- struct{}{} }, 1, true)
	return &Timer{
		C:      c,
		recv:   c,
		taskid: id,
		tw:     tw,
	}
}

func (tw *TimeWheel) NewTicker(d time.Duration) *Ticker {
	c := make(chan struct{})
	stop := make(chan struct{})
	id := tw.addTask(d, func() {
		select {
		case c <- struct{}{}:
		case <-stop:
		}
	}, timesUnlimit, false)
	return &Ticker{
		C:      c,
		stop:   stop,
		taskid: id,
		tw:     tw,
	}
}

func (tw *TimeWheel) Sleep(d time.Duration) {
	c := make(chan struct{}, 1)
	tw.addTask(d, func() { c <- struct{}{} }, 1, true)
	<-c
}

func (tw *TimeWheel) After(d time.Duration) <-chan time.Time {
	c := make(chan time.Time, 1)
	tw.addTask(d, func() { c <- time.Now() }, 1, true)
	return c
}

func (tw *TimeWheel) AfterFunc(d time.Duration, f func()) *Timer {
	c := make(chan struct{}, 1)
	id := tw.addTask(d, func() {
		f()
		c <- struct{}{}
	}, 1, true)
	return &Timer{
		C:      c,
		recv:   c,
		taskid: id,
		tw:     tw,
		fn:     f,
	}
}

func (tw *TimeWheel) run() {
	for {
		select {
		case <-tw.ticker.C:
			tw.tickHandle()

		case task := <-tw.addTaskC:
			tw.add(task)

		case <-tw.stopC:
			tw.ticker.Stop()
			return

		}
	}
}

func (tw *TimeWheel) tickHandle() {
	l := tw.slots[tw.currentPos]

	tw.scanAddRunTask(l)

	if tw.currentPos == tw.slotNum-1 {
		tw.currentPos = 0
	} else {
		tw.currentPos++
	}
}

func (tw *TimeWheel) scanAddRunTask(l *list.List) {
	if l == nil {
		return
	}

	for item := l.Front(); item != nil; {
		obj := item.Value.(*task)
		status := obj.PrepareRun()

		if status.IsSkip() {
			item = item.Next()
			continue
		}

		next := item.Next()
		l.Remove(item)
		item = next

		if status.IsStop() {
			tw.collectTask(obj)
			continue
		}

		if obj.async {
			go obj.callback()
			if status.IsRunAgain() {
				tw.add(obj)
			} else {
				tw.collectTask(obj)
			}
		} else {
			if status.IsRunAgain() {
				go func() {
					obj.callback()
					tw.addRepeatTask(obj)
				}()
			} else {
				go obj.callback()
				tw.collectTask(obj)
			}
		}
	}
}

func (tw *TimeWheel) addTask(delay time.Duration, callback func(), times int32, async bool) TaskId {
	if delay <= 0 {
		delay = tw.interval
	}

	var obj *task
	if times == timesUnlimit {
		obj = new(task)
	} else {
		obj = defaultTaskPool.get()
		obj.pool = true
	}

	obj.delay = delay
	obj.times = times
	obj.async = async
	obj.callback = callback
	obj.id = tw.genUniqueID()

	tw.addTaskC <- obj
	tw.taskRecord.Store(obj.id, obj)

	return obj.id
}

func (tw *TimeWheel) addRepeatTask(obj *task) {
	tw.addTaskC <- obj
}

func (tw *TimeWheel) add(obj *task) {
	pos, circle := tw.getPositionAndCircle(obj.delay)
	obj.circle = circle

	tw.slots[pos].PushBack(obj)
}

func (tw *TimeWheel) collectTask(obj *task) {
	tw.taskRecord.Delete(obj.id)
	if obj.pool {
		defaultTaskPool.put(obj)
	}
}

func (tw *TimeWheel) getPositionAndCircle(d time.Duration) (pos int, circle int) {
	delaySeconds := d.Seconds()
	intervalSeconds := tw.interval.Seconds()
	circle = int(delaySeconds/intervalSeconds) / tw.slotNum
	pos = (tw.currentPos + int(delaySeconds/intervalSeconds)) % tw.slotNum
	return
}

func (tw *TimeWheel) genUniqueID() TaskId {
	return TaskId(atomic.AddUint64(&tw.randomID, 1))
}
