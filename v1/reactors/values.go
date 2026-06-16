package reactors

import "time"

// Bool stores v in a new bool value and returns a pointer to it.
func Bool(v bool) *bool { return &v }

// Int stores v in a new int64 value and returns a pointer to it.
func Int(v int64) *int64 { return &v }

// Float stores v in a new float64 value and returns a pointer to it.
func Float(v float64) *float64 { return &v }

// Duration stores v in a new time.Duration value and returns a pointer to it.
func Duration(v time.Duration) *time.Duration { return &v }
