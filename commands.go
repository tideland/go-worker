// -----------------------------------------------------------------------------
// worker for running tasks enqueued in background (commands.go)
//
// Copyright (C) 2024 Frank Mueller / Oldenburg / Germany / World
// -----------------------------------------------------------------------------

package worker // import "tideland.dev/go/worker"

import "time"

// Enqueue passes a task to a worker.
func Enqueue(w *Worker, task Task) error {
	return w.enqueue(func() (action, Task) {
		return actionProcess, task
	})
}

// EnqueueWaiting passes a task to a worker and waits until it has been processed.
func EnqueueWaiting(w *Worker, task Task) error {
	done := make(chan struct{})
	signaller := func() (action, Task) {
		defer close(done)
		return actionProcess, task
	}
	if err := w.enqueue(signaller); err != nil {
		return err
	}
	// Wait for the worker to finish the task.
	select {
	case <-done:
		return nil
	case <-time.After(w.timeout):
		return TimeoutError{}
	}
}

// Awaiter is returned by AsyncAwait to wait for the worker to finish a task.
type Awaiter func() error

// AsyncAwait passes a task to a worker and returns an Awaiter to wait for the
// worker to finish the task. The timeout is the maximum time to wait for the
// worker to finish the task. After enqueuing the task different work can be done
// before waiting for the worker to finish the task.
func AsyncAwait(w *Worker, task Task, timeout time.Duration) (Awaiter, error) {
	done := make(chan struct{})
	signaller := func() (action, Task) {
		return actionProcess, func() error {
			defer close(done)
			err := task()
			return err
		}
	}
	if err := w.enqueue(signaller); err != nil {
		return nil, err
	}
	// Return a function waiting for the worker to finish the task.
	return func() error {
		select {
		case <-done:
			return nil
		case <-time.After(timeout):
			return TimeoutError{}
		}
	}, nil
}

// Shutdown tells the worker to stop working. It's an asynchronous operation and
// returns immediately. The worker will finish all tasks gracefully before stopping.
func Shutdown(w *Worker) error {
	shutdown := func() (action, Task) {
		return actionShutdown, nil
	}
	return w.enqueue(shutdown)
}

// -----------------------------------------------------------------------------
// end of file
// -----------------------------------------------------------------------------
