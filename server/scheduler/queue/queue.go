package queue

import "container/heap"

type Queue struct {
	d    []interface{}
	less func(i, j interface{}) bool
}

func New(less func(i, j interface{}) bool) *Queue {
	q := &Queue{
		d:    make([]interface{}, 0),
		less: less,
	}

	heap.Init(q)

	return q
}

func (q *Queue) Len() int {
	return len(q.d)
}

func (q *Queue) Less(i, j int) bool {
	return q.less(q.d[i], q.d[j])
}

func (q *Queue) Swap(i, j int) {
	q.d[i], q.d[j] = q.d[j], q.d[i]
}

func (q *Queue) Push(x interface{}) {
	q.d = append(q.d, x)
}

func (q *Queue) Pop() interface{} {
	old := q.d
	n := len(old)
	x := old[n-1]
	q.d = old[0 : n-1]
	return x
}

func (q *Queue) Peek() interface{} {
	return q.d[0]
}
