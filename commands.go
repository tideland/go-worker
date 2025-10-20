// Tideland Go Worker - Commands
//
// Copyright (C) 2014-2025 Frank Mueller / Tideland / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package worker

import (
	"sync"
	"time"
)

// Enqueue passes a task to a WorkProcessor for background processing.
func Enqueue(wp WorkProcessor, task Task) error {
	return wp.enqueue(task)
}

// EnqueueWaiting passes a task to a WorkProcessor and waits until it has been processed.
func EnqueueWaiting(wp WorkProcessor, task Task) error {
	done := make(chan error, 1)
	wrappedTask := func() error {
		err := task()
		select {
		case done <- err:
		default:
		}
		return err
	}

	if err := wp.enqueue(wrappedTask); err != nil {
		return err
	}

	// Wait for the worker to finish the task.
	select {
	case err := <-done:
		return err
	case <-time.After(wp.config().Timeout):
		return TimeoutError{Duration: wp.config().Timeout}
	}
}

// Awaiter is returned by EnqueueAwaiting to wait for the worker to finish a task.
type Awaiter func() error

// EnqueueAwaiting passes a task to a WorkProcessor and returns an Awaiter to wait for the
// worker to finish the task. The timeout is the maximum time to wait for the
// worker to finish the task. After enqueuing the task different work can be done
// before waiting for the worker to finish the task.
func EnqueueAwaiting(wp WorkProcessor, task Task, timeout time.Duration) (Awaiter, error) {
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

	if err := wp.enqueue(wrappedTask); err != nil {
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

// Stop initiates a graceful shutdown of the WorkProcessor. All pending tasks
// will be processed before the worker stops.
func Stop(wp WorkProcessor) error {
	return wp.stop()
}

