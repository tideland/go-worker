// -----------------------------------------------------------------------------
// worker for running tasks enqueued in background
//
// Copyright (C) 2024 Frank Mueller / Oldenburg / Germany / World
// -----------------------------------------------------------------------------

package worker // import "tideland.dev/go/worker"

import (
	"context"
	"time"
)

// Worker help to enqueue tasks and process them in the background in order. Stopping
// the worker with the according command ensures that all tasks before are processed.
type Worker struct {
	rate    int
	burst   int
	timeout time.Duration
	in      input
	out     output
}

// New creates a new worker. The options are used to configure the worker. If no
// options are given the worker is created with default settings.
func New(ctx context.Context, options ...Option) (*Worker, error) {
	worker := &Worker{}

	// Set different options.
	for _, option := range options {
		if err := option(worker); err != nil {
			return nil, err
		}
	}

	// Check if the options are set or use defaults.
	if worker.rate == 0 {
		worker.rate = defaultRate
	}
	if worker.burst == 0 {
		worker.burst = defaultBurst
	}
	if worker.timeout == 0 {
		worker.timeout = defaultTimeout
	}

	// Set input and output for limited buffer.
	in, out := setupRatedBuffer(ctx, worker.rate, worker.burst, worker.timeout)

	worker.in = in
	worker.out = out

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
func (w *Worker) enqueue(task actionTask) error {
	return w.in(task)
}

// processor runs the worker goroutine for processing the tasks.
func (w *Worker) processor(started chan struct{}) {
	close(started)
	for atask := range w.out() {
		action, task := atask()
		// Check action.
		switch action {
		case actionProcess:
			if err := task(); err != nil {
				// Handle the error.
				// TODO: log the error.
				// Continue with the next task.
				continue
			}
		case actionShutdown:
			// Shutdown the worker.
			// TODO: log the shutdown.
			return
		}
	}
}

// -----------------------------------------------------------------------------
// end of file
// -----------------------------------------------------------------------------
