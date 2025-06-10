# Tideland Go Worker

[![GitHub release](https://img.shields.io/github/release/tideland/go-worker.svg)](https://github.com/tideland/go-worker)
[![GitHub license](https://img.shields.io/badge/license-New%20BSD-blue.svg)](https://raw.githubusercontent.com/tideland/go-worker/main/LICENSE)
[![Go Module](https://img.shields.io/github/go-mod/go-version/tideland/go-worker)](https://github.com/tideland/go-worker/blob/main/go.mod)
[![GoDoc](https://godoc.org/tideland.dev/go/worker?status.svg)](https://pkg.go.dev/mod/tideland.dev/go/worker?tab=packages)

## Description

Handle synchronous and asynchronous tasks enqueued in the background using individual workers or worker pools.

### Features

- **Single Worker**: Process tasks sequentially with rate limiting
- **Worker Pool**: Distribute tasks across multiple workers for parallel processing
- **Unified Interface**: Both Worker and WorkerPool implement the `WorkProcessor` interface
- **Graceful Shutdown**: All pending tasks are processed before stopping
- **Error Handling**: Custom error handlers for task failures
- **Flexible Configuration**: Rate limiting, timeouts, and burst capacity

## Usage

### Single Worker

```go
// Create a worker
w, err := worker.New(worker.DefaultConfig())
if err != nil {
    log.Fatal(err)
}
defer worker.Stop(w)

// Enqueue tasks
worker.Enqueue(w, func() error {
    fmt.Println("Task executed")
    return nil
})
```

### Worker Pool

```go
// Create a pool with 5 workers
pool, err := worker.NewWorkerPool(5, worker.DefaultConfig())
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
w, _ := worker.New(worker.DefaultConfig())
processTasksWithAnyProcessor(w)

// Use with pool
pool, _ := worker.NewWorkerPool(5, worker.DefaultConfig())
processTasksWithAnyProcessor(pool)
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

## Contributors

- Frank Mueller (https://github.com/themue / https://github.com/tideland / https://themue.dev)
