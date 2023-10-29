package scheduler

import (
	"container/heap"
	"errors"
	"sync"

	"github.com/hostinger/fireactions"
	"github.com/hostinger/fireactions/server/scheduler/queue"
)

type schedulingQueue struct {
	runnersInQueue map[string]*fireactions.Runner
	waiting        *queue.Queue
	blocked        map[string]*fireactions.Runner
	closed         bool
	closeOnce      sync.Once
	l              *sync.Mutex
	cond           *sync.Cond
}

func newSchedulingQueue() *schedulingQueue {
	return &schedulingQueue{
		waiting: queue.New(func(i, j interface{}) bool {
			return i.(*fireactions.Runner).CreatedAt.Before(j.(*fireactions.Runner).CreatedAt)
		}),
		runnersInQueue: make(map[string]*fireactions.Runner),
		blocked:        make(map[string]*fireactions.Runner),
		closed:         false,
		l:              &sync.Mutex{},
		closeOnce:      sync.Once{},
		cond:           sync.NewCond(&sync.Mutex{}),
	}
}

// Enqueue enqueues a Runner into the queue. If the Runner is already in the
// queue, it is ignored.
func (q *schedulingQueue) Enqueue(m *fireactions.Runner) error {
	q.l.Lock()
	defer q.l.Unlock()

	_, ok := q.runnersInQueue[m.ID]
	if ok {
		return nil
	}

	defer q.cond.Signal()

	q.runnersInQueue[m.ID] = m
	heap.Push(q.waiting, m)

	return nil
}

// Dequeue dequeues a Runner from the queue. If the queue is empty, it blocks
// until a Runner is enqueued.
func (q *schedulingQueue) Dequeue() (*fireactions.Runner, error) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	for q.waiting.Len() == 0 {
		if q.closed {
			return nil, errors.New("dequeue: queue is closed")
		}

		q.cond.Wait()
	}

	return heap.Pop(q.waiting).(*fireactions.Runner), nil
}

// Block blocks a Runner that is currently in the queue and prevents it from
// being scheduled until it is unblocked.
func (q *schedulingQueue) Block(id string) {
	q.l.Lock()
	defer q.l.Unlock()

	m, ok := q.runnersInQueue[id]
	if !ok {
		return
	}

	q.blocked[id] = m
}

// Unblock unblocks a Runner that is currently deemed unschedulable and puts it
// back into the queue.
func (q *schedulingQueue) Unblock(id string) {
	q.l.Lock()
	defer q.l.Unlock()

	m, ok := q.blocked[id]
	if !ok {
		return
	}

	heap.Push(q.waiting, m)
	q.cond.Signal()
}

// UnblockAll unblocks all Runners that are currently deemed unschedulable and
// puts them back into the queue.
func (q *schedulingQueue) UnblockAll() {
	q.l.Lock()
	defer q.l.Unlock()

	for _, m := range q.blocked {
		heap.Push(q.waiting, m)

		q.cond.Signal()
	}

	q.blocked = make(map[string]*fireactions.Runner)
}

// Remove removes a Runner from the queue. If the Runner is not in the queue, it
// is ignored.
func (q *schedulingQueue) Remove(id string) {
	q.l.Lock()
	defer q.l.Unlock()

	_, ok := q.runnersInQueue[id]
	if !ok {
		return
	}

	delete(q.runnersInQueue, id)
}

// Close closes the queue and prevents any further scheduling.
func (q *schedulingQueue) Close() {
	defer q.l.Unlock()
	q.l.Lock()

	q.closeOnce.Do(func() {
		q.closed = true
		q.cond.Broadcast()
	})
}
