// Tideland Go Worker - Configuration
//
// Copyright (C) 2014-2025 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package worker

import (
	"context"
	"time"
)

// Config contains all configuration options for a Worker.
type Config struct {
	// Context defines the lifetime of the Worker. If nil,
	// context.Background() will be used.
	Context context.Context

	// Rate is the number of tasks processed per second. Must be positive,
	// default is one task per second.
	Rate int

	// Burst is the number of tasks processed at once. If 0, Rate is used as Burst.
	Burst int

	// Timeout is the duration for different actions like enqueueing a task.
	// Must be positive, default is 5 seconds.
	Timeout time.Duration

	// ShutdownTimeout is the maximum duration to wait during graceful shutdown.
	// Must be positive, default is 5 seconds.
	ShutdownTimeout time.Duration

	// ErrorHandler allows custom error handling for task errors.
	// If nil, errors will be ignored.
	ErrorHandler ErrorHandler
}

// DefaultConfig returns a Config with default values.
func DefaultConfig() Config {
	return Config{
		Context:         context.Background(),
		Rate:            1,
		Burst:           0,
		Timeout:         5 * time.Second,
		ShutdownTimeout: 5 * time.Second,
	}
}

// Validate checks if the configuration is valid and
// sets default values where needed.
func (c *Config) Validate() error {
	// Set defaults for nil values.
	if c.Context == nil {
		c.Context = context.Background()
	}
	if c.Rate <= 0 {
		c.Rate = 1
	}
	if c.Burst < 0 {
		c.Burst = 0
	}
	if c.Timeout <= 0 {
		c.Timeout = 5 * time.Second
	}
	if c.ShutdownTimeout <= 0 {
		c.ShutdownTimeout = 5 * time.Second
	}
	// Set burst to rate if not configured.
	if c.Burst == 0 {
		c.Burst = c.Rate
	}
	return nil
}

// EOF
