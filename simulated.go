package gotime

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/mgb/gotime/internal/queue"
)

type simulation struct {
	// Mockable time for unit tests
	c Clock

	start time.Time
	drift time.Duration
	ratio float64

	timers      queue.TimeQueue
	timerCancel func()

	sync.RWMutex
}

func (s *simulation) String() string {
	// Doesn't lock, to prevent recursive locking
	return fmt.Sprintf("simulation{now: %s, start: %s, warp: %f, drift: %s, clock: %s, timers: %s}",
		s.lockedNow(),
		s.start,
		s.ratio,
		s.drift,
		s.c,
		s.timers,
	)
}

// lockedNow must only be used when holding the lock
func (s *simulation) lockedNow() time.Time {
	return s.start.Add(s.toSimulatedDuration(s.c.Now().Sub(s.start)) + s.drift)
}

func (s *simulation) toSimulatedDuration(d time.Duration) time.Duration {
	return time.Duration(float64(d) * s.ratio)
}

func (s *simulation) fromSimulatedDuration(d time.Duration) time.Duration {
	return time.Duration(float64(d) / s.ratio)
}

func (s *simulation) Add(d time.Duration) time.Time {
	s.Lock()
	defer s.Unlock()

	old := s.c.Now()
	t := old.Add(d)
	s.start = t
	s.drift = old.Sub(s.start)

	s.triggerTimers(t)

	return old

}

func (s *simulation) SetNow(t time.Time) time.Time {
	s.Lock()
	defer s.Unlock()

	old := s.c.Now()
	s.start = t
	s.drift = old.Sub(s.start)

	s.triggerTimers(t)

	return old
}

func (s *simulation) triggerTimers(t time.Time) {
	// Need to reset timers to the new time
	if oldestT, ok := s.timers.Peek(); ok {
		s.makeTimer(oldestT.Sub(t))
	}
}

func (s *simulation) SetWarpSpeed(ratio float64) error {
	// Require ratio to be a positive, non-infinite, non-NaN number
	if ratio <= 0 || math.IsNaN(ratio) || math.IsInf(ratio, 0) {
		return ErrNegativeRatio
	}

	s.Lock()
	defer s.Unlock()

	now := s.lockedNow()
	s.start = s.c.Now()
	s.drift = now.Sub(s.start)
	s.ratio = ratio

	// Need to reset timers to the new warp speed
	if oldestT, ok := s.timers.Peek(); ok {
		s.makeTimer(oldestT.Sub(now))
	}

	return nil
}

func (s *simulation) After(d time.Duration) <-chan time.Time {
	s.Lock()
	defer s.Unlock()

	oldestT, ok := s.timers.Peek()

	ch := make(chan time.Time, 1)
	t := s.lockedNow().Add(d)
	s.timers.Add(s.lockedNow().Add(d), ch)

	if !ok || oldestT.After(t) {
		// t is older than any other timer, create a new timer
		s.makeTimer(d)
	}
	return ch
}

// makeTimer must be called during a write lock
func (s *simulation) makeTimer(d time.Duration) {
	if s.timerCancel != nil {
		s.timerCancel()
	}

	ch := make(chan struct{}, 1)
	ctx, cancel := context.WithCancel(context.Background())
	s.timerCancel = cancel

	timer := s.c.Timer(s.fromSimulatedDuration(d))
	go func() {
		defer cancel()

		select {
		case <-timer.C():
			ch <- struct{}{}

		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C()
			}
			close(ch)
		}
	}()

	go func() {
		_, ok := <-ch
		if !ok {
			return
		}

		s.Lock()
		defer s.Unlock()

		now := s.lockedNow()
		for _, c := range s.timers.PopBeforeOrEqual(now) {
			c <- now
		}

		// Check to see if we need to make another timer for the next oldest remaining timer
		oldestT, ok := s.timers.Peek()
		if !ok {
			s.timerCancel = nil
			return
		}

		s.makeTimer(oldestT.Sub(now))
	}()
}

func (s *simulation) Now() time.Time {
	s.RLock()
	defer s.RUnlock()

	return s.lockedNow()
}

func (s *simulation) Since(t time.Time) time.Duration {
	s.RLock()
	defer s.RUnlock()

	return s.fromSimulatedDuration(s.Now().Sub(t))
}

func (s *simulation) Sleep(d time.Duration) {
	<-s.After(d)
}

func (s *simulation) Timer(d time.Duration) Timer {
	return s.newTimer(d)
}

func (s *simulation) newTimer(d time.Duration) *fakeTimer {
	s.Lock()
	defer s.Unlock()

	c := make(chan time.Time, 1)
	closeCh := make(chan struct{})
	done := make(chan struct{})

	go func() {
		defer close(done)

		select {
		case now := <-s.After(d):
			c <- now
		case <-closeCh:
		}
	}()

	return &fakeTimer{
		c:        c,
		close:    closeCh,
		done:     done,
		newTimer: s.newTimer,
	}
}
