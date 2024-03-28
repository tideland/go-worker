// -----------------------------------------------------------------------------
// worker for running tasks enqueued in background (options.go)
//
// Copyright (C) 2024 Frank Mueller / Oldenburg / Germany / World
// -----------------------------------------------------------------------------

package worker // import "tideland.dev/go/worker"

import "time"

// Option defines the signature of an option setting function.
type Option func(act *Worker) error

const defaultRate = int(time.Second)
const defaultBurst = 0

// WithRateBurst sets the rate and the burst values of a worker. Rate is the
// number of tasks processed per second, burst the number of tasks processed
// at once (means enqueued to the worker at ounce). If burst is 0 the rate
// is used as burst.
//
// Be aware that a high rate and burst can lead to a high CPU load.
func WithRateBurst(rate, burst int) Option {
	return func(w *Worker) error {
		if rate < 0 {
			rate = defaultRate
		}
		if burst < 0 {
			burst = defaultBurst
		}
		w.rate = rate
		w.burst = burst
		return nil
	}
}

// WithTimeout sets the timeout for different actions like enqueueing a task.
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
