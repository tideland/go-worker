// -----------------------------------------------------------------------------
// worker for running tasks enqueued in background (task.go)
//
// Copyright (C) 2024 Frank Mueller / Oldenburg / Germany / World
// -----------------------------------------------------------------------------

package worker // import "tideland.dev/go/worker"

// Task defines the signature of a task functionto be processed by a worker.
type Task func() error

// action allows to control the worker.
type action int

const (
	actionProcess action = iota
	actionShutdown
)

// actionTask combines an action with the tast to be processed.
type actionTask func() (action, Task)

// -----------------------------------------------------------------------------
// end of file
// -----------------------------------------------------------------------------
