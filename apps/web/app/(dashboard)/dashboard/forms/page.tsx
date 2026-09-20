"use client";

import { useState } from "react";
import { FeatureAccess } from "@/components/dashboard/feature-access";
import { PageHeader } from "@/components/dashboard/page-header";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { useRpc } from "@/lib/use-rpc";

interface Message { id: string; senderName: string; senderEmail: string; createdAt: string; isSpam?: boolean; data?: Record<string, string> }

export default function FormsPage() {
  const messages = useRpc<{ messages?: Message[] }>("halomail.contact.v1.MessageService/ListMessages", { page: { pageSize: 50 } });
  const [refreshTrigger, setRefreshTrigger] = useState(0);
  return <>
    <PageHeader title="Forms" description="Receive your website and portfolio submissions in your email inbox." />
    <FeatureAccess feature="forms" refreshTrigger={refreshTrigger} />
    <div className="mb-3 flex items-center justify-between"><h2 className="font-medium">Recent submissions</h2><Button variant="outline" onClick={() => { void messages.reload(); setRefreshTrigger(t => t + 1); }}>Refresh</Button></div>
    <Card className="divide-y divide-border">
      {messages.error && <p role="alert" className="p-5 text-destructive">{messages.error}</p>}
      {messages.loading ? <p className="p-5">Loading…</p> : !messages.data?.messages?.length ? <p className="p-5 text-muted-foreground">No submissions yet.</p> : messages.data.messages.map(message => <details key={message.id} className="p-5">
        <summary className="cursor-pointer text-sm">{message.senderName || message.senderEmail || "Website visitor"} · {new Date(message.createdAt).toLocaleString()} {message.isSpam ? "· Spam (not emailed)" : ""}</summary>
        <p className="mt-3 text-sm text-muted-foreground">{message.senderEmail}</p>
        <dl className="mt-3 space-y-2 text-sm">{Object.entries(message.data || {}).map(([name, value]) => <div key={name}><dt className="font-medium">{name}</dt><dd className="whitespace-pre-wrap break-words text-muted-foreground">{value}</dd></div>)}</dl>
      </details>)}
    </Card>
  </>;
}
