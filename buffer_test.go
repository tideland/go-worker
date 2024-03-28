// -----------------------------------------------------------------------------
// worker for running tasks enqueued in background (buffer.go)
//
// Copyright (C) 2024 Frank Mueller / Oldenburg / Germany / World
// -----------------------------------------------------------------------------

package worker

import (
	"context"
	"fmt"
	"testing"
	"time"

	"tideland.dev/go/audit/asserts"
)

// -----------------------------------------------------------------------------
// Tests
// -----------------------------------------------------------------------------

// TestRatedBufferSetup tests the simple setup of a rated buffer.
func TestRatedBufferSetup(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.FailStop)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in, out := setupRatedBuffer(ctx, 1, 1, time.Second)

	assert.NotNil(in)
	assert.NotNil(out)
}

// TestRatedBufferIn tests the input of tasks in order to process them.
func TestRatedBufferIn(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.FailStop)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in, out := setupRatedBuffer(ctx, 100, 10, time.Second)

	assert.NotNil(in)
	assert.NotNil(out)

	assert.NoError(in(makeProcessATask("A")))
	assert.NoError(in(makeProcessATask("B")))
	assert.NoError(in(makeProcessATask("C")))
}

// TestRatedBufferInOut tests the input and output of tasks.
func TestRatedBufferInOut(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.FailStop)
	ctx, cancel := context.WithCancel(context.Background())

	in, out := setupRatedBuffer(ctx, 2, 10, time.Second)

	assert.NotNil(in)
	assert.NotNil(out)

	assert.NoError(in(makeProcessATask("A")))
	assert.NoError(in(makeProcessATask("B")))
	assert.NoError(in(makeProcessATask("C")))
	assert.NoError(in(makeShutdownATask()))

	for {
		atask := <-out()
		action, task := atask()
		if action == actionShutdown {
			break
		}
		assert.Logf("task error %q", task().Error())
	}
	cancel()
}

func TestRatedBufferRate(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.FailStop)
	ctx, cancel := context.WithCancel(context.Background())

	in, out := setupRatedBuffer(ctx, 2, 10, time.Second)
	expected := 500 * time.Millisecond

	assert.NotNil(in)
	assert.NotNil(out)

	assert.NoError(in(makeProcessATask("A")))
	assert.NoError(in(makeProcessATask("B")))
	assert.NoError(in(makeProcessATask("C")))
	assert.NoError(in(makeProcessATask("D")))
	assert.NoError(in(makeProcessATask("E")))
	assert.NoError(in(makeShutdownATask()))

	for {
		now := time.Now()
		atask := <-out()
		action, task := atask()
		duration := time.Since(now)
		assert.Logf("duration %v", duration)
		assert.About(float64(duration), float64(expected), float64(100*time.Millisecond))
		if action == actionShutdown {
			break
		}
		assert.Logf("task error %q", task().Error())
	}
	cancel()
}

// -----------------------------------------------------------------------------
// internal helper
// -----------------------------------------------------------------------------

func makeProcessATask(msg string) actionTask {
	return func() (action, Task) {
		return actionProcess, func() error { return fmt.Errorf(msg) }
	}
}

func makeShutdownATask() actionTask {
	return func() (action, Task) {
		return actionShutdown, nil
	}
}

// -----------------------------------------------------------------------------
// end of file
// -----------------------------------------------------------------------------
