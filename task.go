package timewheel

import (
	"sync"
	"time"
)

const (
	timesUnlimit = int32(-1)
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
	stop     bool
	run      bool
}

func (t *task) Stop() (hasRun bool) {
	t.mut.Lock()
	t.stop = true
	hasRun = t.run
	t.mut.Unlock()
	return
}

// Cannot be called concurrently with other functions
func (t *task) Reset() {
	t.id = 0
	t.circle = 0
	t.callback = nil
	t.times = 0
	t.async = false
	t.pool = false
	t.stop = false
	t.run = false
}

// Cannot be called concurrently
func (t *task) Run() (gc bool) {
	if t.circle > 0 {
		t.circle--
		return
	}

	t.mut.Lock()
	if t.stop {
		t.mut.Unlock()
		return true
	}
	t.run = true
	t.mut.Unlock()

	if t.async {
		go t.callback()
	} else {
		t.callback()
	}

	if t.times > 0 {
		t.times--
	}

	return true
}
