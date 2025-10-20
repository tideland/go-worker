// Tideland Go Worker
//
// Copyright (C) 2014-2025 Frank Mueller / Tideland / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

// Package worker provides a simple configurable task queue that processes
// tasks in the background. The worker package supports both single workers
// that process tasks in FIFO order with configurable rate limiting, and
// worker pools that distribute tasks across multiple workers for parallel
// processing.
//
// Features:
//   - Single Worker: Process tasks sequentially with rate limiting
//   - Worker Pool: Distribute tasks across multiple workers for parallel processing
//   - Unified Interface: Both Worker and WorkerPool implement the WorkProcessor interface
//   - Graceful Shutdown: All pending tasks are processed before stopping
//   - Error Handling: Custom error handlers for task failures
//   - Flexible Configuration: Rate limiting, timeouts, and burst capacity
//
// Creating a Single Worker:
//
//	w, err := worker.New(worker.Config{
//		Rate:    10,               // 10 tasks per second
//		Burst:   5,                // Buffer up to 5 tasks
//		Timeout: time.Second,      // 1 second timeout for operations
//	})
//
// Creating a Worker Pool:
//
//	pool, err := worker.NewWorkerPool(5, worker.Config{
//		Rate:    10,               // 10 tasks per second per worker
//		Burst:   5,                // Buffer up to 5 tasks per worker
//		Timeout: time.Second,      // 1 second timeout for operations
//	})
//
// All operations are performed through command functions that work with
// both Worker and WorkerPool via the WorkProcessor interface:
//
// Enqueue a task for background processing:
//
//	err := worker.Enqueue(w, func() error {
//		// Do work
//		return nil
//	})
//
// Enqueue and wait for completion:
//
//	err := worker.EnqueueWaiting(w, func() error {
//		// Do work synchronously
//		return nil
//	})
//
// Enqueue and get an awaiter for later completion checking:
//
//	awaiter, err := worker.AsyncAwait(w, func() error {
//		// Do work
//		return nil
//	}, 5*time.Second)
//	// Do other work...
//	err = awaiter() // Wait for completion
//
// Stop the worker or pool gracefully:
//
//	err := worker.Stop(w)
//
// The worker will process all pending tasks before stopping. Error handling
// can be customized through the ErrorHandler in the configuration.
//
// Choosing Between Worker and WorkerPool:
//
// Use a single Worker when:
//   - Tasks must be processed in order
//   - Rate limiting is more important than throughput
//   - Resource usage needs to be minimal
//
// Use a WorkerPool when:
//   - Tasks can be processed in parallel
//   - Higher throughput is needed
//   - You have CPU-bound or I/O-bound tasks that benefit from concurrency

package worker // import "tideland.dev/go/worker"

