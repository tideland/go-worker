// Tideland Go Worker - Worker Pool
//
// Copyright (C) 2014-2025 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package worker

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// WorkProcessor defines the interface for types that can process tasks.
// This interface is implemented by both Worker and WorkerPool.
type WorkProcessor interface {
	// enqueue is the internal method for adding tasks.
	enqueue(task Task) error

	// stop is the internal method for graceful shutdown.
	stop() error

	// config returns the configuration of the processor.
	config() Config
}

// WorkerPool manages a pool of workers for parallel task processing.
type WorkerPool struct {
	workers  []*Worker
	size     int
	current  atomic.Uint64
	mu       sync.RWMutex
	stopped  bool
	stopOnce sync.Once
	cfg      Config
}

// NewWorkerPool creates a new worker pool with the specified size and configuration.
// Each worker in the pool is created with the provided configuration.
func NewWorkerPool(size int, cfg Config) (*WorkerPool, error) {
	if size <= 0 {
		return nil, fmt.Errorf("pool size must be positive, got %d", size)
	}

	// Handle empty configuration.
	if (Config{}) == cfg {
		cfg = DefaultConfig()
	}

	// Validate configuration.
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	pool := &WorkerPool{
		workers: make([]*Worker, size),
		size:    size,
		cfg:     cfg,
	}

	// Create workers
	for i := range size {
		worker, err := New(cfg)
		if err != nil {
			// Clean up already created workers
			for j := range i {
				_ = pool.workers[j].stop()
			}
			return nil, fmt.Errorf("failed to create worker %d: %w", i, err)
		}
		pool.workers[i] = worker
	}

	return pool, nil
}

// enqueue adds a task to one of the workers in the pool for background processing.
// It uses round-robin distribution to balance the load across workers.
func (p *WorkerPool) enqueue(task Task) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.stopped {
		return ShuttingDownError{}
	}

	// Select worker using round-robin
	index := p.current.Add(1) % uint64(p.size)
	worker := p.workers[index]

	return worker.enqueue(task)
}

// stop initiates a graceful shutdown of all workers in the pool.
// All pending tasks will be processed before the workers stop.
func (p *WorkerPool) stop() error {
	var firstErr error

	p.stopOnce.Do(func() {
		p.mu.Lock()
		p.stopped = true
		p.mu.Unlock()

		// Stop all workers
		var wg sync.WaitGroup
		errors := make(chan error, p.size)

		for _, worker := range p.workers {
			wg.Add(1)
			go func(w *Worker) {
				defer wg.Done()
				if err := w.stop(); err != nil {
					errors <- err
				}
			}(worker)
		}

		wg.Wait()
		close(errors)

		// Collect first error if any
		for err := range errors {
			if firstErr == nil {
				firstErr = err
			}
		}
	})

	return firstErr
}

// config returns the configuration of the worker pool.
func (p *WorkerPool) config() Config {
	return p.cfg
}

// Size returns the number of workers in the pool.
func (p *WorkerPool) Size() int {
	return p.size
}

// Ensure Worker implements WorkProcessor
var _ WorkProcessor = (*Worker)(nil)

// config returns the configuration of the worker.
// This method makes Worker implement the WorkProcessor interface.
func (w *Worker) config() Config {
	return w.cfg
}

// Ensure WorkerPool implements WorkProcessor
var _ WorkProcessor = (*WorkerPool)(nil)

// EOF
