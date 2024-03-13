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

// EnqueueFunc is a shortcut to pass a function as task to a worker.
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
	if err := w.enqueue(signaller); err != nil {
		return err
	}
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
	signaller := TaskFunc(func() error {
		task.Process()
		close(done)
		return nil
	})
	if err := w.enqueue(signaller); err != nil {
		return nil, err
	}
	return func() error {
		select {
		case <-done:
			return nil
		case <-time.After(timeout):
			return TimeoutError{}
		}
	}, nil
}

// AsyncAwaitFunc is a shortcut for AsyncAwait with a task function.
func AsyncAwaitFunc(w *Worker, task TaskFunc, timeout time.Duration) (Awaiter, error) {
	return AsyncAwait(w, task, timeout)
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
