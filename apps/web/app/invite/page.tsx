"use client";

import { Suspense, useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import { Users, CheckCircle2, AlertCircle, ArrowRight } from "lucide-react";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Logo } from "@/components/logo";
import { rpc } from "@/lib/api";
import { getUser, saveSession, type SessionUser } from "@/lib/auth";

function InviteAcceptor() {
	const searchParams = useSearchParams();
	const router = useRouter();
	const token = searchParams.get("token") || "";

	const [currentUser, setCurrentUser] = useState<SessionUser | null>(null);
	const [busy, setBusy] = useState(false);
	const [error, setError] = useState("");
	const [accepted, setAccepted] = useState(false);

	useEffect(() => {
		setCurrentUser(getUser());
	}, []);

	async function handleAccept() {
		if (!token || busy) return;
		setBusy(true);
		setError("");
		try {
			const res = await rpc<{ user: SessionUser }>(
				"halomail.identity.v1.TeamService/AcceptInvitation",
				{ token },
			);
			saveSession("", res.user);
			setAccepted(true);
			setTimeout(() => {
				router.push("/dashboard/team");
			}, 1500);
		} catch (err: unknown) {
			setError(err instanceof Error ? err.message : "Failed to accept invitation");
		} finally {
			setBusy(false);
		}
	}

	if (!token) {
		return (
			<Card className="p-8 text-center space-y-4 max-w-md mx-auto">
				<AlertCircle className="size-10 text-destructive mx-auto" />
				<h2 className="text-xl font-semibold">Invalid Invitation Link</h2>
				<p className="text-sm text-muted-foreground">
					No invitation token was found in the link. Please ask your administrator to send you a new invite.
				</p>
				<Button asChild variant="outline" className="mt-4">
					<Link href="/">Back to Home</Link>
				</Button>
			</Card>
		);
	}

	if (accepted) {
		return (
			<Card className="p-8 text-center space-y-4 max-w-md mx-auto animate-in fade-in-50">
				<CheckCircle2 className="size-12 text-emerald-500 mx-auto" />
				<h2 className="text-2xl font-semibold">Welcome to the Team!</h2>
				<p className="text-sm text-muted-foreground">
					You have successfully joined the workspace. Redirecting you to your team dashboard…
				</p>
				<Button asChild className="mt-4">
					<Link href="/dashboard/team">
						Go to Dashboard <ArrowRight className="size-4 ml-1.5" />
					</Link>
				</Button>
			</Card>
		);
	}

	return (
		<Card className="p-8 space-y-6 max-w-md mx-auto">
			<div className="text-center space-y-2">
				<div className="flex size-12 rounded-full bg-brand/10 text-brand mx-auto items-center justify-center">
					<Users className="size-6" />
				</div>
				<h1 className="text-2xl font-semibold">Team Workspace Invitation</h1>
				<p className="text-sm text-muted-foreground">
					You have been invited to collaborate with your team on HaloMail.
				</p>
			</div>

			{error && (
				<div className="flex items-start gap-2 rounded-lg border border-destructive/30 bg-destructive/10 p-3 text-xs text-destructive">
					<AlertCircle className="size-4 shrink-0 mt-0.5" />
					<span>{error}</span>
				</div>
			)}

			<div className="space-y-3 pt-2">
				{currentUser ? (
					<div className="space-y-4">
						<div className="rounded-md border bg-muted/40 p-3 text-xs text-muted-foreground">
							Signed in as <strong className="text-foreground">{currentUser.email}</strong>
						</div>
						<Button onClick={handleAccept} className="w-full gap-2" disabled={busy}>
							{busy ? "Joining workspace…" : "Accept and Join Workspace"}
						</Button>
					</div>
				) : (
					<div className="space-y-3">
						<p className="text-xs text-muted-foreground text-center">
							Please sign in or register with the email address that received this invitation.
						</p>
						<Button asChild className="w-full">
							<Link href={`/login?redirect=${encodeURIComponent(`/invite?token=${token}`)}`}>
								Sign In to Accept
							</Link>
						</Button>
						<Button asChild variant="outline" className="w-full">
							<Link href={`/register?redirect=${encodeURIComponent(`/invite?token=${token}`)}`}>
								Create an Account
							</Link>
						</Button>
					</div>
				)}
			</div>
		</Card>
	);
}

export default function InvitePage() {
	return (
		<div className="min-h-screen flex flex-col justify-center items-center p-4 bg-background">
			<div className="mb-6">
				<Link href="/" aria-label="HaloMail home">
					<Logo />
				</Link>
			</div>
			<Suspense fallback={<Card className="p-8 text-center max-w-md mx-auto">Loading invitation…</Card>}>
				<InviteAcceptor />
			</Suspense>
		</div>
	);
}
