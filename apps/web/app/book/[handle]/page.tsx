"use client";

import { Suspense, useEffect, useState } from "react";
import { useParams, useSearchParams } from "next/navigation";
import Link from "next/link";
import { Logo } from "@/components/logo";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { rpc } from "@/lib/api";

interface EventType {
  id: string;
  title: string;
  description: string;
  durationMinutes: number;
}

interface Slot {
  start: string;
  end: string;
}

interface HostOwner {
  id?: string;
  name: string;
  handle: string;
  avatar?: string;
  tier?: string;
}

// Clean SVG Icons (replacing all emojis with precise 1.5px stroke geometry)
function ClockIcon({ className = "size-4" }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
      <circle cx="12" cy="12" r="10" />
      <polyline points="12 6 12 12 16 14" />
    </svg>
  );
}

function VideoIcon({ className = "size-4" }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
      <polygon points="23 7 16 12 23 17 23 7" />
      <rect x="1" y="5" width="15" height="14" rx="2" ry="2" />
    </svg>
  );
}

function GlobeIcon({ className = "size-4" }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
      <circle cx="12" cy="12" r="10" />
      <line x1="2" y1="12" x2="22" y2="12" />
      <path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z" />
    </svg>
  );
}

function CheckIcon({ className = "size-4" }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <polyline points="20 6 9 17 4 12" />
    </svg>
  );
}

function ChevronLeftIcon({ className = "size-4" }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round">
      <polyline points="15 18 9 12 15 6" />
    </svg>
  );
}

function ChevronRightIcon({ className = "size-4" }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round">
      <polyline points="9 18 15 12 9 6" />
    </svg>
  );
}

function CalendarIcon({ className = "size-4" }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
      <rect x="3" y="4" width="18" height="18" rx="2" ry="2" />
      <line x1="16" y1="2" x2="16" y2="6" />
      <line x1="8" y1="2" x2="8" y2="6" />
      <line x1="3" y1="10" x2="21" y2="10" />
    </svg>
  );
}

