"use client";

import { useCallback, useEffect, useState } from "react";
import { useWorkspace } from "@/components/dashboard/workspace-context";
import { Button } from "@/components/ui/button";

export function CurrentPlanClient() {
  const { usage: query } = useWorkspace();
  const usage = query.data;
  const error = query.error;
  return (
    <section className="workspace-billing-current">
      <div>
        <h2>Current subscription</h2>
        <p className="mt-2" role="status">
          {error ||
            (!usage ? (
              "Loading subscription…"
            ) : (
              <>
                {usage.forms.limit.toLocaleString()} submissions and{" "}
                {usage.meetings.limit.toLocaleString()} meetings per usage
                period.
                {usage.expiresAt && usage.plan !== "free" && (
                  <>
                    <br />
                    Expires {new Date(usage.expiresAt).toLocaleDateString()}.
                  </>
                )}
              </>
            ))}
        </p>
      </div>
      {usage && <span>{usage.plan}</span>}
    </section>
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

  const loadInvoices = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const res = await fetch("/api/invoices", { cache: "no-store" });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || "Could not load invoices.");
      setInvoices(data.invoices);
    } catch (error) {
      setError(
        error instanceof Error ? error.message : "Could not load invoices.",
      );
    } finally {
      setLoading(false);
    }
  }, []);
  useEffect(() => {
    void loadInvoices();
    setPaymentId(sessionStorage.getItem("halomail_pending_payment") || "");
  }, [loadInvoices]);

  async function syncPayment() {
    setSyncing(true);
    setMessage("");
    try {
      const res = await fetch("/api/razorpay/reconcile", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ paymentId: paymentId.trim() }),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || "Could not sync payment.");
      sessionStorage.removeItem("halomail_pending_payment");
      setMessage(data.refunded
        ? "Full refund synced. Access for this purchase has ended."
        : "Payment synced. Your plan and invoice are up to date.");
      window.dispatchEvent(new Event("billing-updated"));
      await loadInvoices();
    } catch (error) {
      setMessage(
        error instanceof Error
          ? error.message
          : "Could not sync payment. Do not pay again.",
      );
    } finally {
      setSyncing(false);
    }
  }

  return (
    <div>
      <div className="workspace-table">
        {loading ? (
          <p className="py-8 text-xs text-muted-foreground" role="status">
            Loading invoices…
          </p>
        ) : error ? (
          <div className="py-6" role="alert">
            <p className="mb-3 text-xs text-muted-foreground">{error}</p>
            <Button variant="outline" size="sm" onClick={loadInvoices}>
              Retry
            </Button>
          </div>
        ) : !invoices.length ? (
          <p className="py-10 text-sm text-muted-foreground">
            No invoices yet.
          </p>
        ) : (
          invoices.map((inv) => (
            <div
              key={inv.key}
              className="flex flex-wrap items-center justify-between gap-4 py-5"
            >
              <div className="min-w-0">
                <p className="break-all font-mono text-xs">{inv.id}</p>
                <p className="mt-2 text-[11px] text-muted-foreground">
                  {new Date(inv.date).toLocaleDateString()}
                </p>
              </div>
              <Button variant="outline" size="sm" asChild>
                <a
                  href={`/api/invoices/download?key=${encodeURIComponent(inv.key)}`}
                >
                  Download invoice
                </a>
              </Button>
            </div>
          ))
        )}
      </div>
      <details className="workspace-embed max-w-2xl">
        <summary>Missing a payment or invoice?</summary>
        <p className="workspace-key-note">
          If you already paid, enter the Razorpay payment ID from your receipt
          to sync your plan and invoice.
        </p>
        <div className="mb-4 flex flex-wrap gap-2">
          <input
            aria-label="Razorpay payment ID"
            value={paymentId}
            onChange={(event) => setPaymentId(event.target.value)}
            placeholder="pay_..."
            className="min-w-0 rounded border border-border bg-transparent px-3 py-2 font-mono text-xs"
          />
          <Button
            size="sm"
            onClick={syncPayment}
            disabled={syncing || !paymentId.trim()}
          >
            {syncing ? "Syncing…" : "Sync payment"}
          </Button>
        </div>
        {message && (
          <p role="status" className="workspace-key-note">
            {message}
          </p>
        )}
      </details>
    </div>
  );
}
