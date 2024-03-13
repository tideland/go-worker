// -----------------------------------------------------------------------------
// worker for running tasks enqueued in background.
//
// Copyright (C) 2024 Frank Mueller / Oldenburg / Germany / World
// -----------------------------------------------------------------------------

package worker_test

import (
	"testing"

	"tideland.dev/go/audit/asserts"

	"tideland.dev/go/worker"
)

// -----------------------------------------------------------------------------
// Tests
// -----------------------------------------------------------------------------

// TestNewOK tests the simple creation of a worker.
func TestNewOK(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.FailStop)
	w, err := worker.New()
	assert.OK(err)
	assert.NotNil(w)
}

// TestStoppingOK tests the stopping of running worker.
func TestStoppingOK(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.FailStop)
	w, err := worker.New()
	assert.OK(err)
	assert.NotNil(w)

	err = worker.Stop(w)
	assert.OK(err)
}

// TestEnqueueOK tests the enqueuing of tasks in order to process them.
func TestEnqueueOK(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.FailStop)
	w, err := worker.New()
	assert.OK(err)
	assert.NotNil(w)

	count := 0
	task := func() error {
		count++
		return nil
	}
	err = worker.EnqueueFunc(w, task)
	assert.OK(err)
	err = worker.EnqueueFunc(w, task)
	assert.OK(err)
	err = worker.EnqueueFunc(w, task)
	assert.OK(err)

	err = worker.Stop(w)
	assert.OK(err)

	assert.Equal(count, 3)
}

// -----------------------------------------------------------------------------
// end of file
// -----------------------------------------------------------------------------
