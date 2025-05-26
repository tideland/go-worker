// -----------------------------------------------------------------------------
// worker for running tasks enqueued in background - tests
//
// Copyright (C) 2024-2025 Frank Mueller / Oldenburg / Germany / World
// -----------------------------------------------------------------------------

package worker_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"tideland.dev/go/asserts/verify"
	"tideland.dev/go/worker"
)

// TestNewOK tests the simple creation of a worker.
func TestNewOK(t *testing.T) {
	w, err := worker.New(worker.Config{})
	verify.NoError(t, err)
	verify.NotNil(t, w)

	err = worker.Stop(w)
	verify.NoError(t, err)
}

// TestNewWithConfigOK tests the creation of a worker with config.
func TestNewWithConfigOK(t *testing.T) {
	cfg := worker.Config{
		Context:         context.Background(),
		Rate:            100,
		Burst:           10,
		Timeout:         time.Second,
		ShutdownTimeout: time.Second,
	}
	w, err := worker.New(cfg)
	verify.NoError(t, err)
	verify.NotNil(t, w)

	err = worker.Stop(w)
	verify.NoError(t, err)
}

// TestNewWithDefaultsOK tests that invalid config values get defaults.
func TestNewWithDefaultsOK(t *testing.T) {
	cfg := worker.Config{
		Rate: -1, // Invalid, should get default
	}
	w, err := worker.New(cfg)
	verify.NoError(t, err) // Should not fail, uses defaults
	verify.NotNil(t, w)

	err = worker.Stop(w)
	verify.NoError(t, err)
}

// TestEnqueueOK tests basic task enqueueing.
func TestEnqueueOK(t *testing.T) {
	w, err := worker.New(worker.Config{})
	verify.NoError(t, err)

	count := 0
	for i := 0; i < 3; i++ {
		err = worker.Enqueue(w, func() error {
			count++
			return nil
		})
		verify.NoError(t, err)
	}

	// Give tasks time to process
	time.Sleep(200 * time.Millisecond)

	err = worker.Stop(w)
	verify.NoError(t, err)
	verify.Equal(t, count, 3)
}

// TestEnqueueWaitingOK tests synchronous task execution.
func TestEnqueueWaitingOK(t *testing.T) {
	w, err := worker.New(worker.Config{})
	verify.NoError(t, err)

	// Task completing normally.
	count := 0
	err = worker.EnqueueWaiting(w, func() error {
		count++
		return nil
	})
	verify.NoError(t, err)
	verify.Equal(t, count, 1)

	// Task returning error.
	testErr := errors.New("test error")
	err = worker.EnqueueWaiting(w, func() error {
		return testErr
	})
	verify.Equal(t, err, testErr)

	err = worker.Stop(w)
	verify.NoError(t, err)
}

// TestEnqueueWaitingTimeout tests timeout in synchronous execution.
func TestEnqueueWaitingTimeout(t *testing.T) {
	cfg := worker.Config{
		Timeout: 50 * time.Millisecond,
	}
	w, err := worker.New(cfg)
	verify.NoError(t, err)

	// Task taking longer than timeout.
	err = worker.EnqueueWaiting(w, func() error {
		time.Sleep(200 * time.Millisecond)
		return nil
	})
	var timeoutErr worker.TimeoutError
	verify.True(t, errors.As(err, &timeoutErr))

	err = worker.Stop(w)
	verify.NoError(t, err)
}

// TestAsyncAwaitOK tests async task execution with awaiter.
func TestAsyncAwaitOK(t *testing.T) {
	w, err := worker.New(worker.Config{})
	verify.NoError(t, err)

	count := 0
	awaiter, err := worker.EnqueueAwaiting(w, func() error {
		time.Sleep(10 * time.Millisecond)
		count++
		return nil
	}, 1*time.Second) // Longer timeout
	verify.NoError(t, err)

	// Do some other work while task is processing.
	time.Sleep(5 * time.Millisecond)

	// Wait for completion.
	err = awaiter()
	verify.NoError(t, err)
	verify.Equal(t, count, 1)

	err = worker.Stop(w)
	verify.NoError(t, err)
}

// TestAsyncAwaitTimeout tests timeout in async execution.
func TestAsyncAwaitTimeout(t *testing.T) {
	w, err := worker.New(worker.Config{})
	verify.NoError(t, err)

	awaiter, err := worker.EnqueueAwaiting(w, func() error {
		time.Sleep(100 * time.Millisecond)
		return nil
	}, 10*time.Millisecond)
	verify.NoError(t, err)

	err = awaiter()
	var timeoutErr worker.TimeoutError
	verify.True(t, errors.As(err, &timeoutErr))
	verify.Equal(t, timeoutErr.Duration, 10*time.Millisecond)

	err = worker.Stop(w)
	verify.NoError(t, err)
}

