// Package postgres implements the scheduling repository ports over PostgreSQL.
package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/aashishrajdev/halomail/services/scheduling/internal/domain"
	"github.com/aashishrajdev/halomail/services/shared/errs"
	"github.com/aashishrajdev/halomail/services/shared/idgen"
	"github.com/aashishrajdev/halomail/services/shared/usage"
)

// ---- Event types ---------------------------------------------------------

type EventTypes struct{ pool *pgxpool.Pool }

func NewEventTypes(pool *pgxpool.Pool) *EventTypes { return &EventTypes{pool: pool} }

const etColumns = `id, owner_id, slug, title, description, duration_minutes, location_kind,
	location_detail, buffer_before_minutes, buffer_after_minutes, color, active, created_at`

func (r *EventTypes) Create(ctx context.Context, et *domain.EventType) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO event_types (`+etColumns+`)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		et.ID, et.OwnerID, et.Slug, et.Title, et.Description, et.DurationMinutes, et.LocationKind,
		et.LocationDetail, et.BufferBeforeMinutes, et.BufferAfterMinutes, et.Color, et.Active, et.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return errs.Conflict("an event type with that slug already exists")
		}
	}
	return err
}

func (r *EventTypes) GetByID(ctx context.Context, id string) (*domain.EventType, error) {
	return scanEventType(r.pool.QueryRow(ctx, `SELECT `+etColumns+` FROM event_types WHERE id = $1`, id))
}

func (r *EventTypes) ListByOwner(ctx context.Context, ownerID string) ([]domain.EventType, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+etColumns+` FROM event_types WHERE owner_id = $1 ORDER BY created_at`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.EventType
	for rows.Next() {
		et, err := scanEventType(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *et)
	}
	return out, rows.Err()
}

