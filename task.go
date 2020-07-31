package taskq

// Task is the interface that must me implemented by an object so that
// it can be submitted for execution to a Queue.
// The Execute method is called in its own goroutine. Several tasks
// may be executed concurrently.
type Task interface {
	Execute()
}
