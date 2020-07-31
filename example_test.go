package taskq

import "fmt"

// CounterTask is a task which sends an integer value to its out channel
type CounterTask struct {
	out chan<- int
}

// Execute implements the Task interface
func (t *CounterTask) Execute() {
	// This simple task's mission is to send a 1 (one) to its output channel
	t.out <- 1
}

func ExampleQueue() {
	// Create a task queue with a few workers
	queue := New(Workers(5))

	// Create a channel to receive the results sent by the tasks
	results := make(chan int)

	// Collect the results sent by the tasks
	sum := collect(results)

	// Submit some tasks to the queue. When done, signal the queue that
	// no more tasks will be submitted and wait for all the submitted tasks
	// to be executed
	for i := 0; i < 10; i++ {
		task := &CounterTask{out: results}
		queue.Submit(task)
	}
	queue.CloseAndWait()

	// Signal the collector that no more results are to be expected
	close(results)

	// Retrieve the sum of the values sent by the tasks
	fmt.Println(<-sum)
	// Output: 10
}

// collect sums all the integer values received via the in channel and sends
// the computed value to the returned channel
func collect(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		sum := 0
		for r := range in {
			sum += r
		}
		out <- sum
	}()
	return out
}

type SimpleTask int

func (t SimpleTask) Execute() {
	fmt.Printf("Hello I'm task %d\n", t)
}

func ExampleSimple() {
	// Create a queue with 10 workers
	q := New(Workers(2))

	// Submit tasks for execution.
	for i := 0; i < 5; i++ {
		task := SimpleTask(i)
		q.Submit(task)
	}

	// Notify the queue that no more tasks will be submitted
	// and wait for all the already submitted tasks to finish
	q.CloseAndWait()

	// Unordered output:
	// Hello I'm task 0
	// Hello I'm task 1
	// Hello I'm task 2
	// Hello I'm task 3
	// Hello I'm task 4
}
