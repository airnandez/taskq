package taskq

import (
	"math/rand"
	"os"
	"runtime"
	"testing"
	"time"
)

var (
	numWorkers int
	numTasks   int
)

// Simple task
type countTask struct {
	output chan<- int
}

// Execute sends the value 1 to the output channel of task t
func (t *countTask) Execute() {
	t.output <- 1
}

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UTC().UnixNano())
	numWorkers = initRandom(2 * runtime.NumCPU())
	numTasks = initRandom(10 * numWorkers)
	os.Exit(m.Run())
}

// initRandom returns a random number in the interval [1, upper)
func initRandom(upper int) int {
	i := 0
	for {
		if i = rand.Intn(upper); i > 0 {
			return i
		}
	}
}

// collectResults collect all the values sent to the results channel
// and sends the sum of all those values to the returned output channel
func collectResults(results <-chan int) <-chan int {
	done := make(chan int)
	go func() {
		sum := 0
		for r := range results {
			sum += r
		}
		done <- sum
	}()
	return done
}

func TestClose(t *testing.T) {
	// Create a prototype task and get ready to collect the results that the
	// submitted tasks will generate
	results := make(chan int)
	task := &countTask{output: results}
	done := collectResults(results)

	// Create a task queue and signal the downstream collector when
	// the queue has finished processing tasks
	q := New(Workers(numWorkers))

	// Submit tasks
	for i := 0; i < numTasks; i++ {
		q.Submit(task)
	}
	q.CloseAndWait()
	close(results)

	// Check the collected results
	total := <-done
	if total != numTasks {
		t.Fatalf("expected result after normal termination %d got %d", numTasks, total)
	}
}

func TestCancel(t *testing.T) {
	// Create a prototype task and get ready to collect the results that the
	// submitted tasks will generate
	results := make(chan int)
	task := &countTask{output: results}
	done := collectResults(results)

	// Create a task queue and signal the downstream collector when
	// the queue has finished processing tasks
	q := New(Workers(numWorkers))

	// Submit some tasks and then cancel
	cancelAt := rand.Intn(numTasks)
	for i := 0; i < numTasks; i++ {
		if i == cancelAt {
			q.Cancel()
			break
		}
		q.Submit(task)
	}
	q.Wait()
	close(results)

	// Check the collected results
	total := <-done
	if total > cancelAt {
		t.Fatalf("expected result after cancelation <%d got %d", cancelAt, total)
	}
}

func TestOptions(t *testing.T) {
	// Create a prototype task and get ready to collect the results that the
	q := New(Workers(numWorkers))
	if numWorkers != q.workers {
		t.Fatalf("unexpected number of workers: expecting %d got %d", numWorkers, q.workers)
	}
	if q.workers > cap(q.tasks) {
		t.Fatalf("unexpected waiting size: expecting %d got %d", q.workers, cap(q.tasks))
	}

	// Check using options after creating the queue has no effect
	waitSize := cap(q.tasks)
	Waiting(initRandom(100))(q)
	if waitSize != cap(q.tasks) {
		t.Fatalf("option modified waiting size after cqueue initializationreation: expecting %d got %d", waitSize, cap(q.tasks))
	}
	workers := q.workers
	Workers(initRandom(100))(q)
	if workers != q.workers {
		t.Fatalf("option modified number of workers after queue initialization: expecting %d got %d", workers, q.workers)
	}

	// Check coherency between the number of workers and the waiting size
	waitSize = initRandom(100)
	q = New(Waiting(waitSize))
	if DefaultWorkers != q.workers {
		t.Fatalf("unexpected number of workers: expecting %d got %d", DefaultWorkers, q.workers)
	}
	if waitSize != cap(q.tasks) {
		t.Fatalf("unexpected waiting size: expecting %d got %d", waitSize, cap(q.tasks))
	}
	if cap(q.tasks) < q.workers {
		t.Fatalf("waiting size less than the number of workers: waiting size %d workers %d", cap(q.tasks), q.workers)
	}
}
