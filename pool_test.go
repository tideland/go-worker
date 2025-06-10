// Tideland Go Worker - Worker Pool Tests
//
// Copyright (C) 2014-2025 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package worker_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"tideland.dev/go/asserts/verify"
	"tideland.dev/go/worker"
)

// TestWorkerPoolCreation tests the creation of a worker pool.
func TestWorkerPoolCreation(t *testing.T) {
	// Test with valid size
	pool, err := worker.NewWorkerPool(5, worker.DefaultConfig())
	verify.NoError(t, err)
	verify.NotNil(t, pool)
	defer worker.Stop(pool)

	verify.Equal(t, pool.Size(), 5)

	// Test with invalid size
	_, err = worker.NewWorkerPool(0, worker.DefaultConfig())
	verify.NotNil(t, err)

	_, err = worker.NewWorkerPool(-1, worker.DefaultConfig())
	verify.NotNil(t, err)
}

// TestWorkerPoolEnqueue tests basic task enqueueing to the pool.
func TestWorkerPoolEnqueue(t *testing.T) {
	pool, err := worker.NewWorkerPool(3, worker.DefaultConfig())
	verify.NoError(t, err)
	defer worker.Stop(pool)

	executed := atomic.Int32{}
	var wg sync.WaitGroup

	// Enqueue multiple tasks
	for i := 0; i < 10; i++ {
		wg.Add(1)
		task := func() error {
			executed.Add(1)
			wg.Done()
			return nil
		}

		err := worker.Enqueue(pool, task)
		if err != nil {
			wg.Done()
		}
		verify.NoError(t, err)
	}

	// Wait for all tasks to complete
	wg.Wait()

	verify.Equal(t, executed.Load(), int32(10))
}

// TestWorkerPoolEnqueueWaiting tests synchronous task execution.
func TestWorkerPoolEnqueueWaiting(t *testing.T) {
	pool, err := worker.NewWorkerPool(2, worker.DefaultConfig())
	verify.NoError(t, err)
	defer worker.Stop(pool)

	// Test successful task
	result := make(chan int, 1)
	task := func() error {
		result <- 42
		return nil
	}

	err = worker.EnqueueWaiting(pool, task)
	verify.NoError(t, err)

	select {
	case val := <-result:
		verify.Equal(t, val, 42)
	default:
		t.Fatal("Task was not executed")
	}

	// Test task with error
	taskErr := fmt.Errorf("task error")
	errorTask := func() error {
		return taskErr
	}

	err = worker.EnqueueWaiting(pool, errorTask)
	verify.Equal(t, err, taskErr)
}

// TestWorkerPoolDistribution tests that tasks are distributed across workers.
func TestWorkerPoolDistribution(t *testing.T) {
	poolSize := 4
	pool, err := worker.NewWorkerPool(poolSize, worker.DefaultConfig())
	verify.NoError(t, err)
	defer worker.Stop(pool)

	// Track task execution
	executed := atomic.Int32{}
	var wg sync.WaitGroup

	// Enqueue many tasks to ensure distribution
	numTasks := 100

	for i := 0; i < numTasks; i++ {
		wg.Add(1)
		task := func() error {
			executed.Add(1)
			wg.Done()
			time.Sleep(10 * time.Millisecond) // Simulate work
			return nil
		}

		err := worker.Enqueue(pool, task)
		if err != nil {
			wg.Done()
		}
		verify.NoError(t, err)
	}

	wg.Wait()

	verify.Equal(t, executed.Load(), int32(numTasks))
}

// TestWorkerPoolStop tests graceful shutdown of the pool.
func TestWorkerPoolStop(t *testing.T) {
	pool, err := worker.NewWorkerPool(2, worker.DefaultConfig())
	verify.NoError(t, err)

	executed := atomic.Int32{}
	var wg sync.WaitGroup

	// Enqueue tasks that take some time
	for i := 0; i < 5; i++ {
		wg.Add(1)
		task := func() error {
			time.Sleep(50 * time.Millisecond)
			executed.Add(1)
			wg.Done()
			return nil
		}

		err := worker.Enqueue(pool, task)
		if err != nil {
			wg.Done()
		}
		verify.NoError(t, err)
	}

	// Stop the pool
	err = worker.Stop(pool)
	verify.NoError(t, err)

	// Wait to ensure all tasks completed
	wg.Wait()

	// All tasks should have been executed
	verify.Equal(t, executed.Load(), int32(5))

	// Try to enqueue after stop
	err = worker.Enqueue(pool, func() error { return nil })
	verify.NotNil(t, err)
	var shutdownErr worker.ShuttingDownError
	verify.AsError(t, err, &shutdownErr)
}

