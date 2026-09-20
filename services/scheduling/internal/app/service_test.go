package app

import (
	"context"
	"github.com/aashishrajdev/halomail/services/scheduling/internal/domain"
	"testing"
	"time"
)

type testEvents struct {
	EventTypeRepo
	event *domain.EventType
}

func (repo testEvents) GetByID(context.Context, string) (*domain.EventType, error) {
	return repo.event, nil
}

type testAvailability struct {
	AvailabilityRepo
	availability *domain.Availability
}

func (repo testAvailability) Get(context.Context, string) (*domain.Availability, error) {
	return repo.availability, nil
}

type testBookings struct{ BookingRepo }

func (testBookings) ListConfirmedBetween(context.Context, string, time.Time, time.Time) ([]domain.Booking, error) {
	return nil, nil
}

type testCalendar struct {
	Calendar
	busy     []domain.Booking
	excluded string
}

func (calendar *testCalendar) Ready(context.Context, string) error { return nil }
func (calendar *testCalendar) Busy(_ context.Context, _ string, _ time.Time, _ time.Time, exclude string) ([]domain.Booking, error) {
	calendar.excluded = exclude
	return calendar.busy, nil
}

func TestCalendarConflictsBlockNewAndRescheduledBookings(test *testing.T) {
	start := time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC)
	event := &domain.EventType{ID: "event", OwnerID: "owner", Active: true, DurationMinutes: 30}
	calendar := &testCalendar{busy: []domain.Booking{{Start: start, End: start.Add(time.Hour), Status: domain.StatusConfirmed}}}
	service := New(Repos{EventTypes: testEvents{event: event}, Availability: testAvailability{availability: &domain.Availability{Timezone: "UTC", Rules: []domain.Rule{{Weekday: 1, StartMinute: 540, EndMinute: 600}}}}, Bookings: testBookings{}}, Config{Calendar: calendar})
	service.now = func() time.Time { return start.Add(-time.Hour) }
	slots, err := service.ListSlots(context.Background(), "event", "2026-10-05", "2026-10-05", "UTC")
	if err != nil || len(slots) != 0 {
		test.Fatalf("busy calendar returned slots: %v %v", slots, err)
	}
	for _, excluded := range []string{"", "existing-booking"} {
		free, err := service.slotFree(context.Background(), event, start, excluded)
		if err != nil || free {
			test.Fatalf("calendar conflict ignored during booking/reschedule: %v", err)
		}
		if calendar.excluded != excluded {
			test.Fatal("existing remote event exclusion not forwarded")
		}
	}
	calendar.busy = nil
	free, err := service.slotFree(context.Background(), event, start, "")
	if err != nil || !free {
		test.Fatalf("free slot unavailable: %v", err)
	}
}
