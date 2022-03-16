package gotime

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestSetNow(t *testing.T) {
	tests := []time.Time{
		time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
		time.Date(-44, time.March, 1, 15, 15, 14, 16, time.UTC),
	}
	for _, tt := range tests {
		t.Run(fmt.Sprint(tt), func(t *testing.T) {
			c := NewSettableClock()

			prev := c.Now()
			old := c.SetNow(tt)
			if old != prev {
				t.Errorf("got %s, want %s", old, prev)
			}

			t.Logf("SettableClock: %s", c)

			now := c.Now()
			if now != tt {
				t.Errorf("got %s, want %s", now, tt)
			}
		})
	}
}

func TestAdd(t *testing.T) {
	tests := []struct {
		now time.Time
		d   time.Duration
		exp time.Time
	}{
		{
			now: time.Date(2020, time.January, 1, 12, 0, 0, 0, time.UTC),
			d:   time.Hour,
			exp: time.Date(2020, time.January, 1, 13, 0, 0, 0, time.UTC),
		},
		{
			now: time.Date(2020, time.January, 1, 12, 0, 0, 0, time.UTC),
			d:   -time.Second,
			exp: time.Date(2020, time.January, 1, 11, 59, 59, 0, time.UTC),
		},
		{
			now: time.Date(2020, time.January, 1, 12, 0, 0, 0, time.UTC),
			d:   0,
			exp: time.Date(2020, time.January, 1, 12, 0, 0, 0, time.UTC),
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprint(tt.d), func(t *testing.T) {
			c := NewSettableClock()
			c.SetNow(tt.now)

			old := c.Add(tt.d)
			if old != tt.now {
				t.Errorf("got %s, want %s", old, tt.now)
			}

			now := c.Now()
			if now != tt.exp {
				t.Errorf("got %s, want %s", now, tt.exp)
			}
		})
	}
}

func TestAfter(t *testing.T) {
	c := NewSettableClock()

	ch := c.After(time.Second)
	select {
	case <-ch:
		t.Error("got value, want nothing")
	default:
	}

	c.SetNow(time.Now().Add(time.Second))
	select {
	case <-ch:
	default:
		t.Error("got nothing, want value")
	}

	ch = c.After(-time.Second)
	select {
	case <-ch:
	default:
		t.Error("got nothing, want value")
	}

	ch = c.After(0)
	select {
	case <-ch:
	default:
		t.Error("got nothing, want value")
	}
}

func TestNow(t *testing.T) {
	c := NewSettableClock()
	c.SetNow(time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))

	now := c.Now()
	if now != time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC) {
		t.Errorf("got %s, want %s", now, time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))
	}
}

func TestSince(t *testing.T) {
	tests := []struct {
		now      time.Time
		old      time.Time
		expected time.Duration
	}{
		{
			now:      time.Date(2020, time.January, 1, 12, 0, 0, 0, time.UTC),
			old:      time.Date(2020, time.January, 1, 10, 0, 0, 0, time.UTC),
			expected: 2 * time.Hour,
		},
		{
			now:      time.Date(2020, time.January, 1, 12, 0, 0, 0, time.UTC),
			old:      time.Date(2020, time.January, 1, 12, 0, 3, 0, time.UTC),
			expected: -3 * time.Second,
		},
		{
			now:      time.Date(2020, time.January, 1, 12, 0, 0, 0, time.UTC),
			old:      time.Date(2020, time.January, 1, 12, 0, 0, 0, time.UTC),
			expected: 0,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprint(tt.old), func(t *testing.T) {
			c := NewSettableClock()
			c.SetNow(tt.now)

			d := c.Since(tt.old)
			if d != tt.expected {
				t.Errorf("got %s, want %s", d, tt.expected)
			}
		})
	}
}

func TestSleep(t *testing.T) {
	c := NewSettableClock()
	f := c.(*faketime)

	select {
	case <-f.timerAdded:
		t.Error("got value, want nothing")
	default:
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		c.Sleep(time.Second)
	}()

	<-f.timerAdded
	c.SetNow(time.Now().Add(time.Second))
	wg.Wait()
}

func TestTimer(t *testing.T) {
	c := NewSettableClock()
	f := c.(*faketime)

	select {
	case <-f.timerAdded:
		t.Error("got value, want nothing")
	default:
	}

	timer := c.Timer(time.Second)
	select {
	case <-timer.C():
		t.Error("got value, want nothing")
	default:
	}

	c.SetNow(time.Now().Add(time.Second))

	select {
	case <-timer.C():
	case <-time.After(50 * time.Millisecond):
		t.Error("timer took too long to trigger")
	}
}
