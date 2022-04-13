package gotime

import (
	"sync"
	"testing"
	"time"
)

func TestSimulatedTime_After_real(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	t.Parallel()

	tests := []struct {
		name              string
		warpSpeed         float64
		sleepTime         time.Duration
		expectedSleepTime time.Duration
		epsilon           time.Duration
	}{
		{
			name:              "real time",
			sleepTime:         time.Second,
			expectedSleepTime: time.Second,
			epsilon:           100 * time.Millisecond,
		},
		{
			name:              "60x speed of real time",
			warpSpeed:         60.0,
			sleepTime:         time.Minute,
			expectedSleepTime: time.Second,
			epsilon:           100 * time.Millisecond,
		},
		{
			name:              "120x speed of real time",
			warpSpeed:         120.0,
			sleepTime:         time.Minute,
			expectedSleepTime: time.Second / 2,
			epsilon:           100 * time.Millisecond,
		},
		{
			name:              "3600x speed of real time",
			warpSpeed:         3600.0,
			sleepTime:         time.Hour,
			expectedSleepTime: time.Second,
			epsilon:           100 * time.Millisecond,
		},
		{
			name:              "1/1000 speed of real time",
			warpSpeed:         1 / 1000.0,
			sleepTime:         time.Millisecond,
			expectedSleepTime: time.Second,
			epsilon:           100 * time.Millisecond,
		},
		{
			name:              "1/500 speed of real time",
			warpSpeed:         1 / 500.0,
			sleepTime:         time.Millisecond,
			expectedSleepTime: time.Second / 2,
			epsilon:           100 * time.Millisecond,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sim := NewTimeWarpableClock()

			if tt.warpSpeed != 0 {
				if err := sim.SetWarpSpeed(tt.warpSpeed); err != nil {
					t.Fatal(err)
				}
			}

			ch := sim.After(tt.sleepTime)
			select {
			case <-ch:
				t.Error("got value, want nothing")
			default:
			}

			start := time.Now()
			<-ch
			d := time.Since(start)

			select {
			case <-ch:
				t.Error("got value, want nothing")
			default:
			}

			diff := time.Duration(int64(d) - int64(tt.expectedSleepTime))
			if diff < 0 {
				diff = -diff
			}

			if diff > tt.epsilon {
				t.Errorf("got %s, want %s", d, tt.expectedSleepTime)
			}
		})
	}
}

func TestSimulatedTime_After_Warped_real(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	t.Parallel()

	type warptime struct {
		realSleepDur time.Duration
		warpSpeed    float64
	}
	tests := []struct {
		name              string
		initialWarpSpeed  float64
		warptimes         []warptime
		sleepTime         time.Duration
		expectedSleepTime time.Duration
		epsilon           time.Duration
	}{
		{
			name:             "real time, then half speed",
			initialWarpSpeed: 1,
			warptimes: []warptime{
				{
					realSleepDur: time.Second / 2,
					warpSpeed:    2,
				},
			},
			sleepTime:         time.Second,
			expectedSleepTime: time.Second * 3 / 4,
			epsilon:           100 * time.Millisecond,
		},
		{
			name:             "2x speed, then real time",
			initialWarpSpeed: 2,
			warptimes: []warptime{
				{
					realSleepDur: 400 * time.Millisecond,
					warpSpeed:    1,
				},
			},
			sleepTime:         time.Second,
			expectedSleepTime: 600 * time.Millisecond,
			epsilon:           100 * time.Millisecond,
		},
		{
			name:             "1/100th speed, then 2x speed",
			initialWarpSpeed: 1 / 100.0,
			warptimes: []warptime{
				{
					realSleepDur: 50 * time.Millisecond,
					warpSpeed:    2,
				},
			},
			sleepTime:         time.Millisecond,
			expectedSleepTime: 52 * time.Millisecond,
			epsilon:           10 * time.Millisecond,
		},
		{
			name:             "60x speed, then 1x speed, then 60x speed",
			initialWarpSpeed: 60,
			warptimes: []warptime{
				{
					realSleepDur: 500 * time.Millisecond,
					warpSpeed:    1,
				},
				{
					realSleepDur: 10 * time.Millisecond,
					warpSpeed:    60,
				},
			},
			sleepTime:         time.Minute,
			expectedSleepTime: 1010 * time.Millisecond,
			epsilon:           100 * time.Millisecond,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sim := NewTimeWarpableClock()

			if tt.initialWarpSpeed != 0 {
				if err := sim.SetWarpSpeed(tt.initialWarpSpeed); err != nil {
					t.Fatal(err)
				}
			}

			afterCh := sim.After(tt.sleepTime)

			select {
			case <-afterCh:
				t.Error("got value, want nothing")
			default:
			}

			ch := make(chan time.Duration)
			var wg sync.WaitGroup
			wg.Add(2)

			go func() {
				defer wg.Done()
				for _, wt := range tt.warptimes {
					time.Sleep(wt.realSleepDur)
					sim.SetWarpSpeed(wt.warpSpeed)
				}
			}()

			go func() {
				defer wg.Done()

				start := time.Now()
				<-afterCh
				ch <- time.Since(start)
			}()

			d := <-ch
			wg.Wait()

			select {
			case <-afterCh:
				t.Error("got value, want nothing")
			default:
			}

			diff := time.Duration(int64(d) - int64(tt.expectedSleepTime))
			if diff < 0 {
				diff = -diff
			}

			if diff > tt.epsilon {
				t.Errorf("got %s, want %s", d, tt.expectedSleepTime)
			}
		})
	}
}