// TestWorkProcessorInterface tests the WorkProcessor interface with both Worker and WorkerPool.
func TestWorkProcessorInterface(t *testing.T) {
	// Test function that works with any WorkProcessor
	testProcessor := func(wp worker.WorkProcessor, name string) {
		executed := atomic.Int32{}

		// Test Enqueue
		err := worker.Enqueue(wp, func() error {
			executed.Add(1)
			return nil
		})
		verify.NoError(t, err)

		// Test EnqueueWaiting
		err = worker.EnqueueWaiting(wp, func() error {
			executed.Add(1)
			return nil
		})
		verify.NoError(t, err)

		// Give time for async task to complete
		time.Sleep(100 * time.Millisecond)

		verify.Equal(t, executed.Load(), int32(2))
	}

	// Test with single Worker
	w, err := worker.New(worker.DefaultConfig())
	verify.NoError(t, err)
	defer worker.Stop(w)
	testProcessor(w, "Worker")

	// Test with WorkerPool
	pool, err := worker.NewWorkerPool(3, worker.DefaultConfig())
	verify.NoError(t, err)
	defer worker.Stop(pool)
	testProcessor(pool, "WorkerPool")
}

// TestPolymorphicUsage demonstrates using the same function with both Worker and WorkerPool.
func TestPolymorphicUsage(t *testing.T) {
	// Function that processes tasks using any WorkProcessor
	processTasks := func(wp worker.WorkProcessor, taskCount int) (int32, error) {
		executed := atomic.Int32{}
		var wg sync.WaitGroup

		for i := 0; i < taskCount; i++ {
			wg.Add(1)
			err := worker.Enqueue(wp, func() error {
				executed.Add(1)
				wg.Done()
				return nil
			})
			if err != nil {
				wg.Done()
				return executed.Load(), err
			}
		}

		wg.Wait()
		return executed.Load(), nil
	}

	// Test with Worker
	w, err := worker.New(worker.DefaultConfig())
	verify.NoError(t, err)
	defer worker.Stop(w)

	count, err := processTasks(w, 5)
	verify.NoError(t, err)
	verify.Equal(t, count, int32(5))

	// Test with WorkerPool
	pool, err := worker.NewWorkerPool(3, worker.DefaultConfig())
	verify.NoError(t, err)
	defer worker.Stop(pool)

	count, err = processTasks(pool, 10)
	verify.NoError(t, err)
	verify.Equal(t, count, int32(10))
}

// TestEnqueueAwaiting tests the EnqueueAwaiting function with both Worker and WorkerPool.
func TestEnqueueAwaiting(t *testing.T) {
	testAwaiting := func(wp worker.WorkProcessor, name string) {
		// Test successful task
		result := 0
		awaiter, err := worker.EnqueueAwaiting(wp, func() error {
			result = 123
			return nil
		}, 2*time.Second)
		verify.NoError(t, err)

		// Do other work...
		time.Sleep(50 * time.Millisecond)

		// Now wait for the result
		err = awaiter()
		verify.NoError(t, err)
		verify.Equal(t, result, 123)

		// Test timeout
		awaiter, err = worker.EnqueueAwaiting(wp, func() error {
			time.Sleep(200 * time.Millisecond)
			return nil
		}, 100*time.Millisecond)
		verify.NoError(t, err)

		err = awaiter()
		verify.NotNil(t, err)
		var timeoutErr worker.TimeoutError
		verify.AsError(t, err, &timeoutErr)
		verify.Equal(t, timeoutErr.Duration, 100*time.Millisecond)
	}

	// Test with Worker
	w, err := worker.New(worker.DefaultConfig())
	verify.NoError(t, err)
	defer worker.Stop(w)
	testAwaiting(w, "Worker")

	// Test with WorkerPool
	pool, err := worker.NewWorkerPool(2, worker.DefaultConfig())
	verify.NoError(t, err)
	defer worker.Stop(pool)
	testAwaiting(pool, "WorkerPool")
}

