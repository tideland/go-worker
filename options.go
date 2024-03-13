// -----------------------------------------------------------------------------
// worker for running tasks enqueued in background (options.go)
//
// Copyright (C) 2024 Frank Mueller / Oldenburg / Germany / World
// -----------------------------------------------------------------------------

package worker

import "time"

// Option defines the signature of an option setting function.
type Option func(act *Worker) error

// defaultQueueCap defines the default capacity of the task queue.
const defaultQueueCap = 16

// WithQueueCap defines the channel capacity for enqueued.
func WithQueueCap(c int) Option {
	return func(w *Worker) error {
		if c < defaultQueueCap {
			c = defaultQueueCap
		}
		w.tasks = make(chan Task, c)
		return nil
	}
}

const defaultTimeout = 5 * time.Second

func WithTimeout(d time.Duration) Option {
	return func(w *Worker) error {
		if d < 0 {
			d = defaultTimeout
		}
		w.timeout = d
		return nil
	}
}

// -----------------------------------------------------------------------------
// end of file
// -----------------------------------------------------------------------------
