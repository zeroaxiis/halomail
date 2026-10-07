// Package slots computes free meeting slots from availability rules, date
// overrides, and busy intervals. It is pure and timezone-aware: callers pass
// the owner's *time.Location; output slots are absolute UTC instants.
package slots

import "time"

// Rule is a weekly availability window. Weekday is 0=Sunday..6=Saturday and
// the minutes count from midnight in the owner zone.
type Rule struct {
	Weekday     int
	StartMinute int
	EndMinute   int
}

// Override replaces the weekly rules for one date: it either blocks the day
// or opens a single custom window.
type Override struct {
	Date        string // "2006-01-02"
	Unavailable bool
	StartMinute int
	EndMinute   int
}

// Interval is a busy [Start, End) range (UTC).
type Interval struct {
	Start time.Time
	End   time.Time
}

// Slot is a free [Start, End) range (UTC).
type Slot struct {
	Start time.Time
	End   time.Time
}

// Params is the input to Compute.
type Params struct {
	Location        *time.Location
	Rules           []Rule
	Overrides       []Override
	Busy            []Interval
	FromDate        string // inclusive "2006-01-02" (owner tz)
	ToDate          string // inclusive
	DurationMin     int
	BufferBeforeMin int
	BufferAfterMin  int
	StepMin         int // granularity; defaults to DurationMin
	Now             time.Time
	MaxSlots        int // 0 = unlimited
}

// Compute returns the free slots within the date range.
func Compute(p Params) []Slot {
	if p.Location == nil {
		p.Location = time.UTC
	}
	step := p.StepMin
	if step <= 0 {
		step = p.DurationMin
	}
	if step <= 0 || p.DurationMin <= 0 {
		return nil
	}

	from, err1 := time.ParseInLocation("2006-01-02", p.FromDate, p.Location)
	to, err2 := time.ParseInLocation("2006-01-02", p.ToDate, p.Location)
	if err1 != nil || err2 != nil || to.Before(from) {
		return nil
	}

	overrideByDate := make(map[string]Override, len(p.Overrides))
	for _, o := range p.Overrides {
		overrideByDate[o.Date] = o
	}

	var out []Slot
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		for _, w := range windowsFor(d, p.Rules, overrideByDate) {
			for m := w.start; m+p.DurationMin <= w.end; m += step {
				start := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, p.Location).
					Add(time.Duration(m) * time.Minute)
				end := start.Add(time.Duration(p.DurationMin) * time.Minute)
				if !start.After(p.Now) {
					continue
				}
				bs := start.Add(-time.Duration(p.BufferBeforeMin) * time.Minute)
				be := end.Add(time.Duration(p.BufferAfterMin) * time.Minute)
				if overlaps(bs, be, p.Busy) {
					continue
				}
				out = append(out, Slot{Start: start.UTC(), End: end.UTC()})
				if p.MaxSlots > 0 && len(out) >= p.MaxSlots {
					return out
				}
			}
		}
	}
	return out
}

type window struct{ start, end int }

func windowsFor(d time.Time, rules []Rule, ov map[string]Override) []window {
	if o, ok := ov[d.Format("2006-01-02")]; ok {
		if o.Unavailable || o.EndMinute <= o.StartMinute {
			return nil
		}
		return []window{{o.StartMinute, o.EndMinute}}
	}
	wd := int(d.Weekday())
	var ws []window
	for _, r := range rules {
		if r.Weekday == wd && r.EndMinute > r.StartMinute {
			ws = append(ws, window{r.StartMinute, r.EndMinute})
		}
	}
	return ws
}

func overlaps(s, e time.Time, busy []Interval) bool {
	for _, b := range busy {
		if s.Before(b.End) && b.Start.Before(e) {
			return true
		}
	}
	return false
}
