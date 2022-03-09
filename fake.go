package gotime

import (
	"fmt"
	"sync"
	"time"

	"github.com/mgb/gotime/internal/queue"
)

type faketime struct {
	now    time.Time
	timers queue.TimeQueue

	// Useful for unit tests, this buffered channel will signal when at least one timer was added since it was last read
	timerAdded chan struct{}

	sync.RWMutex
}

func (f *faketime) String() string {
	return fmt.Sprintf("fakeTime{now: %s, timers: %s}", f.now, f.timers)
}

func (f *faketime) SetNow(t time.Time) {
	f.Lock()
	defer f.Unlock()

	f.now = t

	// Trigger any timer that would pop with the new time
	for _, c := range f.timers.PopBeforeOrEqual(t) {
		c <- t
	}
}

func (f *faketime) After(d time.Duration) <-chan time.Time {
	if d <= 0 {
		// Trigger immediately if in the future
		ch := make(chan time.Time, 1)
		ch <- f.now
		return ch
	}

	f.Lock()
	defer f.Unlock()

	ch := make(chan time.Time, 1)
	f.timers.Add(f.now.Add(d), ch)
	f.notifyTimer()

	return ch
}

func (f *faketime) Now() time.Time {
	f.RLock()
	defer f.RUnlock()

	return f.now
}

func (f *faketime) Since(t time.Time) time.Duration {
	return f.Now().Sub(t)
}

func (f *faketime) Sleep(d time.Duration) {
	<-f.After(d)
}

func (f *faketime) Timer(d time.Duration) Timer {
	return f.newTimer(d)
}

func (f *faketime) newTimer(d time.Duration) *fakeTimer {
	if d <= 0 {
		// Trigger immediately if in the future
		c := make(chan time.Time, 1)
		c <- f.now
		done := make(chan struct{})
		close(done)
		return &fakeTimer{
			c:        c,
			done:     done,
			newTimer: f.newTimer,
		}
	}

	f.Lock()
	defer f.Unlock()

	c := make(chan time.Time, 1)
	closeCh := make(chan struct{})
	done := make(chan struct{})

	ch := make(chan time.Time, 1)
	cancel := f.timers.Add(f.now.Add(d), ch)
	f.notifyTimer()

	go func() {
		defer close(done)

		select {
		case now := <-ch:
			c <- now
		case <-closeCh:
			c <- f.Now()
			cancel()
		}
	}()

	return &fakeTimer{
		c:        c,
		close:    closeCh,
		done:     done,
		newTimer: f.newTimer,
	}
}

func (f *faketime) notifyTimer() {
	// Notify anyone waiting for a timer to be added in a non-blocking manner.
	// Useful for unit tests to then mutate the current time after a timer was added.
	select {
	case f.timerAdded <- struct{}{}:
	default:
	}
}

type fakeTimer struct {
	c chan time.Time

	// Close this channel to stop the goroutine
	close chan struct{}
	// Will be closed after the goroutine is stopped
	done chan struct{}

	newTimer func(d time.Duration) *fakeTimer
}

func (t *fakeTimer) C() <-chan time.Time {
	return t.c
}

func (t *fakeTimer) Reset(d time.Duration) bool {
	newT := t.newTimer(d)
	t.c = newT.c
	t.close = newT.close
	t.done = newT.done

	return false
}

func (t *fakeTimer) Stop() bool {
	select {
	case <-t.done:
		return true
	default:
	}

	close(t.close)
	return false
}
