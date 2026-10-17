"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { API_URL, rpc } from "@/lib/api";
import { useRpc } from "@/lib/use-rpc";
import { UsageMeter } from "./ui";
import { useWorkspace } from "./workspace-context";

type Feature = "forms" | "meetings";
interface AccessKey {
	id: string;
	name: string;
	prefix: string;
	lastFour: string;
	revoked?: boolean;
	scopes?: string[];
}

export function FeatureAccess({
	feature,
	refreshTrigger,
}: { feature: Feature; refreshTrigger?: number }) {
	const keys = useRpc<{ keys?: AccessKey[] }>(
		"halomail.identity.v1.ApiKeyService/ListApiKeys",
	);
	const { usage } = useWorkspace();
	const reloadUsage = usage.reload;
	useEffect(() => {
		if (refreshTrigger) void reloadUsage();
	}, [refreshTrigger, reloadUsage]);
	const [secret, setSecret] = useState<string | null>(null);
	const [busy, setBusy] = useState(false);
	const [error, setError] = useState("");
	const [copied, setCopied] = useState<"key" | "snippet" | null>(null);
	const scope = feature === "forms" ? "forms:submit" : "meetings:book";
	const active = keys.data?.keys?.find(
		(key) => !key.revoked && key.scopes?.includes(scope),
	);
	const allowance = usage.data?.[feature];

	useEffect(() => {
		if (!copied) return;
		const timer = window.setTimeout(() => setCopied(null), 2200);
		return () => window.clearTimeout(timer);
	}, [copied]);

	async function generate() {
		setBusy(true);
		setError("");
		try {
			const result = await rpc<{ secret: string }>(
				"halomail.identity.v1.ApiKeyService/CreateApiKey",
				{ name: feature },
			);
			setSecret(result.secret);
			await keys.reload();
		} catch (failure) {
			setError(
				failure instanceof Error ? failure.message : "Could not generate key",
			);
		} finally {
			setBusy(false);
		}
	}

	async function remove() {
		if (
			!active ||
			!window.confirm(
				"Delete this key? Websites using it will stop accepting submissions or bookings until you replace it.",
			)
		)
			return;
		setBusy(true);
		setError("");
		try {
			await rpc("halomail.identity.v1.ApiKeyService/RevokeApiKey", {
				id: active.id,
			});
			setSecret(null);
			await keys.reload();
		} catch (failure) {
			setError(
				failure instanceof Error ? failure.message : "Could not delete key",
			);
		} finally {
			setBusy(false);
		}
	}

	async function copy(value: string, kind: "key" | "snippet") {
		try {
			await navigator.clipboard.writeText(value);
			setCopied(kind);
		} catch {
			setError("Clipboard unavailable. Select and copy the code manually.");
		}
	}

	const keyValue = secret || "YOUR_ACCESS_KEY";
	const snippet =
		feature === "forms"
			? `<form action="${API_URL}/submit" method="POST">\n  <input type="hidden" name="access_key" value="${keyValue}">\n  <input name="name" required>\n  <input name="email" type="email" required>\n  <textarea name="message" required></textarea>\n  <input name="_hl_hp" style="display:none" tabindex="-1">\n  <button type="submit">Send message</button>\n</form>`
			: `<a href="${typeof window === "undefined" ? "" : window.location.origin}/book/meeting?key=${keyValue}">Book a meeting</a>`;

	return (
		<section className="workspace-key">
			<div className="workspace-key-header">
				<h2>Access key</h2>
				<span className="workspace-status">
					{keys.loading
						? "Loading"
						: keys.error
							? "Unavailable"
							: active
								? "Active"
								: "Not connected"}
				</span>
			</div>
			{keys.loading ? (
				<p className="workspace-key-note" role="status">
					Loading key…
				</p>
			) : active || secret ? (
				<code>
					{secret || `${active?.prefix}_••••••••••••${active?.lastFour}`}
				</code>
			) : keys.error ? (
				<p className="workspace-key-note">Unable to load your key.</p>
			) : (
				<p className="workspace-key-note">
					No active key. Generate one to connect{" "}
					{feature === "forms" ? "your website forms" : "your booking button"}.
				</p>
			)}
			{(secret || (!active && !keys.loading && !keys.error)) && (
				<p className="workspace-key-note">
					{secret
						? "Shown once. Copy this key before leaving."
						: "Save your key when it appears. It is shown only once."}
				</p>
			)}
			<div className="workspace-key-controls">
				{secret && (
					<Button
						variant="outline"
						size="sm"
						onClick={() => void copy(secret, "key")}
					>
						{copied === "key" ? "Copied" : "Copy key"}
					</Button>
				)}
				{!active && (
					<Button
						onClick={generate}
						size="sm"
						disabled={busy || keys.loading || !!keys.error || !!secret}
					>
						{busy ? "Generating…" : "Generate key"}
					</Button>
				)}
				{active && (
					<button
						type="button"
						className="text-xs text-muted-foreground underline underline-offset-4 disabled:opacity-50"
						onClick={remove}
						disabled={busy}
					>
						{busy ? "Deleting…" : "Delete key"}
					</button>
				)}
				<span className="sr-only" role="status">
					{copied ? `${copied === "key" ? "Key" : "Snippet"} copied` : ""}
				</span>
			</div>
			{(error || keys.error || usage.error) && (
				<div className="mt-5">
					<p role="alert" className="dashboard-error">
						{error || keys.error || usage.error}
					</p>
					{(keys.error || usage.error) && (
						<Button
							variant="outline"
							size="sm"
							onClick={() => {
								void keys.reload();
								void reloadUsage();
							}}
						>
							Retry
						</Button>
					)}
				</div>
			)}

			<div className="workspace-key-usage">
				{allowance ? (
					<>
						<UsageMeter
							allowance={allowance}
							label={feature === "forms" ? "Submissions used" : "Bookings used"}
						/>
						<p className="workspace-key-note">
							{allowance.resetsAt
								? `Resets ${new Date(allowance.resetsAt).toLocaleDateString(undefined, { month: "short", day: "numeric" })}.`
								: "Lifetime allowance."}
						</p>
					</>
				) : (
					<p className="workspace-key-note">
						{usage.loading ? "Loading usage…" : "Usage unavailable"}
					</p>
				)}
			</div>
			<details className="workspace-embed">
				<summary>Website embed code</summary>
				<p className="workspace-key-note">
					{feature === "forms"
						? "Add this form to your website. Named text fields are sent to your account email. File attachments are not supported."
						: "Connect Google Calendar, create an event type, and save availability before sharing this button."}
				</p>
				<pre data-lenis-prevent>
					<code>{snippet}</code>
				</pre>
				<div className="mt-3">
					<Button
						variant="outline"
						size="sm"
						onClick={() => void copy(snippet, "snippet")}
					>
						{copied === "snippet" ? "Copied" : "Copy code"}
					</Button>
				</div>
				{!secret && (
					<p className="workspace-key-note">
						Replace YOUR_ACCESS_KEY with the key you saved when creating it.
					</p>
				)}
				<p className="workspace-key-note">
					Keys are hidden after creation. Deleting a key does not reset usage.
				</p>
			</details>
		</section>
	);
}
