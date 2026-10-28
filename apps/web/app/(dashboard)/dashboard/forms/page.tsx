"use client";

import { useState } from "react";
import { FeatureAccess } from "@/components/dashboard/feature-access";
import { PageHeader } from "@/components/dashboard/page-header";
import {
	EmptyState,
	LoadingRows,
	PanelHeading,
	RefreshButton,
} from "@/components/dashboard/ui";
import { useRpc } from "@/lib/use-rpc";

interface Message {
	id: string;
	senderName: string;
	senderEmail: string;
	createdAt: string;
	isSpam?: boolean;
	data?: Record<string, string>;
}

export default function FormsPage() {
	const messages = useRpc<{ messages?: Message[] }>(
		"halomail.contact.v1.MessageService/ListMessages",
		{ page: { pageSize: 50 } },
	);
	const [refreshTrigger, setRefreshTrigger] = useState(0);
	return (
		<>
			<PageHeader
				title="Forms"
				action={
					<a
						href="/docs/index.html#embed"
						target="_blank"
						rel="noreferrer"
						className="workspace-text-link"
					>
						Integration guide <span aria-hidden="true">↗</span>
					</a>
				}
			/>
			<div className="workspace-content-grid">
				<section className="min-w-0">
					<PanelHeading
						title="Recent submissions"
						action={
							<RefreshButton
								loading={messages.loading}
								onClick={() => {
									void messages.reload();
									setRefreshTrigger((t) => t + 1);
								}}
							/>
						}
					/>
					<div className="workspace-table">
						<div className="workspace-table-labels workspace-label">
							<span>From</span>
							<span>Message</span>
							<span className="text-right">Received</span>
						</div>
						{messages.loading ? (
							<LoadingRows label="Loading submissions" />
						) : messages.error ? (
							<p role="alert" className="dashboard-error">
								{messages.error}
							</p>
						) : !messages.data?.messages?.length ? (
							<EmptyState
								title="No submissions yet."
								description="Add a form to your website using the access key and embed code."
							/>
						) : (
							messages.data.messages.map((message) => (
								<details key={message.id} className="workspace-submission">
									<summary className="workspace-table-row">
										<span className="min-w-0">
											<span className="block truncate font-medium">
												{message.senderName ||
													message.senderEmail ||
													"Website visitor"}
											</span>
										</span>
										<span className="min-w-0">
											<span className="block truncate text-muted-foreground">
												{message.data?.message || "Website submission"}
											</span>
											{message.isSpam && (
												<span className="workspace-status">
													Spam · not emailed
												</span>
											)}
										</span>
										<time
											dateTime={message.createdAt}
											className="text-right text-xs text-muted-foreground"
										>
											{new Date(message.createdAt).toLocaleDateString(
												undefined,
												{ month: "short", day: "numeric" },
											)}
										</time>
									</summary>
									<div className="workspace-submission-body">
										{message.isSpam && (
											<p className="mb-5 text-xs text-muted-foreground">
												Flagged as spam. This submission was not emailed.
											</p>
										)}
										<dl>
											<dt>From</dt>
											<dd>
												{message.senderEmail ||
													message.senderName ||
													"Website visitor"}
											</dd>
											<dt>Received</dt>
											<dd>{new Date(message.createdAt).toLocaleString()}</dd>
											{Object.entries(message.data || {}).map(
												([name, value]) => (
													<div key={name} className="contents">
														<dt>{name}</dt>
														<dd>{value}</dd>
													</div>
												),
											)}
										</dl>
									</div>
								</details>
							))
						)}
					</div>
					{!!messages.data?.messages?.length && (
						<p className="workspace-submission-hint">
							Select a message to read it.
						</p>
					)}
				</section>
				<aside className="workspace-rail">
					<FeatureAccess feature="forms" refreshTrigger={refreshTrigger} />
				</aside>
			</div>
		</>
	);
}
