package timing

import (
	"container/heap"
	"sync"
)

type PriorityQueue interface {
	Add(v interface{}, priority int64) int
	Peek() interface{}
	Shift() interface{}
	PriorityShift(maxPriority int64) (interface{}, int64)
	Size() int
}

func NewPriorityQueue(capacity int) PriorityQueue {
	return &priorityQueue{entries: make([]*entry, 0, capacity), cap: capacity}
}

type entry struct {
	value    interface{}
	priority int64
	index    int
}

var _entryPool = sync.Pool{
	New: func() interface{} {
		return &entry{}
	},
}

func getEntry() *entry {
	return _entryPool.Get().(*entry)
}

func putEntry(et *entry) {
	et.index = 0
	et.priority = 0
	et.value = nil
	_entryPool.Put(et)
}

type priorityQueue struct {
	entries []*entry
	cap     int
}

func assertHeapImplementation() {
	var _ heap.Interface = (*priorityQueue)(nil)
}

func (q *priorityQueue) Len() int {
	return len(q.entries)
}

func (q *priorityQueue) Less(i, j int) bool {
	return q.entries[i].priority < q.entries[j].priority
}

func (q *priorityQueue) Swap(i, j int) {
	q.entries[i], q.entries[j] = q.entries[j], q.entries[i]
	q.entries[i].index = i
	q.entries[j].index = j
}

func (q *priorityQueue) Push(x interface{}) {
	n := len(q.entries)

	i := x.(*entry)
	i.index = n
	q.entries = append(q.entries, i)
}

func (q *priorityQueue) Pop() interface{} {
	n := len(q.entries)
	i := q.entries[n-1]
	i.index = -1

	q.entries[n-1] = nil
	q.entries = q.entries[0 : n-1]

	c := cap(q.entries)
	if n < (c/2) && c > q.cap {
		ne := make([]*entry, n-1, c/2)
		copy(ne, q.entries)
		q.entries = ne
	}

	return i
}

func (q *priorityQueue) Add(v interface{}, priority int64) int {
	e := getEntry()
	//e := &entry{}
	e.value = v
	e.priority = priority
	heap.Push(q, e)
	return e.index
}

func (q *priorityQueue) Peek() interface{} {
	if q.Len() == 0 {
		return nil
	}
	return q.entries[0].value
}

func (q *priorityQueue) Shift() (value interface{}) {
	if q.Len() == 0 {
		return nil
	}
	i := q.entries[0]
	heap.Remove(q, 0)
	value = i.value
	putEntry(i)
	return
}

func (q *priorityQueue) PriorityShift(maxPriority int64) (value interface{}, priority int64) {
	if q.Len() == 0 {
		return nil, 0
	}
	i := q.entries[0]
	if i.priority > maxPriority {
		return nil, i.priority
	}
	heap.Remove(q, 0)
	value, priority = i.value, i.priority
	putEntry(i)
	return
}

func (q *priorityQueue) Size() int {
	return q.Len()
}
