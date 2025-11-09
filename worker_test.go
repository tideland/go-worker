// Tideland Go Worker - Unit Tests
//
// Copyright (C) 2014-2025 Frank Mueller / Tideland / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package worker_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"tideland.dev/go/asserts/verify"

	"tideland.dev/go/worker"
)

// TestNewOK verifies worker creation with default configuration.
func TestNewOK(t *testing.T) {
	w, err := worker.New(nil)
	verify.NoError(t, err)
	verify.NotNil(t, w)

	err = worker.Stop(w)
	verify.NoError(t, err)
}

// TestNewWithConfigOK verifies worker creation with custom configuration.
func TestNewWithConfigOK(t *testing.T) {
	cfg := worker.NewConfig(context.Background()).
		SetRate(100).
		SetBurst(200).
		SetTimeout(time.Second).
		SetShutdownTimeout(time.Second)

	verify.NoError(t, cfg.Error())

	w, err := worker.New(cfg)
	verify.NoError(t, err)
	verify.NotNil(t, w)

	err = worker.Stop(w)
	verify.NoError(t, err)
}

// TestNewWithInvalidConfig verifies error handling for invalid configuration.
func TestNewWithInvalidConfig(t *testing.T) {
	// Test negative rate
	cfg := worker.NewConfig(context.Background()).
		SetRate(-1)

	verify.Error(t, cfg.Error())
	_, err := worker.New(cfg)
	verify.Error(t, err)

	// Test multiple errors
	cfg = worker.NewConfig(context.Background()).
		SetRate(-5).
		SetBurst(-10).
		SetTimeout(-time.Second)

	err = cfg.Error()
	verify.Error(t, err)
	verify.ErrorContains(t, err, "rate must be positive")
	verify.ErrorContains(t, err, "burst must be positive")
	verify.ErrorContains(t, err, "timeout must be positive")

	// Test burst less than rate
	cfg = worker.NewConfig(context.Background()).
		SetRate(100).
		SetBurst(50)

	err = cfg.Error()
	verify.Error(t, err)
	verify.ErrorContains(t, err, "burst (50) cannot be less than rate (100)")
}

// TestNewWithInvalidConfigNilCheck verifies New() checks for configuration errors even if user doesn't.
func TestNewWithInvalidConfigNilCheck(t *testing.T) {
	// Create invalid config but don't check cfg.Error()
	cfg := worker.NewConfig(context.Background()).
		SetRate(-10).
		SetBurst(-5).
		SetTimeout(-time.Second)

	// New() should still catch the errors
	_, err := worker.New(cfg)
	verify.Error(t, err)
	verify.ErrorContains(t, err, "rate must be positive")
	verify.ErrorContains(t, err, "burst must be positive")
	verify.ErrorContains(t, err, "timeout must be positive")
}

// TestEnqueueOK verifies task enqueueing.
func TestEnqueueOK(t *testing.T) {
	w, err := worker.New(nil)
	verify.NoError(t, err)

	count := 0
	for range 3 {
		err = worker.Enqueue(w, func() error {
			count++
			return nil
		})
		verify.NoError(t, err)
	}

	// Give tasks time to complete.
	time.Sleep(100 * time.Millisecond)

	err = worker.Stop(w)
	verify.NoError(t, err)
	verify.Equal(t, 3, count)
}

// TestEnqueueWaitingOK verifies synchronous task execution.
func TestEnqueueWaitingOK(t *testing.T) {
	w, err := worker.New(nil)
	verify.NoError(t, err)

	// Task completing normally.
	count := 0
	err = worker.EnqueueWaiting(w, func() error {
		count++
		return nil
	})
	verify.NoError(t, err)
	verify.Equal(t, 1, count)

	// Task returning an error.
	testErr := errors.New("test error")
	err = worker.EnqueueWaiting(w, func() error {
		return testErr
	})
	verify.Equal(t, err, testErr)

	err = worker.Stop(w)
	verify.NoError(t, err)
}

// TestEnqueueWaitingTimeout verifies timeout during task execution.
func TestEnqueueWaitingTimeout(t *testing.T) {
	cfg := worker.NewConfig(context.Background()).
		SetTimeout(50 * time.Millisecond)

	verify.NoError(t, cfg.Error())

	w, err := worker.New(cfg)
	verify.NoError(t, err)

	// Task taking longer than timeout.
	err = worker.EnqueueWaiting(w, func() error {
		time.Sleep(200 * time.Millisecond)
		return nil
	})
	verify.Error(t, err)

	err = worker.Stop(w)
	verify.NoError(t, err)
}

// TestAsyncAwaitOK verifies asynchronous task execution with await.
func TestAsyncAwaitOK(t *testing.T) {
	w, err := worker.New(nil)
	verify.NoError(t, err)

	count := 0
	awaiter, err := worker.EnqueueAwaiting(w, func() error {
		time.Sleep(10 * time.Millisecond)
		count++
		return nil
	}, 1*time.Second) // Longer timeout
	verify.NoError(t, err)

	// Count should still be 0 as task is running asynchronously.
	verify.Equal(t, 0, count)

	// Wait for task completion.
	err = awaiter()
	verify.NoError(t, err)
	verify.Equal(t, 1, count)

	err = worker.Stop(w)
	verify.NoError(t, err)
}

