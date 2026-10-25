"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import {
	UserPlus,
	Mail,
	Trash2,
	Copy,
	Check,
	AlertCircle,
	ArrowUpRight,
	Clock,
	X,
	ChevronDown,
} from "lucide-react";
import { PageHeader } from "@/components/dashboard/page-header";
import { PanelHeading } from "@/components/dashboard/ui";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { rpc } from "@/lib/api";
import { useRpc } from "@/lib/use-rpc";
import { useWorkspace } from "@/components/dashboard/workspace-context";

interface TeamMember {
	id: string;
	email: string;
	name: string;
	handle: string;
	avatarUrl: string;
	role: "owner" | "admin" | "manager" | "viewer";
	createdAt: string;
}

interface Invitation {
	id: string;
	email: string;
	role: "admin" | "manager" | "viewer";
	invitedByName: string;
	status: string;
	expiresAt: string;
	createdAt: string;
	token?: string;
}

interface TeamResponse {
	members: TeamMember[];
	invitations: Invitation[];
	maxSeats: number;
	usedSeats: number;
	tier: string;
	currentUserRole: string;
}

const ROLE_DETAILS = {
	owner: {
		label: "Owner",
		badge: "brand" as const,
		desc: "Full workspace ownership, subscription billing, seat upgrades, and workspace transfer permissions.",
	},
	admin: {
		label: "Administrator",
		badge: "default" as const,
		desc: "Manage team members, configure workspace settings, forms, integrations, and scheduling rules.",
	},
	manager: {
		label: "Manager",
		badge: "success" as const,
		desc: "Manage bookings, schedule dates, cancel or reschedule meetings, and view form submissions.",
	},
	viewer: {
		label: "Viewer",
		badge: "outline" as const,
		desc: "Read-only access to view bookings, calendar availability, and form submission data.",
	},
};