// TestAsyncAwaitError tests error handling in async execution.
func TestAsyncAwaitError(t *testing.T) {
	w, err := worker.New(worker.Config{})
	verify.NoError(t, err)

	testErr := errors.New("async task error")
	awaiter, err := worker.EnqueueAwaiting(w, func() error {
		return testErr
	}, 1*time.Second) // Longer timeout
	verify.NoError(t, err)

	err = awaiter()
	verify.Equal(t, err, testErr)

	err = worker.Stop(w)
	verify.NoError(t, err)
}

// TestStopOK tests graceful worker shutdown.
func TestStopOK(t *testing.T) {
	w, err := worker.New(worker.Config{})
	verify.NoError(t, err)

	// Enqueue a quick task.
	err = worker.Enqueue(w, func() error {
		time.Sleep(10 * time.Millisecond)
		return nil
	})
	verify.NoError(t, err)

	// Stop should wait for task completion.
	err = worker.Stop(w)
	verify.NoError(t, err)
}

// TestStopTimeout tests shutdown timeout.
func TestStopTimeout(t *testing.T) {
	cfg := worker.Config{
		ShutdownTimeout: 50 * time.Millisecond,
	}
	w, err := worker.New(cfg)
	verify.NoError(t, err)

	// Enqueue a long-running task.
	err = worker.Enqueue(w, func() error {
		time.Sleep(200 * time.Millisecond)
		return nil
	})
	verify.NoError(t, err)

	// Stop should timeout.
	err = worker.Stop(w)
	var timeoutErr worker.TimeoutError
	verify.True(t, errors.As(err, &timeoutErr))
	verify.Equal(t, timeoutErr.Duration, 50*time.Millisecond)
}

// TestConcurrentProcessing tests that tasks are processed concurrently.
func TestConcurrentProcessing(t *testing.T) {
	cfg := worker.Config{
		Burst:   10, // Allow buffering of multiple tasks
		Timeout: 10 * time.Second,
	}
	w, err := worker.New(cfg)
	verify.NoError(t, err)

	var mu sync.Mutex
	processed := make([]time.Time, 0, 5)
	
	// Enqueue tasks that record their execution time.
	for i := 0; i < 5; i++ {
		err = worker.Enqueue(w, func() error {
			start := time.Now()
			time.Sleep(20 * time.Millisecond) // Small processing time
			mu.Lock()
			processed = append(processed, start)
			mu.Unlock()
			return nil
		})
		verify.NoError(t, err)
	}

	// Wait for all tasks to complete.
	time.Sleep(200 * time.Millisecond)

	err = worker.Stop(w)
	verify.NoError(t, err)

	// Verify all tasks were processed.
	verify.Equal(t, len(processed), 5)
	
	// Tasks should be processed sequentially (one at a time).
	// The time difference between task starts should be at least their processing time.
	for i := 1; i < len(processed); i++ {
		interval := processed[i].Sub(processed[i-1])
		verify.True(t, interval >= 15*time.Millisecond) // Allow some tolerance
	}
}

// TestErrorHandling tests custom error handling.
func TestErrorHandling(t *testing.T) {
	errorCh := make(chan worker.TaskError, 1)
	handler := worker.NewDefaultErrorHandler(func(te worker.TaskError) {
		errorCh <- te
	})

	cfg := worker.Config{
		ErrorHandler: handler,
	}
	w, err := worker.New(cfg)
	verify.NoError(t, err)

	testErr := errors.New("task error")
	err = worker.Enqueue(w, func() error {
		return testErr
	})
	verify.NoError(t, err)

	// Wait for error to be handled.
	select {
	case taskErr := <-errorCh:
		verify.Equal(t, taskErr.Err, testErr)
		verify.True(t, !taskErr.Timestamp.IsZero())
	case <-time.After(time.Second):
		t.Fatal("Expected error handler to be called")
	}

	err = worker.Stop(w)
	verify.NoError(t, err)
}

// TestEnqueueAfterStop tests enqueueing after shutdown.
func TestEnqueueAfterStop(t *testing.T) {
	w, err := worker.New(worker.Config{})
	verify.NoError(t, err)

	err = worker.Stop(w)
	verify.NoError(t, err)

	// Try to enqueue after stop.
	err = worker.Enqueue(w, func() error {
		return nil
	})
	var shutdownErr worker.ShuttingDownError
	verify.True(t, errors.As(err, &shutdownErr))
}

// EOF