func (r *EventTypes) Update(ctx context.Context, et *domain.EventType) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE event_types SET title=$2, description=$3, duration_minutes=$4, location_kind=$5,
		 location_detail=$6, buffer_before_minutes=$7, buffer_after_minutes=$8, color=$9, active=$10
		 WHERE id=$1`,
		et.ID, et.Title, et.Description, et.DurationMinutes, et.LocationKind,
		et.LocationDetail, et.BufferBeforeMinutes, et.BufferAfterMinutes, et.Color, et.Active,
	)
	return err
}

func (r *EventTypes) Delete(ctx context.Context, id, ownerID string) error {
	tag, err := r.pool.Exec(ctx, `UPDATE event_types SET active=false WHERE id=$1 AND owner_id=$2`, id, ownerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errs.NotFound("event type not found")
	}
	return nil
}

func scanEventType(row pgx.Row) (*domain.EventType, error) {
	var et domain.EventType
	if err := row.Scan(&et.ID, &et.OwnerID, &et.Slug, &et.Title, &et.Description, &et.DurationMinutes,
		&et.LocationKind, &et.LocationDetail, &et.BufferBeforeMinutes, &et.BufferAfterMinutes,
		&et.Color, &et.Active, &et.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.NotFound("event type not found")
		}
		return nil, err
	}
	return &et, nil
}

// ---- Availability --------------------------------------------------------

type Availability struct{ pool *pgxpool.Pool }

func NewAvailability(pool *pgxpool.Pool) *Availability { return &Availability{pool: pool} }

func (r *Availability) Get(ctx context.Context, ownerID string) (*domain.Availability, error) {
	a := &domain.Availability{OwnerID: ownerID, Timezone: "UTC"}

	err := r.pool.QueryRow(ctx, `SELECT timezone FROM availabilities WHERE owner_id=$1`, ownerID).Scan(&a.Timezone)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	ruleRows, err := r.pool.Query(ctx,
		`SELECT weekday, start_minute, end_minute FROM availability_rules WHERE owner_id=$1 ORDER BY weekday, start_minute`, ownerID)
	if err != nil {
		return nil, err
	}
	defer ruleRows.Close()
	for ruleRows.Next() {
		var rule domain.Rule
		if err := ruleRows.Scan(&rule.Weekday, &rule.StartMinute, &rule.EndMinute); err != nil {
			return nil, err
		}
		a.Rules = append(a.Rules, rule)
	}
	if err := ruleRows.Err(); err != nil {
		return nil, err
	}

	ovRows, err := r.pool.Query(ctx,
		`SELECT on_date, unavailable, start_minute, end_minute FROM date_overrides WHERE owner_id=$1 ORDER BY on_date`, ownerID)
	if err != nil {
		return nil, err
	}
	defer ovRows.Close()
	for ovRows.Next() {
		var (
			o  domain.Override
			dt time.Time
		)
		if err := ovRows.Scan(&dt, &o.Unavailable, &o.StartMinute, &o.EndMinute); err != nil {
			return nil, err
		}
		o.Date = dt.Format("2006-01-02")
		a.Overrides = append(a.Overrides, o)
	}
	return a, ovRows.Err()
}

func (r *Availability) Set(ctx context.Context, a *domain.Availability) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if _, err := tx.Exec(ctx,
		`INSERT INTO availabilities (owner_id, timezone, updated_at) VALUES ($1,$2,now())
		 ON CONFLICT (owner_id) DO UPDATE SET timezone=excluded.timezone, updated_at=now()`,
		a.OwnerID, a.Timezone); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM availability_rules WHERE owner_id=$1`, a.OwnerID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM date_overrides WHERE owner_id=$1`, a.OwnerID); err != nil {
		return err
	}
	for _, rule := range a.Rules {
		if _, err := tx.Exec(ctx,
			`INSERT INTO availability_rules (id, owner_id, weekday, start_minute, end_minute) VALUES ($1,$2,$3,$4,$5)`,
			idgen.Prefixed("avr_"), a.OwnerID, rule.Weekday, rule.StartMinute, rule.EndMinute); err != nil {
			return err
		}
	}
	for _, o := range a.Overrides {
		if _, err := tx.Exec(ctx,
			`INSERT INTO date_overrides (id, owner_id, on_date, unavailable, start_minute, end_minute) VALUES ($1,$2,$3::date,$4,$5,$6)`,
			idgen.Prefixed("ovr_"), a.OwnerID, o.Date, o.Unavailable, o.StartMinute, o.EndMinute); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// ---- Bookings ------------------------------------------------------------

type Bookings struct {
	pool   *pgxpool.Pool
	policy usage.Policy
}

func NewBookings(pool *pgxpool.Pool, policies ...usage.Policy) *Bookings {
	policy := usage.Policy{Forms: 25, Meetings: 10, Monthly: true}
	if len(policies) > 0 {
		policy = policies[0]
	}
	return &Bookings{pool: pool, policy: policy}
}

const bkColumns = `id, event_type_id, owner_id, invitee_name, invitee_email, invitee_timezone,
	start_at, end_at, status, location, notes, reschedule_token, cancel_token, created_at`

func (r *Bookings) Create(ctx context.Context, b *domain.Booking) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err = lockSlot(ctx, tx, b); err != nil {
		return err
	}
	if err = usage.Consume(ctx, tx, r.policy, b.OwnerID, "meetings"); err != nil {
		return err
	}
	_, err = tx.Exec(ctx,
		`INSERT INTO bookings (`+bkColumns+`)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		b.ID, b.EventTypeID, b.OwnerID, b.InviteeName, b.InviteeEmail, b.InviteeTimezone,
		b.Start, b.End, b.Status, b.Location, b.Notes, b.RescheduleToken, b.CancelToken, b.CreatedAt,
	)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `UPDATE bookings SET calendar_pending=true,location='' WHERE id=$1`, b.ID)
	if err != nil {
		return err
	}
	b.Location = ""
	return tx.Commit(ctx)
}

func (r *Bookings) GetByID(ctx context.Context, id string) (*domain.Booking, error) {
	return scanBooking(r.pool.QueryRow(ctx, `SELECT `+bkColumns+` FROM bookings WHERE id=$1`, id))
}

func (r *Bookings) GetByToken(ctx context.Context, kind, token string) (*domain.Booking, error) {
	col := "cancel_token"
	if kind == "reschedule" {
		col = "reschedule_token"
	}
	return scanBooking(r.pool.QueryRow(ctx, `SELECT `+bkColumns+` FROM bookings WHERE `+col+`=$1`, token))
}

func (r *Bookings) ListByOwner(ctx context.Context, ownerID, status string, limit, offset int) ([]domain.Booking, error) {
	q := `SELECT ` + bkColumns + ` FROM bookings WHERE owner_id=$1`
	args := []any{ownerID}
	if status != "" {
		q += ` AND status=$2 ORDER BY start_at DESC LIMIT $3 OFFSET $4`
		args = append(args, status, limit, offset)
	} else {
		q += ` ORDER BY start_at DESC LIMIT $2 OFFSET $3`
		args = append(args, limit, offset)
	}
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectBookings(rows)
}

