"use client";

import { useEffect, useRef, useState } from "react";

import {
	EmptyState,
	LoadingRows,
	PanelHeading,
	RefreshButton,
} from "@/components/dashboard/ui";
import { FeatureAccess } from "@/components/dashboard/feature-access";
import { PageHeader } from "@/components/dashboard/page-header";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select } from "@/components/ui/select";
import { rpc } from "@/lib/api";
import { localTimezone, timezones } from "@/lib/timezones";
import { useRpc } from "@/lib/use-rpc";

interface Booking {
	id: string;
	inviteeName: string;
	inviteeEmail: string;
	start: string;
	status: string;
	location?: string;
}
interface EventType {
	id: string;
	title: string;
	durationMinutes: number;
}
interface Rule {
	weekday: number;
	startMinute: number;
	endMinute: number;
}

export default function MeetingsPage() {
	const bookings = useRpc<{ bookings?: Booking[] }>(
		"halomail.scheduling.v1.BookingService/ListBookings",
	);
	const stats = useRpc<{
		totalBookings?: number;
		upcomingBookings?: number;
		cancelledBookings?: number;
	}>("halomail.scheduling.v1.BookingService/GetUsageStats");
	const events = useRpc<{ eventTypes?: EventType[] }>(
		"halomail.scheduling.v1.EventTypeService/ListEventTypes",
	);
	const connections = useRpc<{ connections?: { id: string; email: string }[] }>(
		"halomail.scheduling.v1.CalendarService/ListConnections",
	);
	const availability = useRpc<{
		availability: { timezone: string; rules?: Rule[]; overrides?: unknown[] };
	}>("halomail.scheduling.v1.AvailabilityService/GetAvailability");
	const [title, setTitle] = useState("Introductory call");
	const [duration, setDuration] = useState(30);
	const [timezone, setTimezone] = useState("UTC");
	const [rules, setRules] = useState<Rule[]>(
		[1, 2, 3, 4, 5].map((weekday) => ({
			weekday,
			startMinute: 540,
			endMinute: 1020,
		})),
	);
	const [busy, setBusy] = useState(false);
	const [notice, setNotice] = useState("");
	const [error, setError] = useState("");
	const [refreshTrigger, setRefreshTrigger] = useState(0);
	const [setupOpen, setSetupOpen] = useState(false);
	const setupRef = useRef<HTMLDetailsElement>(null);

	useEffect(() => {
		function openLinkedSetup() {
			if (window.location.hash === "#scheduling-setup") setSetupOpen(true);
		}
		openLinkedSetup();
		window.addEventListener("hashchange", openLinkedSetup);
		return () => window.removeEventListener("hashchange", openLinkedSetup);
	}, []);

	useEffect(() => {
		setTimezone(localTimezone());
		const result = new URLSearchParams(window.location.search).get("calendar");
		if (result === "connected") setNotice("Google Calendar connected.");
		if (result === "failed")
			setError("Calendar connection failed or expired. Please connect again.");
	}, []);
	useEffect(() => {
		if (availability.data?.availability) {
			setTimezone(availability.data.availability.timezone || localTimezone());
			setRules(
				(availability.data.availability.rules || []).map((rule) => ({
					weekday: rule.weekday ?? 0,
					startMinute: rule.startMinute ?? 0,
					endMinute: rule.endMinute ?? 0,
				})),
			);
		}
	}, [availability.data]);

	async function perform(action: () => Promise<void>) {
		setBusy(true);
		setError("");
		setNotice("");
		try {
			await action();
		} catch (failure) {
			setError(failure instanceof Error ? failure.message : "Request failed");
		} finally {
			setBusy(false);
		}
	}

	function connect() {
		void perform(async () => {
			const result = await rpc<{ authorizationUrl: string }>(
				"halomail.scheduling.v1.CalendarService/StartConnect",
				{ provider: "CALENDAR_PROVIDER_GOOGLE" },
			);
			window.location.assign(result.authorizationUrl);
		});
	}

	function createEvent(event: React.FormEvent) {
		event.preventDefault();
		void perform(async () => {
			await rpc("halomail.scheduling.v1.EventTypeService/CreateEventType", {
				title,
				durationMinutes: duration,
				locationKind: "LOCATION_KIND_GOOGLE_MEET",
				description: "Tell us what you would like to discuss when booking.",
			});
			await events.reload();
			setNotice("Event type created.");
		});
	}

	function saveAvailability(event: React.FormEvent) {
		event.preventDefault();
		void perform(async () => {
			await rpc("halomail.scheduling.v1.AvailabilityService/SetAvailability", {
				timezone,
				rules,
				overrides: availability.data?.availability.overrides || [],
			});
			await availability.reload();
			setNotice("Availability saved.");
		});
	}

	function changeRule(weekday: number, update: Partial<Rule>) {
		setRules((current) =>
			current.map((rule) =>
				rule.weekday === weekday ? { ...rule, ...update } : rule,
			),
		);
	}

	return (
		<>
			<PageHeader
				title="Meetings"
				action={
					<button
						type="button"
						className="workspace-text-link"
						aria-controls="scheduling-setup"
						aria-expanded={setupOpen}
						onClick={() => {
							setSetupOpen(true);
							setupRef.current?.scrollIntoView({ block: "start" });
							setupRef.current?.querySelector("summary")?.focus();
						}}
					>
						Scheduling setup <span aria-hidden="true">↓</span>
					</button>
				}
			/>
			<div className="workspace-metrics">
				{[
					["Total bookings", stats.data?.totalBookings],
					["Upcoming", stats.data?.upcomingBookings],
					["Cancelled", stats.data?.cancelledBookings],
				].map(([label, value]) => (
					<section className="workspace-metric" key={label}>
						<h2 className="workspace-metric-label">{label}</h2>
						<p className="workspace-metric-value">
							{stats.loading ? "…" : stats.error ? "—" : (value ?? 0)}
						</p>
					</section>
				))}
			</div>
			{(error ||
				stats.error ||
				events.error ||
				connections.error ||
				availability.error) && (
				<p role="alert" className="dashboard-error">
					{error ||
						stats.error ||
						events.error ||
						connections.error ||
						availability.error}
				</p>
			)}
			{notice && (
				<p role="status" className="dashboard-notice">
					{notice}
				</p>
			)}
			<div className="workspace-content-grid">
				<section className="min-w-0">
					<PanelHeading
						title="Recent bookings"
						action={
							<RefreshButton
								loading={bookings.loading || stats.loading}
								onClick={() => {
									void bookings.reload();
									void stats.reload();
									setRefreshTrigger((t) => t + 1);
								}}
							/>
						}
					/>
					<div className="workspace-table">
						{bookings.loading ? (
							<LoadingRows label="Loading bookings" />
						) : bookings.error ? (
							<p role="alert" className="dashboard-error">
								{bookings.error}
							</p>
						) : !bookings.data?.bookings?.length ? (
							<EmptyState
								title="No bookings yet."
								description="Open scheduling setup to connect your calendar and set your hours."
							/>
						) : (
							bookings.data.bookings.map((booking) => (
								<article key={booking.id} className="workspace-booking">
									<time
										dateTime={booking.start}
										className="workspace-booking-date"
									>
										{new Date(booking.start).toLocaleDateString(undefined, {
											month: "short",
											day: "numeric",
										})}
										<span>
											{new Date(booking.start).toLocaleTimeString(undefined, {
												hour: "numeric",
												minute: "2-digit",
											})}
										</span>
									</time>
									<div className="min-w-0">
										<h3 className="workspace-booking-name">
											{booking.inviteeName}
										</h3>
										<p className="workspace-booking-email">
											{booking.inviteeEmail}
										</p>
										<span
											className={`workspace-status mt-2 ${booking.status === "BOOKING_STATUS_CONFIRMED" ? "confirmed" : ""}`}
										>
											{booking.status
												.replace("BOOKING_STATUS_", "")
												.replaceAll("_", " ")
												.toLowerCase()}
										</span>
									</div>
									<div>
										{booking.location?.startsWith(
											"https://meet.google.com/",
										) ? (
											<Button variant="outline" size="sm" asChild>
												<a
													href={booking.location}
													target="_blank"
													rel="noreferrer"
												>
													Join meeting
												</a>
											</Button>
										) : (
											booking.status !== "BOOKING_STATUS_CANCELLED" && (
												<p className="max-w-[125px] text-[10px] leading-5 text-muted-foreground">
													Calendar confirmation pending
												</p>
											)
										)}
									</div>
								</article>
							))
						)}
					</div>
				</section>
				<aside className="workspace-rail">
					<FeatureAccess feature="meetings" refreshTrigger={refreshTrigger} />
				</aside>
			</div>

			<details
				ref={setupRef}
				id="scheduling-setup"
				className="workspace-setup"
				open={setupOpen}
				onToggle={(event) => setSetupOpen(event.currentTarget.open)}
			>
				<summary>Scheduling setup</summary>
				<div className="workspace-setup-grid">
					<div className="min-w-0">
						<section className="workspace-settings-section">
							<div className="mb-4 flex flex-wrap items-center justify-between gap-3">
								<h3 className="text-sm font-medium">Google Calendar</h3>
								<span className="workspace-status">
									{connections.loading
										? "Checking…"
										: connections.error
											? "Unavailable"
											: connections.data?.connections?.length
												? "Connected"
												: "Not connected"}
								</span>
							</div>
							<p className="mb-5 text-xs text-muted-foreground">
								{connections.data?.connections?.[0]?.email ||
									"Connect your calendar to check for conflicts."}
							</p>
							<Button
								variant="outline"
								size="sm"
								onClick={connect}
								disabled={busy || connections.loading}
							>
								{connections.data?.connections?.length
									? "Reconnect Google Calendar"
									: "Connect Google Calendar"}
							</Button>
						</section>
						<section className="workspace-settings-section">
							<h3 className="mb-5 text-sm font-medium">Event types</h3>
							{events.loading ? (
								<p className="mb-5 text-xs text-muted-foreground">
									Loading event types…
								</p>
							) : (
								!!events.data?.eventTypes?.length && (
									<ul className="mb-6">
										{events.data.eventTypes.map((event) => (
											<li
												key={event.id}
												className="flex items-center justify-between gap-4 py-3 text-xs"
											>
												<span className="min-w-0 truncate">{event.title}</span>
												<span className="shrink-0 font-mono text-[10px] text-muted-foreground">
													{event.durationMinutes} min
												</span>
											</li>
										))}
									</ul>
								)
							)}
							<form onSubmit={createEvent} className="workspace-form">
								<div className="grid gap-4 sm:grid-cols-[1fr_110px]">
									<div>
										<Label htmlFor="event-title">Meeting title</Label>
										<Input
											id="event-title"
											required
											maxLength={200}
											value={title}
											onChange={(event) => setTitle(event.target.value)}
										/>
									</div>
									<div>
										<Label htmlFor="duration">Minutes</Label>
										<Input
											id="duration"
											type="number"
											min={5}
											max={480}
											required
											value={duration}
											onChange={(event) =>
												setDuration(Number(event.target.value))
											}
										/>
									</div>
								</div>
								<div>
									<Button
										variant="outline"
										size="sm"
										disabled={busy}
										type="submit"
									>
										Create event type
									</Button>
								</div>
							</form>
						</section>
					</div>
					<section className="workspace-settings-section">
						<h3 className="mb-5 text-sm font-medium">Weekly availability</h3>
						<form onSubmit={saveAvailability} className="workspace-form">
							<div>
								<Label htmlFor="meeting-timezone">Timezone</Label>
								<Select
									id="meeting-timezone"
									value={timezone}
									disabled={availability.loading}
									onChange={(event) => setTimezone(event.target.value)}
								>
									{Array.from(new Set([timezone, ...timezones()])).map(
										(zone) => (
											<option key={zone}>{zone}</option>
										),
									)}
								</Select>
							</div>
							{availability.loading ? (
								<LoadingRows label="Loading availability" />
							) : (
								<div>
									{[
										"Sunday",
										"Monday",
										"Tuesday",
										"Wednesday",
										"Thursday",
										"Friday",
										"Saturday",
									].map((day, weekday) => {
										const rule = rules.find(
											(candidate) => candidate.weekday === weekday,
										);
										return (
											<div key={day} className="workspace-schedule-day">
												<label>
													<input
														className="size-3.5"
														type="checkbox"
														checked={!!rule}
														disabled={!!availability.error}
														onChange={(event) =>
															setRules((current) =>
																event.target.checked
																	? [
																			...current,
																			{
																				weekday,
																				startMinute: 540,
																				endMinute: 1020,
																			},
																		]
																	: current.filter(
																			(candidate) =>
																				candidate.weekday !== weekday,
																		),
															)
														}
													/>
													{day}
												</label>
												{rule ? (
													<div className="flex items-center gap-2">
														<input
															aria-label={day + " start"}
															type="time"
															required
															value={clockTime(rule.startMinute)}
															onChange={(event) =>
																changeRule(weekday, {
																	startMinute: minutes(event.target.value),
																})
															}
														/>
														<span className="text-muted-foreground">–</span>
														<input
															aria-label={day + " end"}
															type="time"
															required
															value={clockTime(rule.endMinute)}
															onChange={(event) =>
																changeRule(weekday, {
																	endMinute: minutes(event.target.value),
																})
															}
														/>
													</div>
												) : (
													<span className="text-[11px] text-muted-foreground">
														Unavailable
													</span>
												)}
											</div>
										);
									})}
								</div>
							)}
							<div>
								<Button
									disabled={
										busy || availability.loading || !!availability.error
									}
									size="sm"
									type="submit"
								>
									{busy ? "Saving…" : "Save availability"}
								</Button>
								<p className="mt-4 text-[11px] leading-6 text-muted-foreground">
									Calendar conflicts are excluded automatically.
								</p>
							</div>
						</form>
					</section>
				</div>
			</details>
		</>
	);
}

function clockTime(value: number) {
	return (
		String(Math.floor(value / 60)).padStart(2, "0") +
		":" +
		String(value % 60).padStart(2, "0")
	);
}
function minutes(value: string) {
	const [hours, remainder] = value.split(":").map(Number);
	return hours * 60 + remainder;
}
