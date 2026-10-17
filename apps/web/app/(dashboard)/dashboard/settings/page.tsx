"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useWorkspace } from "@/components/dashboard/workspace-context";
import { PageHeader } from "@/components/dashboard/page-header";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select } from "@/components/ui/select";
import { rpc } from "@/lib/api";
import { getToken } from "@/lib/auth";
import { localTimezone, timezones } from "@/lib/timezones";
import { useRpc } from "@/lib/use-rpc";

interface User {
	id: string;
	email: string;
	name: string;
	handle: string;
	timezone: string;
}

export default function SettingsPage() {
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
		</>
	);
}
