// Tideland Go Worker - Safety Example
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
	"strings"
	"time"

	"tideland.dev/go/worker"
)

func main() {
	fmt.Println("=== Worker Config Safety Example ===")

	// Example 1: Forgetting to check configuration errors
	fmt.Println("Example 1: Configuration with errors, but forgot to check cfg.Error()")

	cfg1 := worker.NewConfig(context.Background()).
		SetRate(-10).                        // Invalid: negative rate
		SetBurst(-5).                        // Invalid: negative burst
		SetTimeout(-1 * time.Second).        // Invalid: negative timeout
		SetShutdownTimeout(-2 * time.Second) // Invalid: negative shutdown timeout

	// Oops! Developer forgot to check cfg.Error()
	// But worker.New() will catch it anyway!

	fmt.Println("Creating worker without checking cfg.Error()...")
	w1, err := worker.New(cfg1)
	if err != nil {
		fmt.Printf("✓ worker.New() caught the configuration errors:\n%v\n", err)
		fmt.Println("Worker was NOT created (w1 is nil)")
	} else {
		fmt.Println("✗ This shouldn't happen - config had errors!")
		defer worker.Stop(w1)
	}

	fmt.Println("\n" + strings.Repeat("-", 50) + "\n")

	// Example 2: Same safety for WorkerPool
	fmt.Println("Example 2: WorkerPool with invalid configuration")

	cfg2 := worker.NewConfig(context.Background()).
		SetRate(100).
		SetBurst(50) // Invalid: burst less than rate

	// Again, forgot to check cfg.Error()
	fmt.Println("Creating worker pool without checking cfg.Error()...")
	pool, err := worker.NewWorkerPool(3, cfg2)
	if err != nil {
		fmt.Printf("✓ NewWorkerPool() caught the configuration error:\n%v\n", err)
		fmt.Println("Pool was NOT created (pool is nil)")
	} else {
		fmt.Println("✗ This shouldn't happen - config had errors!")
		defer worker.Stop(pool)
	}

	fmt.Println("\n" + strings.Repeat("-", 50) + "\n")

	// Example 3: Best practice - check early for better debugging
	fmt.Println("Example 3: Best practice - check cfg.Error() immediately")

	cfg3 := worker.NewConfig(context.Background()).
		SetRate(-10).
		SetBurst(-5)

	// Check immediately after configuration
	if err := cfg3.Error(); err != nil {
		fmt.Printf("✓ Early detection of configuration errors:\n%v\n", err)
		fmt.Println("Can handle this before attempting to create worker")

		// Fix the configuration
		cfg3 = worker.NewConfig(context.Background()).
			SetRate(10).
			SetBurst(20)

		if err := cfg3.Error(); err != nil {
			log.Fatalf("Still has errors: %v", err)
		}
		fmt.Println("\n✓ Configuration fixed!")
	}

	// Now create the worker with valid config
	w3, err := worker.New(cfg3)
	if err != nil {
		log.Fatalf("Unexpected error: %v", err)
	}
	fmt.Println("✓ Worker created successfully with valid configuration")
	defer worker.Stop(w3)

	fmt.Println("\n" + strings.Repeat("-", 50) + "\n")

	// Example 4: Demonstrating error accumulation
	fmt.Println("Example 4: Multiple errors are all reported")

	cfg4 := worker.NewConfig(context.TODO()). // Start with TODO context
							SetContext(nil).               // Intentionally nil to demonstrate validation error
							SetRate(0).                    // Error 2: zero rate
							SetBurst(-100).                // Error 3: negative burst
							SetTimeout(0).                 // Error 4: zero timeout
							SetShutdownTimeout(-time.Hour) // Error 5: negative shutdown timeout

	_, err = worker.New(cfg4)
	if err != nil {
		fmt.Println("✓ All configuration errors were caught:")
		// The error contains all validation errors joined together
		errorLines := strings.Split(err.Error(), "\n")
		for i, line := range errorLines {
			fmt.Printf("  Error %d: %s\n", i+1, line)
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("Summary:")
	fmt.Println("1. worker.New() and NewWorkerPool() always validate configuration")
	fmt.Println("2. Invalid configurations will never create workers")
	fmt.Println("3. All validation errors are reported together")
	fmt.Println("4. Best practice: check cfg.Error() early for better debugging")
	fmt.Println("5. This safety net prevents runtime failures from bad configs")
}
