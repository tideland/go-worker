// Tideland Go Worker - Worker Pool Tests
//
// Copyright (C) 2014-2025 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package worker_test

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"tideland.dev/go/worker"
)

// TestWorkerPoolCreation tests the creation of a worker pool.
func TestWorkerPoolCreation(t *testing.T) {
	// Test with valid size
	pool, err := worker.NewWorkerPool(5, worker.DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create worker pool: %v", err)
	}
	defer worker.Stop(pool)

	if pool.Size() != 5 {
		t.Errorf("Expected pool size 5, got %d", pool.Size())
	}

	// Test with invalid size
	_, err = worker.NewWorkerPool(0, worker.DefaultConfig())
	if err == nil {
		t.Error("Expected error for pool size 0, got nil")
	}

	_, err = worker.NewWorkerPool(-1, worker.DefaultConfig())
	if err == nil {
		t.Error("Expected error for negative pool size, got nil")
	}
}

// TestWorkerPoolEnqueue tests basic task enqueueing to the pool.
func TestWorkerPoolEnqueue(t *testing.T) {
	pool, err := worker.NewWorkerPool(3, worker.DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create worker pool: %v", err)
	}
	defer worker.Stop(pool)

	executed := atomic.Int32{}
	var wg sync.WaitGroup

	// Enqueue multiple tasks
	for i := range 10 {
		wg.Add(1)
		task := func() error {
			executed.Add(1)
			wg.Done()
			return nil
		}

		err := worker.Enqueue(pool, task)
		if err != nil {
			t.Errorf("Failed to enqueue task %d: %v", i, err)
			wg.Done()
		}
	}

	// Wait for all tasks to complete
	wg.Wait()

	if executed.Load() != 10 {
		t.Errorf("Expected 10 tasks executed, got %d", executed.Load())
	}
}

// TestWorkerPoolEnqueueWaiting tests synchronous task execution.
func TestWorkerPoolEnqueueWaiting(t *testing.T) {
	pool, err := worker.NewWorkerPool(2, worker.DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create worker pool: %v", err)
	}
	defer worker.Stop(pool)

	// Test successful task
	result := make(chan int, 1)
	task := func() error {
		result <- 42
		return nil
	}

	err = worker.EnqueueWaiting(pool, task)
	if err != nil {
		t.Errorf("Failed to enqueue waiting task: %v", err)
	}

	select {
	case val := <-result:
		if val != 42 {
			t.Errorf("Expected result 42, got %d", val)
		}
	default:
		t.Error("Task was not executed")
	}

	// Test task with error
	taskErr := fmt.Errorf("task error")
	errorTask := func() error {
		return taskErr
	}

	err = worker.EnqueueWaiting(pool, errorTask)
	if err != taskErr {
		t.Errorf("Expected error %v, got %v", taskErr, err)
	}
}

// TestWorkerPoolDistribution tests that tasks are distributed across workers.
func TestWorkerPoolDistribution(t *testing.T) {
	poolSize := 4
	pool, err := worker.NewWorkerPool(poolSize, worker.DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create worker pool: %v", err)
	}
	defer worker.Stop(pool)

	// Track task execution
	executed := atomic.Int32{}
	var wg sync.WaitGroup

	// Enqueue many tasks to ensure distribution
	numTasks := 100

	for i := range numTasks {
		wg.Add(1)
		task := func() error {
			executed.Add(1)
			wg.Done()
			time.Sleep(10 * time.Millisecond) // Simulate work
			return nil
		}

		err := worker.Enqueue(pool, task)
		if err != nil {
			t.Errorf("Failed to enqueue task %d: %v", i, err)
			wg.Done()
		}
	}

	wg.Wait()

	if executed.Load() != int32(numTasks) {
		t.Errorf("Expected %d tasks executed, got %d", numTasks, executed.Load())
	}
}

// TestWorkerPoolStop tests graceful shutdown of the pool.
func TestWorkerPoolStop(t *testing.T) {
	pool, err := worker.NewWorkerPool(2, worker.DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create worker pool: %v", err)
	}

	executed := atomic.Int32{}
	var wg sync.WaitGroup

	// Enqueue tasks that take some time
	for i := range 5 {
		wg.Add(1)
		task := func() error {
			time.Sleep(50 * time.Millisecond)
			executed.Add(1)
			wg.Done()
			return nil
		}

		err := worker.Enqueue(pool, task)
		if err != nil {
			t.Errorf("Failed to enqueue task %d: %v", i, err)
			wg.Done()
		}
	}

	// Stop the pool
	err = worker.Stop(pool)
	if err != nil {
		t.Errorf("Failed to stop pool: %v", err)
	}

	// Wait to ensure all tasks completed
	wg.Wait()

	// All tasks should have been executed
	if executed.Load() != 5 {
		t.Errorf("Expected 5 tasks executed before stop, got %d", executed.Load())
	}

	// Try to enqueue after stop
	err = worker.Enqueue(pool, func() error { return nil })
	if err == nil {
		t.Error("Expected error when enqueueing to stopped pool")
	}
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
		if err != nil {
			t.Errorf("%s: Failed to enqueue task: %v", name, err)
		}

		// Test EnqueueWaiting
		err = worker.EnqueueWaiting(wp, func() error {
			executed.Add(1)
			return nil
		})
		if err != nil {
			t.Errorf("%s: Failed to enqueue waiting task: %v", name, err)
		}

		// Give time for async task to complete
		time.Sleep(100 * time.Millisecond)

		if executed.Load() != 2 {
			t.Errorf("%s: Expected 2 tasks executed, got %d", name, executed.Load())
		}
	}

	// Test with single Worker
	w, err := worker.New(worker.DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create worker: %v", err)
	}
	defer worker.Stop(w)
	testProcessor(w, "Worker")

	// Test with WorkerPool
	pool, err := worker.NewWorkerPool(3, worker.DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create worker pool: %v", err)
	}
	defer worker.Stop(pool)
	testProcessor(pool, "WorkerPool")
}

