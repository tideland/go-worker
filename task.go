// -----------------------------------------------------------------------------
// worker for running tasks enqueued in background - task
//
// Copyright (C) 2024-2025 Frank Mueller / Oldenburg / Germany / World
// -----------------------------------------------------------------------------

package worker

// Task defines the signature of a task function to be processed by a worker.
type Task func() error

// EOF