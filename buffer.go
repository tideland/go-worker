// -----------------------------------------------------------------------------
// worker for running tasks enqueued in background (buffer.go)
//
// Copyright (C) 2024 Frank Mueller / Oldenburg / Germany / World
// -----------------------------------------------------------------------------

package worker // import "tideland.dev/go/worker"

import (
	"context"
	"sync"
	"time"
)

type input func(at actionTask) error
type output func() <-chan actionTask

// ratedBuffer is a buffer that limits the number of tasks that can be retrieved from
// in a given time frame.
func setupRatedBuffer(
	ctx context.Context,
	rate int,
	burst int,
	timeout time.Duration,
) (input, output) {
	mu := sync.Mutex{}
	if rate < 0 {
		rate = 1
	}
	beat := time.Duration(int(time.Second) / rate)
	if burst <= 0 {
		burst = rate
	}
	buffer := make([]actionTask, 0, burst)
	// Create input function. It accepts a task and tries to add it to the buffer as long
	// as the burst limit is not reached. Otherwise it blocks until the buffer is ready.
	in := func(at actionTask) error {
		wait := time.Nanosecond
		countdown := timeout
		done := false
		for {
			select {
			case <-ctx.Done():
				return ShuttingDownError{}
			case <-time.After(wait):
				mu.Lock()
				if len(buffer) < burst {
					buffer = append(buffer, at)
					done = true
				}
				mu.Unlock()
				if done {
					return nil
				}
				// Increase wait time and decrease countdown.
				wait *= 2
				countdown -= wait
			case <-time.After(countdown):
				return TimeoutError{}
			}
		}
	}
	// Create output function. It returns a channel that is filled with tasks from the
	// buffer. If the buffer is empty it blocks. The delivery is limited to the given rate
	// based limit duration. If the context is done the channel is closed. out can be used
	// as iterator in a for task := range out() { ... } loop.
	out := func() <-chan actionTask {
		outc := make(chan actionTask, 1)
		go func() {
			ticker := time.NewTicker(beat)
			defer ticker.Stop()
			for range ticker.C {
				mu.Lock()
				if len(buffer) > 0 {
					outc <- buffer[0]
					buffer = buffer[1:]
				}
				mu.Unlock()
				// Check if the context is done.
				done, ok := <-ctx.Done()
				if ok && done == struct{}{} {
					close(outc)
					return
				}
			}
		}()
		return outc
	}
	return in, out
}

// -----------------------------------------------------------------------------
// end of file
// -----------------------------------------------------------------------------
