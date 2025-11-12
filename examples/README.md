# Worker Package Examples

This directory contains example programs demonstrating the usage of the Tideland Go Worker package.

## Examples

### Configuration Example (`config/`)

Demonstrates various configuration options and validation scenarios for the Worker package.

**Run:**

```bash
go run ./config/
```

**What it shows:**

- Valid configuration with method chaining
- Invalid configuration with multiple validation errors
- Burst validation (burst must be >= rate)
- Auto-adjustment when setting rate
- Using default configuration
- Partial configuration (override only what you need)

### Safety Example (`safety/`)

Demonstrates the safety features and error handling in the Worker package configuration.

**Run:**

```bash
go run ./safety/
```

**What it shows:**

- How `worker.New()` validates configuration even if you forget to check `cfg.Error()`
- Same safety for `WorkerPool` creation
- Best practices for early error detection
- How multiple configuration errors are accumulated and reported together
- The safety net that prevents runtime failures from bad configurations

### Wait for Tasks Example (`wait/`)

Demonstrates the `WaitForTasks` functionality for synchronizing task completion.

**Run:**

```bash
go run ./wait/
```

**What it shows:**

- Basic usage of `WaitForTasks` to wait for all enqueued tasks
- Handling timeouts when tasks take too long
- Batch processing with synchronization points
- Using `WaitForTasks` with worker pools for parallel processing
- Building pipelines with checkpoints between stages

## Building the Examples

To build standalone executables:

```bash
# Build configuration example
go build -o config_example ./config/

# Build safety example
go build -o safety_example ./safety/

# Build wait example
go build -o wait_example ./wait/

# Run the built executables
./config_example
./safety_example
./wait_example
```

## Key Takeaways

1. **Always validate configuration**: While `worker.New()` and `NewWorkerPool()` will catch configuration errors, it's best practice to check `cfg.Error()` immediately after configuration for better debugging.

2. **Configuration errors are accumulated**: Multiple validation errors are collected and reported together, making it easy to fix all issues at once.

3. **Safe by default**: Invalid configurations will never create workers, preventing runtime failures.

4. **Flexible configuration**: Use method chaining to configure only what you need; sensible defaults are provided for everything else.

5. **Task synchronization**: Use `WaitForTasks` to ensure all enqueued work completes before proceeding, perfect for batch processing or pipeline stages.

## License

Copyright (C) 2014-2025 Frank Mueller / Tideland / Germany

All rights reserved. Use of this source code is governed by the new BSD license.
