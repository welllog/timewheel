package timewheel

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestTimeWheel_AddOnce(t *testing.T) {
	tw := NewTimeWheel(100*time.Millisecond, 50)
	tw.Start()
	defer tw.Stop()

	start := time.Now()
	ch := make(chan struct{}, 1)
	tw.AddOnce(time.Second, func() {
		ch <- struct{}{}
	})

	select {
	case <-time.After(1120 * time.Millisecond):
		t.Error("delay run")
	case <-ch:
		if time.Now().Sub(start) < time.Second {
			t.Error("run ahead")
		}
	}
}

func TestTimeWheel_AddCron(t *testing.T) {
	tw := NewTimeWheel(100*time.Millisecond, 50)
	tw.Start()
	defer tw.Stop()

	start := time.Now()
	ch := make(chan struct{}, 1)
	taskId := tw.AddCron(500*time.Millisecond, func() {
		ch <- struct{}{}
	})

	var incr int
	for {
		<-ch

		end := time.Now()
		checkTime(t, start, end, 480*time.Millisecond, 620*time.Millisecond)
		start = end

		incr++
		if incr == 20 {
			tw.Remove(taskId)
			break
		}
	}

	select {
	case <-ch:
		t.Error("remove task failed")
	case <-time.After(620 * time.Millisecond):
	}
}

func TestTimeWheel_AddWithTimes(t *testing.T) {
	tw := NewTimeWheel(100*time.Millisecond, 50)
	tw.Start()
	defer tw.Stop()

	start := time.Now()
	ch := make(chan struct{}, 1)
	tw.AddWithTimes(500*time.Millisecond, 5, func() {
		ch <- struct{}{}
	})

	for i := 0; i < 5; i++ {
		<-ch

		end := time.Now()
		checkTime(t, start, end, 480*time.Millisecond, 620*time.Millisecond)
		start = end
	}

	select {
	case <-ch:
		t.Error("task exec over times")
	case <-time.After(620 * time.Millisecond):
	}
}

func TestTimeWheel_NewTimer(t *testing.T) {
	tw := NewTimeWheel(100*time.Millisecond, 50)
	tw.Start()
	defer tw.Stop()

	start := time.Now()
	timer := tw.NewTimer(time.Second)
	select {
	case <-time.After(1120 * time.Millisecond):
		t.Error("delay run")
	case <-timer.C:
		if time.Now().Sub(start) < time.Second {
			t.Error("run ahead")
		}
	}

	start = time.Now()
	timer.Reset(500 * time.Millisecond)
	select {
	case <-timer.C:
		if time.Now().Sub(start) < 500*time.Millisecond {
			t.Error("run ahead")
		}
		if timer.Stop() {
			t.Error("timer stop value error")
		}
	case <-time.After(620 * time.Millisecond):
		t.Error("delay run")
	}

	timer = tw.NewTimer(500 * time.Millisecond)
	select {
	case <-timer.C:
		t.Error("run ahead")
	case <-time.After(460 * time.Millisecond):
		if !timer.Stop() {
			t.Error("run ahead")
			return
		}
		select {
		case <-timer.C:
			t.Error("timer stop return ok, but timer.C has value")
		case <-time.After(200 * time.Millisecond):
		}
	}
}

func TestTimeWheel_NewTicker(t *testing.T) {
	tw := NewTimeWheel(100*time.Millisecond, 50)
	tw.Start()
	defer tw.Stop()

	start := time.Now()
	ticker := tw.NewTicker(500 * time.Millisecond)

	var incr int
	for {
		<-ticker.C

		end := time.Now()
		checkTime(t, start, end, 480*time.Millisecond, 620*time.Millisecond)
		start = end

		incr++
		if incr == 20 {
			ticker.Stop()
			break
		}
	}

	select {
	case <-ticker.C:
		t.Error("ticker stop failed")
	case <-time.After(620 * time.Millisecond):
	}
}

func TestTimeWheel_After(t *testing.T) {
	tw := NewTimeWheel(100*time.Millisecond, 50)
	tw.Start()
	defer tw.Stop()

	start := time.Now()

	select {
	case end := <-tw.After(500 * time.Millisecond):
		if end.Sub(start) < 500*time.Millisecond {
			t.Error("run ahead")
		}
	case <-time.After(620 * time.Millisecond):
		t.Error("delay run")
	}
}

func TestTimeWheel_AfterFunc(t *testing.T) {
	tw := NewTimeWheel(100*time.Millisecond, 50)
	tw.Start()
	defer tw.Stop()

	start := time.Now()

	ch := make(chan struct{}, 1)
	tw.AfterFunc(500*time.Millisecond, func() {
		ch <- struct{}{}
	})

	select {
	case <-ch:
		if time.Now().Sub(start) < 500*time.Millisecond {
			t.Error("run ahead")
		}
	case <-time.After(620 * time.Millisecond):
		t.Error("delay run")
	}
}

func TestTimeWheel_Sleep(t *testing.T) {
	tw := NewTimeWheel(100*time.Millisecond, 50)
	tw.Start()
	defer tw.Stop()

	start := time.Now()
	tw.Sleep(500 * time.Millisecond)
	checkTime(t, start, time.Now(), 500*time.Millisecond, 620*time.Millisecond)
}

func TestNewTimeWheelPool(t *testing.T) {
	twp := NewTimeWheelPool(5, 100*time.Millisecond, 50)
	twp.Start()
	defer twp.Stop()

	var w sync.WaitGroup
	w.Add(200)
	var incr int
	for {
		incr++
		go func() {
			defer w.Done()
			tw := twp.Get()
			select {
			case <-time.After(520 * time.Millisecond):
				t.Error("delay run")
			case <-tw.After(400 * time.Millisecond):
			}
		}()
		if incr == 200 {
			break
		}
	}
	w.Wait()
}

func TestTimeWheel_Stop(t *testing.T) {
	tw := NewTimeWheel(100*time.Millisecond, 50)
	tw.Start()

	ch := tw.After(500 * time.Millisecond)
	tw.Stop()

	select {
	case <-ch:
		t.Error("timewheel stop failed")
	case <-time.After(620 * time.Millisecond):
	}
}

func checkTime(t *testing.T, start, end time.Time, min, max time.Duration) {
	interval := end.Sub(start)
	if interval <= min {
		t.Error("run ahead")
	}

	if interval >= max {
		t.Error("delay run")
	}
}

func TestTicker(t *testing.T) {
	tw := NewTimeWheel(100*time.Millisecond, 50)
	tw.Start()
	defer tw.Stop()

	start := time.Now()

	ticker := tw.NewTicker(time.Second)
	defer ticker.Stop()

	tw.Sleep(3 * time.Second)
	<-ticker.C
	fmt.Println(time.Now().Sub(start).Seconds())
	<-ticker.C
	fmt.Println(time.Now().Sub(start).Seconds())
	<-ticker.C
	fmt.Println(time.Now().Sub(start).Seconds())
}
