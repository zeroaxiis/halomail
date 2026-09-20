"use client";

import { useEffect, useMemo, useState } from "react";
import { PageHeader } from "@/components/dashboard/page-header";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select } from "@/components/ui/select";
import { rpc } from "@/lib/api";
import { getToken, saveSession, getUser } from "@/lib/auth";
import { localTimezone, timezones } from "@/lib/timezones";
import { useRpc } from "@/lib/use-rpc";

interface User {
  id: string;
  email: string;
  name: string;
  handle: string;
  timezone: string;
}

export default function SettingsPage() {
  const { data } = useRpc<{ user: User }>("halomail.identity.v1.AuthService/GetCurrentUser");
  const [name, setName] = useState("");
  const [handle, setHandle] = useState("");
  const [timezone, setTimezone] = useState("");
  const [saved, setSaved] = useState(false);
  const [busy, setBusy] = useState(false);
  const [origin, setOrigin] = useState("");

  // Keep the saved value selectable even if it is not in the browser's list.
  const zones = useMemo(() => {
    const all = timezones();
    return timezone && !all.includes(timezone) ? [timezone, ...all] : all;
  }, [timezone]);

  useEffect(() => setOrigin(window.location.origin), []);
  useEffect(() => {
    if (data?.user) {
      setName(data.user.name);
      setHandle(data.user.handle);
      setTimezone(data.user.timezone);
    }
  }, [data]);

  async function save(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setSaved(false);
    try {
      const res = await rpc<{ user: User }>(
        "halomail.identity.v1.UserService/UpdateUser",
        { name, handle, timezone },
        getToken(),
      );
      const current = getUser();
      const token = getToken();
      if (current) saveSession(token || "", { ...current, ...res.user });
      setSaved(true);
    } finally {
      setBusy(false);
    }
  }

  return (
    <>
      <PageHeader title="Settings" description="Manage your profile and public booking page." />

      <div className="grid max-w-2xl gap-6">
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Profile</CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={save} className="space-y-4">
              <div className="space-y-1.5">
                <Label htmlFor="name">Name</Label>
                <Input id="name" value={name} onChange={(e) => setName(e.target.value)} />
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="handle">Handle</Label>
                <Input id="handle" value={handle} onChange={(e) => setHandle(e.target.value)} />
                <p className="text-xs text-muted-foreground">Your booking page: {origin}/book/{handle || "your-handle"}</p>
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="tz">Timezone</Label>
                <Select id="tz" value={timezone} onChange={(e) => setTimezone(e.target.value)}>
                  {zones.map((z) => (
                    <option key={z} value={z}>
                      {z}
                    </option>
                  ))}
                </Select>
                <button
                  type="button"
                  onClick={() => setTimezone(localTimezone())}
                  className="text-xs text-muted-foreground underline-offset-4 hover:underline"
                >
                  Use my timezone ({localTimezone()})
                </button>
              </div>
              <div className="flex items-center gap-3">
                <Button type="submit" disabled={busy}>{busy ? "Saving…" : "Save changes"}</Button>
                {saved && <span className="text-sm text-emerald-500">Saved</span>}
              </div>
            </form>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-base">Account</CardTitle>
          </CardHeader>
          <CardContent className="text-sm text-muted-foreground">
            Signed in as <span className="text-foreground">{data?.user?.email}</span>.
          </CardContent>
        </Card>
      </div>
    </>
  );
}
