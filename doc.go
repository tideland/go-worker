// Tideland Go Worker
//
// Copyright (C) 2014-2025 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

// Package worker provides a simple configurable task queue that processes
// tasks in the background. The worker processes tasks in FIFO order with
// configurable rate limiting.
//
// Creating a Worker:
//
//	w, err := worker.New(worker.Config{
//		Rate:    10,               // 10 tasks per second
//		Burst:   5,                // Buffer up to 5 tasks
//		Timeout: time.Second,      // 1 second timeout for operations
//	})
//
// All operations are performed through command functions:
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
// Stop the worker gracefully:
//
//	err := worker.Stop(w)
//
// The worker will process all pending tasks before stopping. Error handling
// can be customized through the ErrorHandler in the configuration.

package worker // import "tideland.dev/go/worker"

// EOF
