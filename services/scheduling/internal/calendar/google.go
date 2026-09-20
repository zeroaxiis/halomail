package calendar

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aashishrajdev/halomail/services/scheduling/internal/domain"
	"github.com/aashishrajdev/halomail/services/shared/config"
	"github.com/aashishrajdev/halomail/services/shared/errs"
	"github.com/aashishrajdev/halomail/services/shared/outbox"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Google struct {
	pool    *pgxpool.Pool
	oauth   oauth2.Config
	key     string
	webURL  string
	baseURL string
}

func New(pool *pgxpool.Pool, settings config.OAuth, key, webURL string) *Google {
	return &Google{pool: pool, key: key, webURL: strings.TrimRight(webURL, "/"), baseURL: "https://www.googleapis.com/calendar/v3", oauth: oauth2.Config{ClientID: settings.ClientID, ClientSecret: settings.ClientSecret, RedirectURL: settings.RedirectURL, Endpoint: google.Endpoint, Scopes: []string{"https://www.googleapis.com/auth/calendar.events", "https://www.googleapis.com/auth/calendar.readonly"}}}
}

func (client *Google) configured() error {
	if client.oauth.ClientID == "" || client.oauth.ClientSecret == "" || client.oauth.RedirectURL == "" {
		return errs.Invalid("Google Calendar is not configured on this server")
	}
	_, err := client.box()
	return err
}

func (client *Google) box() (cipher.AEAD, error) {
	key, err := base64.StdEncoding.DecodeString(client.key)
	if err != nil || len(key) != 32 {
		return nil, errs.Invalid("configure a 32-byte base64 CALENDAR_ENCRYPTION_KEY")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func (client *Google) seal(owner string, token *oauth2.Token) (string, error) {
	box, err := client.box()
	if err != nil {
		return "", err
	}
	plain, err := json.Marshal(token)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, box.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(box.Seal(nonce, nonce, plain, []byte(owner))), nil
}

func (client *Google) open(owner, encrypted string) (*oauth2.Token, error) {
	box, err := client.box()
	if err != nil {
		return nil, err
	}
	data, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil || len(data) < box.NonceSize() {
		return nil, errs.Internal("invalid stored calendar credential")
	}
	plain, err := box.Open(nil, data[:box.NonceSize()], data[box.NonceSize():], []byte(owner))
	if err != nil {
		return nil, errs.Internal("cannot decrypt calendar credential")
	}
	var token oauth2.Token
	err = json.Unmarshal(plain, &token)
	return &token, err
}

func digest(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func (client *Google) Start(ctx context.Context, owner string) (string, error) {
	if err := client.configured(); err != nil {
		return "", err
	}
	if _, err := client.pool.Exec(ctx, `DELETE FROM calendar_oauth_states WHERE expires_at<=now()`); err != nil {
		return "", err
	}
	state, verifier := oauth2.GenerateVerifier(), oauth2.GenerateVerifier()
	_, err := client.pool.Exec(ctx, `INSERT INTO calendar_oauth_states(state_hash,owner_id,verifier,expires_at) VALUES($1,$2,$3,now()+interval '10 minutes')`, digest(state), owner, verifier)
	if err != nil {
		return "", err
	}
	return client.oauth.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "consent"), oauth2.S256ChallengeOption(verifier)), nil
}

func (client *Google) Callback(ctx context.Context, owner, state, code string) error {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	if err := client.configured(); err != nil {
		return err
	}
	var verifier string
	err := client.pool.QueryRow(ctx, `DELETE FROM calendar_oauth_states WHERE state_hash=$1 AND owner_id=$2 AND expires_at>now() RETURNING verifier`, digest(state), owner).Scan(&verifier)
	if err != nil {
		return errs.Unauthorized("calendar authorization expired; connect again")
	}
	token, err := client.oauth.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return errs.Invalid("Google authorization failed; connect again")
	}
	if token.RefreshToken == "" {
		return errs.Invalid("Google did not grant offline access; reconnect with consent")
	}
	encrypted, err := client.seal(owner, token)
	if err != nil {
		return err
	}
	_, err = client.pool.Exec(ctx, `INSERT INTO google_credentials(owner_id,encrypted_token) VALUES($1,$2)
	ON CONFLICT(owner_id) DO UPDATE SET encrypted_token=excluded.encrypted_token`, owner, encrypted)
	return err
}

