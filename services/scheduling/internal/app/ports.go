// Package app holds scheduling use cases over repository ports.
package app

import (
	"context"
	"time"

	"github.com/aashishrajdev/halomail/services/scheduling/internal/domain"
)

type EventTypeRepo interface {
	Create(ctx context.Context, et *domain.EventType) error
	GetByID(ctx context.Context, id string) (*domain.EventType, error)
	ListByOwner(ctx context.Context, ownerID string) ([]domain.EventType, error)
	Update(ctx context.Context, et *domain.EventType) error
	Delete(ctx context.Context, id, ownerID string) error
}

type AvailabilityRepo interface {
	// Get returns the owner's availability; a zero-value (empty rules) result is
	// valid and means "nothing configured yet".
	Get(ctx context.Context, ownerID string) (*domain.Availability, error)
	Set(ctx context.Context, a *domain.Availability) error
}

type BookingRepo interface {
	Create(ctx context.Context, b *domain.Booking) error
	GetByID(ctx context.Context, id string) (*domain.Booking, error)
	// GetByToken finds a booking by its reschedule or cancel token (kind is
	// "reschedule" or "cancel").
	GetByToken(ctx context.Context, kind, token string) (*domain.Booking, error)
	ListByOwner(ctx context.Context, ownerID, status string, limit, offset int) ([]domain.Booking, error)
	// ListConfirmedBetween returns confirmed bookings overlapping [from, to),
	// used to compute busy intervals.
	ListConfirmedBetween(ctx context.Context, ownerID string, from, to time.Time) ([]domain.Booking, error)
	Update(ctx context.Context, b *domain.Booking) error
	GetUsageStats(ctx context.Context, ownerID string) (*domain.UsageStats, error)
}

type CalendarRepo interface {
	ListByOwner(ctx context.Context, ownerID string) ([]domain.CalendarConnection, error)
	Delete(ctx context.Context, id, ownerID string) error
}

type Repos struct {
	EventTypes   EventTypeRepo
	Availability AvailabilityRepo
	Bookings     BookingRepo
	Calendars    CalendarRepo
}

type Calendar interface {
	Ready(context.Context, string) error
	Busy(context.Context, string, time.Time, time.Time, string) ([]domain.Booking, error)
	Start(context.Context, string) (string, error)
}
