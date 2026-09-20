"use client";

import { useEffect, useState } from "react";
import { FeatureAccess } from "@/components/dashboard/feature-access";
import { PageHeader } from "@/components/dashboard/page-header";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select } from "@/components/ui/select";
import { rpc } from "@/lib/api";
import { localTimezone, timezones } from "@/lib/timezones";
import { useRpc } from "@/lib/use-rpc";

interface Booking { id: string; inviteeName: string; inviteeEmail: string; start: string; status: string; location?: string }
interface EventType { id: string; title: string; durationMinutes: number }
interface Rule { weekday: number; startMinute: number; endMinute: number }

export default function MeetingsPage() {
  const bookings = useRpc<{ bookings?: Booking[] }>("halomail.scheduling.v1.BookingService/ListBookings");
  const stats = useRpc<{ totalBookings?: number; upcomingBookings?: number; cancelledBookings?: number }>("halomail.scheduling.v1.BookingService/GetUsageStats");
  const events = useRpc<{ eventTypes?: EventType[] }>("halomail.scheduling.v1.EventTypeService/ListEventTypes");
  const connections = useRpc<{ connections?: { id: string; email: string }[] }>("halomail.scheduling.v1.CalendarService/ListConnections");
  const availability = useRpc<{ availability: { timezone: string; rules?: Rule[]; overrides?: unknown[] } }>("halomail.scheduling.v1.AvailabilityService/GetAvailability");
  const [title, setTitle] = useState("Introductory call");
  const [duration, setDuration] = useState(30);
  const [timezone, setTimezone] = useState("UTC");
  const [rules, setRules] = useState<Rule[]>([1, 2, 3, 4, 5].map(weekday => ({ weekday, startMinute: 540, endMinute: 1020 })));
  const [busy, setBusy] = useState(false);
  const [notice, setNotice] = useState("");
  const [error, setError] = useState("");

  useEffect(() => {
    setTimezone(localTimezone());
    const result = new URLSearchParams(window.location.search).get("calendar");
    if (result === "connected") setNotice("Google Calendar connected.");
    if (result === "failed") setError("Calendar connection failed or expired. Please connect again.");
  }, []);
  useEffect(() => {
    if (availability.data?.availability) {
      setTimezone(availability.data.availability.timezone || localTimezone());
      setRules((availability.data.availability.rules || []).map(rule => ({ weekday: rule.weekday ?? 0, startMinute: rule.startMinute ?? 0, endMinute: rule.endMinute ?? 0 })));
    }
  }, [availability.data]);

  async function perform(action: () => Promise<void>) {
    setBusy(true); setError(""); setNotice("");
    try { await action(); }
    catch (failure) { setError(failure instanceof Error ? failure.message : "Request failed"); }
    finally { setBusy(false); }
  }

  function connect() {
    void perform(async () => {
      const result = await rpc<{ authorizationUrl: string }>("halomail.scheduling.v1.CalendarService/StartConnect", { provider: "CALENDAR_PROVIDER_GOOGLE" });
      window.location.assign(result.authorizationUrl);
    });
  }

  function createEvent(event: React.FormEvent) {
    event.preventDefault();
    void perform(async () => {
      await rpc("halomail.scheduling.v1.EventTypeService/CreateEventType", { title, durationMinutes: duration, locationKind: "LOCATION_KIND_GOOGLE_MEET", description: "Tell us what you would like to discuss when booking." });
      await events.reload(); setNotice("Event type created.");
    });
  }

  function saveAvailability(event: React.FormEvent) {
    event.preventDefault();
    void perform(async () => {
      await rpc("halomail.scheduling.v1.AvailabilityService/SetAvailability", { timezone, rules, overrides: availability.data?.availability.overrides || [] });
      await availability.reload(); setNotice("Availability saved.");
    });
  }

  function changeRule(weekday: number, update: Partial<Rule>) {
    setRules(current => current.map(rule => rule.weekday === weekday ? { ...rule, ...update } : rule));
  }

  return <>
    <PageHeader title="Meetings" description="Turn a website button into a calendar booking with a Google Meet link." />
    <FeatureAccess feature="meetings" />
    <div className="mb-6 grid gap-4 sm:grid-cols-3">
      {[["Total bookings", stats.data?.totalBookings], ["Upcoming", stats.data?.upcomingBookings], ["Cancelled", stats.data?.cancelledBookings]].map(([label, value]) => <Card key={String(label)} className="p-5"><p className="text-sm text-muted-foreground">{label}</p><p className="mt-2 text-2xl font-semibold">{stats.loading ? "…" : value || 0}</p></Card>)}
    </div>
    {(error || bookings.error || stats.error || events.error || connections.error || availability.error) && <p role="alert" className="mb-4 text-sm text-destructive">{error || bookings.error || stats.error || events.error || connections.error || availability.error}</p>}
    {notice && <p role="status" className="mb-4 text-sm text-brand">{notice}</p>}
    <div className="mb-6 grid gap-6 lg:grid-cols-2">
      <Card className="space-y-4 p-6">
        <h2 className="font-medium">1. Connect your calendar</h2>
        <p className="text-sm text-muted-foreground">Bookings check your primary calendar for conflicts and create a Google Meet event. You and your guest receive confirmation emails.</p>
        <p className="text-sm">{connections.data?.connections?.length ? "Google Calendar connected" : "No calendar connected"}</p>
        <Button onClick={connect} disabled={busy}>{connections.data?.connections?.length ? "Reconnect Google Calendar" : "Connect Google Calendar"}</Button>
        <h2 className="pt-4 font-medium">2. Create an event type</h2>
        <form onSubmit={createEvent} className="space-y-3">
          <Label htmlFor="event-title">Meeting title</Label><Input id="event-title" required maxLength={200} value={title} onChange={event => setTitle(event.target.value)} />
          <Label htmlFor="duration">Duration in minutes</Label><Input id="duration" type="number" min={5} max={480} required value={duration} onChange={event => setDuration(Number(event.target.value))} />
          <Button disabled={busy} type="submit">Create event type</Button>
        </form>
        <ul className="space-y-2 text-sm">{events.data?.eventTypes?.map(event => <li key={event.id}>{event.title} · {event.durationMinutes} minutes</li>)}</ul>
      </Card>
      <Card className="p-6">
        <h2 className="mb-4 font-medium">3. Save weekly availability</h2>
        <form onSubmit={saveAvailability} className="space-y-4">
          <Label htmlFor="meeting-timezone">Your timezone</Label>
          <Select id="meeting-timezone" value={timezone} onChange={event => setTimezone(event.target.value)}>{Array.from(new Set([timezone, ...timezones()])).map(zone => <option key={zone}>{zone}</option>)}</Select>
          {["Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"].map((day, weekday) => {
            const rule = rules.find(candidate => candidate.weekday === weekday);
            return <div key={day} className="flex flex-wrap items-center gap-2 text-sm">
              <label className="w-28"><input type="checkbox" checked={!!rule} onChange={event => setRules(current => event.target.checked ? [...current, { weekday, startMinute: 540, endMinute: 1020 }] : current.filter(candidate => candidate.weekday !== weekday))} /> {day}</label>
              {rule && <><input aria-label={day + " start"} className="rounded border bg-background p-1" type="time" required value={clockTime(rule.startMinute)} onChange={event => changeRule(weekday, { startMinute: minutes(event.target.value) })} />
                <span>to</span><input aria-label={day + " end"} className="rounded border bg-background p-1" type="time" required value={clockTime(rule.endMinute)} onChange={event => changeRule(weekday, { endMinute: minutes(event.target.value) })} /></>}
            </div>;
          })}
          <Button disabled={busy || availability.loading || !!availability.error} type="submit">Save availability</Button>
        </form>
      </Card>
    </div>
    <div className="mb-3 flex items-center justify-between"><h2 className="font-medium">Recent bookings</h2><Button variant="outline" onClick={() => { void bookings.reload(); void stats.reload(); }}>Refresh</Button></div>
    <Card className="divide-y divide-border">
      {bookings.loading ? <p className="p-5">Loading…</p> : !bookings.data?.bookings?.length ? <p className="p-5 text-muted-foreground">No bookings yet.</p> : bookings.data.bookings.map(booking => <div key={booking.id} className="flex flex-wrap justify-between gap-3 p-5 text-sm">
        <div><p className="font-medium">{booking.inviteeName}</p><p className="text-muted-foreground">{booking.inviteeEmail}</p></div>
        <div><p>{new Date(booking.start).toLocaleString()}</p><p className="text-muted-foreground">{booking.status.replace("BOOKING_STATUS_", "").toLowerCase()}</p>
          {booking.location?.startsWith("https://meet.google.com/") ? <a className="text-brand underline" href={booking.location} target="_blank" rel="noreferrer">Join meeting</a> : booking.status !== "BOOKING_STATUS_CANCELLED" && <span className="text-xs text-muted-foreground">Calendar confirmation pending</span>}</div>
      </div>)}
    </Card>
  </>;
}

function clockTime(value: number) { return String(Math.floor(value / 60)).padStart(2, "0") + ":" + String(value % 60).padStart(2, "0"); }
function minutes(value: string) { const [hours, remainder] = value.split(":").map(Number); return hours * 60 + remainder; }
