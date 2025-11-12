# Tideland Go Worker

[![GitHub release](https://img.shields.io/github/release/tideland/go-worker.svg)](https://github.com/tideland/go-worker)
[![GitHub license](https://img.shields.io/badge/license-New%20BSD-blue.svg)](https://raw.githubusercontent.com/tideland/go-worker/main/LICENSE)
[![Go Module](https://img.shields.io/github/go-mod/go-version/tideland/go-worker)](https://github.com/tideland/go-worker/blob/main/go.mod)
[![GoDoc](https://godoc.org/tideland.dev/go/worker?status.svg)](https://pkg.go.dev/mod/tideland.dev/go/worker?tab=packages)

## Description

Handle synchronous and asynchronous tasks enqueued in the background using individual workers or worker pools. The worker package provides a simple configurable task queue that processes tasks in the background. Workers process tasks in FIFO order with configurable rate limiting.

### Features

- **Single Worker**: Process tasks sequentially with rate limiting
- **Worker Pool**: Distribute tasks across multiple workers for parallel processing
- **Unified Interface**: Both Worker and WorkerPool implement the `WorkProcessor` interface
- **Graceful Shutdown**: All pending tasks are processed before stopping
- **Error Handling**: Custom error handlers for task failures
- **Flexible Configuration**: Rate limiting, timeouts, and burst capacity

## Usage

### Creating a Worker

```go
// Create a worker with custom configuration
cfg := worker.NewConfig(context.Background()).
    SetRate(10).                        // 10 tasks per second
    SetBurst(50).                       // Buffer up to 50 tasks
    SetTimeout(time.Second).            // 1 second timeout for operations
    SetShutdownTimeout(5 * time.Second) // 5 seconds for graceful shutdown

w, err := worker.New(cfg)
if err != nil {
    log.Fatal(err)
}
defer worker.Stop(w)

// Or use default configuration
w, err := worker.New(nil) // Uses DefaultConfig()
```

### Enqueue Tasks

All operations are performed through command functions:

#### Enqueue a task for background processing:

```go
err := worker.Enqueue(w, func() error {
    // Do work
    return nil
})
```

#### Enqueue and wait for completion:

```go
err := worker.EnqueueWaiting(w, func() error {
    // Do work synchronously
    return nil
})
```

#### Enqueue and get an awaiter for later completion checking:

```go
awaiter, err := worker.EnqueueAwaiting(w, func() error {
    // Do work
    return nil
}, 5*time.Second)
// Do other work...
err = awaiter() // Wait for completion
```

### Wait for Tasks to Complete

Wait for all currently enqueued tasks to finish processing:

```go
// Wait with timeout
err := worker.WaitForTasks(w, 5*time.Second)
if err != nil {
    // Timeout occurred
    log.Printf("Tasks did not complete in time: %v", err)
}
```

This is useful for:

- Batch processing with synchronization points
- Ensuring work completes before shutdown
- Pipeline stages that need to wait for previous stages
- Testing and verification

Example with batch processing:

```go
// Process first batch
for _, item := range batch1 {
    worker.Enqueue(w, processItem(item))
}
worker.WaitForTasks(w, 10*time.Second) // Wait for batch to complete

// Process second batch only after first is done
for _, item := range batch2 {
    worker.Enqueue(w, processItem(item))
}
worker.WaitForTasks(w, 10*time.Second)
```

### Stop the Worker

```go
err := worker.Stop(w)
```

The worker will process all pending tasks before stopping. Error handling can be customized through the ErrorHandler in the configuration.

### Worker Pool

Create and use a pool of workers for parallel task processing:

```go
// Create a pool with 5 workers
pool, err := worker.NewWorkerPool(5, nil) // Uses DefaultConfig()
if err != nil {
    log.Fatal(err)
}
defer worker.Stop(pool)

// Enqueue tasks - they'll be distributed across workers
worker.Enqueue(pool, func() error {
    fmt.Println("Task executed by one of the pool workers")
    return nil
})
```

### WorkProcessor Interface

Both Worker and WorkerPool implement the `WorkProcessor` interface internally. The package-level functions accept this interface:

```go
// All these functions work with both Worker and WorkerPool:
worker.Enqueue(processor, task)
worker.EnqueueWaiting(processor, task)
worker.EnqueueAwaiting(processor, task, timeout)
worker.WaitForTasks(processor, timeout)
worker.Stop(processor)
```

This allows you to write code that works with both:

```go
func processTasksWithAnyProcessor(wp worker.WorkProcessor) {
    // Works with both Worker and WorkerPool!
    worker.Enqueue(wp, func() error {
        fmt.Println("Processing task...")
        return nil
    })
}

// Use with single worker
w, _ := worker.New(nil) // Uses DefaultConfig()
processTasksWithAnyProcessor(w)

// Use with pool
pool, _ := worker.NewWorkerPool(5, nil) // Uses DefaultConfig()
processTasksWithAnyProcessor(pool)
```

### Configuration Options

The configuration uses a fluent builder pattern:

```go
cfg := worker.NewConfig(context.Background()).
    SetRate(100).                        // Tasks per second (must be > 0)
    SetBurst(500).                       // Maximum burst size (must be >= rate)
    SetTimeout(5 * time.Second).        // Timeout for operations
    SetShutdownTimeout(10 * time.Second). // Timeout for graceful shutdown
    SetErrorHandler(errorHandler)        // Custom error handler (optional)

// Check for configuration errors
if err := cfg.Error(); err != nil {
    log.Fatal(err)
}
```

Use `worker.DefaultConfig()` for sensible defaults, or pass `nil` to `worker.New()`.

### Configuration Validation and Safety

The configuration system provides automatic validation and safety:

```go
// Configuration with errors
cfg := worker.NewConfig(context.TODO()).
    SetContext(nil).                     // Error: context cannot be nil
    SetRate(-10).                        // Error: rate must be positive
    SetBurst(5).                         // Error: burst must be >= rate
    SetTimeout(-1 * time.Second)         // Error: timeout must be positive

// Check for errors explicitly
if err := cfg.Error(); err != nil {
    // All validation errors are reported together:
    // context cannot be nil
    // rate must be positive, got -10
    // burst (5) cannot be less than rate (10)
    // timeout must be positive, got -1s
}

// Even if you forget to check, worker.New() will catch it
w, err := worker.New(cfg)
if err != nil {
    // Same validation errors reported - worker is NOT created
}
```

**Key safety features:**

- All validation errors are accumulated and reported together
- Invalid configurations will never create workers
- `worker.New()` and `NewWorkerPool()` always validate configuration
- Burst is automatically adjusted if needed when setting rate

### Error Handling

By default, errors are logged. You can provide a custom error handler:

```go
cfg := worker.NewConfig(context.Background()).
    SetErrorHandler(worker.NewDefaultErrorHandler(func(te worker.TaskError) {
        // Custom error handling logic
        log.Printf("Task error: %v at %v", te.Err, te.Timestamp)
    }))

w, _ := worker.New(cfg)
```

### Choosing Between Worker and WorkerPool

- Use a **single Worker** when:
  - Tasks must be processed in order
  - Rate limiting is more important than throughput
  - Resource usage needs to be minimal

- Use a **WorkerPool** when:
  - Tasks can be processed in parallel
  - Higher throughput is needed
  - You have CPU-bound or I/O-bound tasks that benefit from concurrency

## Examples

### Basic Task Processing

```go
package main

import (
    "fmt"
    "log"
    "time"

    "tideland.dev/go/worker"
)

func main() {
    // Create worker with default configuration
    w, err := worker.New(nil)
    if err != nil {
        log.Fatal(err)
    }
    defer worker.Stop(w)

    // Enqueue some tasks
    for i := 0; i < 10; i++ {
        taskNum := i
        worker.Enqueue(w, func() error {
            fmt.Printf("Processing task %d\n", taskNum)
            time.Sleep(100 * time.Millisecond)
            return nil
        })
    }

    // Wait a bit for tasks to complete
    time.Sleep(2 * time.Second)
}
```

### Parallel Processing with Worker Pool

```go
package main

import (
    "fmt"
    "log"
    "sync/atomic"
    "time"

    "tideland.dev/go/worker"
)

func main() {
    // Create a pool with 3 workers using default configuration
    pool, err := worker.NewWorkerPool(3, nil)
    if err != nil {
        log.Fatal(err)
    }
    defer worker.Stop(pool)

    var completed int32

    // Enqueue 20 tasks
    for i := 0; i < 20; i++ {
        taskNum := i
        worker.Enqueue(pool, func() error {
            fmt.Printf("Worker processing task %d\n", taskNum)
            time.Sleep(500 * time.Millisecond)
            atomic.AddInt32(&completed, 1)
            return nil
        })
    }

    // Wait for all tasks to complete
    for atomic.LoadInt32(&completed) < 20 {
        time.Sleep(100 * time.Millisecond)
    }

    fmt.Println("All tasks completed!")
}
```

## Contributors

- Frank Mueller (https://github.com/themue / https://github.com/tideland / https://themue.dev)