// TestWorkerPoolWithCustomConfig tests pool creation with custom configuration.
func TestWorkerPoolWithCustomConfig(t *testing.T) {
	// Create custom error handler
	errorCount := atomic.Int32{}
	errorHandler := worker.NewDefaultErrorHandler(func(err worker.TaskError) {
		errorCount.Add(1)
	})

	// Create config with custom settings
	cfg := worker.Config{
		Rate:         10,
		Burst:        20,
		Timeout:      2 * time.Second,
		ErrorHandler: errorHandler,
	}

	pool, err := worker.NewWorkerPool(3, cfg)
	verify.NoError(t, err)
	defer worker.Stop(pool)

	// Enqueue tasks that fail
	for i := 0; i < 5; i++ {
		err := worker.Enqueue(pool, func() error {
			return fmt.Errorf("task error %d", i)
		})
		verify.NoError(t, err)
	}

	// Wait for error handling
	time.Sleep(200 * time.Millisecond)

	verify.Equal(t, errorCount.Load(), int32(5))
}

// TestWorkerPoolConcurrentEnqueue tests concurrent enqueueing to the pool.
func TestWorkerPoolConcurrentEnqueue(t *testing.T) {
	pool, err := worker.NewWorkerPool(5, worker.DefaultConfig())
	verify.NoError(t, err)
	defer worker.Stop(pool)

	executed := atomic.Int32{}
	var wg sync.WaitGroup

	// Start multiple goroutines enqueueing tasks
	numGoroutines := 10
	tasksPerGoroutine := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < tasksPerGoroutine; j++ {
				err := worker.Enqueue(pool, func() error {
					executed.Add(1)
					return nil
				})
				verify.NoError(t, err)
			}
		}(i)
	}

	wg.Wait()

	// Wait for all tasks to complete
	time.Sleep(200 * time.Millisecond)

	verify.Equal(t, executed.Load(), int32(numGoroutines*tasksPerGoroutine))
}

// TestWorkerPoolWithContext tests pool creation with context cancellation.
func TestWorkerPoolWithContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cfg := worker.Config{
		Context: ctx,
	}

	pool, err := worker.NewWorkerPool(2, cfg)
	verify.NoError(t, err)

	executed := atomic.Int32{}

	// Enqueue a task
	err = worker.Enqueue(pool, func() error {
		executed.Add(1)
		return nil
	})
	verify.NoError(t, err)

	// Give task time to execute
	time.Sleep(50 * time.Millisecond)
	verify.Equal(t, executed.Load(), int32(1))

	// Cancel context
	cancel()

	// Give time for shutdown
	time.Sleep(50 * time.Millisecond)

	// Try to enqueue after cancellation
	err = worker.Enqueue(pool, func() error {
		executed.Add(1)
		return nil
	})
	verify.NotNil(t, err)
}

// BenchmarkWorkerPoolEnqueue benchmarks task enqueueing performance.
func BenchmarkWorkerPoolEnqueue(b *testing.B) {
	pool, err := worker.NewWorkerPool(4, worker.DefaultConfig())
	if err != nil {
		b.Fatalf("Failed to create worker pool: %v", err)
	}
	defer worker.Stop(pool)

	task := func() error {
		// Minimal task
		return nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := worker.Enqueue(pool, task)
		if err != nil {
			b.Fatalf("Failed to enqueue task: %v", err)
		}
	}
}

// BenchmarkWorkerPoolVsSingleWorker compares pool performance with single worker.
func BenchmarkWorkerPoolVsSingleWorker(b *testing.B) {
	b.Run("SingleWorker", func(b *testing.B) {
		w, err := worker.New(worker.DefaultConfig())
		if err != nil {
			b.Fatalf("Failed to create worker: %v", err)
		}
		defer worker.Stop(w)

		task := func() error {
			time.Sleep(time.Microsecond) // Simulate work
			return nil
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = worker.Enqueue(w, task)
		}
	})

	b.Run("WorkerPool-4", func(b *testing.B) {
		pool, err := worker.NewWorkerPool(4, worker.DefaultConfig())
		if err != nil {
			b.Fatalf("Failed to create worker pool: %v", err)
		}
		defer worker.Stop(pool)

		task := func() error {
			time.Sleep(time.Microsecond) // Simulate work
			return nil
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = worker.Enqueue(pool, task)
		}
	})
}

// EOF