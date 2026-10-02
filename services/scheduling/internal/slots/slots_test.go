package slots

import (
	"testing"
	"time"
)

func TestCompute(t *testing.T) {
	loc := time.UTC
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, loc) // a Monday, midnight
	// Monday 09:00–11:00 window, 60-min meetings.
	rules := []Rule{{Weekday: 1, StartMinute: 9 * 60, EndMinute: 11 * 60}}

	tests := []struct {
		name string
		p    Params
		want int
	}{
		{
			name: "two back-to-back slots",
			p: Params{
				Location: loc, Rules: rules,
				FromDate: "2026-06-01", ToDate: "2026-06-01",
				DurationMin: 60, Now: now,
			},
			want: 2, // 09:00 and 10:00
		},
		{
			name: "busy 09:00-10:00 removes first slot",
			p: Params{
				Location: loc, Rules: rules,
				FromDate: "2026-06-01", ToDate: "2026-06-01",
				DurationMin: 60, Now: now,
				Busy: []Interval{{
					Start: time.Date(2026, 6, 1, 9, 0, 0, 0, loc),
					End:   time.Date(2026, 6, 1, 10, 0, 0, 0, loc),
				}},
			},
			want: 1, // only 10:00
		},
		{
			name: "unavailable override yields nothing",
			p: Params{
				Location: loc, Rules: rules,
				Overrides: []Override{{Date: "2026-06-01", Unavailable: true}},
				FromDate:  "2026-06-01", ToDate: "2026-06-01",
				DurationMin: 60, Now: now,
			},
			want: 0,
		},
		{
			name: "past slots excluded",
			p: Params{
				Location: loc, Rules: rules,
				FromDate: "2026-06-01", ToDate: "2026-06-01",
				DurationMin: 60,
				Now:         time.Date(2026, 6, 1, 9, 30, 0, 0, loc), // after 09:00
			},
			want: 1, // 10:00 only
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Compute(tc.p)
			if len(got) != tc.want {
				t.Fatalf("got %d slots, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}

// monday is 2026-06-01; mondayRule opens 09:00-11:00 on Mondays.
var (
	monday     = time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	mondayRule = []Rule{{Weekday: 1, StartMinute: 9 * 60, EndMinute: 11 * 60}}
)

func starts(slots []Slot) []string {
	out := make([]string, 0, len(slots))
	for _, s := range slots {
		out = append(out, s.Start.Format("15:04"))
	}
	return out
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestComputeStepGranularity(t *testing.T) {
	got := starts(Compute(Params{
		Rules: mondayRule, FromDate: "2026-06-01", ToDate: "2026-06-01",
		DurationMin: 60, StepMin: 30, Now: monday,
	}))
	if want := []string{"09:00", "09:30", "10:00"}; !equal(got, want) {
		t.Fatalf("starts = %v, want %v", got, want)
	}
}
