// Tideland Go Worker - Wait for Tasks Example
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
	"math/rand"
	"strings"
	"sync/atomic"
	"time"

	"tideland.dev/go/worker"
)

func main() {
	fmt.Println("=== Worker WaitForTasks Example ===")

	// Example 1: Basic usage - wait for all tasks to complete
	fmt.Println("Example 1: Basic WaitForTasks")
	example1BasicWait()

	fmt.Println("\n" + strings.Repeat("─", 50) + "\n")

	// Example 2: WaitForTasks with timeout
	fmt.Println("Example 2: WaitForTasks with timeout")
	example2WaitWithTimeout()

	fmt.Println("\n" + strings.Repeat("─", 50) + "\n")

	// Example 3: Batch processing with wait
	fmt.Println("Example 3: Batch processing")
	example3BatchProcessing()

	fmt.Println("\n" + strings.Repeat("─", 50) + "\n")

	// Example 4: Worker pool with WaitForTasks
	fmt.Println("Example 4: Worker pool parallel processing")
	example4WorkerPool()

	fmt.Println("\n" + strings.Repeat("─", 50) + "\n")

	// Example 5: Pipeline with wait points
	fmt.Println("Example 5: Pipeline processing with checkpoints")
	example5Pipeline()

	fmt.Println("\nAll examples completed!")
}

func example1BasicWait() {
	// Create a worker
	w, err := worker.New(nil) // Use default config
	if err != nil {
		log.Fatalf("Failed to create worker: %v", err)
	}
	defer worker.Stop(w)

	completed := &atomic.Int32{}

	// Enqueue several tasks
	fmt.Println("Enqueuing 5 tasks...")
	for i := 1; i <= 5; i++ {
		taskID := i
		err := worker.Enqueue(w, func() error {
			// Simulate work
			time.Sleep(100 * time.Millisecond)
			completed.Add(1)
			fmt.Printf("  Task %d completed\n", taskID)
			return nil
		})
		if err != nil {
			log.Fatalf("Failed to enqueue task: %v", err)
		}
	}

	fmt.Println("Waiting for all tasks to complete...")

	// Wait for all tasks with a 2-second timeout
	if err := worker.WaitForTasks(w, 2*time.Second); err != nil {
		log.Fatalf("Failed waiting for tasks: %v", err)
	}

	fmt.Printf("✓ All tasks completed! Total: %d\n", completed.Load())
}

func example2WaitWithTimeout() {
	cfg := worker.NewConfig(context.Background()).
		SetRate(2) // Slow processing rate

	w, err := worker.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create worker: %v", err)
	}
	defer worker.Stop(w)

	// Enqueue tasks that take time
	fmt.Println("Enqueuing slow tasks...")
	for i := 1; i <= 5; i++ {
		taskID := i
		err := worker.Enqueue(w, func() error {
			time.Sleep(300 * time.Millisecond)
			fmt.Printf("  Slow task %d completed\n", taskID)
			return nil
		})
		if err != nil {
			log.Fatalf("Failed to enqueue task: %v", err)
		}
	}

	fmt.Println("Waiting with short timeout (500ms)...")

	// Try to wait with a short timeout
	if err := worker.WaitForTasks(w, 500*time.Millisecond); err != nil {
		fmt.Printf("✗ Timeout occurred as expected: %v\n", err)
		fmt.Println("Some tasks may still be running...")

		// Wait longer to let them finish
		fmt.Println("Waiting again with longer timeout...")
		if err := worker.WaitForTasks(w, 5*time.Second); err != nil {
			log.Fatalf("Failed even with longer timeout: %v", err)
		}
		fmt.Println("✓ All tasks eventually completed")
	}
}

func example3BatchProcessing() {
	w, err := worker.New(nil)
	if err != nil {
		log.Fatalf("Failed to create worker: %v", err)
	}
	defer worker.Stop(w)

	// Process data in batches
	batches := [][]int{
		{1, 2, 3, 4, 5},
		{6, 7, 8, 9, 10},
		{11, 12, 13, 14, 15},
	}

	for batchNum, batch := range batches {
		fmt.Printf("Processing batch %d: %v\n", batchNum+1, batch)

		// Enqueue all items in the batch
		for _, item := range batch {
			value := item // Capture for closure
			err := worker.Enqueue(w, func() error {
				// Simulate processing
				time.Sleep(50 * time.Millisecond)
				fmt.Printf("  Processed item: %d\n", value)
				return nil
			})
			if err != nil {
				log.Fatalf("Failed to enqueue item: %v", err)
			}
		}

		// Wait for this batch to complete before moving to next
		fmt.Printf("Waiting for batch %d to complete...\n", batchNum+1)
		if err := worker.WaitForTasks(w, 5*time.Second); err != nil {
			log.Fatalf("Batch %d failed: %v", batchNum+1, err)
		}
		fmt.Printf("✓ Batch %d completed\n\n", batchNum+1)
	}

	fmt.Println("✓ All batches processed successfully")
}

