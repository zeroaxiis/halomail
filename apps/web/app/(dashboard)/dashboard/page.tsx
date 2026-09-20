import Link from "next/link";
import { CalendarClock, Inbox } from "lucide-react";
import { Card } from "@/components/ui/card";
import { PageHeader } from "@/components/dashboard/page-header";

export default function OverviewPage() {
  return <>
    <PageHeader title="Your workspace" description="Choose what you want to connect to your website." />
    <div className="grid gap-6 md:grid-cols-2">
      <Link href="/dashboard/forms"><Card className="h-full space-y-4 p-8 transition-colors hover:border-brand">
        <Inbox className="size-8 text-brand" /><h2 className="text-2xl font-semibold">Forms</h2>
        <p className="text-muted-foreground">Receive contact forms in your inbox. Manage your access key, view submissions, and track your free allowance.</p>
        <p className="text-sm font-medium">Open Forms →</p>
      </Card></Link>
      <Link href="/dashboard/meetings"><Card className="h-full space-y-4 p-8 transition-colors hover:border-brand">
        <CalendarClock className="size-8 text-brand" /><h2 className="text-2xl font-semibold">Meetings</h2>
        <p className="text-muted-foreground">Let visitors book your time. Connect Google Calendar, set availability, and share a booking button.</p>
        <p className="text-sm font-medium">Open Meetings →</p>
      </Card></Link>
    </div>
  </>;
}
