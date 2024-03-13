// -----------------------------------------------------------------------------
// worker for running tasks enqueued in background (commands.go)
//
// Copyright (C) 2024 Frank Mueller / Oldenburg / Germany / World
// -----------------------------------------------------------------------------

package worker

import "time"

// Enqueue passes a task to a worker.
func Enqueue(w *Worker, task Task) error {
	return w.enqueue(task)
}

// EnqueueFunc passes a task function to a worker.
func EnqueueFunc(w *Worker, task TaskFunc) error {
	return w.enqueue(task)
}

// EnqueueWaiting passes a task to a worker and waits until it has been processed.
func EnqueueWaiting(w *Worker, task Task) error {
	done := make(chan struct{})
	signaller := TaskFunc(func() error {
		task.Process()
		close(done)
		return nil
	})
	w.enqueue(signaller)
	select {
	case <-done:
		return nil
	case <-time.After(w.timeout):
		return TimeoutError{}
	}
}

// Stop tells the worker to stop. It's working asynchronously internally
// to let all enqueued tasks be done. Afterward the function returns.
func Stop(w *Worker) error {
	stopTask := TaskFunc(func() error {
		return stopError{}
	})
	return EnqueueWaiting(w, stopTask)
}

// -----------------------------------------------------------------------------
// end of file
// -----------------------------------------------------------------------------
