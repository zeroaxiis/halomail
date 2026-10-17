"use client";

import Link from "next/link";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/dashboard/page-header";
import {
	EmptyState,
	LoadingRows,
	PanelHeading,
	ViewLink,
} from "@/components/dashboard/ui";
import { useWorkspace } from "@/components/dashboard/workspace-context";
import { useRpc } from "@/lib/use-rpc";

interface Message {
	id: string;
	senderName: string;
	senderEmail: string;
	createdAt: string;
	isSpam?: boolean;
	data?: Record<string, string>;
}
interface Key {
	id: string;
	scopes?: string[];
	revoked?: boolean;
}

export default function OverviewPage() {
	const { usage } = useWorkspace();
	const messages = useRpc<{ messages?: Message[] }>(
		"halomail.contact.v1.MessageService/ListMessages",
		{ page: { pageSize: 5 } },
	);
	const keys = useRpc<{ keys?: Key[] }>(
		"halomail.identity.v1.ApiKeyService/ListApiKeys",
	);
	const connections = useRpc<{ connections?: { id: string }[] }>(
		"halomail.scheduling.v1.CalendarService/ListConnections",
	);
	const stats = useRpc<{ upcomingBookings?: number }>(
		"halomail.scheduling.v1.BookingService/GetUsageStats",
	);
	const activeKeys = keys.data?.keys?.filter((key) => !key.revoked) || [];
	const hasForms = activeKeys.some((key) =>
		key.scopes?.includes("forms:submit"),
	);
	const hasMeetings = activeKeys.some((key) =>
		key.scopes?.includes("meetings:book"),
	);
	const hasCalendar = !!connections.data?.connections?.length;
	const date = new Date().toLocaleDateString(undefined, {
		weekday: "long",
		month: "long",
		day: "numeric",
	});
	const connectionItems = [
		{
			title: "Forms key",
			description: hasForms
				? "Accepting website submissions"
				: "Create a key to connect your form",
			done: hasForms,
			href: "/dashboard/forms",
			loading: keys.loading,
			error: keys.error,
		},
		{
			title: "Meetings key",
			description: hasMeetings
				? "Available for your booking link"
				: "Create a key for your booking link",
			done: hasMeetings,
			href: "/dashboard/meetings",
			loading: keys.loading,
			error: keys.error,
		},
		{
			title: "Google Calendar",
			description: hasCalendar
				? "Connected to your schedule"
				: "Connect your primary calendar",
			done: hasCalendar,
			href: "/dashboard/meetings",
			loading: connections.loading,
			error: connections.error,
		},
	];

	return (
		<>
			<PageHeader
				title="Overview"
				action={<time className="font-mono text-[11px]">{date}</time>}
			/>
			{usage.error && (
				<div
					className="dashboard-error flex flex-wrap items-center justify-between gap-3"
					role="alert"
				>
					<span>Usage could not be loaded. {usage.error}</span>
					<Button
						variant="outline"
						size="sm"
						onClick={() => void usage.reload()}
					>
						Retry
					</Button>
				</div>
			)}
			<div className="workspace-metrics">
				{(["forms", "meetings"] as const).map((feature) => {
					const allowance = usage.data?.[feature];
					return (
						<section className="workspace-metric" key={feature}>
							<h2 className="workspace-metric-label">
								{feature === "forms" ? "Form submissions" : "Meetings booked"}
							</h2>
							<p className="workspace-metric-value">
								{usage.loading
									? "…"
									: allowance
										? allowance.used.toLocaleString()
										: "—"}
							</p>
							<p className="workspace-metric-note">
								{allowance ? (
									<>{allowance.remaining.toLocaleString()} remaining</>
								) : usage.loading ? (
									"Loading allowance…"
								) : (
									"Usage unavailable"
								)}
							</p>
						</section>
					);
				})}
				<section className="workspace-metric">
					<h2 className="workspace-metric-label">Upcoming meetings</h2>
					<p className="workspace-metric-value">
						{stats.loading
							? "…"
							: stats.error
								? "—"
								: (stats.data?.upcomingBookings ?? 0)}
					</p>
					{stats.error ? (
						<button
							type="button"
							className="workspace-metric-note underline"
							onClick={() => void stats.reload()}
						>
							Unable to load · retry
						</button>
					) : (
						<Link
							href="/dashboard/meetings"
							className="workspace-metric-note underline underline-offset-4"
						>
							View schedule
						</Link>
					)}
				</section>
			</div>

			<div className="workspace-overview-grid">
				<section className="min-w-0">
					<PanelHeading
						title="Latest submissions"
						action={
							<ViewLink href="/dashboard/forms">All submissions</ViewLink>
						}
					/>
					<div className="workspace-table">
						<div className="workspace-table-labels workspace-label">
							<span>From</span>
							<span>Message</span>
							<span className="text-right">Received</span>
						</div>
						{messages.loading ? (
							<LoadingRows label="Loading submissions" />
						) : messages.error ? (
							<div className="py-6">
								<p role="alert" className="dashboard-error">
									{messages.error}
								</p>
								<Button
									variant="outline"
									size="sm"
									onClick={() => void messages.reload()}
								>
									Retry
								</Button>
							</div>
						) : !messages.data?.messages?.length ? (
							<EmptyState
								title="No submissions yet."
								description="Connect your first form to start receiving messages."
								action={
									<ViewLink href="/dashboard/forms">Set up forms</ViewLink>
								}
							/>
						) : (
							messages.data.messages.map((message) => (
								<Link
									key={message.id}
									href="/dashboard/forms"
									className="workspace-table-row"
								>
									<div>
										<p className="font-medium">
											{message.senderName ||
												message.senderEmail ||
												"Website visitor"}
										</p>
									</div>
									<div>
										<p className="text-muted-foreground">
											{message.data?.message || "Website submission"}
										</p>
										{message.isSpam && (
											<span className="workspace-status">
												Spam · not emailed
											</span>
										)}
									</div>
									<time
										dateTime={message.createdAt}
										className="text-right text-xs text-muted-foreground"
									>
										{new Date(message.createdAt).toLocaleDateString(undefined, {
											month: "short",
											day: "numeric",
										})}
									</time>
								</Link>
							))
						)}
					</div>
				</section>
				<aside className="workspace-rail">
					<PanelHeading title="Connections" />
					{connectionItems.map((item) => (
						<Link
							key={item.title}
							href={item.href}
							className="workspace-connection"
						>
							<span className="connection-name">
								{item.title}
								<span className="connection-status">
									{item.loading
										? "…"
										: item.error
											? "Unavailable"
											: item.done
												? "Connected"
												: "Set up"}
								</span>
							</span>
							{!item.loading && (item.error || !item.done) && (
								<p className="connection-note">
									{item.error ? "Open to retry" : item.description}
								</p>
							)}
						</Link>
					))}
				</aside>
			</div>
			<div className="workspace-shortcuts">
				<Link href="/dashboard/forms" className="workspace-shortcut">
					<span>
						Manage forms<span aria-hidden="true">↗</span>
					</span>
				</Link>
				<Link href="/dashboard/meetings" className="workspace-shortcut">
					<span>
						Manage meetings<span aria-hidden="true">↗</span>
					</span>
				</Link>
			</div>
		</>
	);
}
