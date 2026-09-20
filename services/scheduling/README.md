# scheduling — availability, bookings, calendars

Public booking pages and everything behind them.

- **Dev port:** `8082`
- **Proto:** [`proto/halomail/scheduling/v1`](../../proto/halomail/scheduling/v1/scheduling.proto)
- **Module:** `github.com/aashishrajdev/halomail/services/scheduling`

## Services / RPCs

| Service               | RPCs                                                        |
| --------------------- | ---------------------------------------------------------- |
| `EventTypeService`    | Create/Get/List/Update/Delete EventType                    |
| `AvailabilityService` | GetAvailability, SetAvailability                            |
| `BookingService`      | ListSlots (public), CreateBooking (public), GetBooking, ListBookings, RescheduleBooking (token), CancelBooking (token) |
| `CalendarService`     | StartConnect (OAuth), ListConnections, DisconnectCalendar  |

## How slots are computed

```
weekly AvailabilityRules (owner tz)
  ─ minus DateOverrides (block/extend specific days)
  ─ minus existing bookings (+ buffers)
  ─ minus busy blocks from connected Google calendars
  = available slots (respecting buffer/notice times)
  ─ projected into the invitee's timezone
```

## Calendar integration

- **Google** (Calendar API) via OAuth2.
- `StartConnect` returns a consent URL; the callback (handled at the gateway)
  exchanges the code and stores **encrypted** refresh tokens.
- On booking: a calendar event is created and its busy time feeds back into
  future slot computation.

## Data model

| Table                   | Notes                                              |
| ----------------------- | -------------------------------------------------- |
| `event_types`           | bookable templates (slug per owner)                |
| `availability_rules`    | weekly windows (weekday, start/end minute)         |
| `date_overrides`        | per-date block/extend                              |
| `bookings`              | invitee details, start/end, status, reschedule/cancel tokens |
| `calendar_connections`  | provider, account email, encrypted tokens          |

## Configuration

| Env                | Purpose                         |
| ------------------ | ------------------------------- |
| `GOOGLE_*`         | Google OAuth client            |
| `MICROSOFT_*`      | Microsoft OAuth client         |

## Run

```bash
cd services/scheduling && HTTP_PORT=8082 go run ./cmd/server
```
