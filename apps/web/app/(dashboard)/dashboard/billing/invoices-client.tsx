"use client";

import { useEffect, useState } from "react";
import { rpc } from "@/lib/api";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Download, FileText } from "lucide-react";
import { Badge } from "@/components/ui/badge";

type Usage = { plan: string; expiresAt?: string; forms: { limit: number }; meetings: { limit: number } };
export function CurrentPlanClient() {
  const [usage, setUsage] = useState<Usage | null>(null);
  const [error, setError] = useState("");
  useEffect(() => {
    const refresh = () => rpc<Usage>("v1/usage").then(data => { setUsage(data); setError(""); }).catch(() => setError("Could not load your subscription. Please refresh."));
    void refresh();
    window.addEventListener("billing-updated", refresh);
    return () => window.removeEventListener("billing-updated", refresh);
  }, []);
  return (
    <div className="mb-8 p-5 bg-card/60 backdrop-blur border border-border shadow-sm rounded-xl flex items-center justify-between">
      <div>
        <h3 className="font-semibold text-lg text-foreground">Current Subscription</h3>
        <p className="text-sm text-muted-foreground" role="status">
          {error || (!usage ? "Loading subscription..." : <>You are currently on the <span className="font-medium text-foreground capitalize">{usage.plan}</span> plan.
          {usage.expiresAt && usage.plan !== "free" && ` Your subscription expires on ${new Date(usage.expiresAt).toLocaleDateString()}.`}</>)}
        </p>
        {usage && <p className="text-sm text-muted-foreground mt-2">{usage.forms.limit.toLocaleString()} submissions and {usage.meetings.limit.toLocaleString()} meetings per usage period.</p>}
      </div>
      {usage && <Badge className="uppercase text-sm px-4 py-1.5">{usage.plan}</Badge>}
    </div>
  );
}

type Invoice = { key: string; id: string; date: string };
export function InvoicesClient() {
  const [invoices, setInvoices] = useState<Invoice[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [paymentId, setPaymentId] = useState("");
  const [syncing, setSyncing] = useState(false);
  const [message, setMessage] = useState("");

  async function loadInvoices() {
    setLoading(true); setError("");
    try {
      const res = await fetch("/api/invoices", { cache: "no-store" });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || "Could not load invoices.");
      setInvoices(data.invoices);
    } catch (error) { setError(error instanceof Error ? error.message : "Could not load invoices."); }
    finally { setLoading(false); }
  }
  useEffect(() => {
    void loadInvoices();
    setPaymentId(sessionStorage.getItem("halomail_pending_payment") || "");
  }, []);

  async function syncPayment() {
    setSyncing(true); setMessage("");
    try {
      const res = await fetch("/api/razorpay/reconcile", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ paymentId: paymentId.trim() }) });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || "Could not sync payment.");
      sessionStorage.removeItem("halomail_pending_payment");
      setMessage("Payment synced. Your plan and invoice are up to date.");
      window.dispatchEvent(new Event("billing-updated"));
      await loadInvoices();
    } catch (error) { setMessage(error instanceof Error ? error.message : "Could not sync payment. Do not pay again."); }
    finally { setSyncing(false); }
  }

  return (
    <div className="mt-6 flex flex-col gap-3">
      <div className="rounded-xl border border-border p-4">
        <p className="text-sm mb-3">Already paid but missing your plan or invoice? Enter the Razorpay payment ID from your receipt.</p>
        <div className="flex flex-wrap gap-2">
          <input aria-label="Razorpay payment ID" value={paymentId} onChange={event => setPaymentId(event.target.value)} placeholder="pay_..." className="rounded-md border border-border bg-background px-3 py-2 text-sm" />
          <Button onClick={syncPayment} disabled={syncing || !paymentId.trim()}>{syncing ? "Syncing..." : "Sync payment"}</Button>
        </div>
        {message && <p role="status" className="text-sm mt-3">{message}</p>}
      </div>
      {loading ? <p className="text-sm text-muted-foreground">Loading invoices...</p> :
        error ? <div role="alert"><p className="text-sm">{error}</p><Button variant="outline" onClick={loadInvoices}>Retry</Button></div> :
        invoices.length === 0 ? <p className="text-sm text-muted-foreground">No invoices generated yet.</p> :
        <p className="text-sm font-medium">Total invoices: {invoices.length}</p>}
      {invoices.map(inv => (
        <Card key={inv.key} className="p-4 flex items-center justify-between bg-card/40 border-border">
          <div className="flex items-center gap-4">
            <FileText className="size-5 text-primary" />
            <div><p className="font-medium text-sm">{inv.id}</p><p className="text-xs text-muted-foreground">{new Date(inv.date).toLocaleDateString()}</p></div>
          </div>
          <Button variant="outline" size="sm" asChild>
            <a href={`/api/invoices/download?key=${encodeURIComponent(inv.key)}`}><Download className="size-4 mr-2" /> Download</a>
          </Button>
        </Card>
      ))}
    </div>
  );
}
