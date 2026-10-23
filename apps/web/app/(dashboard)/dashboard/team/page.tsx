"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import {
	Users,
	UserPlus,
	Shield,
	Crown,
	Mail,
	Trash2,
	Copy,
	Check,
	AlertCircle,
	ArrowUpRight,
	Sparkles,
	Clock,
} from "lucide-react";
import { PageHeader } from "@/components/dashboard/page-header";
import { PanelHeading, EmptyState } from "@/components/dashboard/ui";
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
		badge: "brand",
		desc: "Full workspace ownership, billing, and transfer permissions.",
	},
	admin: {
		label: "Administrator",
		badge: "default",
		desc: "Manage team members, configure settings, forms, and calendar rules.",
	},
	manager: {
		label: "Manager",
		badge: "success",
		desc: "Manage bookings, dates, cancel/reschedule meetings, and view forms.",
	},
	viewer: {
		label: "Viewer",
		badge: "outline",
		desc: "Read-only access to bookings, calendars, and form submissions.",
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

	const tier = data?.tier || "free";
	const maxSeats = data?.maxSeats || 0;
	const usedSeats = data?.usedSeats || data?.members?.length || 1;
	const isFreePlan = tier.toLowerCase() === "free";
	const canManage = data?.currentUserRole === "owner" || data?.currentUserRole === "admin";
	const isOwner = data?.currentUserRole === "owner";

	async function handleSendInvite(e: React.FormEvent) {
		e.preventDefault();
		if (!inviteEmail || inviteBusy) return;
		setInviteBusy(true);
		setInviteError("");
		setInviteSuccess("");
		try {
			await rpc("halomail.identity.v1.TeamService/InviteTeamMember", {
				email: inviteEmail.trim(),
				role: inviteRole,
			});
			setInviteSuccess(`Invitation sent to ${inviteEmail}!`);
			setInviteEmail("");
			await reload();
			setTimeout(() => {
				setShowInviteModal(false);
				setInviteSuccess("");
			}, 1800);
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
				title="Team Management"
				description="Collaborate with teammates, assign role-based permissions, and manage your company workspace seats."
			/>

			{/* Plan & Seat Allowance Status */}
			<Card className="p-6">
				<div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
					<div className="space-y-1">
						<div className="flex items-center gap-2">
							<h2 className="text-lg font-medium">Workspace Seats</h2>
							<Badge variant={isFreePlan ? "outline" : "brand"} className="capitalize">
								{tier} Plan
							</Badge>
						</div>
						<p className="text-sm text-muted-foreground">
							{isFreePlan ? (
								"Solo plan: 1 user included. Team management is available on Pro (4 seats) and Business (12 seats)."
							) : (
								<>
									<span className="font-semibold text-foreground">{usedSeats}</span> of{" "}
									<span className="font-semibold text-foreground">{maxSeats}</span> seats used (
									{Math.max(0, maxSeats - usedSeats)} remaining)
								</>
							)}
						</p>
					</div>

					<div>
						{isFreePlan ? (
							<Button asChild size="sm" className="gap-1.5">
								<Link href="/dashboard/billing">
									<Sparkles className="size-4 text-brand" />
									Upgrade to Invite Team
								</Link>
							</Button>
						) : canManage ? (
							<Button
								onClick={() => setShowInviteModal(true)}
								size="sm"
								className="gap-2"
								disabled={usedSeats >= maxSeats}
							>
								<UserPlus className="size-4" />
								Invite Teammate
							</Button>
						) : null}
					</div>
				</div>

				{/* Seat Progress Bar */}
				{!isFreePlan && maxSeats > 0 && (
					<div className="mt-4 pt-4 border-t border-border">
						<div
							className="h-2 w-full rounded-full bg-secondary overflow-hidden"
							role="progressbar"
							aria-valuenow={usedSeats}
							aria-valuemin={0}
							aria-valuemax={maxSeats}
						>
							<div
								className="h-full bg-brand transition-all duration-300 rounded-full"
								style={{ width: `${Math.min(100, (usedSeats / maxSeats) * 100)}%` }}
							/>
						</div>
					</div>
				)}
			</Card>

			{/* Upgrade Callout for Free Tier */}
			{isFreePlan && (
				<Card className="p-6 border-dashed border-brand/30 bg-card/60">
					<div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
						<div className="space-y-1">
							<h3 className="text-base font-semibold flex items-center gap-2">
								<Users className="size-4 text-brand" />
								Need to collaborate with your team?
							</h3>
							<p className="text-sm text-muted-foreground max-w-xl">
								Upgrade to <strong>Pro</strong> to add up to 4 members or <strong>Business</strong> for
								up to 12 members. Assign roles like <em>Administrator</em>, <em>Manager</em>, or{" "}
								<em>Viewer</em> to control access to forms and calendar bookings.
							</p>
						</div>
						<Button asChild variant="outline" size="sm" className="gap-1 shrink-0">
							<Link href="/dashboard/billing">
								View Plans <ArrowUpRight className="size-3.5" />
							</Link>
						</Button>
					</div>
				</Card>
			)}

			{/* Action Error Banner */}
			{(actionError || error) && (
				<div className="flex items-center gap-2 rounded-lg border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive">
					<AlertCircle className="size-4 shrink-0" />
					<span>{actionError || error}</span>
				</div>
			)}

			{/* Invite Modal / Dialog */}
			{showInviteModal && (
				<div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
					<Card className="w-full max-w-md p-6 space-y-5 animate-in fade-in-50 zoom-in-95">
						<div className="flex items-center justify-between">
							<div className="space-y-1">
								<h2 className="text-xl font-semibold">Invite Team Member</h2>
								<p className="text-xs text-muted-foreground">
									Send an email invitation to join your company workspace.
								</p>
							</div>
							<button
								type="button"
								onClick={() => setShowInviteModal(false)}
								className="text-muted-foreground hover:text-foreground text-sm p-1"
							>
								✕
							</button>
						</div>

						<form onSubmit={handleSendInvite} className="space-y-4">
							<div>
								<Label htmlFor="invite-email">Teammate Email Address</Label>
								<Input
									id="invite-email"
									type="email"
									placeholder="colleague@company.com"
									value={inviteEmail}
									onChange={(e) => setInviteEmail(e.target.value)}
									required
									autoFocus
									disabled={inviteBusy}
									className="mt-1"
								/>
							</div>

							<div className="space-y-2">
								<Label>Assign Access Role</Label>
								<div className="space-y-2">
									{(["manager", "admin", "viewer"] as const).map((r) => {
										const detail = ROLE_DETAILS[r];
										const isSelected = inviteRole === r;
										return (
											<button
												key={r}
												type="button"
												onClick={() => setInviteRole(r)}
												className={`w-full text-left p-3 rounded-md border text-sm transition-colors ${
													isSelected
														? "border-brand bg-brand/5 ring-1 ring-brand"
														: "border-border hover:bg-muted/50"
												}`}
											>
												<div className="flex items-center justify-between font-medium">
													<span>{detail.label}</span>
													{isSelected && <Check className="size-4 text-brand" />}
												</div>
												<p className="text-xs text-muted-foreground mt-0.5">
													{detail.desc}
												</p>
											</button>
										);
									})}
								</div>
							</div>

							{inviteError && (
								<p role="alert" className="text-xs text-destructive">
									{inviteError}
								</p>
							)}
							{inviteSuccess && (
								<p role="status" className="text-xs text-emerald-500 font-medium">
									{inviteSuccess}
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
						</form>
					</Card>
				</div>
			)}

			{/* Active Members Section */}
			<section className="space-y-3">
				<PanelHeading
					title={`Active Members (${data?.members?.length || 0})`}
					description="Users who currently have access to this workspace."
				/>

				<div className="divide-y divide-border rounded-lg border bg-card overflow-hidden">
					{data?.members?.map((member) => {
						const isSelf = member.id === user.id;
						const isTargetOwner = member.role === "owner";
						const isTargetAdmin = member.role === "admin";
						const canEditTarget =
							canManage && !isTargetOwner && (isOwner || !isTargetAdmin);

						return (
							<div
								key={member.id}
								className="flex flex-col sm:flex-row sm:items-center justify-between p-4 gap-3"
							>
								<div className="flex items-center gap-3 min-w-0">
									<div className="flex size-10 shrink-0 items-center justify-center rounded-full bg-secondary font-medium text-sm text-foreground">
										{member.name
											? member.name.slice(0, 2).toUpperCase()
											: member.email.slice(0, 2).toUpperCase()}
									</div>
									<div className="min-w-0">
										<div className="flex items-center gap-2">
											<p className="font-medium text-sm truncate">
												{member.name || member.email}
											</p>
											{isSelf && (
												<span className="text-[11px] font-mono text-muted-foreground">
													(You)
												</span>
											)}
										</div>
										<p className="text-xs text-muted-foreground truncate">{member.email}</p>
									</div>
								</div>

								<div className="flex items-center gap-3 shrink-0 self-end sm:self-center">
									{/* Role Badge or Selector */}
									{canEditTarget ? (
										<select
											value={member.role}
											onChange={(e) => handleRoleChange(member.id, e.target.value)}
											disabled={actionBusyId === member.id}
											aria-label={`Change role for ${member.name || member.email}`}
											className="rounded border border-input bg-background px-2.5 py-1 text-xs font-medium focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
										>
											{isOwner && <option value="admin">Administrator</option>}
											<option value="manager">Manager</option>
											<option value="viewer">Viewer</option>
										</select>
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
											className="capitalize gap-1"
										>
											{member.role === "owner" && <Crown className="size-3" />}
											{member.role === "admin" && <Shield className="size-3" />}
											{ROLE_DETAILS[member.role]?.label || member.role}
										</Badge>
									)}

									{/* Remove Action */}
									{canEditTarget && (
										<Button
											variant="ghost"
											size="sm"
											className="h-8 w-8 p-0 text-muted-foreground hover:text-destructive"
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

					<div className="divide-y divide-border rounded-lg border bg-card overflow-hidden">
						{data.invitations.map((inv) => (
							<div
								key={inv.id}
								className="flex flex-col sm:flex-row sm:items-center justify-between p-4 gap-3"
							>
								<div className="flex items-center gap-3 min-w-0">
									<div className="flex size-9 shrink-0 items-center justify-center rounded-full bg-secondary/60 text-muted-foreground">
										<Mail className="size-4" />
									</div>
									<div className="min-w-0">
										<p className="font-medium text-sm truncate">{inv.email}</p>
										<p className="text-xs text-muted-foreground flex items-center gap-1.5 mt-0.5">
											<Clock className="size-3" />
											Invited by {inv.invitedByName} · Expires{" "}
											{new Date(inv.expiresAt).toLocaleDateString()}
										</p>
									</div>
								</div>

								<div className="flex items-center gap-2 shrink-0 self-end sm:self-center">
									<Badge variant="outline" className="capitalize text-xs">
										{ROLE_DETAILS[inv.role]?.label || inv.role}
									</Badge>

									{canManage && (
										<Button
											variant="outline"
											size="sm"
											className="text-xs h-7 px-2.5 text-muted-foreground hover:text-destructive"
											disabled={actionBusyId === inv.id}
											onClick={() => handleRevokeInvite(inv.id)}
										>
											Revoke
										</Button>
									)}
								</div>
							</div>
						))}
					</div>
				</section>
			)}

			{/* Roles & Permissions Explanation Card */}
			<section className="space-y-3 pt-4">
				<PanelHeading title="Role Permissions Reference" />
				<div className="grid grid-cols-1 md:grid-cols-3 gap-4">
					<Card className="p-4 space-y-2">
						<div className="flex items-center gap-2">
							<Shield className="size-4 text-brand" />
							<h3 className="font-medium text-sm">Administrator</h3>
						</div>
						<p className="text-xs text-muted-foreground">
							Can invite and remove team members, configure forms, calendar booking rules, integrations, and company settings.
						</p>
					</Card>
					<Card className="p-4 space-y-2">
						<div className="flex items-center gap-2">
							<Users className="size-4 text-amber-500" />
							<h3 className="font-medium text-sm">Manager (Operator)</h3>
						</div>
						<p className="text-xs text-muted-foreground">
							Operational staff: can view and manage calendar bookings, cancel or reschedule meetings, and manage form submissions.
						</p>
					</Card>
					<Card className="p-4 space-y-2">
						<div className="flex items-center gap-2">
							<Users className="size-4 text-muted-foreground" />
							<h3 className="font-medium text-sm">Viewer</h3>
						</div>
						<p className="text-xs text-muted-foreground">
							Read-only observer: can check booking availability, event details, and form responses without making modifications.
						</p>
					</Card>
				</div>
			</section>
		</div>
	);
}