// TestPolymorphicUsage demonstrates using the same function with both Worker and WorkerPool.
func TestPolymorphicUsage(t *testing.T) {
	// Function that processes tasks using any WorkProcessor
	processTasks := func(wp worker.WorkProcessor, taskCount int) (int32, error) {
		executed := atomic.Int32{}
		var wg sync.WaitGroup

		for range taskCount {
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
	if err != nil {
		t.Fatalf("Failed to create worker: %v", err)
	}
	defer worker.Stop(w)

	count, err := processTasks(w, 5)
	if err != nil {
		t.Errorf("Worker: Failed to process tasks: %v", err)
	}
	if count != 5 {
		t.Errorf("Worker: Expected 5 tasks executed, got %d", count)
	}

	// Test with WorkerPool
	pool, err := worker.NewWorkerPool(3, worker.DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create worker pool: %v", err)
	}
	defer worker.Stop(pool)

	count, err = processTasks(pool, 10)
	if err != nil {
		t.Errorf("WorkerPool: Failed to process tasks: %v", err)
	}
	if count != 10 {
		t.Errorf("WorkerPool: Expected 10 tasks executed, got %d", count)
	}
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
		if err != nil {
			t.Errorf("%s: Failed to enqueue awaiting task: %v", name, err)
			return
		}

		// Do other work...
		time.Sleep(50 * time.Millisecond)

		// Now wait for the result
		err = awaiter()
		if err != nil {
			t.Errorf("%s: Awaiter returned error: %v", name, err)
		}
		if result != 123 {
			t.Errorf("%s: Expected result 123, got %d", name, result)
		}

		// Test timeout
		awaiter, err = worker.EnqueueAwaiting(wp, func() error {
			time.Sleep(200 * time.Millisecond)
			return nil
		}, 100*time.Millisecond)
		if err != nil {
			t.Errorf("%s: Failed to enqueue awaiting task: %v", name, err)
			return
		}

		err = awaiter()
		if err == nil {
			t.Errorf("%s: Expected timeout error, got nil", name)
		}
	}

	// Test with Worker
	w, err := worker.New(worker.DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create worker: %v", err)
	}
	defer worker.Stop(w)
	testAwaiting(w, "Worker")

	// Test with WorkerPool
	pool, err := worker.NewWorkerPool(2, worker.DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create worker pool: %v", err)
	}
	defer worker.Stop(pool)
	testAwaiting(pool, "WorkerPool")
}

// TestWorkerPoolWithCustomConfig tests pool creation with custom configuration.
func TestWorkerPoolWithCustomConfig(t *testing.T) {
	// Create custom error handler
	errorCount := atomic.Int32{}
	errorHandler := &testErrorHandler{
		handler: func(err worker.TaskError) {
			errorCount.Add(1)
		},
	}

	// Create config with custom settings
	cfg := worker.Config{
		Rate:         10,
		Burst:        20,
		Timeout:      2 * time.Second,
		ErrorHandler: errorHandler,
	}

	pool, err := worker.NewWorkerPool(3, cfg)
	if err != nil {
		t.Fatalf("Failed to create worker pool with custom config: %v", err)
	}
	defer worker.Stop(pool)

	// Enqueue tasks that fail
	for i := range 5 {
		err := worker.Enqueue(pool, func() error {
			return fmt.Errorf("task error %d", i)
		})
		if err != nil {
			t.Errorf("Failed to enqueue task %d: %v", i, err)
		}
	}

	// Wait for error handling
	time.Sleep(200 * time.Millisecond)

	if errorCount.Load() != 5 {
		t.Errorf("Expected 5 errors handled, got %d", errorCount.Load())
	}
}

// testErrorHandler implements ErrorHandler for testing.
type testErrorHandler struct {
	handler func(worker.TaskError)
}

func (h *testErrorHandler) HandleError(err worker.TaskError) {
	h.handler(err)
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

	for b.Loop() {
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
		for b.Loop() {
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
		for b.Loop() {
			_ = worker.Enqueue(pool, task)
		}
	})
}

// EOF
