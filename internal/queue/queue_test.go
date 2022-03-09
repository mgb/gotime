package queue

import (
	"math/rand"
	"testing"
	"time"
)

func TestTimeQueue_Add(t *testing.T) {
	tests := []struct {
		name        string
		count       int
		expectedLen int
	}{
		{
			name: "empty",
		},
		{
			name:        "one",
			count:       1,
			expectedLen: 1,
		},
		{
			name:        "lots",
			count:       100,
			expectedLen: 100,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := NewTimeQueue()
			for i := 0; i < tt.count; i++ {
				q.Add(time.Now(), nil)
			}

			l := q.Len()
			if l != tt.expectedLen {
				t.Errorf("got %d, want %d", l, tt.expectedLen)
			}
		})
	}
}

func TestTimeQueue_PopBeforeOrEqual(t *testing.T) {
	type test struct {
		name              string
		times             []time.Time
		popTime           time.Time
		expectedPopCount  int
		expectedRemaining int
	}
	tests := []test{
		{
			name: "empty",
		},
		{
			name: "one timer newer than pop time",
			times: []time.Time{
				time.Date(2020, time.February, 1, 0, 0, 0, 0, time.UTC),
			},
			popTime:           time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
			expectedPopCount:  0,
			expectedRemaining: 1,
		},
		{
			name: "one timer before pop time, returned",
			times: []time.Time{
				time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
			},
			popTime:           time.Date(2020, time.February, 1, 0, 0, 0, 0, time.UTC),
			expectedPopCount:  1,
			expectedRemaining: 0,
		},
		{
			name: "one timer equal to pop time, returned",
			times: []time.Time{
				time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
			},
			popTime:           time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
			expectedPopCount:  1,
			expectedRemaining: 0,
		},
		{
			name: "one timer one nanosecond after pop time, not returned",
			times: []time.Time{
				time.Date(2020, time.January, 1, 0, 0, 0, 1, time.UTC),
			},
			popTime:           time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
			expectedPopCount:  0,
			expectedRemaining: 1,
		},
		{
			name: "bunch of timers",
			times: []time.Time{
				time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
				time.Date(2020, time.January, 2, 0, 0, 0, 0, time.UTC),
				time.Date(2020, time.January, 3, 0, 0, 0, 0, time.UTC),
				time.Date(2020, time.January, 4, 0, 0, 0, 0, time.UTC),
				time.Date(2020, time.January, 5, 0, 0, 0, 0, time.UTC),
				time.Date(2020, time.January, 6, 0, 0, 0, 0, time.UTC),
			},
			popTime:           time.Date(2020, time.January, 3, 12, 0, 0, 0, time.UTC),
			expectedPopCount:  3,
			expectedRemaining: 3,
		},
	}
	if !testing.Short() {
		n := 100000
		lottaT := test{
			name:              "a lot of timers",
			times:             make([]time.Time, 0, n),
			popTime:           time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC),
			expectedPopCount:  n - 1,
			expectedRemaining: 1,
		}
		t := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
		for i := 0; i < n-1; i++ {
			lottaT.times = append(lottaT.times, t)
			t = t.Add(time.Second)
		}
		lottaT.times = append(lottaT.times, time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC))

		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		r.Shuffle(len(lottaT.times), func(i, j int) {
			lottaT.times[i], lottaT.times[j] = lottaT.times[j], lottaT.times[i]
		})

		tests = append(tests, lottaT)
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			q := NewTimeQueue()
			for _, i := range tt.times {
				q.Add(i, nil)
			}

			chs := q.PopBeforeOrEqual(tt.popTime)
			if len(chs) != tt.expectedPopCount {
				t.Errorf("pop count: got %d, want %d", len(chs), tt.expectedPopCount)
			}
			l := q.Len()
			if l != tt.expectedRemaining {
				t.Errorf("remaining count: got %d, want %d", l, tt.expectedRemaining)
			}
		})
	}
}

func TestTimeQueue_Peek(t *testing.T) {
	tests := []struct {
		name   string
		times  []time.Time
		want   time.Time
		wantOk bool
	}{
		{
			name: "empty",
		},
		{
			name:   "one",
			times:  []time.Time{time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)},
			want:   time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
			wantOk: true,
		},
		{
			name: "oldest inserted first",
			times: []time.Time{
				time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
				time.Date(2020, time.February, 1, 0, 0, 0, 0, time.UTC),
			},
			want:   time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
			wantOk: true,
		},
		{
			name: "oldest inserted last",
			times: []time.Time{
				time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
				time.Date(2020, time.February, 1, 0, 0, 0, 0, time.UTC),
			},
			want:   time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
			wantOk: true,
		},
		{
			name: "lots of times in random order",
			times: func() []time.Time {
				var times []time.Time
				for t := time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC); t.Before(time.Date(2021, time.February, 1, 0, 0, 0, 0, time.UTC)); t = t.AddDate(0, 0, 1) {
					times = append(times, t)
				}
				r := rand.New(rand.NewSource(time.Now().UnixNano()))
				r.Shuffle(len(times), func(i, j int) {
					times[i], times[j] = times[j], times[i]
				})
				return times
			}(),
			want:   time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC),
			wantOk: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := NewTimeQueue()
			for _, t := range tt.times {
				q.Add(t, nil)
			}

			got, gotOk := q.Peek()
			if gotOk != tt.wantOk {
				t.Errorf("got %v, want %v", gotOk, tt.wantOk)
			} else if gotOk && got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