func (client *Google) Ready(ctx context.Context, owner string) error {
	if err := client.configured(); err != nil {
		return err
	}
	var exists bool
	if err := client.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM google_credentials WHERE owner_id=$1)`, owner).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return errs.Invalid("connect Google Calendar before accepting bookings")
	}
	return nil
}

func (client *Google) request(ctx context.Context, owner, method, path string, body any, output any) (int, error) {
	var encrypted string
	if err := client.pool.QueryRow(ctx, `SELECT encrypted_token FROM google_credentials WHERE owner_id=$1`, owner).Scan(&encrypted); err != nil {
		return 0, errs.Invalid("connect Google Calendar before accepting bookings")
	}
	token, err := client.open(owner, encrypted)
	if err != nil {
		return 0, err
	}
	httpClient := &http.Client{Timeout: 15 * time.Second}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)
	fresh, err := client.oauth.TokenSource(ctx, token).Token()
	if err != nil {
		return 0, errs.Invalid("calendar access expired; reconnect Google Calendar")
	}
	if fresh.AccessToken != token.AccessToken {
		if fresh.RefreshToken == "" {
			fresh.RefreshToken = token.RefreshToken
		}
		sealed, err := client.seal(owner, fresh)
		if err != nil {
			return 0, err
		}
		if _, err = client.pool.Exec(ctx, `UPDATE google_credentials SET encrypted_token=$2 WHERE owner_id=$1 AND encrypted_token=$3`, owner, sealed, encrypted); err != nil {
			return 0, err
		}
	}
	var buffer bytes.Buffer
	if body != nil {
		if err = json.NewEncoder(&buffer).Encode(body); err != nil {
			return 0, err
		}
	}
	request, err := http.NewRequestWithContext(ctx, method, client.baseURL+path, &buffer)
	if err != nil {
		return 0, err
	}
	fresh.SetAuthHeader(request)
	request.Header.Set("Content-Type", "application/json")
	response, err := httpClient.Do(request)
	if err != nil {
		return 0, errs.Internal("calendar provider unavailable")
	}
	defer response.Body.Close()
	if response.StatusCode >= 300 {
		return response.StatusCode, errs.Internal("calendar request failed (%d)", response.StatusCode)
	}
	if output != nil {
		err = json.NewDecoder(io.LimitReader(response.Body, 2<<20)).Decode(output)
	}
	return response.StatusCode, err
}

func (client *Google) Busy(ctx context.Context, owner string, from, to time.Time, excludeBookingID string) ([]domain.Booking, error) {
	if excludeBookingID != "" {
		return client.busyExcept(ctx, owner, from, to, excludeBookingID)
	}
	var response struct {
		Calendars map[string]struct {
			Busy []struct {
				Start time.Time
				End   time.Time
			}
			Errors []json.RawMessage
		}
	}
	_, err := client.request(ctx, owner, "POST", "/freeBusy", map[string]any{"timeMin": from.UTC().Format(time.RFC3339), "timeMax": to.UTC().Format(time.RFC3339), "items": []map[string]string{{"id": "primary"}}}, &response)
	if err != nil {
		return nil, err
	}
	calendar, exists := response.Calendars["primary"]
	if !exists || len(calendar.Errors) > 0 {
		return nil, errs.Internal("could not read calendar availability")
	}
	busy := make([]domain.Booking, 0, len(calendar.Busy))
	for _, interval := range calendar.Busy {
		busy = append(busy, domain.Booking{Start: interval.Start, End: interval.End, Status: domain.StatusConfirmed})
	}
	return busy, nil
}

func (client *Google) busyExcept(ctx context.Context, owner string, from, to time.Time, excludeBookingID string) ([]domain.Booking, error) {
	query := url.Values{"timeMin": {from.UTC().Format(time.RFC3339)}, "timeMax": {to.UTC().Format(time.RFC3339)}, "singleEvents": {"true"}, "maxResults": {"2500"}}
	busy := []domain.Booking{}
	for page := 0; page < 20; page++ {
		var response struct {
			TimeZone      string `json:"timeZone"`
			NextPageToken string `json:"nextPageToken"`
			Items         []struct {
				ID           string    `json:"id"`
				Status       string    `json:"status"`
				Transparency string    `json:"transparency"`
				Start        eventTime `json:"start"`
				End          eventTime `json:"end"`
			} `json:"items"`
		}
		if _, err := client.request(ctx, owner, "GET", "/calendars/primary/events?"+query.Encode(), nil, &response); err != nil {
			return nil, err
		}
		location, err := time.LoadLocation(response.TimeZone)
		if err != nil {
			return nil, errs.Internal("calendar timezone is unavailable")
		}
		for _, event := range response.Items {
			if event.ID == digest(excludeBookingID) || event.Status == "cancelled" || event.Transparency == "transparent" {
				continue
			}
			start, err := event.Start.value(location)
			if err != nil {
				return nil, err
			}
			end, err := event.End.value(location)
			if err != nil {
				return nil, err
			}
			busy = append(busy, domain.Booking{Start: start, End: end, Status: domain.StatusConfirmed})
		}
		if response.NextPageToken == "" {
			return busy, nil
		}
		query.Set("pageToken", response.NextPageToken)
	}
	return nil, errs.Internal("calendar has too many events to check safely")
}

type eventTime struct {
	DateTime string `json:"dateTime"`
	Date     string `json:"date"`
}

func (value eventTime) value(location *time.Location) (time.Time, error) {
	if value.DateTime != "" {
		return time.Parse(time.RFC3339, value.DateTime)
	}
	return time.ParseInLocation("2006-01-02", value.Date, location)
}

type eventResponse struct {
	HangoutLink    string    `json:"hangoutLink"`
	Start          eventTime `json:"start"`
	End            eventTime `json:"end"`
	ConferenceData struct {
		CreateRequest struct {
			Status struct {
				StatusCode string `json:"statusCode"`
			} `json:"status"`
		} `json:"createRequest"`
	} `json:"conferenceData"`
}

func (client *Google) sync(ctx context.Context, booking domain.Booking, title string) (string, error) {
	eventID := digest(booking.ID)
	path := "/calendars/primary/events/" + eventID
	if booking.Status == domain.StatusCancelled {
		status, err := client.request(ctx, booking.OwnerID, "DELETE", path+"?sendUpdates=all", nil, nil)
		if status == 404 || status == 410 {
			return "", nil
		}
		return "", err
	}
	body := map[string]any{"summary": title, "description": html.EscapeString(booking.Notes), "start": map[string]string{"dateTime": booking.Start.Format(time.RFC3339)}, "end": map[string]string{"dateTime": booking.End.Format(time.RFC3339)}, "attendees": []map[string]string{{"email": booking.InviteeEmail, "displayName": booking.InviteeName}}, "guestsCanInviteOthers": false, "guestsCanSeeOtherGuests": false}
	var existing eventResponse
	status, err := client.request(ctx, booking.OwnerID, "GET", path, nil, &existing)
	if status == 404 {
		body["id"] = eventID
		body["conferenceData"] = map[string]any{"createRequest": map[string]any{"requestId": eventID, "conferenceSolutionKey": map[string]string{"type": "hangoutsMeet"}}}
		_, err = client.request(ctx, booking.OwnerID, "POST", "/calendars/primary/events?conferenceDataVersion=1&sendUpdates=all", body, &existing)
	} else if err == nil && !sameTime(existing, booking) {
		_, err = client.request(ctx, booking.OwnerID, "PATCH", path+"?conferenceDataVersion=1&sendUpdates=all", body, &existing)
	}
	if err != nil {
		return "", err
	}
	if existing.HangoutLink == "" {
		return "", errs.Internal("Google Meet link is not ready; check calendar conferencing support")
	}
	return existing.HangoutLink, nil
}

func sameTime(event eventResponse, booking domain.Booking) bool {
	start, startErr := time.Parse(time.RFC3339, event.Start.DateTime)
	end, endErr := time.Parse(time.RFC3339, event.End.DateTime)
	return startErr == nil && endErr == nil && start.Equal(booking.Start) && end.Equal(booking.End)
}

func (client *Google) Run(ctx context.Context, logger *slog.Logger) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := client.tick(ctx); err != nil {
				logger.ErrorContext(ctx, "calendar sync pending", "error", err.Error())
			}
		}
	}
}

func (client *Google) tick(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	tx, err := client.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var booking domain.Booking
	var title, ownerEmail string
	var revision int
	err = tx.QueryRow(ctx, `SELECT b.id,b.owner_id,b.invitee_name,b.invitee_email,b.start_at,b.end_at,b.status,b.notes,b.cancel_token,b.reschedule_token,b.revision,e.title,u.email
	FROM bookings b JOIN event_types e ON e.id=b.event_type_id JOIN users u ON u.id=b.owner_id
	WHERE b.calendar_pending AND b.calendar_retry_at<=now() ORDER BY b.calendar_retry_at FOR UPDATE OF b SKIP LOCKED LIMIT 1`).Scan(&booking.ID, &booking.OwnerID, &booking.InviteeName, &booking.InviteeEmail, &booking.Start, &booking.End, &booking.Status, &booking.Notes, &booking.CancelToken, &booking.RescheduleToken, &revision, &title, &ownerEmail)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	link, syncErr := client.sync(ctx, booking, title)
	if syncErr != nil {
		if _, err = tx.Exec(ctx, `UPDATE bookings SET calendar_retry_at=now()+interval '1 minute' WHERE id=$1`, booking.ID); err != nil {
			return err
		}
		if err = tx.Commit(ctx); err != nil {
			return err
		}
		return syncErr
	}
	fields := map[string]string{"Event": title, "Starts (UTC)": booking.Start.UTC().Format(time.RFC1123), "Ends (UTC)": booking.End.UTC().Format(time.RFC1123), "Guest": booking.InviteeName, "Guest email": booking.InviteeEmail, "Meeting link": link, "Answers": booking.Notes, "Status": booking.Status}
	subject := "Meeting confirmed: " + title
	if booking.Status == domain.StatusCancelled {
		subject = "Meeting cancelled: " + title
	} else {
		fields["Cancel booking"] = client.webURL + "/booking/manage?cancel=" + url.QueryEscape(booking.CancelToken)
	}
	body := outbox.Fields(subject, fields)
	for index, recipient := range []string{ownerEmail, booking.InviteeEmail} {
		if err = outbox.Queue(ctx, tx, fmt.Sprintf("%s-%d-%d", booking.ID, revision, index), recipient, "", subject, body); err != nil {
			return err
		}
	}
	if _, err = tx.Exec(ctx, `UPDATE bookings SET calendar_pending=false,location=$2 WHERE id=$1`, booking.ID, link); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
