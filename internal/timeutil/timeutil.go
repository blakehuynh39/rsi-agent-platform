package timeutil

import "time"

// PtrTime helps initialize optional time fields in struct literals.
func PtrTime(value time.Time) *time.Time {
	return &value
}
