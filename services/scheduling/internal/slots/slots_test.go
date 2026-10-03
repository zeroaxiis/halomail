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

func TestComputeBuffersWidenBusyIntervals(t *testing.T) {
	p := Params{
		Rules: mondayRule, FromDate: "2026-06-01", ToDate: "2026-06-01",
		DurationMin: 30, Now: monday,
		Busy: []Interval{{Start: monday.Add(10 * time.Hour), End: monday.Add(10*time.Hour + 30*time.Minute)}},
	}
	if got, want := starts(Compute(p)), []string{"09:00", "09:30", "10:30"}; !equal(got, want) {
		t.Fatalf("without buffers: starts = %v, want %v", got, want)
	}

	p.BufferBeforeMin, p.BufferAfterMin = 30, 30
	if got, want := starts(Compute(p)), []string{"09:00"}; !equal(got, want) {
		t.Fatalf("with buffers: starts = %v, want %v", got, want)
	}
}

func TestComputeMaxSlots(t *testing.T) {
	got := starts(Compute(Params{
		Rules: mondayRule, FromDate: "2026-06-01", ToDate: "2026-06-01",
		DurationMin: 30, Now: monday, MaxSlots: 2,
	}))
	if want := []string{"09:00", "09:30"}; !equal(got, want) {
		t.Fatalf("starts = %v, want %v", got, want)
	}
}

func TestComputeOverrideReplacesRules(t *testing.T) {
	base := Params{Rules: mondayRule, DurationMin: 60, Now: monday}

	// A custom window replaces the weekly rule for that date.
	p := base
	p.FromDate, p.ToDate = "2026-06-01", "2026-06-01"
	p.Overrides = []Override{{Date: "2026-06-01", StartMinute: 14 * 60, EndMinute: 15 * 60}}
	if got, want := starts(Compute(p)), []string{"14:00"}; !equal(got, want) {
		t.Fatalf("custom window: starts = %v, want %v", got, want)
	}

	// An override can open a day that has no weekly rule (Tuesday).
	p = base
	p.FromDate, p.ToDate = "2026-06-02", "2026-06-02"
	p.Overrides = []Override{{Date: "2026-06-02", StartMinute: 10 * 60, EndMinute: 11 * 60}}
	if got, want := starts(Compute(p)), []string{"10:00"}; !equal(got, want) {
		t.Fatalf("opened day: starts = %v, want %v", got, want)
	}

	// An empty or inverted window blocks the day.
	p = base
	p.FromDate, p.ToDate = "2026-06-01", "2026-06-01"
	p.Overrides = []Override{{Date: "2026-06-01", StartMinute: 15 * 60, EndMinute: 14 * 60}}
	if got := Compute(p); len(got) != 0 {
		t.Fatalf("inverted window produced slots: %+v", got)
	}
}

func TestComputeReturnsUTCInstants(t *testing.T) {
	ist := time.FixedZone("IST", 5*3600+1800)
	got := Compute(Params{
		Location: ist,
		Rules:    []Rule{{Weekday: 1, StartMinute: 9 * 60, EndMinute: 10 * 60}},
		FromDate: "2026-06-01", ToDate: "2026-06-01",
		DurationMin: 60, Now: monday.AddDate(0, 0, -1),
	})
	if len(got) != 1 {
		t.Fatalf("got %d slots, want 1", len(got))
	}
	// 09:00 IST is 03:30 UTC.
	if want := time.Date(2026, 6, 1, 3, 30, 0, 0, time.UTC); !got[0].Start.Equal(want) || got[0].Start.Location() != time.UTC {
		t.Fatalf("start = %v, want %v in UTC", got[0].Start, want)
	}

	// A nil location is treated as UTC.
	got = Compute(Params{Rules: mondayRule, FromDate: "2026-06-01", ToDate: "2026-06-01", DurationMin: 120, Now: monday})
	if len(got) != 1 || got[0].Start.Hour() != 9 {
		t.Fatalf("nil location: %+v", got)
	}
}
