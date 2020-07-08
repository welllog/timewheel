package timewheel

import "sync"

var defaultTaskPool = newTaskPool()

type taskPool struct {
	p sync.Pool
}

func newTaskPool() *taskPool {
	return &taskPool{
		p: sync.Pool{
			New: func() interface{} {
				return &task{}
			},
		},
	}
}

func (tp *taskPool) get() *task {
	return tp.p.Get().(*task)
}

func (tp *taskPool) put(task *task) {
	task.Reset()
	tp.p.Put(task)
}
