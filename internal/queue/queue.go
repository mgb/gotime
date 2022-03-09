package queue

import (
	"container/heap"
	"fmt"
	"sort"
	"strings"
	"time"
)

// TimeQueue is a priority queue of channels order by their trigger time. Not concurrent safe.
type TimeQueue interface {
	Add(t time.Time, ch chan<- time.Time) func() bool
	PopBeforeOrEqual(t time.Time) []chan<- time.Time
	Peek() (time.Time, bool)
	Len() int
}

// NewTimeQueue returns a new Timer
func NewTimeQueue() TimeQueue {
	return &timeQueue{}
}

func (o *timeQueue) String() string {
	if o.Len() == 0 {
		return "timeQueue{}"
	}

	var ts []string
	for _, t := range o.items {
		ts = append(ts, fmt.Sprint(t.t))
	}
	sort.Strings(ts)

	return fmt.Sprintf("timeQueue{len:%d, timers: %s}",
		o.Len(),
		strings.Join(ts, ", "),
	)
}

func (o *timeQueue) Add(t time.Time, ch chan<- time.Time) func() bool {
	id := o.counter
	o.counter++

	heap.Push(o, &item{
		id: id,
		t:  t,
		ch: ch,
	})

	return func() bool { return o.remove(id) }
}

func (o *timeQueue) PopBeforeOrEqual(t time.Time) []chan<- time.Time {
	var chs []chan<- time.Time
	for i, ok := o.Peek(); ok && !i.After(t); i, ok = o.Peek() {
		chs = append(chs, heap.Pop(o).(*item).ch)
	}
	return chs
}

func (o *timeQueue) Peek() (time.Time, bool) {
	if o.Len() == 0 {
		return time.Time{}, false
	}
	return o.items[0].t, true
}

func (o *timeQueue) remove(id int) bool {
	for _, i := range o.items {
		if i.id == id {
			heap.Remove(o, i.index)
			return true
		}
	}
	return false
}

type item struct {
	// Used for cleanup
	id int

	t  time.Time
	ch chan<- time.Time

	// The index of the item in the heap. It is needed by remove and is maintained by the heap.Interface methods.
	index int
}

// timeQueue implements heap.Interface and holds timers
type timeQueue struct {
	counter int
	items   []*item
}

func (o *timeQueue) Len() int           { return len(o.items) }
func (o *timeQueue) Less(i, j int) bool { return o.items[i].t.Before(o.items[j].t) }
func (o *timeQueue) Swap(i, j int) {
	o.items[i], o.items[j] = o.items[j], o.items[i]
	o.items[i].index = i
	o.items[j].index = j
}

func (o *timeQueue) Push(x interface{}) {
	n := len(o.items)
	t := x.(*item)
	t.index = n
	o.items = append(o.items, t)
}

func (o *timeQueue) Pop() interface{} {
	old := o.items
	n := len(old)
	t := old[n-1]
	old[n-1] = nil // avoid memory leak
	t.index = -1   // for safety
	o.items = old[0 : n-1]
	return t
}
