// -----------------------------------------------------------------------------
// worker for running tasks enqueued in background - documentation
//
// Copyright (C) 2024 Frank Mueller / Oldenburg / Germany / World
// -----------------------------------------------------------------------------

// Package worker provides a worker for running tasks enqueued in
// background. Each worker runs in its own goroutine and processes
// the tasks in a FIFO order. Once a task is enqueued using the Run
// method, it is processed by the worker and the methode immediately
// returns. The worker can be stopped using the Stop method.
//
// The EnqueueWaiting function is a convenience function for running
// a task and stops once the task is done.

package worker // import "tideland.dev/go/worker"

// -----------------------------------------------------------------------------
// end of file
// -----------------------------------------------------------------------------
