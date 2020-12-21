package timing

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestNewDelayQueue(t *testing.T) {
	q := NewDelayQueue(10, time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	now := time.Now()
	go func() {
		q.Poll(ctx, time.Now)
	}()

	data := dqData()
	for _, v := range data {
		item := v
		go func() {
			q.Offer(item.key, now.Add(item.exp))
		}()
	}
	var w sync.WaitGroup
	w.Add(1)
	go func() {
		defer w.Done()
		time.Sleep(5 * time.Second)
		cancel()

		for _, v := range data {
			w.Add(1)
			item := v
			go func() {
				defer w.Done()
				q.Offer(item.key, now.Add(item.exp))
			}()
		}
	}()
	var i int64
	for {
		select {
		case val := <-q.Chan():
			if val.(int64) != i {
				t.Fatal("elem must be ", i)
			}
			i++
		case <-ctx.Done():
			goto exit
		}
	}
exit:
	w.Wait()
}

type dqdata struct {
	key int64
	exp time.Duration
}

func dqData() []dqdata {
	var data []dqdata
	for i := 1000; i >= 0; i-- {
		data = append(data, dqdata{key: int64(i), exp: time.Duration(i*3) * time.Millisecond})
	}
	for i := 1001; i < 1200; i++ {
		data = append(data, dqdata{key: int64(i), exp: time.Duration(i*3) * time.Millisecond})
	}
	return data
}
