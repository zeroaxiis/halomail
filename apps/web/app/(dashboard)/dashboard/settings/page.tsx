"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useWorkspace } from "@/components/dashboard/workspace-context";
import { PageHeader } from "@/components/dashboard/page-header";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select } from "@/components/ui/select";
import { rpc } from "@/lib/api";
import { getToken, clearSession } from "@/lib/auth";
import { localTimezone, timezones } from "@/lib/timezones";
import { useRpc } from "@/lib/use-rpc";
import { Trash2, AlertTriangle, LogOut } from "lucide-react";

interface User {
	id: string;
	email: string;
	name: string;
	handle: string;
	timezone: string;
	role?: string;
	scheduledDeleteAt?: { seconds: string } | string;
	deletionScheduledAt?: { seconds: string } | string;
}

export default function SettingsPage() {
	const router = useRouter();
	const { user, updateUser, usage } = useWorkspace();
	const {
		data,
		loading,
		error: loadError,
		reload,
	} = useRpc<{ user: User }>("halomail.identity.v1.AuthService/GetCurrentUser");
	const [name, setName] = useState("");
	const [handle, setHandle] = useState("");
	const [timezone, setTimezone] = useState("");
	const [error, setError] = useState("");
	const [saved, setSaved] = useState(false);
	const [busy, setBusy] = useState(false);
	const [origin, setOrigin] = useState("");

	const [showDeleteModal, setShowDeleteModal] = useState(false);
	const [deleteConfirmInput, setDeleteConfirmInput] = useState("");
	const [showLeaveModal, setShowLeaveModal] = useState(false);
	const [dangerBusy, setDangerBusy] = useState(false);
	const [dangerError, setDangerError] = useState("");

	const userHandle = (data?.user?.handle || handle || user?.handle || "").trim();
	const isDeleteConfirmed = userHandle
		? deleteConfirmInput.trim().toLowerCase() === userHandle.toLowerCase()
		: deleteConfirmInput.trim().toUpperCase() === "DELETE";

	const scheduledDeleteAt = data?.user?.scheduledDeleteAt;

	function formatDeletionDate(raw: { seconds: string } | string | undefined): string {
		if (!raw) return "in 24 hours";
		const ms = typeof raw === "object" && raw.seconds ? Number(raw.seconds) * 1000 : new Date(raw as string).getTime();
		return new Date(ms).toLocaleString(undefined, {
			month: "short",
			day: "numeric",
			hour: "2-digit",
			minute: "2-digit",
		});
	}

	async function handleScheduleDeletion() {
		if (!isDeleteConfirmed) return;
		setDangerBusy(true);
		setDangerError("");
		try {
			await rpc("halomail.identity.v1.UserService/ScheduleDeleteAccount");
			clearSession();
			router.push("/login");
		} catch (err) {
			setDangerError(err instanceof Error ? err.message : "Failed to schedule account deletion.");
			setDangerBusy(false);
		}
	}

	async function handleCancelDeletion() {
		setDangerBusy(true);
		setDangerError("");
		try {
			await rpc("halomail.identity.v1.UserService/CancelDeleteAccount");
			await reload();
		} catch (err) {
			setDangerError(err instanceof Error ? err.message : "Failed to cancel account deletion.");
		} finally {
			setDangerBusy(false);
		}
	}

	async function handleLeaveWorkspace() {
		setDangerBusy(true);
		setDangerError("");
		try {
			await rpc("halomail.identity.v1.TeamService/LeaveTeam");
			clearSession();
			router.push("/login");
		} catch (err) {
			setDangerError(err instanceof Error ? err.message : "Failed to leave workspace.");
			setDangerBusy(false);
		}
	}

	// Keep the saved value selectable even if it is not in the browser's list.
	const zones = useMemo(() => {
		const all = timezones();
		return timezone && !all.includes(timezone) ? [timezone, ...all] : all;
	}, [timezone]);

	useEffect(() => setOrigin(window.location.origin), []);
	useEffect(() => {
		if (data?.user) {
			setName(data.user.name);
			setHandle(data.user.handle);
			setTimezone(data.user.timezone);
		}
	}, [data]);

	async function save(e: React.FormEvent) {
		e.preventDefault();
		setBusy(true);
		setSaved(false);
		setError("");
		try {
			const res = await rpc<{ user: User }>(
				"halomail.identity.v1.UserService/UpdateUser",
				{ name, handle, timezone },
				getToken(),
			);
			updateUser({ ...user, ...res.user });
			setSaved(true);
		} catch (failure) {
			setError(
				failure instanceof Error
					? failure.message
					: "Could not save your profile. Please try again.",
			);
		} finally {
			setBusy(false);
		}
	}

	return (
		<>
			<PageHeader title="Settings" />
			<section className="workspace-settings-grid">
				<div>
					<h2>Profile</h2>
				</div>
				<form onSubmit={save} className="workspace-form">
					{loadError && (
						<div role="alert" className="dashboard-error">
							{loadError}
							<button
								type="button"
								className="ml-2 underline"
								onClick={() => void reload()}
							>
								Retry
							</button>
						</div>
					)}
					{error && (
						<p role="alert" className="dashboard-error">
							{error}
						</p>
					)}
					<fieldset
						disabled={loading || !!loadError || busy}
						className="workspace-form"
					>
						<div>
							<Label htmlFor="name">Full name</Label>
							<Input
								id="name"
								autoComplete="name"
								value={name}
								onChange={(e) => {
									setName(e.target.value);
									setSaved(false);
								}}
							/>
						</div>
						<div>
							<Label htmlFor="handle">Handle</Label>
							<Input
								id="handle"
								value={handle}
								onChange={(e) => {
									setHandle(e.target.value);
									setSaved(false);
								}}
							/>
							<p className="mt-2 break-all font-mono text-[10px] leading-6 text-muted-foreground">
								{origin}/book/{handle || "your-handle"}
							</p>
						</div>
						<div>
							<Label htmlFor="tz">Timezone</Label>
							<Select
								id="tz"
								value={timezone}
								onChange={(e) => {
									setTimezone(e.target.value);
									setSaved(false);
								}}
							>
								{zones.map((z) => (
									<option key={z} value={z}>
										{z}
									</option>
								))}
							</Select>
							<button
								type="button"
								onClick={() => {
									setTimezone(localTimezone());
									setSaved(false);
								}}
								className="mt-3 text-[11px] text-muted-foreground underline underline-offset-4"
							>
								Use my timezone ({localTimezone()})
							</button>
						</div>
					</fieldset>
					<div className="flex flex-wrap items-center gap-4">
						<Button
							type="submit"
							size="sm"
							disabled={busy || loading || !!loadError}
						>
							{busy ? "Saving…" : "Save changes"}
						</Button>
						{saved && (
							<span role="status" className="text-xs text-muted-foreground">
								Changes saved.
							</span>
						)}
					</div>
				</form>
			</section>
			<section className="workspace-settings-grid">
				<div>
					<h2>Account</h2>
				</div>
				<div>
					<div className="workspace-account-row">
						<span className="text-muted-foreground">Email address</span>
						<span className="break-all">{user.email}</span>
					</div>
					<div className="workspace-account-row">
						<span className="text-muted-foreground">Current plan</span>
						<span className="capitalize">
							{usage.data?.plan || (usage.loading ? "Loading…" : "Unavailable")}
						</span>
					</div>
					<div className="mt-5">
						<Link href="/dashboard/billing" className="workspace-text-link">
							Manage billing<span aria-hidden="true">↗</span>
						</Link>
					</div>
				</div>
			</section>
			<section className="workspace-settings-grid">
				<div>
					<h2>Booking availability</h2>
				</div>
				<div>
					<p className="!mt-0">
						Set booking hours and their timezone in Meetings.
					</p>
					<Link
						href="/dashboard/meetings#scheduling-setup"
						className="workspace-text-link mt-5"
					>
						Manage availability<span aria-hidden="true">↗</span>
					</Link>
				</div>
			</section>

			{/* Danger Zone */}
			<section className="workspace-settings-grid border-t border-destructive/20 pt-8 mt-8">
				<div>
					<h2 className="text-destructive font-semibold">Danger Zone</h2>
					<p className="text-xs text-muted-foreground mt-1">
						Irreversible actions and workspace departure.
					</p>
				</div>

				<div className="space-y-6">
					{/* Active Deletion Banner if scheduled */}
					{scheduledDeleteAt ? (
						<div className="rounded-lg border border-amber-500/30 bg-amber-500/10 p-4 space-y-3">
							<div className="flex items-start gap-2.5">
								<AlertTriangle className="size-5 text-amber-500 shrink-0 mt-0.5" />
								<div className="space-y-1">
									<h3 className="text-sm font-semibold text-amber-600 dark:text-amber-400">
										Account Deletion Scheduled
									</h3>
									<p className="text-xs text-muted-foreground leading-relaxed">
										Your account is scheduled for permanent deletion on{" "}
										<strong>{formatDeletionDate(scheduledDeleteAt)}</strong> (within 24 hours).
										After this window expires, all your account data will be permanently wiped and your email will be freed.
									</p>
								</div>
							</div>
							<div className="flex items-center gap-3 pt-1">
								<Button
									size="sm"
									variant="outline"
									disabled={dangerBusy}
									onClick={handleCancelDeletion}
									className="text-xs border-amber-500/40 hover:bg-amber-500/20"
								>
									{dangerBusy ? "Cancelling…" : "Cancel Deletion Request"}
								</Button>
							</div>
						</div>
					) : (
						/* Normal Delete Account Card */
						<div className="flex flex-col sm:flex-row sm:items-center justify-between p-4 rounded-lg border border-destructive/20 bg-destructive/5 gap-4">
							<div className="space-y-0.5">
								<h3 className="text-sm font-medium text-foreground">Delete Account</h3>
								<p className="text-xs text-muted-foreground max-w-md">
									Schedule permanent account deletion with a 24-hour grace period. You can cancel at any time within 24 hours. After 24 hours, all data is permanently purged and your email is freed.
								</p>
							</div>
							<Button
								type="button"
								variant="destructive"
								size="sm"
								onClick={() => {
									setDeleteConfirmInput("");
									setDangerError("");
									setShowDeleteModal(true);
								}}
								className="shrink-0 gap-1.5"
							>
								<Trash2 className="size-3.5" />
								<span>Delete Account</span>
							</Button>
						</div>
					)}

					{/* Leave Workspace Option if not owner */}
					{data?.user?.role && data.user.role !== "owner" && (
						<div className="flex flex-col sm:flex-row sm:items-center justify-between p-4 rounded-lg border border-border bg-card gap-4">
							<div className="space-y-0.5">
								<h3 className="text-sm font-medium text-foreground">Leave Team Workspace</h3>
								<p className="text-xs text-muted-foreground max-w-md">
									Revoke your membership from this workspace immediately. This frees your email so you can create your own workspace or accept another invitation.
								</p>
							</div>
							<Button
								type="button"
								variant="outline"
								size="sm"
								onClick={() => setShowLeaveModal(true)}
								className="shrink-0 gap-1.5 text-muted-foreground hover:text-destructive hover:border-destructive/40"
							>
								<LogOut className="size-3.5" />
								<span>Leave Workspace</span>
							</Button>
						</div>
					)}
				</div>
			</section>

			{/* Delete Account Modal */}
			{showDeleteModal && (
				<div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4 animate-in fade-in-50">
					<div className="w-full max-w-md rounded-lg border border-border bg-card p-6 shadow-2xl space-y-4">
						<div className="flex items-start gap-3">
							<div className="flex size-10 rounded-full bg-destructive/10 text-destructive items-center justify-center shrink-0">
								<AlertTriangle className="size-5" />
							</div>
							<div className="space-y-1">
								<h3 className="text-lg font-semibold text-foreground">Delete Account (24-Hour Delay)</h3>
								<p className="text-xs text-muted-foreground leading-relaxed">
									Scheduling deletion begins a <strong>24-hour grace period</strong>. Your active sessions will be signed out.
								</p>
							</div>
						</div>

						<div className="rounded-md border bg-muted/40 p-3.5 text-xs text-muted-foreground space-y-2">
							<p className="text-foreground font-medium">How 24-hour deletion works:</p>
							<ul className="list-disc list-inside space-y-1">
								<li>Within 24 hours: Log in to cancel deletion and keep your account.</li>
								<li>After 24 hours: All meetings, templates, forms, and data are permanently purged.</li>
								<li>Your email is released, allowing you to accept invitations or register fresh.</li>
							</ul>
						</div>

						<div className="space-y-2 pt-1">
							<Label htmlFor="delete-confirm-handle" className="text-xs text-muted-foreground font-normal">
								To confirm deletion, please type your handle{" "}
								<strong className="text-foreground font-semibold">
									{userHandle || "DELETE"}
								</strong>{" "}
								below:
							</Label>
							<Input
								id="delete-confirm-handle"
								type="text"
								autoComplete="off"
								spellCheck={false}
								value={deleteConfirmInput}
								onChange={(e) => setDeleteConfirmInput(e.target.value)}
								placeholder={userHandle || "DELETE"}
								className="h-9 text-xs"
							/>
						</div>

						{dangerError && (
							<p className="text-xs text-destructive font-medium">{dangerError}</p>
						)}

						<div className="flex justify-end gap-2 pt-2">
							<Button
								type="button"
								variant="outline"
								size="sm"
								disabled={dangerBusy}
								onClick={() => {
									setShowDeleteModal(false);
									setDeleteConfirmInput("");
									setDangerError("");
								}}
							>
								Cancel
							</Button>
							<Button
								type="button"
								variant="destructive"
								size="sm"
								disabled={dangerBusy || !isDeleteConfirmed}
								onClick={handleScheduleDeletion}
							>
								{dangerBusy ? "Scheduling…" : "Schedule 24h Deletion"}
							</Button>
						</div>
					</div>
				</div>
			)}

			{/* Leave Workspace Modal */}
			{showLeaveModal && (
				<div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4 animate-in fade-in-50">
					<div className="w-full max-w-md rounded-lg border border-border bg-card p-6 shadow-2xl space-y-4">
						<div className="flex items-start gap-3">
							<div className="flex size-10 rounded-full bg-amber-500/10 text-amber-500 items-center justify-center shrink-0">
								<LogOut className="size-5" />
							</div>
							<div className="space-y-1">
								<h3 className="text-lg font-semibold text-foreground">Leave Workspace</h3>
								<p className="text-xs text-muted-foreground leading-relaxed">
									Are you sure you want to leave this workspace? You will lose access to team meetings, forms, and templates immediately.
								</p>
							</div>
						</div>

						<p className="text-xs text-muted-foreground">
							After leaving, your email will no longer be bound to this workspace, allowing you to create your own account or accept an invitation to another team.
						</p>

						{dangerError && (
							<p className="text-xs text-destructive font-medium">{dangerError}</p>
						)}

						<div className="flex justify-end gap-2 pt-2">
							<Button
								type="button"
								variant="outline"
								size="sm"
								disabled={dangerBusy}
								onClick={() => setShowLeaveModal(false)}
							>
								Stay in Workspace
							</Button>
							<Button
								type="button"
								variant="destructive"
								size="sm"
								disabled={dangerBusy}
								onClick={handleLeaveWorkspace}
							>
								{dangerBusy ? "Leaving…" : "Leave Workspace"}
							</Button>
						</div>
					</div>
				</div>
			)}
		</>
	);
}
