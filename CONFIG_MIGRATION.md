# Worker Config Migration Guide

## Overview

The `Config` struct in the worker package has been refactored to use a builder pattern with private fields, setter methods, and comprehensive error accumulation. This provides better encapsulation, validation at the point of setting values, and the ability to detect multiple configuration errors at once.

## Key Changes

### 1. Private Fields

All configuration fields are now private and can only be accessed through getter methods:

- `Context()` - returns the configured context
- `Rate()` - returns the rate limit
- `Burst()` - returns the burst capacity
- `Timeout()` - returns the timeout duration
- `ShutdownTimeout()` - returns the shutdown timeout
- `ErrorHandler()` - returns the error handler

### 2. Constructor Function

Instead of creating a struct literal, use the constructor:

```go
// Old
cfg := worker.Config{
    Rate: 10,
    Burst: 20,
}

// New
cfg := worker.NewConfig(context.Background()).
    SetRate(10).
    SetBurst(20)
```

### 3. Method Chaining

All setter methods return `*Config` for fluent interface:

```go
cfg := worker.NewConfig(ctx).
    SetRate(100).
    SetBurst(200).
    SetTimeout(5 * time.Second).
    SetShutdownTimeout(10 * time.Second).
    SetErrorHandler(errorHandler)
```

### 4. Error Accumulation

Multiple validation errors are accumulated using `errors.Join`:

```go
cfg := worker.NewConfig(nil).
    SetRate(-1).        // Error: rate must be positive
    SetBurst(-5).       // Error: burst must be positive
    SetTimeout(-1)      // Error: timeout must be positive

if err := cfg.Error(); err != nil {
    // err contains ALL validation errors, not just the last one
    fmt.Println(err)
    // Output:
    // rate must be positive, got -1
    // burst must be positive, got -5
    // timeout must be positive, got -1s
}
```

### 5. Automatic Error Checking in Constructors

Both `worker.New()` and `worker.NewWorkerPool()` automatically check for configuration errors:

```go
cfg := worker.NewConfig(context.Background()).
    SetRate(-10).       // Invalid!
    SetBurst(-5)        // Invalid!

// Even if you forget to check cfg.Error()...
w, err := worker.New(cfg)  // This will return the configuration errors
if err != nil {
    // err contains all validation errors from the config
}
```

### 6. Immediate Validation

Values are validated when set, not when the worker is created:

```go
cfg := worker.NewConfig(context.Background()).
    SetRate(-10) // Validation happens here

// Check for errors immediately
if err := cfg.Error(); err != nil {
    // Handle configuration error
}
```

## Migration Examples

### Basic Migration

```go
// Old
cfg := worker.Config{
    Context:         ctx,
    Rate:            10,
    Burst:           20,
    Timeout:         5 * time.Second,
    ShutdownTimeout: 10 * time.Second,
}
w, err := worker.New(cfg)

// New
cfg := worker.NewConfig(ctx).
    SetRate(10).
    SetBurst(20).
    SetTimeout(5 * time.Second).
    SetShutdownTimeout(10 * time.Second)

if err := cfg.Error(); err != nil {
    // Handle configuration errors
    return err
}
w, err := worker.New(cfg)
```

### Using Default Config

```go
// Old
cfg := worker.DefaultConfig()

// New (same function, but returns *Config now)
cfg := worker.DefaultConfig()
```

### Worker Pool Migration

```go
// Old
pool, err := worker.NewWorkerPool(5, worker.Config{
    Rate:  10,
    Burst: 20,
})

// New
cfg := worker.NewConfig(context.Background()).
    SetRate(10).
    SetBurst(20)

if err := cfg.Error(); err != nil {
    return err
}
pool, err := worker.NewWorkerPool(5, cfg)
```

## Validation Rules

1. **Context**: Cannot be nil (defaults to context.Background() in constructor)
2. **Rate**: Must be positive (> 0)
3. **Burst**: Must be positive and >= rate
4. **Timeout**: Must be positive duration
5. **ShutdownTimeout**: Must be positive duration
6. **ErrorHandler**: Can be nil (errors will be ignored)

## Special Behaviors

### Auto-adjustment of Burst

When setting rate, if the current burst is less than the new rate, burst is automatically adjusted:

```go
cfg := worker.NewConfig(context.Background()).
    SetBurst(50).   // Set burst to 50
    SetRate(100)    // Rate is 100, burst auto-adjusts to 100

// cfg.Burst() returns 100, not 50
```

### Error Checking Options

You can check errors at any point:

```go
// Option 1: Check after all setters (recommended for early feedback)
cfg := worker.NewConfig(ctx).
    SetRate(10).
    SetBurst(20)

if err := cfg.Error(); err != nil {
    // Handle error
}

// Option 2: Check via Validate() (same as Error())
if err := cfg.Validate(); err != nil {
    // Handle error
}

// Option 3: Let New() check for you (safety net)
w, err := worker.New(cfg) // Automatically returns error if config is invalid

// Note: Even if you forget to check cfg.Error(), New() and NewWorkerPool()
// will catch configuration errors and return them, preventing invalid workers
// from being created.
```

## Benefits of the New Approach

1. **Better Encapsulation**: Private fields prevent direct manipulation
2. **Immediate Validation**: Errors are caught when values are set
3. **Error Accumulation**: All validation errors are reported, not just the first one
4. **Fluent Interface**: Clean, readable configuration code
5. **Type Safety**: Setter methods ensure correct types
6. **Immutable After Creation**: Once passed to New(), config cannot be accidentally modified
7. **Safety Net**: `New()` and `NewWorkerPool()` automatically check for configuration errors, preventing creation of invalid workers even if the user forgets to check

## Backward Compatibility

This is a breaking change. Code using the old `Config` struct literal syntax must be updated to use the new setter methods.