function Booking() {
  const params = useParams();
  const searchParams = useSearchParams();

  const handle = (params?.handle as string) || "";
  const key = searchParams.get("key") || "";
  const isEmbed = searchParams.get("embed") === "true";

  const [info, setInfo] = useState<{
    owner: HostOwner;
    watermark?: boolean;
    events: EventType[];
  } | null>(null);

  const [eventID, setEventID] = useState("");
  const [slots, setSlots] = useState<Slot[]>([]);
  const [picked, setPicked] = useState<Slot | null>(null);
  const [offset, setOffset] = useState(0);
  const [timezone, setTimezone] = useState("UTC");

  // Form Fields (Standard + Custom Questions)
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [company, setCompany] = useState("");
  const [budget, setBudget] = useState("$5,000 - $15,000");
  const [notes, setNotes] = useState("");

  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [done, setDone] = useState(false);
  const [confirmedBooking, setConfirmedBooking] = useState<{
    id?: string;
    start?: string;
    meetLink?: string;
  } | null>(null);

  // 1. Initial Load of Host & Event Types
  useEffect(() => {
    try {
      setTimezone(Intl.DateTimeFormat().resolvedOptions().timeZone);
    } catch {}

    if (!handle && !key) {
      setError("Please specify a host handle (e.g. /book/drdead).");
      setLoading(false);
      return;
    }

    setLoading(true);
    rpc<{
      owner: HostOwner;
      watermark?: boolean;
      events: EventType[];
    }>("v1/meetings/info", { handle: handle }, null, key)
      .then((result) => {
        setInfo(result);
        if (result.events && result.events.length > 0) {
          setEventID(result.events[0].id);
        }
      })
      .catch((err) => {
        setError(err.message || "Host not found or unavailable.");
      })
      .finally(() => setLoading(false));
  }, [handle, key]);

  // 2. Fetch Available Slots for Selected Event
  useEffect(() => {
    if (!eventID) return;
    let current = true;
    setLoading(true);
    setError("");
    setPicked(null);

    const from = new Date(Date.now() + offset * 86400000);
    const to = new Date(from.getTime() + 6 * 86400000);

    rpc<{ slots?: Slot[] }>(
      "halomail.scheduling.v1.BookingService/ListSlots",
      {
        eventTypeId: eventID,
        fromDate: from.toISOString().slice(0, 10),
        toDate: to.toISOString().slice(0, 10),
        inviteeTimezone: timezone,
      },
      null,
      key
    )
      .then((result) => {
        if (current) setSlots(result.slots || []);
      })
      .catch((err) => {
        if (current) {
          setSlots([]);
          setError(err.message || "Failed to load slots.");
        }
      })
      .finally(() => {
        if (current) setLoading(false);
      });

    return () => {
      current = false;
    };
  }, [eventID, key, offset, timezone]);

  // 3. Confirm Booking with Custom Notes
  async function book(event: React.FormEvent) {
    event.preventDefault();
    if (!picked || busy) return;
    setBusy(true);
    setError("");

    // Compile custom fields into booking notes
    const formattedAnswers = [
      `Company: ${company || "Not provided"}`,
      `Preferred Budget: ${budget}`,
      `Agenda / Discussion Points:\n${notes}`,
    ].join("\n\n");

    try {
      const res = await rpc<{ booking?: { id: string; location?: string } }>(
        "halomail.scheduling.v1.BookingService/CreateBooking",
        {
          eventTypeId: eventID,
          inviteeName: name,
          inviteeEmail: email,
          inviteeTimezone: timezone,
          start: picked.start,
          notes: formattedAnswers,
        },
        null,
        key
      );

      setConfirmedBooking({
        id: res?.booking?.id || "bkg_confirmed",
        start: picked.start,
        meetLink: res?.booking?.location,
      });
      setDone(true);
    } catch (failure) {
      setError(
        failure instanceof Error ? failure.message : "Could not complete booking"
      );
    } finally {
      setBusy(false);
    }
  }

  const selectedEvent = info?.events.find((e) => e.id === eventID);
  const showWatermark = info?.watermark !== false;

  return (
    <div
      className={`min-h-screen ${isEmbed ? "p-3 sm:p-5" : "p-6 sm:p-12"} flex flex-col items-center justify-between text-foreground antialiased`}
      style={{
        backgroundColor: isEmbed ? "transparent" : "hsl(var(--background))",
      }}
    >
      <div className={`w-full max-w-4xl mx-auto ${isEmbed ? "my-0" : "my-6"} space-y-6`}>
        {!isEmbed && (
          <header className="flex items-center justify-between pb-4 border-b border-border">
            <Link href="/" className="inline-block">
              <Logo />
            </Link>
            <span className="text-xs text-muted-foreground font-mono">
              halomail scheduling
            </span>
          </header>
        )}

        {/* Done / Confirmed View */}
        {done ? (
          <Card className="p-8 sm:p-12 text-center space-y-6 bg-card border-border shadow-2xl rounded-2xl">
            <div className="size-12 rounded-full bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 flex items-center justify-center mx-auto">
              <CheckIcon className="size-5" />
            </div>
            <div className="space-y-1.5">
              <h1 className="text-xl sm:text-2xl font-semibold tracking-tight text-foreground">
                Booking Confirmed
              </h1>
              <p className="text-muted-foreground text-sm max-w-md mx-auto">
                Your consultation with{" "}
                <strong className="text-foreground">{info?.owner.name || handle}</strong>{" "}
                has been scheduled.
              </p>
            </div>

            <div className="bg-secondary/40 border border-border rounded-xl p-5 text-left max-w-md mx-auto text-sm space-y-3">
              <div className="flex justify-between py-1 border-b border-border/60">
                <span className="text-muted-foreground">Date &amp; Time</span>
                <span className="font-medium text-foreground">
                  {confirmedBooking?.start
                    ? new Date(confirmedBooking.start).toLocaleString(undefined, {
                        dateStyle: "full",
                        timeStyle: "short",
                      })
                    : ""}
                </span>
              </div>
              <div className="flex justify-between py-1 border-b border-border/60">
                <span className="text-muted-foreground">Guest</span>
                <span className="font-medium text-foreground">{email}</span>
              </div>
              <div className="flex justify-between py-1 border-b border-border/60">
                <span className="text-muted-foreground">Conferencing</span>
                <span className="text-brand font-medium">Google Meet</span>
              </div>
              <div className="flex justify-between py-1">
                <span className="text-muted-foreground">Calendar Event</span>
                <span className="text-emerald-400 font-medium">Synced with Google Calendar</span>
              </div>
            </div>

            <p className="text-xs text-muted-foreground max-w-sm mx-auto">
              A calendar invite and conferencing link have been emailed to both attendee and host.
            </p>

            <Button
              variant="outline"
              className="border-border bg-transparent hover:bg-secondary/60 text-foreground"
              onClick={() => {
                setDone(false);
                setPicked(null);
              }}
            >
              Book Another Slot
            </Button>
          </Card>
        ) : (
          /* Main Booking Card — Website Themed */
          <Card className="overflow-hidden border border-border bg-card rounded-2xl shadow-2xl">
            <div className="grid grid-cols-1 md:grid-cols-12 min-h-[500px]">
              
              {/* Left Column: Host Profile & Event Selection */}
              <div className="md:col-span-5 p-6 sm:p-7 border-b md:border-b-0 md:border-r border-border flex flex-col justify-between bg-card/60">
                <div className="space-y-6">
                  {/* Host Header */}
                  <div className="flex items-center gap-3.5">
                    <div className="size-12 rounded-full bg-secondary border border-border flex items-center justify-center text-foreground font-semibold text-sm shadow-inner overflow-hidden">
                      {info?.owner?.avatar ? (
                        <img
                          src={info.owner.avatar}
                          alt={info.owner.name}
                          className="size-full object-cover"
                        />
                      ) : (
                        (info?.owner.name || handle).slice(0, 2).toUpperCase()
                      )}
                    </div>
                    <div>
                      <h2 className="font-semibold text-base leading-snug text-foreground">
                        {info?.owner.name || handle}
                      </h2>
                      <span className="text-xs text-muted-foreground font-mono">
                        @{info?.owner.handle || handle}
                      </span>
                    </div>
                  </div>

                  {/* Active Event Types */}
                  <div className="space-y-2">
                    <span className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
                      Select Service
                    </span>
                    <div className="space-y-1.5">
                      {info?.events?.map((ev) => (
                        <button
                          key={ev.id}
                          type="button"
                          onClick={() => setEventID(ev.id)}
                          className={`w-full text-left p-3 rounded-lg border text-sm transition-all ${
                            eventID === ev.id
                              ? "border-brand/60 bg-brand/10 text-foreground"
                              : "border-border/80 hover:border-border bg-secondary/40 text-muted-foreground hover:text-foreground"
                          }`}
                        >
                          <div className="font-medium flex justify-between items-center">
                            <span>{ev.title}</span>
                            <span className="font-mono text-xs text-muted-foreground">
                              {ev.durationMinutes}m
                            </span>
                          </div>
                          {ev.description && (
                            <p className="text-xs text-muted-foreground mt-1 line-clamp-2">
                              {ev.description}
                            </p>
                          )}
                        </button>
                      ))}
                      {info && !info.events.length && (
                        <p className="text-xs text-muted-foreground">
                          No active event types published by host.
                        </p>
                      )}
                    </div>
                  </div>

                  {/* Event Details */}
                  {selectedEvent && (
                    <div className="pt-4 border-t border-border space-y-2.5 text-xs text-muted-foreground">
                      <div className="flex items-center gap-2.5">
                        <ClockIcon className="size-4 text-muted-foreground" />
                        <span>{selectedEvent.durationMinutes} Minutes Consultation</span>
                      </div>
                      <div className="flex items-center gap-2.5">
                        <VideoIcon className="size-4 text-muted-foreground" />
                        <span>Google Meet Video Conferencing</span>
                      </div>
                    </div>
                  )}
                </div>

                {/* Timezone display */}
                <div className="pt-5 border-t border-border flex items-center gap-2 text-xs text-muted-foreground">
                  <GlobeIcon className="size-3.5 text-muted-foreground" />
                  <span>Timezone:</span>
                  <span className="text-foreground font-medium">{timezone}</span>
                </div>
              </div>

              {/* Right Column: Slot Picker / Booking Form */}
              <div className="md:col-span-7 p-6 sm:p-7 flex flex-col justify-between bg-card">
                {error && (
                  <div className="mb-4 p-3 rounded-lg bg-red-500/10 border border-red-500/20 text-red-400 text-xs">
                    {error}
                  </div>
                )}

                {/* Step 1: Calendar Slot Picker */}
                {!picked && (
                  <div className="space-y-4 flex-1">
                    <div className="flex items-center justify-between pb-3 border-b border-border">
                      <div className="flex items-center gap-2">
                        <CalendarIcon className="size-4 text-muted-foreground" />
                        <h3 className="font-semibold text-sm text-foreground">Select Date &amp; Time</h3>
                      </div>
                      <div className="flex gap-1.5">
                        <Button
                          variant="outline"
                          size="sm"
                          className="h-8 px-2.5 border-border bg-secondary/40 hover:bg-secondary/70 text-foreground"
                          disabled={offset === 0 || loading}
                          onClick={() => setOffset((v) => Math.max(0, v - 7))}
                        >
                          <ChevronLeftIcon className="size-3.5" />
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          className="h-8 px-2.5 border-border bg-secondary/40 hover:bg-secondary/70 text-foreground"
                          disabled={offset >= 56 || loading}
                          onClick={() => setOffset((v) => v + 7)}
                        >
                          <ChevronRightIcon className="size-3.5" />
                        </Button>
                      </div>
                    </div>

                    {loading ? (
                      <div className="py-20 text-center text-xs text-muted-foreground">
                        Checking calendar availability...
                      </div>
                    ) : slots.length ? (
                      <div className="grid grid-cols-2 sm:grid-cols-3 gap-2 max-h-[340px] overflow-y-auto pr-1">
                        {slots.map((slot) => (
                          <button
                            key={slot.start}
                            type="button"
                            className="p-2.5 rounded-lg border border-border bg-secondary/30 hover:bg-brand/10 hover:border-brand/40 text-left transition-all group"
                            onClick={() => setPicked(slot)}
                          >
                            <span className="block text-[11px] text-muted-foreground group-hover:text-foreground">
                              {new Date(slot.start).toLocaleDateString(undefined, {
                                weekday: "short",
                                month: "short",
                                day: "numeric",
                              })}
                            </span>
                            <span className="block text-xs font-semibold text-foreground mt-0.5">
                              {new Date(slot.start).toLocaleTimeString(undefined, {
                                hour: "numeric",
                                minute: "2-digit",
                              })}
                            </span>
                          </button>
                        ))}
                      </div>
                    ) : (
                      <div className="py-20 text-center text-xs text-muted-foreground">
                        No available slots for this period. Try the next week.
                      </div>
                    )}
                  </div>
                )}

                {/* Step 2: Custom Question Form */}
                {picked && (
                  <form onSubmit={book} className="space-y-3.5 flex-1">
                    <div className="flex items-center justify-between pb-3 border-b border-border">
                      <div>
                        <h3 className="font-semibold text-sm text-foreground">Your Booking Details</h3>
                        <p className="text-xs text-brand mt-0.5">
                          {new Date(picked.start).toLocaleString(undefined, {
                            dateStyle: "full",
                            timeStyle: "short",
                          })}
                        </p>
                      </div>
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        disabled={busy}
                        onClick={() => setPicked(null)}
                        className="text-xs text-muted-foreground hover:text-foreground"
                      >
                        Change Slot
                      </Button>
                    </div>

                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                      <div>
                        <Label htmlFor="guest-name" className="text-xs text-muted-foreground">Full Name *</Label>
                        <Input
                          id="guest-name"
                          required
                          value={name}
                          onChange={(e) => setName(e.target.value)}
                          placeholder="Your name"
                          maxLength={200}
                          className="bg-secondary/40 border-border text-sm h-9 text-foreground focus-visible:ring-ring"
                        />
                      </div>
                      <div>
                        <Label htmlFor="guest-email" className="text-xs text-muted-foreground">Email Address *</Label>
                        <Input
                          id="guest-email"
                          type="email"
                          required
                          value={email}
                          onChange={(e) => setEmail(e.target.value)}
                          placeholder="name@example.com"
                          className="bg-secondary/40 border-border text-sm h-9 text-foreground focus-visible:ring-ring"
                        />
                      </div>
                    </div>

                    {/* Custom Questions configured by host */}
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                      <div>
                        <Label htmlFor="company" className="text-xs text-muted-foreground">Company / Project</Label>
                        <Input
                          id="company"
                          value={company}
                          onChange={(e) => setCompany(e.target.value)}
                          placeholder="Company name"
                          className="bg-secondary/40 border-border text-sm h-9 text-foreground focus-visible:ring-ring"
                        />
                      </div>
                      <div>
                        <Label htmlFor="budget" className="text-xs text-muted-foreground">Estimated Budget</Label>
                        <select
                          id="budget"
                          value={budget}
                          onChange={(e) => setBudget(e.target.value)}
                          className="w-full bg-secondary/40 border border-border rounded-md px-3 py-1.5 text-sm h-9 text-foreground focus:outline-none focus:border-brand"
                        >
                          <option>$1,000 - $5,000</option>
                          <option>$5,000 - $15,000</option>
                          <option>$15,000 - $50,000</option>
                          <option>$50,000+</option>
                        </select>
                      </div>
                    </div>

                    <div>
                      <Label htmlFor="notes" className="text-xs text-muted-foreground">Discussion Topics &amp; Notes *</Label>
                      <textarea
                        id="notes"
                        required
                        rows={3}
                        value={notes}
                        onChange={(e) => setNotes(e.target.value)}
                        placeholder="Please share any topics or context for our meeting..."
                        className="w-full bg-secondary/40 border border-border rounded-md p-2.5 text-xs text-foreground focus:outline-none focus:border-brand placeholder:text-muted-foreground/60"
                        maxLength={5000}
                      />
                    </div>

                    <Button
                      type="submit"
                      disabled={busy}
                      className="w-full mt-1 h-9 font-medium bg-primary text-primary-foreground hover:opacity-90"
                    >
                      {busy ? "Confirming..." : "Confirm Booking"}
                    </Button>
                  </form>
                )}
              </div>
            </div>

            {/* Watermark Section (Shown ONLY on Free tier, completely hidden for Paid/Business tier) */}
            {showWatermark && (
              <footer className="py-2.5 px-4 bg-card border-t border-border/80 flex items-center justify-center text-[11px] text-muted-foreground">
                <a
                  href="https://halomail.zeroaxiis.tech"
                  target="_blank"
                  rel="noreferrer"
                  className="hover:text-foreground transition-colors inline-flex items-center gap-1"
                >
                  <span>Powered by</span>
                  <strong className="font-semibold text-foreground">HaloMail</strong>
                </a>
              </footer>
            )}
          </Card>
        )}
      </div>
    </div>
  );
}

export default function BookingPage() {
  return (
    <Suspense fallback={<div className="p-12 text-center text-xs text-zinc-500">Loading calendar...</div>}>
      <Booking />
    </Suspense>
  );
}
