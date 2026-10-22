"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useTheme } from "next-themes";
import { useEffect, useState } from "react";
import { Sun, Moon, LogOut } from "lucide-react";
import { clearSession } from "@/lib/auth";
import { rpc } from "@/lib/api";
import { LogoMark } from "@/components/logo";
import { Sidebar } from "./sidebar";
import { useWorkspace } from "./workspace-context";

export function Topbar() {
	const router = useRouter();
	const { usage } = useWorkspace();
	const { resolvedTheme, setTheme } = useTheme();
	const [mounted, setMounted] = useState(false);
	useEffect(() => setMounted(true), []);

	async function logout() {
		await rpc("halomail.identity.v1.AuthService/Logout").catch(() => { });
		clearSession();
		router.replace("/login");
	}

	return (
		<header className="workspace-masthead">
			<div className="workspace-masthead-inner">
				<Link
					href="/dashboard"
					className="workspace-wordmark"
					aria-label="HaloMail dashboard"
				>
					<LogoMark className="size-7 shrink-0" />
					<span>HaloMail</span>
				</Link>
				<Sidebar />
				<div className="workspace-utilities">
					<a
						href="/docs/index.html"
						target="_blank"
						rel="noreferrer"
						className="hidden sm:inline"
					>
						Documentation
					</a>
					<Link
						href="/dashboard/billing"
						className="workspace-plan-tag"
						title="Manage billing"
					>
						{usage.data?.plan || "Free"}
					</Link>
					<div
						className="workspace-action-pill"
						role="group"
						aria-label="Workspace controls"
					>
						<button
							type="button"
							onClick={() =>
								setTheme(resolvedTheme === "dark" ? "light" : "dark")
							}
							className="workspace-pill-btn workspace-theme-toggle"
							aria-label={
								mounted && resolvedTheme === "dark"
									? "Switch to light mode"
									: "Switch to dark mode"
							}
							title={
								mounted && resolvedTheme === "dark"
									? "Switch to light mode"
									: "Switch to dark mode"
							}
						>
							{mounted && resolvedTheme === "dark" ? (
								<Sun className="size-3.5 sun-icon" />
							) : (
								<Moon className="size-3.5 moon-icon" />
							)}
						</button>
						<span className="workspace-pill-divider" aria-hidden="true" />
						<button
							type="button"
							onClick={logout}
							className="workspace-pill-btn workspace-signout-btn"
							aria-label="Sign out"
							title="Sign out"
						>
							<LogOut className="size-3.5 signout-icon" />
						</button>
					</div>
				</div>
			</div>
		</header>
	);
}
