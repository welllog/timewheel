package timing

import "time"

type Timing interface {
	Start()
	Stop()
	AddTask(delay time.Duration, task func()) Timer
	ScheduleTask(s Scheduler, task func()) Timer
}

type Scheduler interface {
	Next(time.Time) time.Time
}

type Timer interface {
	Stop() bool
}
