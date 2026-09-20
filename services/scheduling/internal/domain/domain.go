// Package domain holds scheduling entities and invariants. Pure: no I/O.
package domain

import "time"

// Booking statuses.
const (
	StatusConfirmed   = "confirmed"
	StatusCancelled   = "cancelled"
	StatusRescheduled = "rescheduled"
)

// Calendar providers.
const (
	ProviderGoogle  = "google"
	ProviderOutlook = "outlook"
)

// EventType is a bookable meeting template owned by a user.
type EventType struct {
	ID                  string
	OwnerID             string
	Slug                string
	Title               string
	Description         string
	DurationMinutes     int
	LocationKind        string
	LocationDetail      string
	BufferBeforeMinutes int
	BufferAfterMinutes  int
	Color               string
	Active              bool
	CreatedAt           time.Time
}

// Rule is a weekly recurring availability window, in the owner's timezone.
type Rule struct {
	Weekday     int // 0=Sunday .. 6=Saturday
	StartMinute int // minutes from midnight
	EndMinute   int
}

// Override opens or blocks a specific date.
type Override struct {
	Date        string // ISO "2006-01-02"
	Unavailable bool
	StartMinute int
	EndMinute   int
}

// Availability is an owner's complete schedule definition.
type Availability struct {
	OwnerID   string
	Timezone  string
	Rules     []Rule
	Overrides []Override
}

// Booking is a confirmed (or cancelled) meeting.
type Booking struct {
	ID              string
	EventTypeID     string
	OwnerID         string
	InviteeName     string
	InviteeEmail    string
	InviteeTimezone string
	Start           time.Time
	End             time.Time
	Status          string
	Location        string
	Notes           string
	RescheduleToken string
	CancelToken     string
	CreatedAt       time.Time
}

// CalendarConnection is a linked external calendar (tokens omitted here).
type CalendarConnection struct {
	ID        string
	OwnerID   string
	Provider  string
	Email     string
	CreatedAt time.Time
}

// Slot is a free [Start, End) instant range (UTC).
type Slot struct {
	Start time.Time
	End   time.Time
}

// UsageStats holds aggregate metrics for the dashboard.
type UsageStats struct {
	TotalBookings     int
	UpcomingBookings  int
	CancelledBookings int
}
