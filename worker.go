// Tideland Go Worker - Worker
//
// Copyright (C) 2014-2025 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package worker

import (
	"context"
	"sync"
	"time"
)

// Worker is a simple configurable task queue that processes tasks in background.
type Worker struct {
	cfg    Config
	ctx    context.Context
	cancel context.CancelFunc

	taskCh   chan Task
	done     chan struct{}
	stopOnce sync.Once
	running  bool
	mu       sync.RWMutex
}

// New creates and starts a new worker with the given configuration.
// If no configuration is provided, default configuration is used.
func New(cfg Config) (*Worker, error) {
	// Handle empty configuration.
	if (Config{}) == cfg {
		cfg = DefaultConfig()
	}

	// Validate configuration.
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Create worker context.
	ctx, cancel := context.WithCancel(cfg.Context)

	// Create worker.
	w := &Worker{
		cfg:     cfg,
		ctx:     ctx,
		cancel:  cancel,
		taskCh:  make(chan Task, cfg.Burst),
		done:    make(chan struct{}),
		running: true,
	}

	// Start processing in background.
	go w.run()

	return w, nil
}

// enqueue adds a task to the worker's queue.
// This is an internal method used by commands.
func (w *Worker) enqueue(task Task) error {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if !w.running {
		return ShuttingDownError{}
	}

	select {
	case <-w.ctx.Done():
		return ShuttingDownError{}
	case w.taskCh <- task:
		return nil
	case <-time.After(w.cfg.Timeout):
		return TimeoutError{Duration: w.cfg.Timeout}
	}
}

// stop initiates graceful shutdown of the worker.
// This is an internal method used by commands.
func (w *Worker) stop() error {
	var err error
	w.stopOnce.Do(func() {
		w.mu.Lock()
		w.running = false
		w.mu.Unlock()

		// Cancel context to signal shutdown.
		w.cancel()

		// Wait for completion or timeout.
		select {
		case <-w.done:
			// Clean shutdown.
		case <-time.After(w.cfg.ShutdownTimeout):
			// Timeout during shutdown.
			err = TimeoutError{Duration: w.cfg.ShutdownTimeout}
		}
	})
	return err
}

// run is the main processing loop that runs in a separate goroutine.
func (w *Worker) run() {
	defer close(w.done)

	for {
		select {
		case <-w.ctx.Done():
			// Context cancelled, process remaining tasks and shutdown.
			w.processPendingTasks()
			return

		case task := <-w.taskCh:
			// Process task immediately.
			w.processTask(task)
		}
	}
}

// processTask executes a single task with error handling.
func (w *Worker) processTask(task Task) {
	if task == nil {
		return
	}

	// Execute task and handle any error.
	if err := task(); err != nil && w.cfg.ErrorHandler != nil {
		w.cfg.ErrorHandler.HandleError(TaskError{
			Err:       err,
			Timestamp: time.Now(),
		})
	}
}

// processPendingTasks processes all remaining tasks during shutdown.
func (w *Worker) processPendingTasks() {
	for {
		select {
		case task := <-w.taskCh:
			w.processTask(task)
		default:
			// No more tasks.
			return
		}
	}
}

// EOF
