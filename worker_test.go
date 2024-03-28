// -----------------------------------------------------------------------------
// worker for running tasks enqueued in background.
//
// Copyright (C) 2024 Frank Mueller / Oldenburg / Germany / World
// -----------------------------------------------------------------------------

package worker_test

import (
	"context"
	"testing"
	"time"

	"tideland.dev/go/audit/asserts"

	"tideland.dev/go/worker"
)

// -----------------------------------------------------------------------------
// Tests
// -----------------------------------------------------------------------------

// TestNewOK tests the simple creation of a worker.
func TestNewOK(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.FailStop)
	w, err := worker.New(context.TODO())
	assert.OK(err)
	assert.NotNil(w)
}

// TestStoppingOK tests the stopping of running worker.
func TestStoppingOK(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.FailStop)
	w, err := worker.New(context.TODO())
	assert.OK(err)
	assert.NotNil(w)

	err = worker.Shutdown(w)
	assert.OK(err)
}

// TestEnqueueOK tests the enqueuing of tasks in order to process them.
func TestEnqueueOK(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.FailStop)
	w, err := worker.New(context.TODO())
	assert.OK(err)
	assert.NotNil(w)

	// Enqueue some counter tasks.
	count := 0
	task := func() error {
		count++
		return nil
	}
	err = worker.Enqueue(w, task)
	assert.OK(err)
	err = worker.Enqueue(w, task)
	assert.OK(err)
	err = worker.Enqueue(w, task)
	assert.OK(err)

	// Stop the worker and check the count.
	err = worker.Shutdown(w)
	assert.OK(err)

	assert.Equal(count, 3)
}

func TestAsyncAwaitOK(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.FailStop)
	w, err := worker.New(context.TODO())
	assert.OK(err)
	assert.NotNil(w)

	count := 0
	task := func() error {
		count++
		return nil
	}
	await, err := worker.AsyncAwait(w, task, 1*time.Second)
	assert.OK(err)

	// Simulate some work.
	simulator := 0
	for i := 0; i < 50; i++ {
		simulator++
	}

	// Wait for the worker to finish the task.
	err = await()
	assert.OK(err)

	err = worker.Shutdown(w)
	assert.OK(err)

	assert.Equal(count, 1)
}

// -----------------------------------------------------------------------------
// end of file
// -----------------------------------------------------------------------------
