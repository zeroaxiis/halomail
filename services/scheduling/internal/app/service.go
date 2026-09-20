package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/mail"
	"strings"
	"time"

	"github.com/aashishrajdev/halomail/services/scheduling/internal/domain"
	"github.com/aashishrajdev/halomail/services/scheduling/internal/slots"
	"github.com/aashishrajdev/halomail/services/shared/config"
	"github.com/aashishrajdev/halomail/services/shared/errs"
	"github.com/aashishrajdev/halomail/services/shared/idgen"
)

// Config carries calendar OAuth settings.
type Config struct {
	Google    config.OAuth
	Microsoft config.Microsoft
	Calendar  Calendar
}

type Service struct {
	eventTypes   EventTypeRepo
	availability AvailabilityRepo
	bookings     BookingRepo
	calendars    CalendarRepo
	cfg          Config
	now          func() time.Time
}

func New(r Repos, cfg Config) *Service {
	return &Service{
		eventTypes:   r.EventTypes,
		availability: r.Availability,
		bookings:     r.Bookings,
		calendars:    r.Calendars,
		cfg:          cfg,
		now:          time.Now,
	}
}

// ---- Event types ---------------------------------------------------------

type EventTypeInput struct {
	Title               string
	Slug                string
	Description         string
	DurationMinutes     int
	LocationKind        string
	LocationDetail      string
	BufferBeforeMinutes int
	BufferAfterMinutes  int
	Color               string
}

func (s *Service) CreateEventType(ctx context.Context, ownerID string, in EventTypeInput) (*domain.EventType, error) {
	if strings.TrimSpace(in.Title) == "" {
		return nil, errs.Invalid("title is required")
	}
	if in.DurationMinutes <= 0 {
		in.DurationMinutes = 30
	}
	if in.DurationMinutes > 480 || in.BufferBeforeMinutes < 0 || in.BufferAfterMinutes < 0 || in.BufferBeforeMinutes > 240 || in.BufferAfterMinutes > 240 {
		return nil, errs.Invalid("invalid duration or buffers")
	}
	slug := in.Slug
	if slug == "" {
		slug = slugify(in.Title)
	}
	lk := in.LocationKind
	if lk == "" {
		lk = "custom"
	}
	et := &domain.EventType{
		ID:                  idgen.Prefixed("evt_"),
		OwnerID:             ownerID,
		Slug:                slug,
		Title:               in.Title,
		Description:         in.Description,
		DurationMinutes:     in.DurationMinutes,
		LocationKind:        lk,
		LocationDetail:      in.LocationDetail,
		BufferBeforeMinutes: in.BufferBeforeMinutes,
		BufferAfterMinutes:  in.BufferAfterMinutes,
		Color:               in.Color,
		Active:              true,
		CreatedAt:           s.now(),
	}
	if err := s.eventTypes.Create(ctx, et); err != nil {
		return nil, err
	}
	return et, nil
}

func (s *Service) GetEventType(ctx context.Context, id string) (*domain.EventType, error) {
	return s.eventTypes.GetByID(ctx, id)
}

func (s *Service) ListEventTypes(ctx context.Context, ownerID string) ([]domain.EventType, error) {
	return s.eventTypes.ListByOwner(ctx, ownerID)
}

func (s *Service) UpdateEventType(ctx context.Context, ownerID, id string, in EventTypeInput, active bool) (*domain.EventType, error) {
	if in.DurationMinutes > 480 || in.DurationMinutes < 0 || in.BufferBeforeMinutes < 0 || in.BufferAfterMinutes < 0 || in.BufferBeforeMinutes > 240 || in.BufferAfterMinutes > 240 {
		return nil, errs.Invalid("invalid duration or buffers")
	}
	et, err := s.eventTypes.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if et.OwnerID != ownerID {
		return nil, errs.Forbidden("not your event type")
	}
	if in.Title != "" {
		et.Title = in.Title
	}
	if in.Description != "" {
		et.Description = in.Description
	}
	if in.DurationMinutes > 0 {
		et.DurationMinutes = in.DurationMinutes
	}
	if in.LocationKind != "" {
		et.LocationKind = in.LocationKind
	}
	if in.LocationDetail != "" {
		et.LocationDetail = in.LocationDetail
	}
	if in.Color != "" {
		et.Color = in.Color
	}
	et.BufferBeforeMinutes = in.BufferBeforeMinutes
	et.BufferAfterMinutes = in.BufferAfterMinutes
	et.Active = active
	if err := s.eventTypes.Update(ctx, et); err != nil {
		return nil, err
	}
	return et, nil
}