func TestSimulatedTime_Now_real(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	t.Parallel()

	tests := []struct {
		name             string
		warpSpeed        float64
		sleepTime        time.Duration
		expectedDuration time.Duration
		epsilon          time.Duration
	}{
		{
			name:             "real time",
			sleepTime:        time.Second,
			expectedDuration: time.Second,
			epsilon:          100 * time.Millisecond,
		},
		{
			name:             "60x speed of real time",
			warpSpeed:        60.0,
			sleepTime:        time.Second,
			expectedDuration: time.Minute,
			epsilon:          time.Second,
		},
		{
			name:             "120x speed of real time",
			warpSpeed:        120.0,
			sleepTime:        time.Second / 2,
			expectedDuration: time.Minute,
			epsilon:          time.Second,
		},
		{
			name:             "3600x speed of real time",
			warpSpeed:        3600.0,
			sleepTime:        time.Second,
			expectedDuration: time.Hour,
			epsilon:          time.Minute,
		},
		{
			name:             "1/1000 speed of real time",
			warpSpeed:        1 / 1000.0,
			sleepTime:        time.Second,
			expectedDuration: time.Millisecond,
			epsilon:          100 * time.Millisecond,
		},
		{
			name:             "1/500 speed of real time",
			warpSpeed:        1 / 500.0,
			sleepTime:        time.Second / 2,
			expectedDuration: time.Millisecond,
			epsilon:          100 * time.Millisecond,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sim := NewTimeWarpableClock()

			if tt.warpSpeed != 0 {
				if err := sim.SetWarpSpeed(tt.warpSpeed); err != nil {
					t.Fatal(err)
				}
			}

			start := sim.Now()
			time.Sleep(tt.sleepTime)
			d := sim.Now().Sub(start)

			diff := time.Duration(int64(d) - int64(tt.expectedDuration))
			if diff < 0 {
				diff = -diff
			}

			if diff > tt.epsilon {
				t.Errorf("got %s, want %s", d, tt.expectedDuration)
			}
		})
	}
}