// TestAsyncAwaitTimeout verifies timeout during async task await.
func TestAsyncAwaitTimeout(t *testing.T) {
	w, err := worker.New(nil)
	verify.NoError(t, err)

	awaiter, err := worker.EnqueueAwaiting(w, func() error {
		time.Sleep(100 * time.Millisecond)
		return nil
	}, 10*time.Millisecond)
	verify.NoError(t, err)

	err = awaiter()
	verify.Error(t, err)

	time.Sleep(150 * time.Millisecond) // Let task complete

	err = worker.Stop(w)
	verify.NoError(t, err)
}

// TestAsyncAwaitError verifies error propagation in async tasks.
func TestAsyncAwaitError(t *testing.T) {
	w, err := worker.New(nil)
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

// TestStopOK verifies graceful worker shutdown.
func TestStopOK(t *testing.T) {
	w, err := worker.New(nil)
	verify.NoError(t, err)

	// Enqueue a quick task.
	err = worker.Enqueue(w, func() error {
		time.Sleep(10 * time.Millisecond)
		return nil
	})
	verify.NoError(t, err)

	err = worker.Stop(w)
	verify.NoError(t, err)
}

// TestStopTimeout verifies timeout during shutdown.
func TestStopTimeout(t *testing.T) {
	cfg := worker.NewConfig(context.Background()).
		SetShutdownTimeout(50 * time.Millisecond)

	verify.NoError(t, cfg.Error())

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
	verify.Error(t, err)
}

// TestContextCancellation verifies worker stops when context is cancelled.
func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cfg := worker.NewConfig(ctx)

	w, err := worker.New(cfg)
	verify.NoError(t, err)

	// Enqueue tasks.
	count := atomic.Int32{}
	for range 5 {
		err = worker.Enqueue(w, func() error {
			count.Add(1)
			time.Sleep(100 * time.Millisecond)
			return nil
		})
		verify.NoError(t, err)
	}

	// Cancel context should stop worker.
	cancel()

	// Give time for processing to stop.
	time.Sleep(50 * time.Millisecond)

	// New tasks should fail.
	err = worker.Enqueue(w, func() error {
		return nil
	})
	verify.Error(t, err)

	err = worker.Stop(w)
	verify.NoError(t, err)
}

// TestRateLimiting verifies task rate limiting (basic check).
func TestRateLimiting(t *testing.T) {
	cfg := worker.NewConfig(context.Background()).
		SetRate(10) // 10 tasks per second

	verify.NoError(t, cfg.Error())

	w, err := worker.New(cfg)
	verify.NoError(t, err)

	start := time.Now()
	count := atomic.Int32{}

	// Enqueue multiple tasks
	for range 5 {
		err = worker.Enqueue(w, func() error {
			count.Add(1)
			return nil
		})
		verify.NoError(t, err)
	}

	// Wait for tasks to complete
	time.Sleep(100 * time.Millisecond)

	elapsed := time.Since(start)
	verify.True(t, elapsed >= 50*time.Millisecond) // Rate limiting should apply
	verify.Equal(t, int32(5), count.Load())

	err = worker.Stop(w)
	verify.NoError(t, err)
}

// TestErrorHandling verifies custom error handler.
func TestErrorHandling(t *testing.T) {
	errorCount := atomic.Int32{}
	errorHandler := worker.NewDefaultErrorHandler(func(err worker.TaskError) {
		errorCount.Add(1)
	})

	cfg := worker.NewConfig(context.Background()).
		SetErrorHandler(errorHandler)

	verify.NoError(t, cfg.Error())

	w, err := worker.New(cfg)
	verify.NoError(t, err)

	// Enqueue failing tasks.
	for range 3 {
		err = worker.Enqueue(w, func() error {
			return errors.New("task error")
		})
		verify.NoError(t, err)
	}

	// Give time for error handling.
	time.Sleep(100 * time.Millisecond)

	verify.Equal(t, int32(3), errorCount.Load())

	err = worker.Stop(w)
	verify.NoError(t, err)
}

// TestConfigChaining verifies fluent configuration API.
func TestConfigChaining(t *testing.T) {
	// Test successful chaining
	cfg := worker.NewConfig(context.Background()).
		SetRate(100).
		SetBurst(200).
		SetTimeout(2 * time.Second).
		SetShutdownTimeout(3 * time.Second).
		SetErrorHandler(nil)

	verify.NoError(t, cfg.Error())
	verify.Equal(t, 100, cfg.Rate())
	verify.Equal(t, 200, cfg.Burst())
	verify.Equal(t, 2*time.Second, cfg.Timeout())
	verify.Equal(t, 3*time.Second, cfg.ShutdownTimeout())
	verify.Nil(t, cfg.ErrorHandler())

	// Test error accumulation - intentionally using nil context to test validation
	cfg = worker.NewConfig(context.TODO()).
		SetContext(nil). // Intentionally nil to test validation error
		SetRate(0).
		SetBurst(-1).
		SetTimeout(0).
		SetShutdownTimeout(-time.Second)

	err := cfg.Error()
	verify.Error(t, err)
	// Should contain multiple errors
	verify.ErrorContains(t, err, "context cannot be nil")
	verify.ErrorContains(t, err, "rate must be positive")
	verify.ErrorContains(t, err, "burst must be positive")
	verify.ErrorContains(t, err, "timeout must be positive")
	verify.ErrorContains(t, err, "shutdown timeout must be positive")
}
