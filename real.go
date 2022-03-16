package gotime

import (
	"fmt"
	"time"
)

type realtime struct{}

func (realtime) String() string {
	return fmt.Sprintf("realtime{now: %s}", time.Now())
}

func (realtime) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}

func (realtime) Now() time.Time {
	return time.Now()
}

func (realtime) Since(t time.Time) time.Duration {
	return time.Since(t)
}

func (realtime) Sleep(d time.Duration) {
	time.Sleep(d)
}

func (realtime) Timer(d time.Duration) Timer {
	return &timerWrapper{
		t: time.NewTimer(d),
	}
}

type timerWrapper struct {
	t *time.Timer
}

func (t *timerWrapper) C() <-chan time.Time        { return t.t.C }
func (t *timerWrapper) Reset(d time.Duration) bool { return t.t.Reset(d) }
func (t *timerWrapper) Stop() bool                 { return t.t.Stop() }
