"use client";

import { Suspense, useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import {
	Users,
	CheckCircle2,
	AlertCircle,
	ArrowRight,
	Eye,
	EyeOff,
	KeyRound,
	Sparkles,
	ShieldAlert,
	Loader2,
	XCircle,
} from "lucide-react";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Logo } from "@/components/logo";
import { rpc } from "@/lib/api";
import { getUser, saveSession, type SessionUser } from "@/lib/auth";

interface InvitationData {
	id: string;
	email: string;
	role: string;
	invitedByName: string;
	status: string;
	expiresAt: { seconds: string } | string;
}

interface GetInvitationResponse {
	invitation: InvitationData;
	orgName: string;
	userExists: boolean;
}

function generateRandomPassword() {
	const chars = "abcdefghjkmnpqrstuvwxyzABCDEFGHJKMNPQRSTUVWXYZ23456789!@#$%^&*";
	let pwd = "";
	for (let i = 0; i < 14; i++) {
		pwd += chars.charAt(Math.floor(Math.random() * chars.length));
	}
	return pwd;
}

function InviteAcceptor() {
	const searchParams = useSearchParams();
	const router = useRouter();
	const token = searchParams.get("token") || "";

	const [currentUser, setCurrentUser] = useState<SessionUser | null>(null);
	const [loading, setLoading] = useState(true);
	const [invData, setInvData] = useState<GetInvitationResponse | null>(null);
	const [fetchError, setFetchError] = useState("");

	// Step states: "choice" -> "credentials" -> "otp" -> "completed" | "declined"
	const [step, setStep] = useState<"choice" | "credentials" | "otp" | "completed" | "declined">("choice");

	// Form states
	const [name, setName] = useState("");
	const [password, setPassword] = useState("");
	const [showPassword, setShowPassword] = useState(false);
	const [otpCode, setOtpCode] = useState("");
	const [challengeId, setChallengeId] = useState("");
	const [busy, setBusy] = useState(false);
	const [formError, setFormError] = useState("");
	const [copiedPassword, setCopiedPassword] = useState(false);

	useEffect(() => {
		setCurrentUser(getUser());
	}, []);

	useEffect(() => {
		if (!token) {
			setLoading(false);
			return;
		}
		let cancelled = false;
		async function fetchInv() {
			try {
				const res = await rpc<GetInvitationResponse>(
					"halomail.identity.v1.TeamService/GetInvitation",
					{ token },
				);
				if (!cancelled) {
					setInvData(res);
				}
			} catch (err) {
				if (!cancelled) {
					setFetchError(err instanceof Error ? err.message : "Invitation link is invalid or has expired.");
				}
			} finally {
				if (!cancelled) {
					setLoading(false);
				}
			}
		}
		fetchInv();
		return () => {
			cancelled = true;
		};
	}, [token]);

	// Fast track if already signed in with matching email
	async function handleDirectAccept() {
		if (!token || busy) return;
		setBusy(true);
		setFormError("");
		try {
			const res = await rpc<{ user: SessionUser }>(
				"halomail.identity.v1.TeamService/AcceptInvitation",
				{ token },
			);
			saveSession("", res.user);
			setStep("completed");
			setTimeout(() => {
				router.push("/dashboard/team");
			}, 1500);
		} catch (err) {
			setFormError(err instanceof Error ? err.message : "Failed to accept invitation");
		} finally {
			setBusy(false);
		}
	}

	function handleChoiceYes() {
		if (currentUser && invData && currentUser.email.toLowerCase() === invData.invitation.email.toLowerCase()) {
			handleDirectAccept();
			return;
		}
		setStep("credentials");
	}

	function handleChoiceNo() {
		setStep("declined");
	}

	function handleGeneratePassword() {
		const newPwd = generateRandomPassword();
		setPassword(newPwd);
		setShowPassword(true);
		setCopiedPassword(true);
		navigator.clipboard.writeText(newPwd).catch(() => undefined);
		setTimeout(() => setCopiedPassword(false), 3000);
	}

	async function handleSubmitCredentials(e: React.FormEvent) {
		e.preventDefault();
		if (!invData || busy) return;
		setBusy(true);
		setFormError("");

		const email = invData.invitation.email;

		try {
			if (!invData.userExists) {
				// 1. Create account first
				await rpc("halomail.identity.v1.AuthService/Register", {
					name: name.trim() || email.split("@")[0],
					email,
					password,
				});
			}

			// 2. Request OTP verification code using the password
			const otpRes = await rpc<{ challengeId: string }>("v1/auth/otp/request", {
				email,
				password,
			});

			setChallengeId(otpRes.challengeId);
			setStep("otp");
		} catch (err) {
			setFormError(err instanceof Error ? err.message : "Could not process request. Please try again.");
		} finally {
			setBusy(false);
		}
	}

	async function handleVerifyOtp(e: React.FormEvent) {
		e.preventDefault();
		if (!invData || busy) return;
		setBusy(true);
		setFormError("");

		try {
			// 1. Verify OTP code and establish authenticated session cookies
			await rpc("v1/auth/otp/verify", {
				challengeId,
				code: otpCode.trim(),
			});

			// 2. Accept the team invitation with the newly authenticated session
			const res = await rpc<{ user: SessionUser }>(
				"halomail.identity.v1.TeamService/AcceptInvitation",
				{ token },
			);

			saveSession("", res.user);
			setStep("completed");
			setTimeout(() => {
				router.push("/dashboard/team");
			}, 1500);
		} catch (err) {
			setFormError(err instanceof Error ? err.message : "Invalid or expired verification code.");
		} finally {
			setBusy(false);
		}
	}

	if (loading) {
		return (
			<Card className="p-8 text-center space-y-4 max-w-md mx-auto">
				<Loader2 className="size-8 animate-spin text-muted-foreground mx-auto" />
				<p className="text-sm text-muted-foreground">Loading invitation details…</p>
			</Card>
		);
	}

	if (!token || fetchError) {
		return (
			<Card className="p-8 text-center space-y-4 max-w-md mx-auto">
				<AlertCircle className="size-10 text-destructive mx-auto" />
				<h2 className="text-xl font-semibold">Invalid Invitation</h2>
				<p className="text-sm text-muted-foreground">
					{fetchError || "No invitation token was found in the link. Please ask your administrator to send you a new invite."}
				</p>
				<Button asChild variant="outline" className="mt-4">
					<Link href="/">Back to Home</Link>
				</Button>
			</Card>
		);
	}

	if (step === "completed") {
		return (
			<Card className="p-8 text-center space-y-4 max-w-md mx-auto animate-in fade-in-50">
				<CheckCircle2 className="size-12 text-emerald-500 mx-auto" />
				<h2 className="text-2xl font-semibold">Welcome to {invData?.orgName || "the Team"}!</h2>
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

	if (step === "declined") {
		return (
			<Card className="p-8 text-center space-y-4 max-w-md mx-auto">
				<XCircle className="size-10 text-muted-foreground mx-auto" />
				<h2 className="text-xl font-semibold">Invitation Declined</h2>
				<p className="text-sm text-muted-foreground">
					You have declined the invitation to join <strong>{invData?.orgName}</strong>. No changes were made to your account.
				</p>
				<Button asChild variant="outline" className="mt-4">
					<Link href="/">Return to Home</Link>
				</Button>
			</Card>
		);
	}

	const roleLabel =
		invData?.invitation.role === "admin"
			? "Administrator"
			: invData?.invitation.role === "manager"
			? "Manager"
			: "Viewer";

	return (
		<Card className="p-8 space-y-6 max-w-md mx-auto border-border/80 shadow-lg">
			{/* Choice Step */}
			{step === "choice" && (
				<div className="space-y-6">
					<div className="text-center space-y-2.5">
						<div className="flex size-12 rounded-full bg-primary/10 text-primary mx-auto items-center justify-center">
							<Users className="size-6" />
						</div>
						<div className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium bg-muted text-muted-foreground">
							Team Invitation
						</div>
						<h1 className="text-2xl font-semibold tracking-tight">Join {invData?.orgName}</h1>
						<p className="text-sm text-muted-foreground leading-relaxed">
							<strong>{invData?.invitation.invitedByName || "An administrator"}</strong> has invited you (
							<span className="font-mono text-foreground font-medium">{invData?.invitation.email}</span>) to join their workspace as a{" "}
							<span className="font-medium text-foreground">{roleLabel}</span>.
						</p>
					</div>

					{currentUser && currentUser.email.toLowerCase() !== invData?.invitation.email.toLowerCase() && (
						<div className="flex items-start gap-2.5 rounded-lg border border-amber-500/30 bg-amber-500/10 p-3 text-xs text-amber-600 dark:text-amber-400">
							<ShieldAlert className="size-4 shrink-0 mt-0.5" />
							<div>
								You are signed in as <strong>{currentUser.email}</strong>. This invitation was sent to{" "}
								<strong>{invData?.invitation.email}</strong>.
							</div>
						</div>
					)}

					<div className="rounded-lg border bg-muted/40 p-4 text-center space-y-1">
						<p className="text-xs font-medium text-foreground">Would you like to accept this invitation?</p>
						<p className="text-[11px] text-muted-foreground">
							Accepting will grant you access to {invData?.orgName}&apos;s team resources and meetings.
						</p>
					</div>

					{formError && (
						<div className="flex items-start gap-2 rounded-lg border border-destructive/30 bg-destructive/10 p-3 text-xs text-destructive">
							<AlertCircle className="size-4 shrink-0 mt-0.5" />
							<span>{formError}</span>
						</div>
					)}

					<div className="grid grid-cols-2 gap-3 pt-1">
						<Button
							type="button"
							variant="outline"
							onClick={handleChoiceNo}
							disabled={busy}
							className="w-full text-muted-foreground hover:text-foreground"
						>
							No, Decline
						</Button>
						<Button
							type="button"
							onClick={handleChoiceYes}
							disabled={busy}
							className="w-full"
						>
							{busy ? "Accepting…" : "Yes, Accept"}
						</Button>
					</div>
				</div>
			)}

			{/* Credentials Step (Password Input / Generation) */}
			{step === "credentials" && (
				<form onSubmit={handleSubmitCredentials} className="space-y-5">
					<div className="space-y-1.5 text-center">
						<div className="flex size-10 rounded-full bg-primary/10 text-primary mx-auto items-center justify-center mb-2">
							<KeyRound className="size-5" />
						</div>
						<h2 className="text-xl font-semibold">
							{invData?.userExists ? "Account Verification" : "Create Your Password"}
						</h2>
						{invData?.userExists ? (
							<div className="text-xs text-muted-foreground space-y-1">
								<p className="font-medium text-foreground">You already have an account on this email.</p>
								<p>Please enter your existing password for <strong>{invData.invitation.email}</strong> to join {invData.orgName}.</p>
							</div>
						) : (
							<div className="text-xs text-muted-foreground space-y-1">
								<p>Set a password for <strong>{invData?.invitation.email}</strong> to create your team account.</p>
							</div>
						)}
					</div>

					{formError && (
						<div className="flex items-start gap-2 rounded-lg border border-destructive/30 bg-destructive/10 p-3 text-xs text-destructive">
							<AlertCircle className="size-4 shrink-0 mt-0.5" />
							<span>{formError}</span>
						</div>
					)}

					{!invData?.userExists && (
						<div className="space-y-1.5">
							<Label htmlFor="inv-name" className="text-xs">Your Name</Label>
							<Input
								id="inv-name"
								autoComplete="name"
								placeholder="Full Name"
								value={name}
								onChange={(e) => setName(e.target.value)}
								disabled={busy}
							/>
						</div>
					)}

					<div className="space-y-1.5">
						<div className="flex items-center justify-between">
							<Label htmlFor="inv-pwd" className="text-xs">
								{invData?.userExists ? "Your Password" : "Create Password"}
							</Label>
							{!invData?.userExists && (
								<button
									type="button"
									onClick={handleGeneratePassword}
									className="text-[11px] text-primary hover:underline flex items-center gap-1 font-medium"
								>
									<Sparkles className="size-3" />
									{copiedPassword ? "Generated & Copied!" : "Generate Password"}
								</button>
							)}
						</div>

						<div className="relative">
							<Input
								id="inv-pwd"
								type={showPassword ? "text" : "password"}
								autoComplete={invData?.userExists ? "current-password" : "new-password"}
								required
								minLength={8}
								value={password}
								onChange={(e) => setPassword(e.target.value)}
								disabled={busy}
								placeholder={invData?.userExists ? "Enter your password" : "At least 8 characters"}
								className="pr-10"
							/>
							<button
								type="button"
								onClick={() => setShowPassword(!showPassword)}
								className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
								tabIndex={-1}
								aria-label={showPassword ? "Hide password" : "Show password"}
							>
								{showPassword ? <EyeOff className="size-4" /> : <Eye className="size-4" />}
							</button>
						</div>
					</div>

					<div className="pt-2 flex flex-col gap-2">
						<Button type="submit" disabled={busy || password.length < 8} className="w-full">
							{busy ? "Processing…" : "Continue to Verification"}
						</Button>
						<Button
							type="button"
							variant="ghost"
							size="sm"
							onClick={() => setStep("choice")}
							disabled={busy}
							className="text-xs text-muted-foreground"
						>
							Back
						</Button>
					</div>
				</form>
			)}

			{/* OTP Verification Step */}
			{step === "otp" && (
				<form onSubmit={handleVerifyOtp} className="space-y-5">
					<div className="space-y-1.5 text-center">
						<div className="flex size-10 rounded-full bg-primary/10 text-primary mx-auto items-center justify-center mb-2">
							<CheckCircle2 className="size-5" />
						</div>
						<h2 className="text-xl font-semibold">Verify Your Email</h2>
						<p className="text-xs text-muted-foreground">
							We sent a 6-digit verification code to <strong>{invData?.invitation.email}</strong>. Enter it below to join {invData?.orgName}.
						</p>
					</div>

					{formError && (
						<div className="flex items-start gap-2 rounded-lg border border-destructive/30 bg-destructive/10 p-3 text-xs text-destructive">
							<AlertCircle className="size-4 shrink-0 mt-0.5" />
							<span>{formError}</span>
						</div>
					)}

					<div className="space-y-1.5">
						<Label htmlFor="inv-otp" className="text-xs">6-Digit Verification Code</Label>
						<Input
							id="inv-otp"
							type="text"
							inputMode="numeric"
							maxLength={6}
							required
							autoFocus
							value={otpCode}
							onChange={(e) => setOtpCode(e.target.value.replace(/\D/g, ""))}
							disabled={busy}
							placeholder="123456"
							className="text-center font-mono text-lg tracking-widest h-12"
						/>
					</div>

					<div className="pt-2 flex flex-col gap-2">
						<Button type="submit" disabled={busy || otpCode.length !== 6} className="w-full">
							{busy ? "Verifying & Joining…" : "Verify & Join Team"}
						</Button>
						<Button
							type="button"
							variant="ghost"
							size="sm"
							onClick={() => setStep("credentials")}
							disabled={busy}
							className="text-xs text-muted-foreground"
						>
							Change Password or Try Again
						</Button>
					</div>
				</form>
			)}
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
