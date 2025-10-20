// Tideland Go Worker - Task
//
// Copyright (C) 2014-2025 Frank Mueller / Tideland / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package worker

// Task defines the signature of a task function to be processed by a worker.
type Task func() error