func (r *Bookings) ListConfirmedBetween(ctx context.Context, ownerID string, from, to time.Time) ([]domain.Booking, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+bkColumns+` FROM bookings
		 WHERE owner_id=$1 AND status=$2 AND start_at < $3 AND end_at > $4`,
		ownerID, domain.StatusConfirmed, to, from)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectBookings(rows)
}

func (r *Bookings) Update(ctx context.Context, b *domain.Booking) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err = lockSlot(ctx, tx, b); err != nil {
		return err
	}
	_, err = tx.Exec(ctx,
		`UPDATE bookings SET start_at=$2, end_at=$3, status=$4, notes=$5,location='',calendar_pending=true,calendar_retry_at=now(),revision=revision+1 WHERE id=$1`,
		b.ID, b.Start, b.End, b.Status, b.Notes)
	if err != nil {
		return err
	}
	b.Location = ""
	return tx.Commit(ctx)
}

func lockSlot(ctx context.Context, tx pgx.Tx, booking *domain.Booking) error {
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, booking.OwnerID); err != nil {
		return err
	}
	if booking.Status == domain.StatusCancelled {
		return nil
	}
	var currentStatus string
	statusErr := tx.QueryRow(ctx, `SELECT status FROM bookings WHERE id=$1 FOR UPDATE`, booking.ID).Scan(&currentStatus)
	if statusErr != nil && !errors.Is(statusErr, pgx.ErrNoRows) {
		return statusErr
	}
	if currentStatus == domain.StatusCancelled {
		return errs.Conflict("cannot reschedule a cancelled booking")
	}
	var overlaps bool
	err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM bookings b
	JOIN event_types existing ON existing.id=b.event_type_id JOIN event_types requested ON requested.id=$5
	WHERE b.owner_id=$1 AND b.id<>$2 AND b.status='confirmed'
	AND b.start_at-make_interval(mins=>existing.buffer_before_minutes)<$4::timestamptz+make_interval(mins=>requested.buffer_after_minutes)
	AND b.end_at+make_interval(mins=>existing.buffer_after_minutes)>$3::timestamptz-make_interval(mins=>requested.buffer_before_minutes))`, booking.OwnerID, booking.ID, booking.Start, booking.End, booking.EventTypeID).Scan(&overlaps)
	if err != nil {
		return err
	}
	if overlaps {
		return errs.Conflict("that time was just booked; choose another slot")
	}
	return nil
}

func (r *Bookings) GetUsageStats(ctx context.Context, ownerID string) (*domain.UsageStats, error) {
	var stats domain.UsageStats
	err := r.pool.QueryRow(ctx, `
		SELECT 
			COUNT(*),
			COALESCE(COUNT(*) FILTER (WHERE status = 'confirmed' AND start_at > NOW()), 0),
			COALESCE(COUNT(*) FILTER (WHERE status = 'cancelled'), 0)
		FROM bookings 
		WHERE owner_id = $1
	`, ownerID).Scan(&stats.TotalBookings, &stats.UpcomingBookings, &stats.CancelledBookings)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

func collectBookings(rows pgx.Rows) ([]domain.Booking, error) {
	var out []domain.Booking
	for rows.Next() {
		b, err := scanBooking(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *b)
	}
	return out, rows.Err()
}

func scanBooking(row pgx.Row) (*domain.Booking, error) {
	var b domain.Booking
	if err := row.Scan(&b.ID, &b.EventTypeID, &b.OwnerID, &b.InviteeName, &b.InviteeEmail, &b.InviteeTimezone,
		&b.Start, &b.End, &b.Status, &b.Location, &b.Notes, &b.RescheduleToken, &b.CancelToken, &b.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.NotFound("booking not found")
		}
		return nil, err
	}
	return &b, nil
}

// ---- Calendars -----------------------------------------------------------

type Calendars struct{ pool *pgxpool.Pool }

func NewCalendars(pool *pgxpool.Pool) *Calendars { return &Calendars{pool: pool} }

func (r *Calendars) ListByOwner(ctx context.Context, ownerID string) ([]domain.CalendarConnection, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT 'google:'||owner_id,owner_id,'google',email,created_at FROM google_credentials WHERE owner_id=$1`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.CalendarConnection
	for rows.Next() {
		var c domain.CalendarConnection
		if err := rows.Scan(&c.ID, &c.OwnerID, &c.Provider, &c.Email, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *Calendars) Delete(ctx context.Context, id, ownerID string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM google_credentials WHERE 'google:'||owner_id=$1 AND owner_id=$2`, id, ownerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errs.NotFound("calendar connection not found")
	}
	return nil
}
