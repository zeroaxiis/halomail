"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { API_URL, rpc } from "@/lib/api";
import { useRpc } from "@/lib/use-rpc";

type Feature = "forms" | "meetings";
interface AccessKey { id: string; name: string; prefix: string; lastFour: string; revoked?: boolean; scopes?: string[] }
interface Allowance { used: number; limit: number; remaining: number; resetsAt: string | null }

export function FeatureAccess({ feature }: { feature: Feature }) {
  const keys = useRpc<{ keys?: AccessKey[] }>("halomail.identity.v1.ApiKeyService/ListApiKeys");
  const usage = useRpc<Record<Feature, Allowance>>("v1/usage");
  const [secret, setSecret] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const scope = feature === "forms" ? "forms:submit" : "meetings:book";
  const active = keys.data?.keys?.find(key => !key.revoked && key.scopes?.includes(scope));
  const allowance = usage.data?.[feature];

  async function generate() {
    setBusy(true); setError("");
    try {
      const result = await rpc<{ secret: string }>("halomail.identity.v1.ApiKeyService/CreateApiKey", { name: feature });
      setSecret(result.secret);
      await keys.reload();
    } catch (failure) { setError(failure instanceof Error ? failure.message : "Could not generate key"); }
    finally { setBusy(false); }
  }

  async function remove() {
    if (!active || !window.confirm("Delete this key? Websites using it will stop accepting submissions or bookings until you replace it.")) return;
    setBusy(true); setError("");
    try {
      await rpc("halomail.identity.v1.ApiKeyService/RevokeApiKey", { id: active.id });
      setSecret(null);
      await keys.reload();
    } catch (failure) { setError(failure instanceof Error ? failure.message : "Could not delete key"); }
    finally { setBusy(false); }
  }

  const keyValue = secret || "YOUR_ACCESS_KEY";
  const snippet = feature === "forms"
    ? `<form action="${API_URL}/submit" method="POST">\n  <input type="hidden" name="access_key" value="${keyValue}">\n  <input name="name" required>\n  <input name="email" type="email" required>\n  <textarea name="message" required></textarea>\n  <input name="_hl_hp" style="display:none" tabindex="-1">\n  <button type="submit">Send message</button>\n</form>`
    : `<a href="${typeof window === "undefined" ? "" : window.location.origin}/book/meeting?key=${keyValue}">Book a meeting</a>`;

  return <Card className="mb-6 space-y-4 p-6">
    <div className="flex flex-wrap items-start justify-between gap-4">
      <div><h2 className="font-medium">{feature === "forms" ? "Forms" : "Meetings"} access key</h2>
        <p className="mt-1 text-sm text-muted-foreground">One active key for this feature. Copy a new key now; it cannot be viewed again after leaving this page.</p></div>
      {allowance && <div className="text-sm"><strong>{allowance.used} / {allowance.limit}</strong> used · {allowance.remaining} remaining
        <p className="text-xs text-muted-foreground">{allowance.resetsAt ? `Resets ${new Date(allowance.resetsAt).toLocaleDateString()}` : "Lifetime free allowance"}</p></div>}
    </div>
    {keys.loading ? <p>Loading key…</p> : active ? <code className="block break-all rounded bg-secondary p-3">{secret || `${active.prefix}_••••••••••••${active.lastFour}`}</code> : <p className="text-sm text-muted-foreground">No active key. Generate one to connect your website.</p>}
    {secret && <p className="text-sm text-brand">Shown once. Copy and store this key before leaving this page.</p>}
    <div className="flex gap-3">
      <Button onClick={generate} disabled={busy || keys.loading || !!keys.error || !!active}>Generate key</Button>
      {active && <Button variant="outline" onClick={remove} disabled={busy}>Delete key</Button>}
    </div>
    {(error || keys.error || usage.error) && <p role="alert" className="text-sm text-destructive">{error || keys.error || usage.error}</p>}
    <details className="text-sm"><summary className="cursor-pointer font-medium">Add to your website</summary>
      <p className="my-3 text-muted-foreground">{feature === "forms" ? "All named text fields are included in the email to your account address. File attachments are not supported." : "Connect your calendar, create an event type, and save availability below before sharing this button."}</p>
      <pre className="overflow-auto rounded bg-secondary p-4 text-xs"><code>{snippet}</code></pre>
      <p className="mt-2 text-xs text-muted-foreground">This key is visible in your website code and only permits this feature’s public submissions. Deleting it does not reset usage.</p>
    </details>
  </Card>;
}
