import Link from "next/link";
import { Button } from "@/components/ui/button";
import type { Allowance } from "./workspace-context";

export function EmptyState({
	title,
	description,
	action,
}: { title: string; description: string; action?: React.ReactNode }) {
	return (
		<div className="workspace-empty">
			<h3>{title}</h3>
			<p>{description}</p>
			{action && <div className="mt-5">{action}</div>}
		</div>
	);
}

export function PanelHeading({
	title,
	description,
	action,
}: { title: string; description?: string; action?: React.ReactNode }) {
	return (
		<div className="workspace-section-heading">
			<div>
				<h2>{title}</h2>
				{description && <p>{description}</p>}
			</div>
			{action}
		</div>
	);
}

export function UsageMeter({
	allowance,
	label,
}: { allowance: Allowance; label: string }) {
	const percentage =
		allowance.limit > 0
			? Math.min(100, Math.max(0, (allowance.used / allowance.limit) * 100))
			: 0;
	return (
		<div className="workspace-usage">
			<div>
				<span>{label}</span>
				<span className="font-mono tabular-nums">
					{allowance.used.toLocaleString()}{" "}
					<span className="text-muted-foreground">
						/ {allowance.limit.toLocaleString()}
					</span>
				</span>
			</div>
			<div
				className="workspace-usage-track"
				role="progressbar"
				aria-label={label}
				aria-valuemin={0}
				aria-valuemax={allowance.limit}
				aria-valuenow={Math.min(allowance.used, allowance.limit)}
				aria-valuetext={`${allowance.used} of ${allowance.limit} used`}
			>
				<span style={{ width: `${percentage}%` }} />
			</div>
		</div>
	);
}

export function RefreshButton({
	loading,
	onClick,
}: { loading: boolean; onClick: () => void }) {
	return (
		<Button variant="ghost" size="sm" disabled={loading} onClick={onClick}>
			{loading ? "Refreshing…" : "Refresh"}
		</Button>
	);
}

export function ViewLink({
	href,
	children,
}: { href: string; children: React.ReactNode }) {
	return (
		<Link href={href} className="workspace-text-link">
			{children}
			<span aria-hidden="true">↗</span>
		</Link>
	);
}

export function LoadingRows({ label = "Loading" }: { label?: string }) {
	return (
		<div role="status" aria-label={label} className="workspace-loading-rows">
			{[0, 1, 2].map((row) => (
				<div key={row} className="motion-safe:animate-pulse">
					<span />
					<span />
				</div>
			))}
			<span className="sr-only">{label}…</span>
		</div>
	);
}
