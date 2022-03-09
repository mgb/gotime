package gotime

import (
	"time"

	"github.com/mgb/gotime/internal/queue"
)

// Clock is a interface for common time functions for faking or simulatable purposes
// TODO: Add AfterFunc, Tick, Ticker
type Clock interface {
	After(d time.Duration) <-chan time.Time
	Now() time.Time
	Since(t time.Time) time.Duration
	Sleep(d time.Duration)
	Timer(d time.Duration) Timer
}

// Timer is an interface for time.Timer
type Timer interface {
	C() <-chan time.Time
	Reset(d time.Duration) bool
	Stop() bool
}

// SettableClock is a Clock that can be set to a specific time
type SettableClock interface {
	// SetNow sets the clock to the specified time. Timers will not be adjusted and will immediately trigger if time skips ahead of them.
	SetNow(t time.Time)

	Clock
}

// TimeWarpableClock is a Clock that can tick faster or slower than real time
type TimeWarpableClock interface {
	SetWarpSpeed(ratio float64) error

	SettableClock
}

// NewRealClock returns a realtime clock
func NewRealClock() Clock {
	return realtime{}
}

// NewSimulatedClock returns a clock that can be set to a specific time
func NewSettableClock() SettableClock {
	return &faketime{
		now:        time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC), // Obviously the start of the universe
		timers:     queue.NewTimeQueue(),
		timerAdded: make(chan struct{}, 1),
	}
}

// NewTimeWarpableClock returns a clock set to the current time with no warping
func NewTimeWarpableClock() TimeWarpableClock {
	return &simulation{
		c:      NewRealClock(),
		start:  time.Now(),
		ratio:  1,
		timers: queue.NewTimeQueue(),
	}
}