func TestSimulatedTime_Now_Warped_real(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	t.Parallel()

	type warptime struct {
		realSleepDur time.Duration
		warpSpeed    float64
	}
	tests := []struct {
		name             string
		initialWarpSpeed float64
		warptimes        []warptime
		sleepTime        time.Duration
		expectedDuration time.Duration
		epsilon          time.Duration
	}{
		{
			name:             "real time, then half speed",
			initialWarpSpeed: 1,
			warptimes: []warptime{
				{
					realSleepDur: time.Second / 2,
					warpSpeed:    2,
				},
			},
			expectedDuration: time.Second,
			sleepTime:        time.Second * 3 / 4,
			epsilon:          100 * time.Millisecond,
		},
		{
			name:             "2x speed, then real time",
			initialWarpSpeed: 2,
			warptimes: []warptime{
				{
					realSleepDur: 400 * time.Millisecond,
					warpSpeed:    1,
				},
			},
			expectedDuration: time.Second,
			sleepTime:        600 * time.Millisecond,
			epsilon:          100 * time.Millisecond,
		},
		{
			name:             "1/100th speed, then 2x speed",
			initialWarpSpeed: 1 / 100.0,
			warptimes: []warptime{
				{
					realSleepDur: 50 * time.Millisecond,
					warpSpeed:    2,
				},
			},
			expectedDuration: time.Millisecond,
			sleepTime:        52 * time.Millisecond,
			epsilon:          10 * time.Millisecond,
		},
		{
			name:             "60x speed, then 1x speed, then 60x speed",
			initialWarpSpeed: 60,
			warptimes: []warptime{
				{
					realSleepDur: 500 * time.Millisecond,
					warpSpeed:    1,
				},
				{
					realSleepDur: 10 * time.Millisecond,
					warpSpeed:    60,
				},
			},
			expectedDuration: time.Minute,
			sleepTime:        1010 * time.Millisecond,
			epsilon:          100 * time.Millisecond,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sim := NewTimeWarpableClock()

			if tt.initialWarpSpeed != 0 {
				if err := sim.SetWarpSpeed(tt.initialWarpSpeed); err != nil {
					t.Fatal(err)
				}
			}

			var wg sync.WaitGroup
			wg.Add(2)

			go func() {
				defer wg.Done()

				for _, wt := range tt.warptimes {
					time.Sleep(wt.realSleepDur)
					sim.SetWarpSpeed(wt.warpSpeed)
				}
			}()

			var d time.Duration
			go func() {
				defer wg.Done()

				start := sim.Now()
				time.Sleep(tt.sleepTime)
				d = sim.Now().Sub(start)
			}()

			wg.Wait()

			diff := time.Duration(int64(d) - int64(tt.expectedDuration))
			if diff < 0 {
				diff = -diff
			}

			if diff > tt.epsilon {
				t.Errorf("got %s, want %s", d, tt.expectedDuration)
			}
		})
	}
}

func TestSimulatedTime_Since_real(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	t.Parallel()

	tests := []struct {
		name              string
		warpSpeed         float64
		sleepTime         time.Duration
		expectedSleepTime time.Duration
		epsilon           time.Duration
	}{
		{
			name:              "real time",
			sleepTime:         time.Second,
			expectedSleepTime: time.Second,
			epsilon:           100 * time.Millisecond,
		},
		{
			name:              "60x speed of real time",
			warpSpeed:         60.0,
			sleepTime:         time.Minute,
			expectedSleepTime: time.Second,
			epsilon:           100 * time.Millisecond,
		},
		{
			name:              "120x speed of real time",
			warpSpeed:         120.0,
			sleepTime:         time.Minute,
			expectedSleepTime: time.Second / 2,
			epsilon:           time.Second,
		},
		{
			name:              "3600x speed of real time",
			warpSpeed:         3600.0,
			sleepTime:         time.Hour,
			expectedSleepTime: time.Second,
			epsilon:           time.Minute,
		},
		{
			name:              "1/1000 speed of real time",
			warpSpeed:         1 / 1000.0,
			sleepTime:         time.Millisecond,
			expectedSleepTime: time.Second,
			epsilon:           100 * time.Millisecond,
		},
		{
			name:              "1/500 speed of real time",
			warpSpeed:         1 / 500.0,
			sleepTime:         time.Millisecond,
			expectedSleepTime: time.Second / 2,
			epsilon:           100 * time.Millisecond,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sim := NewTimeWarpableClock()

			if tt.warpSpeed != 0 {
				if err := sim.SetWarpSpeed(tt.warpSpeed); err != nil {
					t.Fatal(err)
				}
			}

			realStart := time.Now()
			start := sim.Now()
			sim.Sleep(tt.sleepTime)
			d := sim.Since(start)
			realD := time.Since(realStart)

			diff := d - tt.sleepTime
			if diff < 0 {
				diff = -diff
			}

			if diff > tt.epsilon {
				t.Errorf("got %s, want %s", d, tt.sleepTime)
			}

			diff = time.Duration(int64(realD) - int64(tt.expectedSleepTime))
			if diff < 0 {
				diff = -diff
			}

			if diff > 100*time.Millisecond {
				t.Errorf("got %s, want %s", d, tt.expectedSleepTime)
			}
		})
	}
}

