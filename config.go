// Tideland Go Worker - Configuration
//
// Copyright (C) 2014-2025 Frank Mueller / Tideland / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package worker

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Config contains all configuration options for a Worker.
// All fields are private and can only be set through setter methods
// that validate the values. Setters return the config for chaining
// and accumulate any validation errors internally.
type Config struct {
	// Private fields
	context         context.Context
	rate            int
	burst           int
	timeout         time.Duration
	shutdownTimeout time.Duration
	errorHandler    ErrorHandler

	// Internal error for chaining - wraps all validation errors
	err error
}

// NewConfig creates a new Config with default values.
// If ctx is nil, context.Background() will be used.
func NewConfig(ctx context.Context) *Config {
	if ctx == nil {
		ctx = context.Background()
	}
	return &Config{
		context:         ctx,
		rate:            1,
		burst:           1,
		timeout:         5 * time.Second,
		shutdownTimeout: 5 * time.Second,
	}
}

// DefaultConfig returns a Config with default values using context.Background().
func DefaultConfig() *Config {
	return NewConfig(context.Background())
}

// SetContext sets the context for the Worker's lifetime.
// If ctx is nil, an error is recorded.
func (c *Config) SetContext(ctx context.Context) *Config {
	if ctx == nil {
		err := fmt.Errorf("context cannot be nil")
		c.wrapError(err)
		return c
	}
	c.context = ctx
	return c
}

// SetRate sets the number of tasks processed per second.
// Rate must be positive.
func (c *Config) SetRate(rate int) *Config {
	if rate <= 0 {
		err := fmt.Errorf("rate must be positive, got %d", rate)
		c.wrapError(err)
		return c
	}
	c.rate = rate
	// Adjust burst if it was set to match the old rate
	if c.burst < rate {
		c.burst = rate
	}
	return c
}

// SetBurst sets the number of tasks that can be buffered.
// Burst must be positive and cannot be less than rate.
func (c *Config) SetBurst(burst int) *Config {
	if burst <= 0 {
		err := fmt.Errorf("burst must be positive, got %d", burst)
		c.wrapError(err)
		return c
	}
	if burst < c.rate {
		err := fmt.Errorf("burst (%d) cannot be less than rate (%d)", burst, c.rate)
		c.wrapError(err)
		return c
	}
	c.burst = burst
	return c
}

// SetTimeout sets the duration for different actions like enqueueing a task.
// Timeout must be positive.
func (c *Config) SetTimeout(timeout time.Duration) *Config {
	if timeout <= 0 {
		err := fmt.Errorf("timeout must be positive, got %v", timeout)
		c.wrapError(err)
		return c
	}
	c.timeout = timeout
	return c
}

// SetShutdownTimeout sets the maximum duration to wait during graceful shutdown.
// ShutdownTimeout must be positive.
func (c *Config) SetShutdownTimeout(timeout time.Duration) *Config {
	if timeout <= 0 {
		err := fmt.Errorf("shutdown timeout must be positive, got %v", timeout)
		c.wrapError(err)
		return c
	}
	c.shutdownTimeout = timeout
	return c
}

// SetErrorHandler sets the error handler for task errors.
// ErrorHandler can be nil, in which case errors will be ignored.
func (c *Config) SetErrorHandler(handler ErrorHandler) *Config {
	c.errorHandler = handler
	return c
}

// Error returns any validation errors that occurred during configuration.
// Multiple errors are wrapped together.
func (c *Config) Error() error {
	return c.err
}

// Validate checks if the configuration is valid and returns any accumulated errors.
// This is typically called after chaining multiple setters.
func (c *Config) Validate() error {
	return c.err
}

// Context returns the configured context.
func (c *Config) Context() context.Context {
	return c.context
}

// Rate returns the configured rate.
func (c *Config) Rate() int {
	return c.rate
}

// Burst returns the configured burst.
func (c *Config) Burst() int {
	return c.burst
}

// Timeout returns the configured timeout.
func (c *Config) Timeout() time.Duration {
	return c.timeout
}

// ShutdownTimeout returns the configured shutdown timeout.
func (c *Config) ShutdownTimeout() time.Duration {
	return c.shutdownTimeout
}

// ErrorHandler returns the configured error handler.
func (c *Config) ErrorHandler() ErrorHandler {
	return c.errorHandler
}

// wrapError accumulates validation errors using error wrapping.
// This ensures all errors are preserved and can be unwrapped later.
func (c *Config) wrapError(err error) {
	if c.err == nil {
		c.err = err
	} else {
		c.err = errors.Join(c.err, err)
	}
}
