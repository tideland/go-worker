// Tideland Go Worker - Configuration Example
//
// Copyright (C) 2014-2025 Frank Mueller / Tideland / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"tideland.dev/go/worker"
)

func main() {
	// Example 1: Valid configuration with method chaining
	fmt.Println("=== Example 1: Valid Configuration ===")
	cfg1 := worker.NewConfig(context.Background()).
		SetRate(100).  // Process 100 tasks per second
		SetBurst(500). // Buffer up to 500 tasks
		SetTimeout(5 * time.Second).
		SetShutdownTimeout(10 * time.Second).
		SetErrorHandler(worker.NewDefaultErrorHandler(func(err worker.TaskError) {
			log.Printf("Task failed: %v at %v", err.Err, err.Timestamp)
		}))

	if err := cfg1.Error(); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	fmt.Printf("Rate: %d tasks/second\n", cfg1.Rate())
	fmt.Printf("Burst: %d tasks\n", cfg1.Burst())
	fmt.Printf("Timeout: %v\n", cfg1.Timeout())
	fmt.Printf("Shutdown Timeout: %v\n", cfg1.ShutdownTimeout())

	// Create a worker with the valid configuration
	w1, err := worker.New(cfg1)
	if err != nil {
		log.Fatalf("Failed to create worker: %v", err)
	}
	defer worker.Stop(w1)

	fmt.Println("\n=== Example 2: Invalid Configuration with Multiple Errors ===")
	// Example 2: Configuration with multiple validation errors
	// All errors are accumulated and wrapped together
	cfg2 := worker.NewConfig(context.TODO()).
		SetContext(nil).                     // Intentionally nil to demonstrate validation error
		SetRate(-10).                        // Error: rate must be positive
		SetBurst(-5).                        // Error: burst must be positive
		SetTimeout(-1 * time.Second).        // Error: timeout must be positive
		SetShutdownTimeout(-2 * time.Second) // Error: shutdown timeout must be positive

	// Check for accumulated errors
	if err := cfg2.Error(); err != nil {
		fmt.Printf("Configuration has errors:\n%v\n", err)

		// The error contains all validation errors joined together
		// You can also check for specific errors if needed
	}

	// Attempting to create a worker with invalid config will fail
	_, err = worker.New(cfg2)
	if err != nil {
		fmt.Printf("\nFailed to create worker (as expected): %v\n", err)
	}

	fmt.Println("\n=== Example 3: Burst Validation ===")
	// Example 3: Burst must be >= rate
	cfg3 := worker.NewConfig(context.Background()).
		SetRate(100).
		SetBurst(50) // Error: burst cannot be less than rate

	if err := cfg3.Error(); err != nil {
		fmt.Printf("Burst validation error: %v\n", err)
	}

	fmt.Println("\n=== Example 4: Auto-adjustment When Setting Rate ===")
	// Example 4: When rate is increased, burst is auto-adjusted if needed
	cfg4 := worker.NewConfig(context.Background()).
		SetBurst(50). // Set burst first
		SetRate(100)  // This auto-adjusts burst to 100 if it was less

	if err := cfg4.Error(); err != nil {
		fmt.Printf("Unexpected error: %v\n", err)
	} else {
		fmt.Printf("Rate: %d, Burst: %d (auto-adjusted)\n", cfg4.Rate(), cfg4.Burst())
	}

	fmt.Println("\n=== Example 5: Using Default Configuration ===")
	// Example 5: Using default configuration
	cfg5 := worker.DefaultConfig()

	fmt.Printf("Default Rate: %d\n", cfg5.Rate())
	fmt.Printf("Default Burst: %d\n", cfg5.Burst())
	fmt.Printf("Default Timeout: %v\n", cfg5.Timeout())
	fmt.Printf("Default Shutdown Timeout: %v\n", cfg5.ShutdownTimeout())

	// Create worker with defaults
	w5, err := worker.New(cfg5)
	if err != nil {
		log.Fatalf("Failed to create worker with defaults: %v", err)
	}
	defer worker.Stop(w5)

	fmt.Println("\n=== Example 6: Partial Configuration ===")
	// Example 6: Only set what you need, rest uses defaults
	cfg6 := worker.NewConfig(context.Background()).
		SetRate(50) // Only override rate, rest stays default

	if err := cfg6.Error(); err != nil {
		log.Fatalf("Unexpected error: %v", err)
	}

	fmt.Printf("Custom Rate: %d\n", cfg6.Rate())
	fmt.Printf("Default Burst (adjusted): %d\n", cfg6.Burst())
	fmt.Printf("Default Timeout: %v\n", cfg6.Timeout())

	fmt.Println("\nAll examples completed!")
}