func TestSimulatedTime_Sleep_real(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	t.Parallel()

	tests := []struct {
		name              string
		warpSpeed         float64
		sleepTime         time.Duration
		expectedSleepTime time.Duration
		epsilon           time.Duration
	}{
		{
			name:              "real time",
			sleepTime:         time.Second,
			expectedSleepTime: time.Second,
			epsilon:           100 * time.Millisecond,
		},
		{
			name:              "60x speed of real time",
			warpSpeed:         60.0,
			sleepTime:         time.Minute,
			expectedSleepTime: time.Second,
			epsilon:           100 * time.Millisecond,
		},
		{
			name:              "120x speed of real time",
			warpSpeed:         120.0,
			sleepTime:         time.Minute,
			expectedSleepTime: time.Second / 2,
			epsilon:           100 * time.Millisecond,
		},
		{
			name:              "3600x speed of real time",
			warpSpeed:         3600.0,
			sleepTime:         time.Hour,
			expectedSleepTime: time.Second,
			epsilon:           100 * time.Millisecond,
		},
		{
			name:              "1/1000 speed of real time",
			warpSpeed:         1 / 1000.0,
			sleepTime:         time.Millisecond,
			expectedSleepTime: time.Second,
			epsilon:           100 * time.Millisecond,
		},
		{
			name:              "1/500 speed of real time",
			warpSpeed:         1 / 500.0,
			sleepTime:         time.Millisecond,
			expectedSleepTime: time.Second / 2,
			epsilon:           100 * time.Millisecond,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sim := NewTimeWarpableClock()

			if tt.warpSpeed != 0 {
				if err := sim.SetWarpSpeed(tt.warpSpeed); err != nil {
					t.Fatal(err)
				}
			}

			start := time.Now()
			sim.Sleep(tt.sleepTime)
			d := time.Since(start)

			diff := time.Duration(int64(d) - int64(tt.expectedSleepTime))
			if diff < 0 {
				diff = -diff
			}

			if diff > tt.epsilon {
				t.Errorf("got %s, want %s", d, tt.expectedSleepTime)
			}
		})
	}
}

func TestSimulatedTime_Sleep_Warped_real(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	t.Parallel()

	type warptime struct {
		realSleepDur time.Duration
		warpSpeed    float64
	}
	tests := []struct {
		name              string
		initialWarpSpeed  float64
		warptimes         []warptime
		sleepTime         time.Duration
		expectedSleepTime time.Duration
		epsilon           time.Duration
	}{
		{
			name:             "real time, then half speed",
			initialWarpSpeed: 1,
			warptimes: []warptime{
				{
					realSleepDur: time.Second / 2,
					warpSpeed:    2,
				},
			},
			sleepTime:         time.Second,
			expectedSleepTime: time.Second * 3 / 4,
			epsilon:           100 * time.Millisecond,
		},
		{
			name:             "2x speed, then real time",
			initialWarpSpeed: 2,
			warptimes: []warptime{
				{
					realSleepDur: 400 * time.Millisecond,
					warpSpeed:    1,
				},
			},
			sleepTime:         time.Second,
			expectedSleepTime: 600 * time.Millisecond,
			epsilon:           100 * time.Millisecond,
		},
		{
			name:             "1/100th speed, then 2x speed",
			initialWarpSpeed: 1 / 100.0,
			warptimes: []warptime{
				{
					realSleepDur: 50 * time.Millisecond,
					warpSpeed:    2,
				},
			},
			sleepTime:         time.Millisecond,
			expectedSleepTime: 52 * time.Millisecond,
			epsilon:           10 * time.Millisecond,
		},
		{
			name:             "60x speed, then 1x speed, then 60x speed",
			initialWarpSpeed: 60,
			warptimes: []warptime{
				{
					realSleepDur: 500 * time.Millisecond,
					warpSpeed:    1,
				},
				{
					realSleepDur: 10 * time.Millisecond,
					warpSpeed:    60,
				},
			},
			sleepTime:         time.Minute,
			expectedSleepTime: 1010 * time.Millisecond,
			epsilon:           100 * time.Millisecond,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sim := NewTimeWarpableClock()

			if tt.initialWarpSpeed != 0 {
				if err := sim.SetWarpSpeed(tt.initialWarpSpeed); err != nil {
					t.Fatal(err)
				}
			}

			ch := make(chan time.Duration)
			var wg sync.WaitGroup
			wg.Add(2)

			go func() {
				defer wg.Done()
				for _, wt := range tt.warptimes {
					time.Sleep(wt.realSleepDur)
					sim.SetWarpSpeed(wt.warpSpeed)
				}
			}()

			go func() {
				defer wg.Done()

				start := time.Now()
				sim.Sleep(tt.sleepTime)
				ch <- time.Since(start)
			}()

			d := <-ch
			wg.Wait()

			diff := time.Duration(int64(d) - int64(tt.expectedSleepTime))
			if diff < 0 {
				diff = -diff
			}

			if diff > tt.epsilon {
				t.Errorf("got %s, want %s", d, tt.expectedSleepTime)
			}
		})
	}
}
