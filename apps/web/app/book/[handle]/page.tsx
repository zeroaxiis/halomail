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

function Booking() {
  const params = useParams();
  const searchParams = useSearchParams();

  const handle = (params?.handle as string) || "";
  const key = searchParams.get("key") || "";
  const isEmbed = searchParams.get("embed") === "true";
  const themeParam = searchParams.get("theme") || "obsidian";

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

  // Theme configuration
  const themeClasses: Record<string, string> = {
    obsidian: "bg-[#0b0914] text-[#f4f1fa] border-purple-500/20",
    classic: "bg-[#0a1120] text-[#f0f4f8] border-indigo-500/20",
    neon: "bg-[#050e14] text-[#eefcfd] border-cyan-500/20",
    minimal: "bg-[#18181b] text-[#fafafa] border-white/20",
  };
  const themeClass = themeClasses[themeParam] || themeClasses.obsidian;

  return (
    <div
      className={`min-h-screen ${isEmbed ? "p-4" : "p-6 sm:p-12"} flex flex-col items-center justify-between transition-colors`}
      style={{
        backgroundColor: isEmbed ? "transparent" : undefined,
      }}
    >
      <div className={`w-full max-w-4xl mx-auto ${isEmbed ? "my-0" : "my-6"} space-y-6`}>
        {!isEmbed && (
          <header className="flex items-center justify-between pb-4 border-b border-white/10">
            <Link href="/" className="inline-block">
              <Logo />
            </Link>
            <span className="text-xs text-muted-foreground font-mono">
              cal.com-style booking
            </span>
          </header>
        )}

        {/* Done / Confirmed View */}
        {done ? (
          <Card className="p-8 sm:p-12 text-center space-y-6 bg-card/90 border-white/10 backdrop-blur-xl shadow-2xl rounded-2xl">
            <div className="size-16 rounded-full bg-emerald-500/10 border border-emerald-500/30 text-emerald-400 text-3xl flex items-center justify-center mx-auto">
              ✓
            </div>
            <div className="space-y-2">
              <h1 className="text-2xl sm:text-3xl font-bold tracking-tight">
                Meeting Confirmed!
              </h1>
              <p className="text-muted-foreground text-sm max-w-md mx-auto">
                Your consultation with{" "}
                <strong className="text-foreground">{info?.owner.name || handle}</strong>{" "}
                is officially scheduled.
              </p>
            </div>

            <div className="bg-white/[0.03] border border-white/10 rounded-xl p-5 text-left max-w-md mx-auto text-sm space-y-3">
              <div className="flex justify-between py-1 border-b border-white/5">
                <span className="text-muted-foreground">Date &amp; Time:</span>
                <span className="font-medium">
                  {confirmedBooking?.start
                    ? new Date(confirmedBooking.start).toLocaleString(undefined, {
                        dateStyle: "full",
                        timeStyle: "short",
                      })
                    : ""}
                </span>
              </div>
              <div className="flex justify-between py-1 border-b border-white/5">
                <span className="text-muted-foreground">Attendee:</span>
                <span className="font-medium">{email}</span>
              </div>
              <div className="flex justify-between py-1 border-b border-white/5">
                <span className="text-muted-foreground">Conferencing:</span>
                <span className="text-primary font-medium">Google Meet</span>
              </div>
              <div className="flex justify-between py-1">
                <span className="text-muted-foreground">Calendar:</span>
                <span className="text-emerald-400 font-medium">Added to Google Calendar</span>
              </div>
            </div>

            <p className="text-xs text-muted-foreground">
              A calendar invitation with the Google Meet link and full agenda has been sent to both your email and the host.
            </p>

            <Button
              variant="outline"
              onClick={() => {
                setDone(false);
                setPicked(null);
              }}
            >
              Book Another Slot
            </Button>
          </Card>
        ) : (
          /* Main Cal.com Booking Card */
          <Card className={`overflow-hidden border rounded-2xl shadow-2xl backdrop-blur-xl ${themeClass}`}>
            <div className="grid grid-cols-1 md:grid-cols-12 min-h-[520px]">
              {/* Left Column: Host & Event Profile */}
              <div className="md:col-span-5 p-6 sm:p-8 border-b md:border-b-0 md:border-r border-white/10 flex flex-col justify-between">
                <div className="space-y-6">
                  {/* Host Header */}
                  <div className="flex items-center gap-4">
                    <div className="size-14 rounded-full bg-gradient-to-tr from-primary to-indigo-600 flex items-center justify-center text-white font-bold text-lg shadow-md">
                      {info?.owner?.avatar ? (
                        <img
                          src={info.owner.avatar}
                          alt={info.owner.name}
                          className="size-full rounded-full object-cover"
                        />
                      ) : (
                        (info?.owner.name || handle).slice(0, 2).toUpperCase()
                      )}
                    </div>
                    <div>
                      <h2 className="font-semibold text-lg leading-tight">
                        {info?.owner.name || handle}
                      </h2>
                      <span className="text-xs text-muted-foreground font-mono">
                        @{info?.owner.handle || handle}
                      </span>
                    </div>
                  </div>

                  {/* Active Event Types */}
                  <div className="space-y-2">
                    <span className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                      Event Types
                    </span>
                    <div className="space-y-2">
                      {info?.events?.map((ev) => (
                        <button
                          key={ev.id}
                          type="button"
                          onClick={() => setEventID(ev.id)}
                          className={`w-full text-left p-3 rounded-xl border text-sm transition-all ${
                            eventID === ev.id
                              ? "border-primary bg-primary/10 shadow-sm"
                              : "border-white/5 hover:border-white/20 bg-white/[0.02]"
                          }`}
                        >
                          <div className="font-medium flex justify-between">
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

                  {/* Selected Event Details */}
                  {selectedEvent && (
                    <div className="pt-4 border-t border-white/10 space-y-2 text-xs text-muted-foreground">
                      <div className="flex items-center gap-2">
                        <span>⏱</span>
                        <span>{selectedEvent.durationMinutes} Minutes Consultation</span>
                      </div>
                      <div className="flex items-center gap-2">
                        <span>📹</span>
                        <span>Google Meet Video Conferencing</span>
                      </div>
                    </div>
                  )}
                </div>

                {/* Timezone picker info */}
                <div className="pt-6 border-t border-white/10 text-xs text-muted-foreground">
                  <span>🌐 Timezone: </span>
                  <strong className="text-foreground">{timezone}</strong>
                </div>
              </div>

              {/* Right Column: Dynamic Scheduler / Custom Questions */}
              <div className="md:col-span-7 p-6 sm:p-8 flex flex-col justify-between">
                {error && (
                  <div className="mb-4 p-3 rounded-lg bg-destructive/10 border border-destructive/20 text-destructive text-sm">
                    {error}
                  </div>
                )}

                {/* Step 1: Slot Picker */}
                {!picked && (
                  <div className="space-y-5 flex-1">
                    <div className="flex items-center justify-between pb-3 border-b border-white/10">
                      <h3 className="font-semibold text-base">Select Date &amp; Time</h3>
                      <div className="flex gap-2">
                        <Button
                          variant="outline"
                          size="sm"
                          disabled={offset === 0 || loading}
                          onClick={() => setOffset((v) => Math.max(0, v - 7))}
                        >
                          ‹ Prev Week
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          disabled={offset >= 56 || loading}
                          onClick={() => setOffset((v) => v + 7)}
                        >
                          Next Week ›
                        </Button>
                      </div>
                    </div>

                    {loading ? (
                      <div className="py-20 text-center text-sm text-muted-foreground">
                        Checking calendar availability…
                      </div>
                    ) : slots.length ? (
                      <div className="grid grid-cols-2 sm:grid-cols-3 gap-2 max-h-[360px] overflow-y-auto pr-1">
                        {slots.map((slot) => (
                          <Button
                            key={slot.start}
                            variant="outline"
                            className="h-auto py-2.5 flex flex-col items-center border-white/10 hover:border-primary hover:bg-primary/10 transition-all"
                            onClick={() => setPicked(slot)}
                          >
                            <span className="text-[11px] text-muted-foreground">
                              {new Date(slot.start).toLocaleDateString(undefined, {
                                weekday: "short",
                                month: "short",
                                day: "numeric",
                              })}
                            </span>
                            <span className="text-xs font-semibold mt-0.5">
                              {new Date(slot.start).toLocaleTimeString(undefined, {
                                hour: "numeric",
                                minute: "2-digit",
                              })}
                            </span>
                          </Button>
                        ))}
                      </div>
                    ) : (
                      <div className="py-20 text-center text-sm text-muted-foreground">
                        No available slots for this week. Try advancing to next week.
                      </div>
                    )}
                  </div>
                )}

                {/* Step 2: Custom Questions Form */}
                {picked && (
                  <form onSubmit={book} className="space-y-4 flex-1">
                    <div className="flex items-center justify-between pb-3 border-b border-white/10">
                      <div>
                        <h3 className="font-semibold text-base">Your Booking Details</h3>
                        <p className="text-xs text-primary font-medium">
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
                      >
                        Change Time
                      </Button>
                    </div>

                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                      <div>
                        <Label htmlFor="guest-name">Full Name *</Label>
                        <Input
                          id="guest-name"
                          required
                          value={name}
                          onChange={(e) => setName(e.target.value)}
                          placeholder="Jane Doe"
                          maxLength={200}
                        />
                      </div>
                      <div>
                        <Label htmlFor="guest-email">Email Address *</Label>
                        <Input
                          id="guest-email"
                          type="email"
                          required
                          value={email}
                          onChange={(e) => setEmail(e.target.value)}
                          placeholder="jane@example.com"
                        />
                      </div>
                    </div>

                    {/* Custom Questions configured by host */}
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                      <div>
                        <Label htmlFor="company">Company / Organization</Label>
                        <Input
                          id="company"
                          value={company}
                          onChange={(e) => setCompany(e.target.value)}
                          placeholder="Acme Inc."
                        />
                      </div>
                      <div>
                        <Label htmlFor="budget">Project Budget</Label>
                        <select
                          id="budget"
                          value={budget}
                          onChange={(e) => setBudget(e.target.value)}
                          className="w-full bg-background border border-input rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-1 focus:ring-primary"
                        >
                          <option>$1,000 - $5,000</option>
                          <option>$5,000 - $15,000</option>
                          <option>$15,000 - $50,000</option>
                          <option>$50,000+</option>
                        </select>
                      </div>
                    </div>

                    <div>
                      <Label htmlFor="notes">What would you like to discuss? *</Label>
                      <textarea
                        id="notes"
                        required
                        rows={3}
                        value={notes}
                        onChange={(e) => setNotes(e.target.value)}
                        placeholder="Please share any topics or context for our call..."
                        className="w-full bg-background border border-input rounded-md p-3 text-sm focus:outline-none focus:ring-1 focus:ring-primary"
                        maxLength={5000}
                      />
                    </div>

                    <Button
                      type="submit"
                      disabled={busy}
                      className="w-full mt-2 font-semibold"
                    >
                      {busy ? "Scheduling Meeting…" : "Confirm Booking"}
                    </Button>
                  </form>
                )}
              </div>
            </div>

            {/* Watermark Section (Shown on Free tier, completely hidden on Pro/Paid) */}
            {showWatermark && (
              <footer className="py-3 px-6 bg-black/30 border-t border-white/5 flex items-center justify-center text-xs text-muted-foreground">
                <a
                  href="https://halomail.zeroaxiis.tech"
                  target="_blank"
                  rel="noreferrer"
                  className="hover:text-primary transition-colors inline-flex items-center gap-1.5"
                >
                  ⚡ Powered by <strong className="font-semibold text-foreground">HaloMail</strong>
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
    <Suspense fallback={<div className="p-12 text-center text-muted-foreground">Loading calendar…</div>}>
      <Booking />
    </Suspense>
  );
}
