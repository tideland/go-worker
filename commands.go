// -----------------------------------------------------------------------------
// worker for running tasks enqueued in background - commands
//
// Copyright (C) 2024-2025 Frank Mueller / Oldenburg / Germany / World
// -----------------------------------------------------------------------------

package worker

import (
	"sync"
	"time"
)

// Enqueue passes a task to a worker for background processing.
func Enqueue(w *Worker, task Task) error {
	return w.enqueue(task)
}

// EnqueueWaiting passes a task to a worker and waits until it has been processed.
func EnqueueWaiting(w *Worker, task Task) error {
	done := make(chan error, 1)
	wrappedTask := func() error {
		err := task()
		select {
		case done <- err:
		default:
		}
		return err
	}

	if err := w.enqueue(wrappedTask); err != nil {
		return err
	}

	// Wait for the worker to finish the task.
	select {
	case err := <-done:
		return err
	case <-time.After(w.config.Timeout):
		return TimeoutError{Duration: w.config.Timeout}
	}
}

// Awaiter is returned by AsyncAwait to wait for the worker to finish a task.
type Awaiter func() error

// EnqueueAwaiting passes a task to a worker and returns an Awaiter to wait for the
// worker to finish the task. The timeout is the maximum time to wait for the
// worker to finish the task. After enqueuing the task different work can be done
// before waiting for the worker to finish the task.
func EnqueueAwaiting(w *Worker, task Task, timeout time.Duration) (Awaiter, error) {
	done := make(chan error, 1)
	var once sync.Once
	
	wrappedTask := func() error {
		err := task()
		once.Do(func() {
			select {
			case done <- err:
			default:
			}
		})
		return err
	}

	if err := w.enqueue(wrappedTask); err != nil {
		return nil, err
	}

	// Return a function waiting for the worker to finish the task.
	return func() error {
		select {
		case err := <-done:
			return err
		case <-time.After(timeout):
			once.Do(func() {
				select {
				case done <- TimeoutError{Duration: timeout}:
				default:
				}
			})
			return TimeoutError{Duration: timeout}
		}
	}, nil
}

// Stop initiates a graceful shutdown of the worker. All pending tasks
// will be processed before the worker stops.
func Stop(w *Worker) error {
	return w.stop()
}

// EOF