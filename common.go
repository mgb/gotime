package gotime

import "errors"

var (
	// ErrNegativeRatio is returned when the ratio is negative
	ErrNegativeRatio = errors.New("ratio must be positive")

	// ErrTimeInPast is returned when the time is in the past
	ErrTimeInPast = errors.New("time cannot go backwards")
)
