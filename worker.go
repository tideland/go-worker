// -----------------------------------------------------------------------------
// worker for running tasks enqueued in background
//
// Copyright (C) 2024 Frank Mueller / Oldenburg / Germany / World
// -----------------------------------------------------------------------------

package worker

import "time"

// Worker is the main type of the package. It runs tasks enqueued in background
// and does not have any methods beside the constructor. It allows the different
// commands to run special kind of tasks in the background. Beside only enqueuing
// tasks it's also possible to wait for the execution of a task or retrieve a waiter
// to get a notification when the worker has finished the task.
type Worker struct {
	tasks   chan Task
	timeout time.Duration
}

// New creates a new worker. queueCap is the capacity of the task queue.
func New(options ...Option) (*Worker, error) {
	worker := &Worker{}

	// Set different options.
	for _, option := range options {
		if err := option(worker); err != nil {
			return nil, err
		}
	}

	// Check if the task has been configured.
	if worker.tasks == nil {
		worker.tasks = make(chan Task, defaultQueueCap)
	}
	if worker.timeout == 0 {
		worker.timeout = defaultTimeout
	}

	// Start the worker as goroutine. It's ready when the started channel is closed.
	started := make(chan struct{})

	go worker.processor(started)

	select {
	case <-started:
	case <-time.After(worker.timeout):
		return nil, NotStartedError{}
	}

	return worker, nil
}

// enqueue passes a task to the worker.
func (w *Worker) enqueue(task Task) error {
	select {
	case w.tasks <- task:
		return nil
	case <-time.After(w.timeout):
		return TimeoutError{}
	}
}

// processor runs the worker goroutine for processing the tasks.
func (w *Worker) processor(started chan struct{}) {
	close(started)
	for {
		select {
		case task := <-w.tasks:
			// Run the task.
			err := task.Process()
			switch err.(type) {
			case stopError:
				return
			default:
				// Hande the error.
				// TODO: log the error.
			}
		}
	}
}

// -----------------------------------------------------------------------------
// end of file
// -----------------------------------------------------------------------------