func (s *Service) DeleteEventType(ctx context.Context, ownerID, id string) error {
	return s.eventTypes.Delete(ctx, id, ownerID)
}

// ---- Availability --------------------------------------------------------

func (s *Service) GetAvailability(ctx context.Context, ownerID string) (*domain.Availability, error) {
	a, err := s.availability.Get(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	if a.Timezone == "" {
		a.Timezone = "UTC"
	}
	return a, nil
}

func (s *Service) SetAvailability(ctx context.Context, ownerID, tz string, rules []domain.Rule, overrides []domain.Override) (*domain.Availability, error) {
	if len(rules) > 100 || len(overrides) > 366 {
		return nil, errs.Invalid("too many availability rules")
	}
	for _, rule := range rules {
		if rule.Weekday < 0 || rule.Weekday > 6 || rule.StartMinute < 0 || rule.EndMinute > 1440 || rule.EndMinute <= rule.StartMinute {
			return nil, errs.Invalid("invalid availability rule")
		}
	}
	for _, override := range overrides {
		if _, err := time.Parse("2006-01-02", override.Date); err != nil {
			return nil, errs.Invalid("invalid override date")
		}
		if !override.Unavailable && (override.StartMinute < 0 || override.EndMinute > 1440 || override.EndMinute <= override.StartMinute) {
			return nil, errs.Invalid("invalid override hours")
		}
	}
	if tz == "" {
		tz = "UTC"
	}
	if _, err := time.LoadLocation(tz); err != nil {
		return nil, errs.Invalid("invalid timezone %q", tz)
	}
	a := &domain.Availability{OwnerID: ownerID, Timezone: tz, Rules: rules, Overrides: overrides}
	if err := s.availability.Set(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

// ---- Slots & bookings ----------------------------------------------------

func (s *Service) ListSlots(ctx context.Context, eventTypeID, fromDate, toDate, inviteeTZ string) ([]domain.Slot, error) {
	et, err := s.eventTypes.GetByID(ctx, eventTypeID)
	if err != nil {
		return nil, err
	}
	avail, err := s.availability.Get(ctx, et.OwnerID)
	if err != nil {
		return nil, err
	}
	loc := mustLoc(avail.Timezone)

	from, fromErr := time.ParseInLocation("2006-01-02", fromDate, loc)
	to, toErr := time.ParseInLocation("2006-01-02", toDate, loc)
	if fromErr != nil || toErr != nil || to.Before(from) || to.Sub(from) > 31*24*time.Hour {
		return nil, errs.Invalid("choose a valid date range of at most 31 days")
	}
	if !et.Active {
		return nil, errs.Invalid("event type is inactive")
	}
	busyBookings, err := s.bookings.ListConfirmedBetween(ctx, et.OwnerID, from, to.AddDate(0, 0, 1))
	if err != nil {
		return nil, err
	}
	if s.cfg.Calendar == nil {
		return nil, errs.Invalid("connect a calendar before accepting bookings")
	}
	external, err := s.cfg.Calendar.Busy(ctx, et.OwnerID, from, to.AddDate(0, 0, 1), "")
	if err != nil {
		return nil, err
	}
	busyBookings = append(busyBookings, external...)

	computed := slots.Compute(slots.Params{
		Location:        loc,
		Rules:           toSlotRules(avail.Rules),
		Overrides:       toSlotOverrides(avail.Overrides),
		Busy:            toBusy(busyBookings, ""),
		FromDate:        fromDate,
		ToDate:          toDate,
		DurationMin:     et.DurationMinutes,
		BufferBeforeMin: et.BufferBeforeMinutes,
		BufferAfterMin:  et.BufferAfterMinutes,
		Now:             s.now(),
		MaxSlots:        500,
	})

	out := make([]domain.Slot, len(computed))
	for i, sl := range computed {
		out[i] = domain.Slot{Start: sl.Start, End: sl.End}
	}
	return out, nil
}

func (s *Service) CreateBooking(ctx context.Context, eventTypeID, name, email, inviteeTZ string, start time.Time, notes string) (*domain.Booking, error) {
	if strings.TrimSpace(name) == "" || len(name) > 200 || len(notes) > 10000 {
		return nil, errs.Invalid("provide a name and valid booking answers")
	}
	if !start.Equal(start.Truncate(time.Minute)) {
		return nil, errs.Invalid("choose an offered slot")
	}
	if !validEmail(email) {
		return nil, errs.Invalid("a valid invitee email is required")
	}
	et, err := s.eventTypes.GetByID(ctx, eventTypeID)
	if err != nil {
		return nil, err
	}
	if !et.Active {
		return nil, errs.Invalid("this event type is not bookable")
	}
	free, err := s.slotFree(ctx, et, start, "")
	if err != nil {
		return nil, err
	}
	if !free {
		return nil, errs.Conflict("the requested time is no longer available")
	}
	if inviteeTZ == "" {
		inviteeTZ = "UTC"
	}
	b := &domain.Booking{
		ID:              idgen.Prefixed("bkg_"),
		EventTypeID:     et.ID,
		OwnerID:         et.OwnerID,
		InviteeName:     name,
		InviteeEmail:    email,
		InviteeTimezone: inviteeTZ,
		Start:           start.UTC(),
		End:             start.Add(time.Duration(et.DurationMinutes) * time.Minute).UTC(),
		Status:          domain.StatusConfirmed,
		Location:        et.LocationDetail,
		Notes:           notes,
		RescheduleToken: randomToken(),
		CancelToken:     randomToken(),
		CreatedAt:       s.now(),
	}
	if err := s.bookings.Create(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *Service) GetBooking(ctx context.Context, id string) (*domain.Booking, error) {
	return s.bookings.GetByID(ctx, id)
}

func (s *Service) ListBookings(ctx context.Context, ownerID, status string, limit, offset int) ([]domain.Booking, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return s.bookings.ListByOwner(ctx, ownerID, status, limit, offset)
}

func (s *Service) RescheduleBooking(ctx context.Context, rescheduleToken string, newStart time.Time) (*domain.Booking, error) {
	if !newStart.Equal(newStart.Truncate(time.Minute)) {
		return nil, errs.Invalid("choose an offered slot")
	}
	b, err := s.bookings.GetByToken(ctx, "reschedule", rescheduleToken)
	if err != nil {
		return nil, err
	}
	if b.Status == domain.StatusCancelled {
		return nil, errs.Invalid("cannot reschedule a cancelled booking")
	}
	et, err := s.eventTypes.GetByID(ctx, b.EventTypeID)
	if err != nil {
		return nil, err
	}
	free, err := s.slotFree(ctx, et, newStart, b.ID)
	if err != nil {
		return nil, err
	}
	if !free {
		return nil, errs.Conflict("the requested time is not available")
	}
	b.Start = newStart.UTC()
	b.End = newStart.Add(time.Duration(et.DurationMinutes) * time.Minute).UTC()
	if err := s.bookings.Update(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *Service) CancelBooking(ctx context.Context, cancelToken, reason string) (*domain.Booking, error) {
	if len(reason) > 1000 {
		return nil, errs.Invalid("cancellation reason is too long")
	}
	b, err := s.bookings.GetByToken(ctx, "cancel", cancelToken)
	if err != nil {
		return nil, err
	}
	if b.Status == domain.StatusCancelled {
		return b, nil
	}
	b.Status = domain.StatusCancelled
	if reason != "" {
		b.Notes = strings.TrimSpace(b.Notes + "\nCancelled: " + reason)
	}
	if err := s.bookings.Update(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

// slotFree reports whether start is a currently-bookable slot for et,
// optionally ignoring an existing booking (used on reschedule).
func (s *Service) slotFree(ctx context.Context, et *domain.EventType, start time.Time, excludeBookingID string) (bool, error) {
	if !et.Active || start.After(s.now().AddDate(0, 0, 90)) {
		return false, errs.Invalid("choose an active event within the next 90 days")
	}
	if s.cfg.Calendar == nil {
		return false, errs.Invalid("connect a calendar before accepting bookings")
	}
	if err := s.cfg.Calendar.Ready(ctx, et.OwnerID); err != nil {
		return false, err
	}
	avail, err := s.availability.Get(ctx, et.OwnerID)
	if err != nil {
		return false, err
	}
	loc := mustLoc(avail.Timezone)
	date := start.In(loc).Format("2006-01-02")

	busy, err := s.bookings.ListConfirmedBetween(ctx, et.OwnerID, start.Add(-24*time.Hour), start.Add(24*time.Hour))
	if err != nil {
		return false, err
	}
	external, err := s.cfg.Calendar.Busy(ctx, et.OwnerID, start.Add(-24*time.Hour), start.Add(24*time.Hour), excludeBookingID)
	if err != nil {
		return false, err
	}
	busy = append(busy, external...)
	free := slots.Compute(slots.Params{
		Location:        loc,
		Rules:           toSlotRules(avail.Rules),
		Overrides:       toSlotOverrides(avail.Overrides),
		Busy:            toBusy(busy, excludeBookingID),
		FromDate:        date,
		ToDate:          date,
		DurationMin:     et.DurationMinutes,
		BufferBeforeMin: et.BufferBeforeMinutes,
		BufferAfterMin:  et.BufferAfterMinutes,
		Now:             s.now(),
	})
	target := start.UTC().Truncate(time.Minute)
	for _, sl := range free {
		if sl.Start.Equal(target) {
			return true, nil
		}
	}
	return false, nil
}

func (s *Service) GetUsageStats(ctx context.Context, ownerID string) (*domain.UsageStats, error) {
	return s.bookings.GetUsageStats(ctx, ownerID)
}

// ---- Calendars -----------------------------------------------------------

func (s *Service) StartConnect(ctx context.Context, ownerID, provider string) (string, error) {
	if provider != domain.ProviderGoogle {
		return "", errs.Invalid("Google Calendar is supported for meeting bookings")
	}
	if s.cfg.Calendar == nil {
		return "", errs.Invalid("calendar integration is not configured")
	}
	return s.cfg.Calendar.Start(ctx, ownerID)
}

func (s *Service) ListConnections(ctx context.Context, ownerID string) ([]domain.CalendarConnection, error) {
	return s.calendars.ListByOwner(ctx, ownerID)
}

func (s *Service) DisconnectCalendar(ctx context.Context, ownerID, id string) error {
	return s.calendars.Delete(ctx, id, ownerID)
}

// ---- helpers -------------------------------------------------------------

func toSlotRules(rs []domain.Rule) []slots.Rule {
	out := make([]slots.Rule, len(rs))
	for i, r := range rs {
		out[i] = slots.Rule{Weekday: r.Weekday, StartMinute: r.StartMinute, EndMinute: r.EndMinute}
	}
	return out
}

func toSlotOverrides(os []domain.Override) []slots.Override {
	out := make([]slots.Override, len(os))
	for i, o := range os {
		out[i] = slots.Override{Date: o.Date, Unavailable: o.Unavailable, StartMinute: o.StartMinute, EndMinute: o.EndMinute}
	}
	return out
}

func toBusy(bs []domain.Booking, exclude string) []slots.Interval {
	out := make([]slots.Interval, 0, len(bs))
	for _, b := range bs {
		if (exclude != "" && b.ID == exclude) || b.Status == domain.StatusCancelled {
			continue
		}
		out = append(out, slots.Interval{Start: b.Start, End: b.End})
	}
	return out
}

func mustLoc(tz string) *time.Location {
	if loc, err := time.LoadLocation(tz); err == nil {
		return loc
	}
	return time.UTC
}

func randomToken() string {
	b := make([]byte, 18)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func slugify(s string) string {
	var b strings.Builder
	prevDash := false
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevDash = false
		case r == ' ' || r == '-' || r == '_':
			if !prevDash {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		out = "event"
	}
	return out + "-" + randomToken()[:6]
}

func validEmail(e string) bool {
	address, err := mail.ParseAddress(e)
	return err == nil && address.Address == e && len(e) <= 254 && !strings.ContainsAny(e, "\r\n")
}
