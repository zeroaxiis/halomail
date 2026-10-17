"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useTheme } from "next-themes";
import { useEffect, useState } from "react";
import { clearSession } from "@/lib/auth";
import { rpc } from "@/lib/api";
import { Sidebar } from "./sidebar";
import { useWorkspace } from "./workspace-context";

export function Topbar() {
	const router = useRouter();
	const { usage } = useWorkspace();
	const { resolvedTheme, setTheme } = useTheme();
	const [mounted, setMounted] = useState(false);
	useEffect(() => setMounted(true), []);

	async function logout() {
		await rpc("halomail.identity.v1.AuthService/Logout").catch(() => {});
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
					HaloMail
				</Link>
				<Sidebar />
				<div className="workspace-utilities">
					<button
						type="button"
						onClick={() =>
							setTheme(resolvedTheme === "dark" ? "light" : "dark")
						}
						aria-label="Toggle theme"
					>
						{mounted && resolvedTheme === "dark" ? "Light mode" : "Dark mode"}
					</button>
					<button type="button" onClick={logout}>
						Sign out
					</button>
				</div>
			</div>
			<div className="workspace-context-line">
				<div className="workspace-context-inner">
					<Link href="/dashboard/settings" className="truncate">
						Personal workspace
					</Link>
					<div className="flex shrink-0 items-center gap-5">
						<a
							href="/docs/index.html"
							target="_blank"
							rel="noreferrer"
							className="hidden hover:underline sm:inline"
						>
							Documentation
						</a>
						<Link
							href="/dashboard/billing"
							className="capitalize hover:underline"
						>
							{usage.data ? `${usage.data.plan} plan` : "Your plan"}
						</Link>
					</div>
				</div>
			</div>
		</header>
	);
}
