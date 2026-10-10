"use client";

import { useEffect, useState } from "react";
import { getUser } from "@/lib/auth";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Download, FileText } from "lucide-react";
import { Badge } from "@/components/ui/badge";

export function CurrentPlanClient() {
  const [plan, setPlan] = useState("Free");
  const [expiresAt, setExpiresAt] = useState<string | null>(null);

  useEffect(() => {
    fetch("/api/rpc/v1/usage", { method: "POST", body: "{}" })
      .then(res => res.json())
      .then(data => {
        if (data.plan) {
          setPlan(data.plan);
        }
        if (data.expiresAt) {
          setExpiresAt(new Date(data.expiresAt).toLocaleDateString());
        }
      })
      .catch(() => {});
  }, []);

  return (
    <div className="mb-8 p-5 bg-card/60 backdrop-blur border border-border shadow-sm rounded-xl flex items-center justify-between">
      <div>
        <h3 className="font-semibold text-lg text-foreground">Current Subscription</h3>
        <p className="text-sm text-muted-foreground">
          You are currently on the <span className="font-medium text-foreground capitalize">{plan}</span> plan.
          {expiresAt && plan !== "free" && ` Your subscription expires on ${expiresAt}.`}
        </p>
      </div>
      <Badge className="uppercase text-sm px-4 py-1.5">{plan}</Badge>
    </div>
  );
}

export function InvoicesClient() {
  const [invoices, setInvoices] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const user = getUser();
    if (!user) return;

    fetch(`/api/invoices?orgId=${user.org_id}`)
      .then(res => res.json())
      .then(data => {
        if (data.invoices) setInvoices(data.invoices);
        setLoading(false);
      })
      .catch(() => setLoading(false));
  }, []);

  if (loading) return <div className="text-sm text-muted-foreground mt-4">Loading invoices...</div>;

  if (invoices.length === 0) return <div className="text-sm text-muted-foreground mt-4">No invoices generated yet.</div>;

  return (
    <div className="mt-6 flex flex-col gap-3">
      <div className="mb-2 text-sm font-medium text-foreground">
        Total Invoices: {invoices.length}
      </div>
      {invoices.map((inv) => (
        <Card key={inv.key} className="p-4 flex items-center justify-between bg-card/40 backdrop-blur border-border hover:bg-card/60 transition-colors">
          <div className="flex items-center gap-4">
            <div className="p-2.5 bg-primary/10 rounded-xl">
              <FileText className="size-5 text-primary" />
            </div>
            <div>
              <p className="font-medium text-sm text-foreground">{inv.id}</p>
              <p className="text-xs text-muted-foreground">{new Date(inv.date).toLocaleDateString()}</p>
            </div>
          </div>
          <Button variant="outline" size="sm" asChild className="rounded-full">
            <a href={`/api/invoices/download?key=${encodeURIComponent(inv.key)}`} target="_blank" rel="noreferrer">
              <Download className="size-4 mr-2" /> Download
            </a>
          </Button>
        </Card>
      ))}
    </div>
  );
}
