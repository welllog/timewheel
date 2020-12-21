package timewheel

import (
	"github.com/welllog/timewheel/timing/dqdriver"
	"testing"
	"time"
)

func genD(i int) time.Duration {
	return time.Duration(i%10000) * time.Millisecond
}

func BenchmarkDqTimingWheel_StartStop(b *testing.B) {
	tw := dqdriver.NewTimingWheel(time.Millisecond, 50)
	defer tw.Stop()
	
	cases := []struct {
		name string
		N    int // the data size (i.e. number of existing timers)
	}{
		{"N-1m", 1000000},
		{"N-5m", 5000000},
		{"N-10m", 10000000},
	}
	for _, c := range cases {
		b.Run(c.name, func(b *testing.B) {
			for i := 0; i < c.N; i++ {
				tw.AddTask(genD(i), func(){})
			}
			b.ResetTimer()
			
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					tw.AddTask(time.Second, func(){}).Stop()
				}
			})
			//for i := 0; i < b.N; i++ {
			//	tw.AddTask(time.Second, func(){}).Stop()
			//}
			
			b.StopTimer()
		})
	}
}

func BenchmarkTimingWheel_StartStop(b *testing.B) {
	InitTiming(time.Millisecond, 50, _testDrv)
	defer StopTiming()

	cases := []struct {
		name string
		N    int // the data size (i.e. number of existing timers)
	}{
		{"N-1m", 1000000},
		{"N-5m", 5000000},
		{"N-10m", 10000000},
	}
	for _, c := range cases {
		b.Run(c.name, func(b *testing.B) {
			for i := 0; i < c.N; i++ {
				Timing.AddTask(genD(i), func(){})
			}
			b.ResetTimer()

			//for i := 0; i < b.N; i++ {
			//	Timing.AddTask(time.Second, func(){}).Stop()
			//}
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					Timing.AddTask(time.Second, func(){}).Stop()
				}
			})

			b.StopTimer()
		})
	}
}

func BenchmarkStandardTimer_StartStop(b *testing.B) {
	cases := []struct {
		name string
		N    int // the data size (i.e. number of existing timers)
	}{
		{"N-1m", 1000000},
		{"N-5m", 5000000},
		{"N-10m", 10000000},
	}
	for _, c := range cases {
		b.Run(c.name, func(b *testing.B) {
			for i := 0; i < c.N; i++ {
				time.AfterFunc(genD(i), func() {})
			}
			b.ResetTimer()

			//for i := 0; i < b.N; i++ {
			//	time.AfterFunc(time.Second, func() {}).Stop()
			//}
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					time.AfterFunc(time.Second, func() {}).Stop()
				}
			})

			b.StopTimer()
		})
	}
}
