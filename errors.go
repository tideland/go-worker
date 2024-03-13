// -----------------------------------------------------------------------------
// worker for running tasks enqueued in background (errors.go)
//
// Copyright (C) 2024 Frank Mueller / Oldenburg / Germany / World
// -----------------------------------------------------------------------------

package worker

// stopError signals the worker to stop.
type stopError struct{}

func (stopError) Error() string {
	return "sopping worker"
}

// NotStartedError signals that the worker processor did not start.
type NotStartedError struct{}

func (NotStartedError) Error() string {
	return "worker processor did not start in time"
}

// TimeoutError signals that the worker processor did not start in time.
type TimeoutError struct{}

func (TimeoutError) Error() string {
	return "processing task lasts too long"
}

// -----------------------------------------------------------------------------
// end of file
// -----------------------------------------------------------------------------
