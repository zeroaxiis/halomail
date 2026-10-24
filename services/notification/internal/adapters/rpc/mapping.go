package rpc

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	notificationv1 "github.com/zeroaxiis/halomail/services/shared/gen/halomail/notification/v1"

	"github.com/zeroaxiis/halomail/services/notification/internal/domain"
)

func ts(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

func eventToProto(s string) notificationv1.WebhookEvent {
	switch s {
	case domain.EventBookingCreated:
		return notificationv1.WebhookEvent_WEBHOOK_EVENT_BOOKING_CREATED
	case domain.EventBookingCancelled:
		return notificationv1.WebhookEvent_WEBHOOK_EVENT_BOOKING_CANCELLED
	case domain.EventBookingRescheduled:
		return notificationv1.WebhookEvent_WEBHOOK_EVENT_BOOKING_RESCHEDULED
	case domain.EventMessageReceived:
		return notificationv1.WebhookEvent_WEBHOOK_EVENT_MESSAGE_RECEIVED
	default:
		return notificationv1.WebhookEvent_WEBHOOK_EVENT_UNSPECIFIED
	}
}

func eventFromProto(e notificationv1.WebhookEvent) string {
	switch e {
	case notificationv1.WebhookEvent_WEBHOOK_EVENT_BOOKING_CREATED:
		return domain.EventBookingCreated
	case notificationv1.WebhookEvent_WEBHOOK_EVENT_BOOKING_CANCELLED:
		return domain.EventBookingCancelled
	case notificationv1.WebhookEvent_WEBHOOK_EVENT_BOOKING_RESCHEDULED:
		return domain.EventBookingRescheduled
	case notificationv1.WebhookEvent_WEBHOOK_EVENT_MESSAGE_RECEIVED:
		return domain.EventMessageReceived
	default:
		return ""
	}
}

func eventsFromProto(in []notificationv1.WebhookEvent) []string {
	out := make([]string, 0, len(in))
	for _, e := range in {
		if s := eventFromProto(e); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func eventsToProto(in []string) []notificationv1.WebhookEvent {
	out := make([]notificationv1.WebhookEvent, 0, len(in))
	for _, s := range in {
		out = append(out, eventToProto(s))
	}
	return out
}

func toProtoWebhook(w *domain.Webhook) *notificationv1.Webhook {
	if w == nil {
		return nil
	}
	return &notificationv1.Webhook{
		Id:             w.ID,
		OwnerId:        w.OwnerID,
		Url:            w.URL,
		Events:         eventsToProto(w.Events),
		SecretLastFour: w.SecretLastFour(),
		Active:         w.Active,
		CreatedAt:      ts(w.CreatedAt),
	}
}
