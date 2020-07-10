package timewheel

import (
	"sync"
	"time"
)

const (
	timesUnlimit = int32(-1)
)

type TaskStatus uint8

func (s TaskStatus) IsStop() bool {
	return s == stop
}

func (s TaskStatus) IsLast() bool {
	return s == last
}

func (s TaskStatus) IsSkip() bool {
	return s == skip
}

func (s TaskStatus) IsRunAgain() bool {
	return s == runAgain
}

const (
	stop TaskStatus = iota
	last
	skip
	runAgain
)

type TaskId uint64

type task struct {
	id       TaskId
	delay    time.Duration
	circle   int
	callback func()
	mut      sync.Mutex
	times    int32 //-1:no limit >=1:run times
	async    bool
	pool     bool
	run      bool
}

func (t *task) Stop(id TaskId) (hasRun bool) {
	t.mut.Lock()
	if t.id != id {
		// May be recycled, which means it has been run. also it recycled before run,
		// this only occurs when stop is called multiple times.
		t.mut.Unlock()
		return true
	}
	t.times = 0
	hasRun = t.run
	t.mut.Unlock()
	return
}

func (t *task) Reset() {
	t.mut.Lock()
	t.id = 0
	t.run = false
	t.mut.Unlock()
}

func (t *task) PrepareRun() TaskStatus {
	var times int32

	t.mut.Lock()

	if t.times == 0 {
		t.mut.Unlock()
		return stop
	}

	if t.circle > 0 {
		t.circle--
		t.mut.Unlock()
		return skip
	}

	t.run = true
	if t.times > 0 {
		t.times--
	}
	times = t.times

	t.mut.Unlock()

	if times == 0 {
		return last
	}

	return runAgain
}
