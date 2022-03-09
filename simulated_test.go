package gotime

import (
	"sync"
	"testing"
	"time"
)

func TestSimulatedTime_After_fake(t *testing.T) {
	sim, f := newTimeWarpableClockWithFake(t)
	select {
	case <-f.timerAdded:
		t.Error("got timer added, want nothing")
	default:
	}

	ch := sim.After(time.Minute)
	select {
	case <-ch:
		t.Error("got value, want nothing")
	default:
	}

	var wg sync.WaitGroup
	wg.Add(1)

	var d time.Duration
	go func() {
		defer wg.Done()

		t := time.Now()
		<-ch
		d = time.Since(t)
	}()

	select {
	case <-f.timerAdded:
	case <-time.After(100 * time.Millisecond):
		t.Errorf("timer never added")
	}
	f.SetNow(f.Now().Add(time.Minute))

	wg.Wait()

	if d >= time.Second {
		t.Errorf("got %s, want < 1s", d)
	}
}

func TestSimulatedTime_Now_fake(t *testing.T) {
	sim, f := newTimeWarpableClockWithFake(t)

	if sim.Now() != f.Now() {
		t.Errorf("got %s, want %s", sim.Now(), f.Now())
	}
}

func TestSimulatedTime_Now_Warped_fake(t *testing.T) {
	sim, f := newTimeWarpableClockWithFake(t)

	expectedSimTime := f.Now()
	if sim.Now() != expectedSimTime {
		t.Errorf("got %s, want %s", sim.Now(), expectedSimTime)
	}

	f.SetNow(f.Now().Add(time.Minute))
	expectedSimTime = expectedSimTime.Add(time.Minute)
	if sim.Now() != expectedSimTime {
		t.Errorf("got %s, want %s", sim.Now(), expectedSimTime)
	}

	sim.SetWarpSpeed(60)
	if sim.Now() != expectedSimTime {
		t.Errorf("got %s, want %s", sim.Now(), expectedSimTime)
	}

	f.SetNow(f.Now().Add(time.Minute))
	expectedSimTime = expectedSimTime.Add(time.Hour)
	if sim.Now() != expectedSimTime {
		t.Errorf("got %s, want %s", sim.Now(), expectedSimTime)
	}

	sim.SetWarpSpeed(1)
	if sim.Now() != expectedSimTime {
		t.Errorf("got %s, want %s", sim.Now(), expectedSimTime)
	}

	f.SetNow(f.Now().Add(time.Minute))
	expectedSimTime = expectedSimTime.Add(time.Minute)
	if sim.Now() != expectedSimTime {
		t.Errorf("got %s, want %s", sim.Now(), expectedSimTime)
	}

	sim.SetWarpSpeed(60)
	if sim.Now() != expectedSimTime {
		t.Errorf("got %s, want %s", sim.Now(), expectedSimTime)
	}

	f.SetNow(f.Now().Add(time.Minute))
	expectedSimTime = expectedSimTime.Add(time.Hour)
	if sim.Now() != expectedSimTime {
		t.Errorf("got %s, want %s", sim.Now(), expectedSimTime)
	}

	sim.SetWarpSpeed(1 / 60.)
	if sim.Now() != expectedSimTime {
		t.Errorf("got %s, want %s", sim.Now(), expectedSimTime)
	}

	f.SetNow(f.Now().Add(time.Minute))
	expectedSimTime = expectedSimTime.Add(time.Second)
	if sim.Now() != expectedSimTime {
		t.Errorf("got %s, want %s", sim.Now(), expectedSimTime)
	}

}

func TestSimulatedTime_Since_fake(t *testing.T) {
	sim, f := newTimeWarpableClockWithFake(t)

	past := sim.Now()
	f.SetNow(f.Now().Add(time.Minute))
	if sim.Since(past) != time.Minute {
		t.Errorf("got %s, want %s", sim.Since(past), time.Minute)
	}
}

func TestSimulatedTime_Sleep_fake(t *testing.T) {
	sim, f := newTimeWarpableClockWithFake(t)

	var wg sync.WaitGroup
	wg.Add(1)

	var d time.Duration
	go func() {
		defer wg.Done()

		t := time.Now()
		sim.Sleep(time.Minute)
		d = time.Since(t)
	}()

	<-f.timerAdded
	f.SetNow(f.Now().Add(time.Minute))

	wg.Wait()

	if d >= time.Second {
		t.Errorf("got %s, want < 1s", d)
	}
}

func TestSimulatedTime_Sleep_Warped_fake(t *testing.T) {
	sim, f := newTimeWarpableClockWithFake(t)

	var wg sync.WaitGroup
	wg.Add(1)

	var d time.Duration
	go func() {
		defer wg.Done()

		t := time.Now()
		sim.Sleep(3 * time.Minute)
		d = time.Since(t)
	}()

	<-f.timerAdded

	sim.SetWarpSpeed(60)
	f.SetNow(f.Now().Add(time.Second))

	sim.SetWarpSpeed(1)
	f.SetNow(f.Now().Add(time.Minute))

	sim.SetWarpSpeed(60)
	f.SetNow(f.Now().Add(time.Second))

	wg.Wait()

	if d >= time.Second {
		t.Errorf("got %s, want < 1s", d)
	}
}

func TestSimulatedTime_Timer_fake(t *testing.T) {
	sim, f := newTimeWarpableClockWithFake(t)

	timer := sim.Timer(time.Minute)

	select {
	case <-timer.C():
		t.Error("got value, want nothing")
	default:
	}
	select {
	case <-f.timerAdded:
		t.Error("got timer added, want nothing")
	default:
	}

	var wg sync.WaitGroup
	wg.Add(1)

	var d time.Duration
	go func() {
		defer wg.Done()

		t := time.Now()
		<-timer.C()
		d = time.Since(t)
	}()

	<-f.timerAdded
	f.SetNow(f.Now().Add(time.Minute))

	wg.Wait()

	if d >= time.Second {
		t.Errorf("got %s, want < 1s", d)
	}
}

func newTimeWarpableClockWithFake(t *testing.T) (TimeWarpableClock, *faketime) {
	s := NewSettableClock()
	f, ok := s.(*faketime)
	if !ok {
		t.Fatalf("got %T, want *faketime", s)
	}
	c := NewTimeWarpableClock()
	sim := c.(*simulation)
	sim.c = f
	sim.start = f.Now()
	return c, f
}
