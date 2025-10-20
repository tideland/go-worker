// Tideland Go Worker - Errors
//
// Copyright (C) 2014-2025 Frank Mueller / Tideland / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package worker

import (
	"fmt"
	"time"
)

// NotStartedError signals that the worker processor did not start.
type NotStartedError struct{}

func (NotStartedError) Error() string {
	return "worker processor did not start in time"
}

// TimeoutError signals that a task processing exceeded its timeout.
type TimeoutError struct {
	Duration time.Duration
}

func (e TimeoutError) Error() string {
	return fmt.Sprintf("task processing exceeded timeout of %v", e.Duration)
}

// ShuttingDownError signals that the worker processor is shutting down.
type ShuttingDownError struct{}

func (ShuttingDownError) Error() string {
	return "worker processor is shutting down"
}

// TaskError represents an error that occurred during task execution.
type TaskError struct {
	Err       error
	Timestamp time.Time
}

func (e TaskError) Error() string {
	return fmt.Sprintf("task failed at %v: %v", e.Timestamp.Format(time.RFC3339), e.Err)
}

func (e TaskError) Unwrap() error {
	return e.Err
}

// ErrorHandler defines the interface for custom error handling.
type ErrorHandler interface {
	HandleError(TaskError)
}

// DefaultErrorHandler provides a basic implementation of ErrorHandler.
type DefaultErrorHandler struct {
	onError func(TaskError)
}

// NewDefaultErrorHandler creates a new DefaultErrorHandler with the given error callback.
func NewDefaultErrorHandler(onError func(TaskError)) *DefaultErrorHandler {
	return &DefaultErrorHandler{
		onError: onError,
	}
}

func (h *DefaultErrorHandler) HandleError(err TaskError) {
	if h.onError != nil {
		h.onError(err)
	}
}

