package timing

import (
	"testing"
)

func TestNewPriorityQueue(t *testing.T) {
	queue := NewPriorityQueue(8)
	data := pqData()
	for i := range data {
		queue.Add(data[i].key, data[i].pri)
	}
	if queue.Peek().(int64) != 0 {
		t.Fatal("first elem must be 0")
	}
	maxPriority := int64(1100)
	i := int64(0)
	for {
		key, priority := queue.PriorityShift(maxPriority)
		if key == nil {
			if priority != (maxPriority + 1) {
				t.Fatal("termination priority must be ", maxPriority+1)
			}
			break
		}
		if i != priority {
			t.Fatal("this elem must be ", i)
		}
		if key.(int64) != priority {
			t.Fatal("key error")
		}
		i++
	}
}

type pqdata struct {
	key int64
	pri int64
}

func pqData() []pqdata {
	var data []pqdata
	for i := 1000; i >= 0; i-- {
		data = append(data, pqdata{key: int64(i), pri: int64(i)})
	}
	for i := 1001; i < 1200; i++ {
		data = append(data, pqdata{key: int64(i), pri: int64(i)})
	}
	return data
}