export default function TeamPage() {
	const { user } = useWorkspace();
	const { data, loading, error, reload } = useRpc<TeamResponse>(
		"halomail.identity.v1.TeamService/ListTeamMembers",
	);

	const [showInviteModal, setShowInviteModal] = useState(false);
	const [inviteEmail, setInviteEmail] = useState("");
	const [inviteRole, setInviteRole] = useState<"admin" | "manager" | "viewer">("manager");
	const [inviteBusy, setInviteBusy] = useState(false);
	const [inviteError, setInviteError] = useState("");
	const [inviteSuccess, setInviteSuccess] = useState("");

	const [actionBusyId, setActionBusyId] = useState<string | null>(null);
	const [actionError, setActionError] = useState("");
	const [copiedToken, setCopiedToken] = useState<string | null>(null);
	const [createdInviteLink, setCreatedInviteLink] = useState("");
	const [copiedModalLink, setCopiedModalLink] = useState(false);

	const tier = data?.tier || "free";
	const maxSeats = data?.maxSeats || 0;
	const usedSeats = data?.usedSeats || data?.members?.length || 1;
	const isFreePlan = tier.toLowerCase() === "free";
	const canManage = data?.currentUserRole === "owner" || data?.currentUserRole === "admin";
	const isOwner = data?.currentUserRole === "owner";
	const seatsAvailable = Math.max(0, maxSeats - usedSeats);
	const seatsPercentage =
		maxSeats > 0 ? Math.min(100, Math.round((usedSeats / maxSeats) * 100)) : 100;

	function copyInviteLink(token: string) {
		const link = `${window.location.origin}/invite?token=${token}`;
		navigator.clipboard.writeText(link);
		setCopiedToken(token);
		setTimeout(() => setCopiedToken(null), 2000);
	}

	function copyModalLink(link: string) {
		navigator.clipboard.writeText(link);
		setCopiedModalLink(true);
		setTimeout(() => setCopiedModalLink(false), 2000);
	}

	async function handleSendInvite(e: React.FormEvent) {
		e.preventDefault();
		if (!inviteEmail || inviteBusy) return;
		setInviteBusy(true);
		setInviteError("");
		setInviteSuccess("");
		setCreatedInviteLink("");
		try {
			const res = await rpc<{ invitation?: { id: string; token?: string } }>(
				"halomail.identity.v1.TeamService/InviteTeamMember",
				{
					email: inviteEmail.trim(),
					role: inviteRole,
				},
			);
			const token = res.invitation?.token;
			if (token) {
				setCreatedInviteLink(`${window.location.origin}/invite?token=${token}`);
			}
			setInviteSuccess(`Invitation email sent to ${inviteEmail}!`);
			setInviteEmail("");
			await reload();
		} catch (err: unknown) {
			setInviteError(err instanceof Error ? err.message : "Failed to send invitation");
		} finally {
			setInviteBusy(false);
		}
	}

	async function handleRoleChange(memberId: string, newRole: string) {
		setActionBusyId(memberId);
		setActionError("");
		try {
			await rpc("halomail.identity.v1.TeamService/UpdateMemberRole", {
				userId: memberId,
				role: newRole,
			});
			await reload();
		} catch (err: unknown) {
			setActionError(err instanceof Error ? err.message : "Failed to update role");
		} finally {
			setActionBusyId(null);
		}
	}

	async function handleRemoveMember(member: TeamMember) {
		if (
			!window.confirm(
				`Remove ${member.name || member.email} from your company workspace? They will lose access immediately.`,
			)
		) {
			return;
		}
		setActionBusyId(member.id);
		setActionError("");
		try {
			await rpc("halomail.identity.v1.TeamService/RemoveTeamMember", {
				userId: member.id,
			});
			await reload();
		} catch (err: unknown) {
			setActionError(err instanceof Error ? err.message : "Failed to remove member");
		} finally {
			setActionBusyId(null);
		}
	}

	async function handleRevokeInvite(invitationId: string) {
		setActionBusyId(invitationId);
		setActionError("");
		try {
			await rpc("halomail.identity.v1.TeamService/RevokeInvitation", {
				invitationId,
			});
			await reload();
		} catch (err: unknown) {
			setActionError(err instanceof Error ? err.message : "Failed to revoke invitation");
		} finally {
			setActionBusyId(null);
		}
	}

	return (
		<div className="space-y-8">
			<PageHeader
				title="Team"
				description="Manage your team members, workspace permissions, and seat availability."
				action={
					!isFreePlan && canManage ? (
						<Button
							onClick={() => setShowInviteModal(true)}
							size="sm"
							className="gap-2"
							disabled={usedSeats >= maxSeats}
						>
							<UserPlus className="size-4" />
							Invite Member
						</Button>
					) : isFreePlan ? (
						<Button asChild size="sm" variant="outline" className="gap-1.5">
							<Link href="/dashboard/billing">
								Upgrade Plan <ArrowUpRight className="size-3.5" />
							</Link>
						</Button>
					) : null
				}
			/>

			{/* Seat Allowance Status */}
			<Card className="p-5">
				<div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
					<div className="space-y-1">
						<div className="flex items-center gap-2.5">
							<h2 className="text-sm font-semibold tracking-tight">Seat Allowance</h2>
							<Badge variant={isFreePlan ? "outline" : "default"} className="capitalize font-mono text-[11px]">
								{tier} Plan
							</Badge>
						</div>
						<p className="text-xs text-muted-foreground">
							{isFreePlan ? (
								"Solo workspaces are limited to 1 member. Upgrade to Pro or Business to collaborate."
							) : (
								<>
									<span className="font-semibold text-foreground">{usedSeats}</span> of{" "}
									<span className="font-semibold text-foreground">{maxSeats}</span> seats used (
									{seatsAvailable} {seatsAvailable === 1 ? "seat" : "seats"} available)
								</>
							)}
						</p>
					</div>

					{!isFreePlan && maxSeats > 0 && (
						<div className="flex items-center gap-3 w-full sm:w-64">
							<div
								className="h-2 w-full rounded-full bg-secondary overflow-hidden"
								role="progressbar"
								aria-label="Workspace seats used"
								aria-valuenow={usedSeats}
								aria-valuemin={0}
								aria-valuemax={maxSeats}
							>
								<div
									className="h-full bg-foreground/80 transition-all duration-300 rounded-full"
									style={{ width: `${seatsPercentage}%` }}
								/>
							</div>
							<span className="text-xs font-mono text-muted-foreground whitespace-nowrap">
								{seatsPercentage}%
							</span>
						</div>
					)}
				</div>
			</Card>

			{/* Upgrade Callout for Free Tier */}
			{isFreePlan && (
				<div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 p-5 rounded-lg border border-border bg-card">
					<div className="space-y-1">
						<h3 className="text-sm font-semibold">Need to collaborate with your team?</h3>
						<p className="text-xs text-muted-foreground max-w-xl leading-relaxed">
							Upgrade to <strong>Pro</strong> (up to 4 seats) or <strong>Business</strong> (up to 12 seats)
							to invite team members, assign role-based permissions, and share booking availability.
						</p>
					</div>
					<Button asChild size="sm" variant="outline" className="gap-1.5 shrink-0">
						<Link href="/dashboard/billing">
							View Plans <ArrowUpRight className="size-3.5" />
						</Link>
					</Button>
				</div>
			)}

			{/* Action Error Banner */}
			{(actionError || error) && (
				<div className="flex items-center gap-2 rounded-lg border border-destructive/30 bg-destructive/10 p-3 text-xs text-destructive">
					<AlertCircle className="size-4 shrink-0" />
					<span>{actionError || error}</span>
				</div>
			)}

			{/* Invite Modal / Dialog */}
			{showInviteModal && (
				<div
					className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4"
					role="dialog"
					aria-modal="true"
					aria-labelledby="invite-modal-title"
				>
					<Card className="w-full max-w-md p-6 space-y-5 animate-in fade-in-50 zoom-in-95">
						<div className="flex items-start justify-between">
							<div className="space-y-1">
								<h2 id="invite-modal-title" className="text-lg font-semibold text-foreground">
									Invite Team Member
								</h2>
								<p className="text-xs text-muted-foreground">
									Send an email invitation to collaborate in your workspace.
								</p>
							</div>
							<button
								type="button"
								onClick={() => {
									setShowInviteModal(false);
									setCreatedInviteLink("");
									setInviteError("");
									setInviteSuccess("");
								}}
								className="rounded-sm p-1 text-muted-foreground hover:text-foreground hover:bg-muted transition-colors"
								aria-label="Close modal"
							>
								<X className="size-4" />
							</button>
						</div>

						<form onSubmit={handleSendInvite} className="space-y-4">
							<div className="space-y-1.5">
								<Label htmlFor="invite-email" className="text-xs font-medium">
									Teammate Email Address
								</Label>
								<Input
									id="invite-email"
									type="email"
									placeholder="colleague@company.com"
									value={inviteEmail}
									onChange={(e) => setInviteEmail(e.target.value)}
									required
									autoFocus
									disabled={inviteBusy}
								/>
							</div>

							<div className="space-y-2">
								<Label className="text-xs font-medium">Assign Access Role</Label>
								<div className="space-y-2">
									{(["manager", "admin", "viewer"] as const).map((r) => {
										const detail = ROLE_DETAILS[r];
										const isSelected = inviteRole === r;
										return (
											<button
												key={r}
												type="button"
												onClick={() => setInviteRole(r)}
												className={`w-full text-left p-3 rounded-md border text-sm transition-all ${
													isSelected
														? "border-foreground/30 bg-muted/60 ring-1 ring-foreground/20"
														: "border-border hover:bg-muted/40"
												}`}
											>
												<div className="flex items-center justify-between font-medium text-xs">
													<span className="text-foreground">{detail.label}</span>
													{isSelected && <Check className="size-3.5 text-foreground" />}
												</div>
												<p className="text-[11px] text-muted-foreground mt-1 leading-normal">
													{detail.desc}
												</p>
											</button>
										);
									})}
								</div>
							</div>

							{createdInviteLink ? (
								<div className="space-y-3 p-3.5 rounded-lg border border-emerald-500/30 bg-emerald-500/10 text-xs">
									<div className="flex items-center gap-1.5 font-medium text-emerald-600 dark:text-emerald-400">
										<Check className="size-4 shrink-0" />
										<span>{inviteSuccess || "Invitation email dispatched!"}</span>
									</div>
									<p className="text-muted-foreground">
										An email has been sent. You can also copy and share the invite link directly:
									</p>
									<div className="flex items-center gap-2">
										<Input
											readOnly
											value={createdInviteLink}
											className="text-xs h-8 bg-background selection:bg-brand/20 font-mono"
										/>
										<Button
											type="button"
											size="sm"
											variant="outline"
											className="h-8 px-2.5 gap-1.5 shrink-0"
											onClick={() => copyModalLink(createdInviteLink)}
										>
											{copiedModalLink ? (
												<>
													<Check className="size-3 text-emerald-500" />
													<span>Copied</span>
												</>
											) : (
												<>
													<Copy className="size-3" />
													<span>Copy</span>
												</>
											)}
										</Button>
									</div>
									<div className="pt-2 flex justify-end">
										<Button
											type="button"
											size="sm"
											onClick={() => {
												setShowInviteModal(false);
												setCreatedInviteLink("");
												setInviteSuccess("");
											}}
										>
											Done
										</Button>
									</div>
								</div>
							) : (
								<>
									{inviteError && (
										<p role="alert" className="text-xs text-destructive">
											{inviteError}
										</p>
									)}

									<div className="flex justify-end gap-2 pt-2">
										<Button
											type="button"
											variant="outline"
											size="sm"
											disabled={inviteBusy}
											onClick={() => setShowInviteModal(false)}
										>
											Cancel
										</Button>
										<Button type="submit" size="sm" disabled={inviteBusy || !inviteEmail}>
											{inviteBusy ? "Sending…" : "Send Invitation"}
										</Button>
									</div>
								</>
							)}
						</form>
					</Card>
				</div>
			)}

			{/* Active Members Section */}
			<section className="space-y-3">
				<PanelHeading
					title={`Active Members (${data?.members?.length || 0})`}
					description="Teammates who currently have access to this workspace."
				/>

				<div className="divide-y divide-border rounded-lg border border-border bg-card overflow-hidden">
					{data?.members?.map((member) => {
						const isSelf = member.id === user.id;
						const isTargetOwner = member.role === "owner";
						const isTargetAdmin = member.role === "admin";
						const canEditTarget = canManage && !isTargetOwner && (isOwner || !isTargetAdmin);
						const initials = (member.name || member.email)
							.slice(0, 2)
							.toUpperCase();

						return (
							<div
								key={member.id}
								className="flex flex-col sm:flex-row sm:items-center justify-between p-4 gap-3.5 hover:bg-muted/20 transition-colors"
							>
								<div className="flex items-center gap-3 min-w-0">
									{member.avatarUrl ? (
										<img
											src={member.avatarUrl}
											alt={member.name || member.email}
											className="size-9 shrink-0 rounded-full object-cover border border-border"
										/>
									) : (
										<div className="flex size-9 shrink-0 items-center justify-center rounded-full bg-secondary text-xs font-semibold text-foreground/80 border border-border/50">
											{initials}
										</div>
									)}
									<div className="min-w-0 space-y-0.5">
										<div className="flex items-center gap-2">
											<p className="font-medium text-sm text-foreground truncate">
												{member.name || member.email}
											</p>
											{isSelf && (
												<span className="text-[10px] font-medium tracking-wide uppercase px-1.5 py-0.5 rounded bg-secondary text-muted-foreground border border-border/50">
													You
												</span>
											)}
										</div>
										<p className="text-xs text-muted-foreground truncate">{member.email}</p>
									</div>
								</div>

								<div className="flex items-center gap-3 shrink-0 self-end sm:self-center">
									{canEditTarget ? (
										<div className="relative">
											<select
												value={member.role}
												onChange={(e) => handleRoleChange(member.id, e.target.value)}
												disabled={actionBusyId === member.id}
												aria-label={`Change role for ${member.name || member.email}`}
												className="h-8 appearance-none rounded-md border border-input bg-background pl-2.5 pr-7 text-xs font-medium focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring cursor-pointer hover:bg-muted/40 transition-colors"
											>
												{isOwner && <option value="admin">Administrator</option>}
												<option value="manager">Manager</option>
												<option value="viewer">Viewer</option>
											</select>
											<ChevronDown className="pointer-events-none absolute right-2 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" />
										</div>
									) : (
										<Badge
											variant={
												member.role === "owner"
													? "brand"
													: member.role === "admin"
													? "default"
													: member.role === "manager"
													? "success"
													: "outline"
											}
											className="text-xs font-medium capitalize"
										>
											{ROLE_DETAILS[member.role]?.label || member.role}
										</Badge>
									)}

									{canEditTarget && (
										<Button
											variant="ghost"
											size="sm"
											className="h-8 w-8 p-0 text-muted-foreground hover:text-destructive hover:bg-destructive/10"
											disabled={actionBusyId === member.id}
											onClick={() => handleRemoveMember(member)}
											title="Remove member"
										>
											<Trash2 className="size-3.5" />
											<span className="sr-only">Remove member</span>
										</Button>
									)}
								</div>
							</div>
						);
					})}
				</div>
			</section>

			{/* Pending Invitations Section */}
			{!!data?.invitations?.length && (
				<section className="space-y-3">
					<PanelHeading
						title={`Pending Invitations (${data.invitations.length})`}
						description="Invited colleagues who have not yet accepted their invitation."
					/>

					<div className="divide-y divide-border rounded-lg border border-border bg-card overflow-hidden">
						{data.invitations.map((inv) => (
							<div
								key={inv.id}
								className="flex flex-col sm:flex-row sm:items-center justify-between p-4 gap-3 hover:bg-muted/20 transition-colors"
							>
								<div className="flex items-center gap-3 min-w-0">
									<div className="flex size-9 shrink-0 items-center justify-center rounded-full bg-secondary/70 text-muted-foreground border border-border/40">
										<Mail className="size-4" />
									</div>
									<div className="min-w-0 space-y-0.5">
										<p className="font-medium text-sm text-foreground truncate">{inv.email}</p>
										<p className="text-xs text-muted-foreground flex items-center gap-1.5">
											<Clock className="size-3 shrink-0" />
											<span>
												Invited by {inv.invitedByName} · Expires{" "}
												{new Date(inv.expiresAt).toLocaleDateString()}
											</span>
										</p>
									</div>
								</div>

								<div className="flex items-center gap-2.5 shrink-0 self-end sm:self-center">
									<Badge variant="outline" className="text-xs font-medium capitalize">
										{ROLE_DETAILS[inv.role]?.label || inv.role}
									</Badge>

									{canManage && (
										<div className="flex items-center gap-1.5">
											{inv.token && (
												<Button
													variant="outline"
													size="sm"
													className="text-xs h-7 px-2.5 gap-1.5"
													onClick={() => copyInviteLink(inv.token!)}
													title="Copy invitation link"
												>
													{copiedToken === inv.token ? (
														<>
															<Check className="size-3 text-emerald-500" />
															<span>Copied</span>
														</>
													) : (
														<>
															<Copy className="size-3" />
															<span>Copy Link</span>
														</>
													)}
												</Button>
											)}
											<Button
												variant="outline"
												size="sm"
												className="text-xs h-7 px-2.5 text-muted-foreground hover:text-destructive hover:border-destructive/30"
												disabled={actionBusyId === inv.id}
												onClick={() => handleRevokeInvite(inv.id)}
											>
												Revoke
											</Button>
										</div>
									)}
								</div>
							</div>
						))}
					</div>
				</section>
			)}

			{/* Roles & Permissions Reference */}
			<section className="space-y-3 pt-2">
				<PanelHeading
					title="Role Permissions Reference"
					description="Access and capability levels for each workspace role."
				/>
				<div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3.5">
					<Card className="p-4 space-y-2 border-border/80">
						<div className="flex items-center justify-between">
							<span className="text-sm font-semibold text-foreground">Owner</span>
							<Badge variant="brand" className="text-[10px]">Primary</Badge>
						</div>
						<p className="text-xs text-muted-foreground leading-relaxed">
							Full workspace ownership, subscription billing, plan seat upgrades, and workspace transfer rights.
						</p>
					</Card>

					<Card className="p-4 space-y-2 border-border/80">
						<div className="flex items-center justify-between">
							<span className="text-sm font-semibold text-foreground">Administrator</span>
							<Badge variant="default" className="text-[10px]">Admin</Badge>
						</div>
						<p className="text-xs text-muted-foreground leading-relaxed">
							Invite and remove team members, configure booking rules, custom forms, integrations, and workspace settings.
						</p>
					</Card>

					<Card className="p-4 space-y-2 border-border/80">
						<div className="flex items-center justify-between">
							<span className="text-sm font-semibold text-foreground">Manager</span>
							<Badge variant="success" className="text-[10px]">Operator</Badge>
						</div>
						<p className="text-xs text-muted-foreground leading-relaxed">
							Operational control: schedule and manage calendar bookings, handle cancellations, and view form submissions.
						</p>
					</Card>

					<Card className="p-4 space-y-2 border-border/80">
						<div className="flex items-center justify-between">
							<span className="text-sm font-semibold text-foreground">Viewer</span>
							<Badge variant="outline" className="text-[10px]">Read-only</Badge>
						</div>
						<p className="text-xs text-muted-foreground leading-relaxed">
							Read-only observer: view calendar schedules, event bookings, and submitted form responses without edit access.
						</p>
					</Card>
				</div>
			</section>
		</div>
	);
}

