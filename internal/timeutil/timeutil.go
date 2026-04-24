package timeutil

import "time"

// PtrTime returns a pointer to the given time.Time value.
func PtrTime(value time.Time) *time.Time {
	return &value
}