func example4WorkerPool() {
	// Create a pool of 3 workers
	cfg := worker.NewConfig(context.Background()).
		SetRate(10).
		SetBurst(20)

	pool, err := worker.NewWorkerPool(3, cfg)
	if err != nil {
		log.Fatalf("Failed to create worker pool: %v", err)
	}
	defer worker.Stop(pool)

	start := time.Now()
	taskCount := 15
	completed := &atomic.Int32{}

	fmt.Printf("Distributing %d tasks across 3 workers...\n", taskCount)

	// Enqueue many tasks
	for i := 1; i <= taskCount; i++ {
		taskID := i
		err := worker.Enqueue(pool, func() error {
			// Random processing time
			duration := time.Duration(100+rand.Intn(200)) * time.Millisecond
			time.Sleep(duration)

			count := completed.Add(1)
			fmt.Printf("  Worker processed task %d (total completed: %d)\n", taskID, count)
			return nil
		})
		if err != nil {
			log.Fatalf("Failed to enqueue task: %v", err)
		}
	}

	fmt.Println("Waiting for all workers to complete their tasks...")

	// Wait for all tasks across all workers
	if err := worker.WaitForTasks(pool, 10*time.Second); err != nil {
		log.Fatalf("Pool tasks failed: %v", err)
	}

	elapsed := time.Since(start)
	fmt.Printf("✓ All %d tasks completed in %v (parallel processing)\n", taskCount, elapsed)
}

func example5Pipeline() {
	// Create workers for different pipeline stages
	fetcher, _ := worker.New(nil)
	processor, _ := worker.New(nil)
	saver, _ := worker.New(nil)

	defer worker.Stop(fetcher)
	defer worker.Stop(processor)
	defer worker.Stop(saver)

	// Shared state for pipeline
	fetchedData := make([]string, 0)
	processedData := make([]string, 0)

	fmt.Println("Starting 3-stage pipeline...")

	// Stage 1: Fetch data
	fmt.Println("\nStage 1: Fetching data...")
	for i := 1; i <= 5; i++ {
		id := i
		worker.Enqueue(fetcher, func() error {
			time.Sleep(100 * time.Millisecond)
			data := fmt.Sprintf("data-%d", id)
			fetchedData = append(fetchedData, data)
			fmt.Printf("  Fetched: %s\n", data)
			return nil
		})
	}

	// Wait for all fetching to complete
	if err := worker.WaitForTasks(fetcher, 5*time.Second); err != nil {
		log.Fatalf("Fetching failed: %v", err)
	}
	fmt.Printf("✓ Stage 1 complete: %d items fetched\n", len(fetchedData))

	// Stage 2: Process fetched data
	fmt.Println("\nStage 2: Processing data...")
	for _, data := range fetchedData {
		item := data
		worker.Enqueue(processor, func() error {
			time.Sleep(150 * time.Millisecond)
			processed := "processed-" + item
			processedData = append(processedData, processed)
			fmt.Printf("  Processed: %s\n", processed)
			return nil
		})
	}

	// Wait for all processing to complete
	if err := worker.WaitForTasks(processor, 5*time.Second); err != nil {
		log.Fatalf("Processing failed: %v", err)
	}
	fmt.Printf("✓ Stage 2 complete: %d items processed\n", len(processedData))

	// Stage 3: Save processed data
	fmt.Println("\nStage 3: Saving data...")
	savedCount := &atomic.Int32{}
	for _, data := range processedData {
		item := data
		worker.Enqueue(saver, func() error {
			time.Sleep(50 * time.Millisecond)
			savedCount.Add(1)
			fmt.Printf("  Saved: %s\n", item)
			return nil
		})
	}

	// Wait for all saving to complete
	if err := worker.WaitForTasks(saver, 5*time.Second); err != nil {
		log.Fatalf("Saving failed: %v", err)
	}
	fmt.Printf("✓ Stage 3 complete: %d items saved\n", savedCount.Load())

	fmt.Println("\n✓ Pipeline completed successfully!")
}
