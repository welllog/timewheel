package timewheel

import (
	"fmt"
	"testing"
	"time"
)

var _testDrv = DELAY_QUEUE_DRV

func TestTimeWheel_NewTimer(t *testing.T) {
	InitTiming(time.Millisecond, 50, _testDrv)
	defer StopTiming()

	start := time.Now()
	timer := NewTimer(time.Second)
	select {
	case <-time.After(1120 * time.Millisecond):
		t.Error("delay run")
	case <-timer.C:
		if time.Now().Sub(start) < 980*time.Millisecond {
			t.Error("run ahead")
		}
	}

	start = time.Now()
	timer.Reset(500 * time.Millisecond)
	select {
	case <-timer.C:
		if time.Now().Sub(start) < 480*time.Millisecond {
			t.Error("run ahead")
		}
		if timer.Stop() {
			t.Error("timer stop value error")
		}
	case <-time.After(620 * time.Millisecond):
		t.Error("delay run")
	}

	timer = NewTimer(500 * time.Millisecond)
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
	InitTiming(time.Millisecond, 50, _testDrv)
	defer StopTiming()

	start := time.Now()
	ticker := NewTicker(500 * time.Millisecond)

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
	InitTiming(time.Millisecond, 50, _testDrv)
	defer StopTiming()

	start := time.Now()

	select {
	case end := <-After(500 * time.Millisecond):
		if end.Sub(start) < 500*time.Millisecond {
			t.Error("run ahead")
		}
	case <-time.After(620 * time.Millisecond):
		t.Error("delay run")
	}
}

func TestTimeWheel_AfterFunc(t *testing.T) {
	InitTiming(time.Millisecond, 50, _testDrv)
	defer StopTiming()

	start := time.Now()

	ch := make(chan struct{}, 1)
	AfterFunc(500*time.Millisecond, func() {
		ch <- struct{}{}
	})

	select {
	case <-ch:
		if time.Now().Sub(start) < 480*time.Millisecond {
			t.Error("run ahead")
		}
	case <-time.After(620 * time.Millisecond):
		t.Error("delay run")
	}
}

func TestTimeWheel_Sleep(t *testing.T) {
	InitTiming(time.Millisecond, 50, _testDrv)
	defer StopTiming()

	start := time.Now()
	Sleep(500 * time.Millisecond)
	checkTime(t, start, time.Now(), 480*time.Millisecond, 620*time.Millisecond)
}

func TestTimeWheel_Stop(t *testing.T) {
	InitTiming(time.Millisecond, 50, _testDrv)

	ch := After(500 * time.Millisecond)
	StopTiming()

	select {
	case <-ch:
		t.Error("timewheel stop failed")
	case <-time.After(620 * time.Millisecond):
	}
}

func checkTime(t *testing.T, start, end time.Time, min, max time.Duration) {
	interval := end.Sub(start)
	if interval <= min {
		t.Error("run ahead ", interval.Seconds())
	}

	if interval >= max {
		t.Error("delay run ", interval.Seconds())
	}
}

func TestTicker(t *testing.T) {
	InitTiming(time.Millisecond, 50, _testDrv)
	defer StopTiming()

	start := time.Now()

	ticker := NewTicker(time.Second)
	defer ticker.Stop()

	Sleep(3 * time.Second)
	<-ticker.C
	fmt.Println(time.Now().Sub(start).Seconds())
	<-ticker.C
	fmt.Println(time.Now().Sub(start).Seconds())
	<-ticker.C
	fmt.Println(time.Now().Sub(start).Seconds())
}
