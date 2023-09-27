package scheduler

import (
	"container/heap"
	"errors"
	"sync"
	"time"

	"github.com/hostinger/fireactions/internal/server/scheduler/queue"
	"github.com/hostinger/fireactions/internal/structs"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// ErrQueueClosed is returned when an operation is attempted on a closed
	// queue.
	ErrQueueClosed = errors.New("queue is closed")

	// ErrRunnerAlreadyInQueue is returned when a Runner is attempted to be
	// enqueued multiple times.
	ErrRunnerAlreadyInQueue = errors.New("runner is already in queue")

	// ErrRunnerNotInQueue is returned when a Runner is attempted to be
	// dequeued but is not in the queue.
	ErrRunnerNotInQueue = errors.New("runner is not in queue")
)

type QueuedRunnerInfo struct {
	*structs.Runner

	// EnqueueTime is the time at which the Runner was enqueued.
	EnqueueTime time.Time

	// Attempts is the number of times the Runner has been attempted to be
	// scheduled.
	Attempts int
}

type SchedulingQueue struct {
	waiting   *queue.Queue
	runners   map[string]*QueuedRunnerInfo
	blocked   map[string]*QueuedRunnerInfo
	closed    bool
	closeOnce sync.Once
	l         *sync.Mutex
	cond      *sync.Cond

	runnersCount *prometheus.GaugeVec
}

// NewSchedulingQueue creates a new SchedulingQueue.
func NewSchedulingQueue() *SchedulingQueue {
	q := &SchedulingQueue{
		waiting: queue.New(func(i, j interface{}) bool {
			return i.(*QueuedRunnerInfo).EnqueueTime.Before(j.(*QueuedRunnerInfo).EnqueueTime)
		}),
		runners:   make(map[string]*QueuedRunnerInfo),
		blocked:   make(map[string]*QueuedRunnerInfo),
		closed:    false,
		closeOnce: sync.Once{},
		cond:      sync.NewCond(&sync.Mutex{}),
		l:         &sync.Mutex{},
	}

	q.runnersCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "runners_queued_total",
		Namespace: "fireactions",
		Subsystem: "scheduler",
		Help: `The number of Runners in the queue. 
The state label indicates whether the Runner is waiting to be scheduled or is blocked (i.e. unschedulable).`,
	}, []string{"state"})

	return q
}

// Enqueue enqueues a Runner into the queue. If the Runner is already in the
// queue, it is ignored.
func (q *SchedulingQueue) Enqueue(r *structs.Runner) error {
	q.l.Lock()
	defer q.l.Unlock()

	_, ok := q.runners[r.ID]
	if ok {
		return ErrRunnerAlreadyInQueue
	}

	defer q.cond.Signal()

	i := &QueuedRunnerInfo{Runner: r, EnqueueTime: time.Now(), Attempts: 0}
	q.runners[r.ID] = i

	heap.Push(q.waiting, i)

	return nil
}

// Dequeue dequeues a Runner from the queue. If the queue is empty, it blocks
// until a Runner is enqueued.
func (q *SchedulingQueue) Dequeue() (*structs.Runner, error) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	for q.waiting.Len() == 0 {
		if q.closed {
			return nil, ErrQueueClosed
		}

		q.cond.Wait()
	}

	i := heap.Pop(q.waiting).(*QueuedRunnerInfo).Runner
	return i, nil
}

// Block blocks a Runner that is currently in the queue and prevents it from
// being scheduled until it is unblocked.
func (q *SchedulingQueue) Block(id string) {
	q.l.Lock()
	defer q.l.Unlock()

	r, ok := q.runners[id]
	if !ok {
		return
	}

	r.Attempts++

	q.blocked[id] = r
}

// Unblock unblocks a Runner that is currently deemed unschedulable and puts it
// back into the queue.
func (q *SchedulingQueue) Unblock(id string) {
	q.l.Lock()
	defer q.l.Unlock()

	r, ok := q.blocked[id]
	if !ok {
		return
	}

	heap.Push(q.waiting, r)

	q.cond.Signal()
}

// UnblockAll unblocks all Runners that are currently deemed unschedulable and
// puts them back into the queue.
func (q *SchedulingQueue) UnblockAll() {
	q.l.Lock()
	defer q.l.Unlock()

	for _, r := range q.blocked {
		heap.Push(q.waiting, r)

		q.cond.Signal()
	}

	q.blocked = make(map[string]*QueuedRunnerInfo)
}

// Remove removes a Runner from the queue. If the Runner is not in the queue, it
// is ignored.
func (q *SchedulingQueue) Remove(id string) {
	q.l.Lock()
	defer q.l.Unlock()

	_, ok := q.runners[id]
	if !ok {
		return
	}

	delete(q.runners, id)
}

// Close closes the queue and prevents any further scheduling.
func (q *SchedulingQueue) Close() {
	defer q.l.Unlock()
	q.l.Lock()

	q.closeOnce.Do(func() {
		q.closed = true
		q.cond.Broadcast()
	})
}

// Collect implements prometheus.Collector.
func (q *SchedulingQueue) Collect(ch chan<- prometheus.Metric) {
	q.runnersCount.Collect(ch)
}

// Describe implements prometheus.Collector.
func (q *SchedulingQueue) Describe(ch chan<- *prometheus.Desc) {
	q.runnersCount.Describe(ch)
}
