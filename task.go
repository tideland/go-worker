// -----------------------------------------------------------------------------
// worker for running tasks enqueued in background (task.go)
//
// Copyright (C) 2024 Frank Mueller / Oldenburg / Germany / World
// -----------------------------------------------------------------------------

package worker

// Task represents the interface any task has to implement. Process has to
// implement the work to be done.
type Task interface {
	Process() error
}

// TaskFunc allows to use a function as a task. This way it's possible to
// create a task using a function or a closure.
type TaskFunc func() error

func (f TaskFunc) Process() error {
	return f()
}

// -----------------------------------------------------------------------------
// end of file
// -----------------------------------------------------------------------------
