package rpc

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	schedulingv1 "github.com/zeroaxiis/halomail/services/shared/gen/halomail/scheduling/v1"

	"github.com/zeroaxiis/halomail/services/scheduling/internal/domain"
)

func ts(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

// ---- enums ---------------------------------------------------------------

func locKindToProto(s string) schedulingv1.LocationKind {
	switch s {
	case "google_meet":
		return schedulingv1.LocationKind_LOCATION_KIND_GOOGLE_MEET
	case "zoom":
		return schedulingv1.LocationKind_LOCATION_KIND_ZOOM
	case "phone":
		return schedulingv1.LocationKind_LOCATION_KIND_PHONE
	case "in_person":
		return schedulingv1.LocationKind_LOCATION_KIND_IN_PERSON
	case "custom":
		return schedulingv1.LocationKind_LOCATION_KIND_CUSTOM
	default:
		return schedulingv1.LocationKind_LOCATION_KIND_UNSPECIFIED
	}
}

func locKindFromProto(k schedulingv1.LocationKind) string {
	switch k {
	case schedulingv1.LocationKind_LOCATION_KIND_GOOGLE_MEET:
		return "google_meet"
	case schedulingv1.LocationKind_LOCATION_KIND_ZOOM:
		return "zoom"
	case schedulingv1.LocationKind_LOCATION_KIND_PHONE:
		return "phone"
	case schedulingv1.LocationKind_LOCATION_KIND_IN_PERSON:
		return "in_person"
	case schedulingv1.LocationKind_LOCATION_KIND_CUSTOM:
		return "custom"
	default:
		return ""
	}
}

func statusToProto(s string) schedulingv1.BookingStatus {
	switch s {
	case domain.StatusConfirmed:
		return schedulingv1.BookingStatus_BOOKING_STATUS_CONFIRMED
	case domain.StatusCancelled:
		return schedulingv1.BookingStatus_BOOKING_STATUS_CANCELLED
	case domain.StatusRescheduled:
		return schedulingv1.BookingStatus_BOOKING_STATUS_RESCHEDULED
	default:
		return schedulingv1.BookingStatus_BOOKING_STATUS_UNSPECIFIED
	}
}

func statusFromProto(s schedulingv1.BookingStatus) string {
	switch s {
	case schedulingv1.BookingStatus_BOOKING_STATUS_CONFIRMED:
		return domain.StatusConfirmed
	case schedulingv1.BookingStatus_BOOKING_STATUS_CANCELLED:
		return domain.StatusCancelled
	case schedulingv1.BookingStatus_BOOKING_STATUS_RESCHEDULED:
		return domain.StatusRescheduled
	default:
		return ""
	}
}

func providerFromProto(p schedulingv1.CalendarProvider) string {
	switch p {
	case schedulingv1.CalendarProvider_CALENDAR_PROVIDER_GOOGLE:
		return domain.ProviderGoogle
	case schedulingv1.CalendarProvider_CALENDAR_PROVIDER_OUTLOOK:
		return domain.ProviderOutlook
	default:
		return ""
	}
}

func providerToProto(s string) schedulingv1.CalendarProvider {
	switch s {
	case domain.ProviderGoogle:
		return schedulingv1.CalendarProvider_CALENDAR_PROVIDER_GOOGLE
	case domain.ProviderOutlook:
		return schedulingv1.CalendarProvider_CALENDAR_PROVIDER_OUTLOOK
	default:
		return schedulingv1.CalendarProvider_CALENDAR_PROVIDER_UNSPECIFIED
	}
}

// ---- entities ------------------------------------------------------------

func toProtoEventType(et *domain.EventType) *schedulingv1.EventType {
	if et == nil {
		return nil
	}
	return &schedulingv1.EventType{
		Id:                  et.ID,
		OwnerId:             et.OwnerID,
		Slug:                et.Slug,
		Title:               et.Title,
		Description:         et.Description,
		DurationMinutes:     int32(et.DurationMinutes),
		LocationKind:        locKindToProto(et.LocationKind),
		LocationDetail:      et.LocationDetail,
		BufferBeforeMinutes: int32(et.BufferBeforeMinutes),
		BufferAfterMinutes:  int32(et.BufferAfterMinutes),
		Color:               et.Color,
		Active:              et.Active,
		CreatedAt:           ts(et.CreatedAt),
	}
}

func toProtoBooking(b *domain.Booking) *schedulingv1.Booking {
	if b == nil {
		return nil
	}
	return &schedulingv1.Booking{
		Id:              b.ID,
		EventTypeId:     b.EventTypeID,
		OwnerId:         b.OwnerID,
		InviteeName:     b.InviteeName,
		InviteeEmail:    b.InviteeEmail,
		InviteeTimezone: b.InviteeTimezone,
		Start:           ts(b.Start),
		End:             ts(b.End),
		Status:          statusToProto(b.Status),
		Location:        b.Location,
		Notes:           b.Notes,
		RescheduleToken: b.RescheduleToken,
		CancelToken:     b.CancelToken,
		CreatedAt:       ts(b.CreatedAt),
	}
}

func toProtoSlot(s domain.Slot) *schedulingv1.Slot {
	return &schedulingv1.Slot{Start: ts(s.Start), End: ts(s.End)}
}

func toProtoAvailability(a *domain.Availability) *schedulingv1.Availability {
	if a == nil {
		return nil
	}
	rules := make([]*schedulingv1.AvailabilityRule, len(a.Rules))
	for i, r := range a.Rules {
		rules[i] = &schedulingv1.AvailabilityRule{
			Weekday:     int32(r.Weekday),
			StartMinute: int32(r.StartMinute),
			EndMinute:   int32(r.EndMinute),
		}
	}
	overrides := make([]*schedulingv1.DateOverride, len(a.Overrides))
	for i, o := range a.Overrides {
		overrides[i] = &schedulingv1.DateOverride{
			Date:        o.Date,
			Unavailable: o.Unavailable,
			StartMinute: int32(o.StartMinute),
			EndMinute:   int32(o.EndMinute),
		}
	}
	return &schedulingv1.Availability{
		OwnerId:   a.OwnerID,
		Timezone:  a.Timezone,
		Rules:     rules,
		Overrides: overrides,
	}
}

func toProtoConnection(c domain.CalendarConnection) *schedulingv1.CalendarConnection {
	return &schedulingv1.CalendarConnection{
		Id:        c.ID,
		OwnerId:   c.OwnerID,
		Provider:  providerToProto(c.Provider),
		Email:     c.Email,
		ConnectedAt: ts(c.CreatedAt),
	}
}

func rulesFromProto(rs []*schedulingv1.AvailabilityRule) []domain.Rule {
	out := make([]domain.Rule, len(rs))
	for i, r := range rs {
		out[i] = domain.Rule{
			Weekday:     int(r.GetWeekday()),
			StartMinute: int(r.GetStartMinute()),
			EndMinute:   int(r.GetEndMinute()),
		}
	}
	return out
}

func overridesFromProto(os []*schedulingv1.DateOverride) []domain.Override {
	out := make([]domain.Override, len(os))
	for i, o := range os {
		out[i] = domain.Override{
			Date:        o.GetDate(),
			Unavailable: o.GetUnavailable(),
			StartMinute: int(o.GetStartMinute()),
			EndMinute:   int(o.GetEndMinute()),
		}
	}
	return out
}
