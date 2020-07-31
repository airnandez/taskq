// Package taskq enqueues tasks and executes them asynchronously and concurrently
package taskq

import (
	"context"
	"sync"
)

var (
	// DefaultWorkers is the default number of worker goroutines a Queue will
	// create for executing tasks
	DefaultWorkers = 1
)

// Queue represents a queue of tasks submitted for asynchronous processing by
// a pool of workers owned by the queue.
//
// Use 'Submit' to submit new tasks to the queue and 'Close' to notify the
// queue that no more tasks will be submitted.
// Each Task is executed asynchronously in a separate goroutine and several tasks
// may be processed concurrently. The queue cleans up the spawned goroutines when
// processing is normally or abnormally terminated as a result of calling
// 'Close', 'CloseAndWait' or 'Cancel'.
type Queue struct {
	// workers is the number of worker goroutines spawned for processing
	// tasks submitted to this queue
	workers int

	// tasks is a buffered channel for sending Tasks to the workers.
	// It can hold at least one task per worker
	tasks chan Task

	// done is the channel to signal clients that this Queue will no
	// longer process tasks, either because it finished processing all
	// the tasks submitted or because the queue was canceled
	done chan struct{}

	// cancel is a cancel function to signal workers that they must immediately
	// stop procesing tasks
	cancel context.CancelFunc

	// wait group for detecting when all workers have terminated
	wg sync.WaitGroup

	// is this queue already initialized ?
	initialized bool
}

// New creates a new Queue object. 'options' are the functional
// options you can use to configure the Queue (e.g. Workers, Waiting)
func New(options ...func(q *Queue)) *Queue {
	q := &Queue{
		workers: max(1, DefaultWorkers),
		done:    make(chan struct{}),
	}
	for _, opt := range options {
		opt(q)
	}

	// Ensure the size of the waiting buffer is at
	// least the number of workers
	if cap(q.tasks) < q.workers {
		q.tasks = make(chan Task, q.workers)
	}

	// When all workers have terminated, close the 'done' channel
	q.wg.Add(q.workers)
	go func() {
		q.wg.Wait()
		close(q.done)
	}()

	// Start the workers
	var ctx context.Context
	ctx, q.cancel = context.WithCancel(context.Background())
	for i := 0; i < q.workers; i++ {
		go q.executeTasks(ctx)
	}

	q.initialized = true
	return q
}

// Workers is an optional function to be used when creating a new Queue via New()
// to specify the number of worker goroutines to spawn for a Queue.
//
// Applying this option to an already initialized queue has
// no effect.
func Workers(count int) func(q *Queue) {
	return func(q *Queue) {
		if !q.initialized {
			q.workers = max(count, 1)
		}
	}
}

// Waiting is an optional function to be used when creating a new Queue via New()
// to specify the maximum number of tasks that can be in the queue's waiting area.
// After reaching that limit, submiting a new task will block (see Submit).
//
// The waiting area size is ensured to be at least equivalent to the number of
// workers so that all workers can perform work concurrently. This option
// should be used after specifying the number of workers (see Workers).
//
// Applying this option to an already initialized queue has no effect.
func Waiting(count int) func(q *Queue) {
	return func(q *Queue) {
		if !q.initialized {
			// Allocate a buffered channel to ensure there is at least one slot per worker
			q.tasks = make(chan Task, max(q.workers, count))
		}
	}
}

// Submit enqueues a new Task for asynchronous processing. It will block if
// all the slots in the waiting area of the queue are occupied.
// Calling 'Submit' after calling 'Close' or 'Cancel' results in a panic.
func (q *Queue) Submit(t Task) {
	q.tasks <- t
}

// Close notifies the queue that no more tasks will be submitted. The tasks
// already queued at the time Close is called will be executed.
// It returns immediately without waiting for the execution of the tasks already
// queued.
//
// Use Wait or CloseAndWait to synchronously wait for the queue to terminate.
// If a new task is submitted after calling Close, it results in a panic.
func (q *Queue) Close() {
	close(q.tasks)
}

// Wait waits until the queue has finished processing all the queued tasks,
// as a result of calling either 'Close' or 'Cancel'.
func (q *Queue) Wait() {
	<-q.done
}

// CloseAndWait closes the queue and waits until all the queued tasks finish
// their execution
func (q *Queue) CloseAndWait() {
	q.Close()
	q.Wait()
}

// Cancel notifies the queue that no more tasks should be processed including those
// already queued. After calling Cancel, tasks already queued will be ignored.
// Calling 'Submit' after calling 'Cancel' on a queue will result in a panic.
func (q *Queue) Cancel() {
	q.cancel()
	close(q.tasks)
}

// executeTasks receives new tasks from the 'tasks' channel and executes them. It terminates
// either because there are no more tasks to execute (i.e. the 'tasks' channel is closed)
// or the context is canceled. In both cases it stops immediately.
// executeTasks is executed by each worker goroutine managed by a Queue.
func (q *Queue) executeTasks(ctx context.Context) {
	defer q.wg.Done()
	for task := range q.tasks {
		select {
		case <-ctx.Done():
			return
		default:
			task.Execute()
		}
	}
}

// max returns the maximum of a and b
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
