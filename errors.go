// -----------------------------------------------------------------------------
// worker for running tasks enqueued in background (errors.go)
//
// Copyright (C) 2024 Frank Mueller / Oldenburg / Germany / World
// -----------------------------------------------------------------------------

package worker // import "tideland.dev/go/worker"

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

// ShuttingDownError signals that the worker processor is shutting down.
type ShuttingDownError struct{}

func (ShuttingDownError) Error() string {
	return "worker processor is shutting down"
}

// -----------------------------------------------------------------------------
// end of file
// -----------------------------------------------------------------------------
