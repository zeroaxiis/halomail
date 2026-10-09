# Neon database with two hourly R2 backup batches

## Design

Neon is the application's **only active database**. All reads, inserts, updates,
deletes and background jobs use `DATABASE_URL`. R2 stores recovery files only.
There is no second live database or automatic database switching.

One full, consistent PostgreSQL backup is taken every hour. Each batch is
independently restorable and includes all application tables and schemas:
users, organizations, sessions, API keys, audit logs, forms, messages, bookings,
calendars, templates, billing, usage, delivery history and queued work.
New tables are included automatically. Updates/deletions appear in the next batch.

Retention is fixed at **two completed, verified batches**:

| Time | New backup | Batches kept |
| --- | --- | --- |
| 10:00 | Backup 1 | 1 |
| 11:00 | Backup 2 | 1, 2 |
| 12:00 | Backup 3 | 2, 3; backup 1 deleted |
| 13:00 | Backup 4 | 3, 4; backup 2 deleted |

The worker uploads a compressed archive, downloads it to verify its SHA-256 and
size, then publishes a JSON manifest. Only after verifying the two newest batches
does it remove older archives and manifests. Two batches normally mean four R2
objects: two archives and two small manifests.

There is briefly a third copy during upload/verification. A failed backup never
replaces a good one. Cleanup failures temporarily leave extra copies and are
retried after five minutes. If a retained batch is corrupt, older copies are
preserved. Abandoned uploads are removed only when older than both retained
batches and twice the operation timeout. Other objects/prefixes are left alone.

Startup checks R2: if no batch exists, the first backup is immediate. Otherwise
the next run is due one hour after the latest snapshot's start time. Restarts do
not reset this schedule or create extra batches. The hour is between starts,
not one hour after upload completion. Failures retry after five minutes. Long
backups do not overlap within a worker and can delay the next run. Missed
snapshots cannot be recreated retroactively.

## Configure the backend

Use a private R2 bucket and an R2 S3 token with Object Read & Write permission
scoped to that bucket, including deletion for rotation. Disable public development
access and public custom domains. Use the Access Key ID and Secret Access Key,
not the raw Cloudflare API token value.

Set these in the backend environment and local root `.env`:

```dotenv
DATABASE_URL=postgresql://YOUR_NEON_CONNECTION
R2_BACKUP_ENABLED=true
R2_ENDPOINT=https://YOUR_ACCOUNT.r2.cloudflarestorage.com
R2_ACCESS_KEY_ID=YOUR_ACCESS_KEY_ID
R2_SECRET_ACCESS_KEY=YOUR_SECRET_ACCESS_KEY
R2_BUCKET_NAME=YOUR_PRIVATE_BUCKET
R2_REGION=auto
R2_BACKUP_PREFIX=halomail/production
R2_BACKUP_INTERVAL=1h
R2_BACKUP_TIMEOUT=15m
R2_BACKUP_MAX_AGE=2h
```

Use the exact endpoint for the bucket's account/jurisdiction. Give each unrelated
application/environment/database its own prefix. Old batches from the previous
implementation in this prefix will be pruned after verifying the two newest
batches. Credentials belong in backend secrets, not Git or browser variables.

`BACKUP_DATABASE_URL` optionally supplies Neon's direct/unpooled connection for
dumps. It must point to the **same database** as `DATABASE_URL`; leave it empty
to use `DATABASE_URL`.

`deploy/Dockerfile` includes PostgreSQL 18 clients and the `/backup` command.
Local execution needs `pg_dump` and `pg_restore` on PATH or `PG_DUMP_PATH` and
`PG_RESTORE_PATH`. The dump client must be at least as new as the source server.
The database role must be allowed to read all data. Temporary disk space must
hold one compressed archive. Partial R2 configuration or missing tools is logged
and reported through `/backupz`; the application continues using Neon. Correct
the backup setup and restart the worker to enable protection.

Run **one backup worker per prefix**, either gateway or standalone, including
across hosts/replicas. For a standalone worker, set `R2_BACKUP_ENABLED=false` in
the gateway environment and `true` in the worker's separate environment.
Only the gateway starts one automatically; split services need the standalone
worker. Keep it always on: a sleeping or scale-to-zero host cannot run hourly
jobs. Restart/redeploy after environment changes.

Disable lifecycle expiration on these R2 prefixes if it can remove the two
retained batches independently. The application handles two-batch rotation.

## Commands and monitoring

Run from the repository root to load `.env`:

```powershell
# List completed batches, newest first
go run ./services/shared/cmd/backup -action list

# Verify the latest batch without connecting to Neon
go run ./services/shared/cmd/backup -action verify

# Force an immediate backup and rotate to the newest two
go run ./services/shared/cmd/backup -action snapshot

# Run the hourly worker separately from the gateway
go run ./services/shared/cmd/backup -action run
```

Manual snapshots use the same retention and reset the next hourly due time.
Do not run them concurrently with another worker for the same prefix.
The container exposes the equivalent CLI:

```sh
docker run --rm --env-file .env --entrypoint /backup halomail-api -action list
```

Monitor gateway `/backupz`: it returns 503 if backups are disabled, misconfigured,
or unavailable, before a batch is verified, when the
latest snapshot exceeds `R2_BACKUP_MAX_AGE`, or when cleanup fails. On restart
the worker verifies existing backups for health. `/readyz` still checks Neon.
Monitor worker logs and process uptime as well.

## Recover into a replacement Neon database

1. **Stop the gateway, all background writers and the backup worker.** Otherwise
   empty/damaged snapshots could rotate away the good recovery batches.
2. Create a new, empty Neon database with the same or a newer compatible
   PostgreSQL version and required extensions. Do not run application migrations
   or start the app against this empty database yet.
3. Keep the original R2 credentials/bucket/prefix in your recovery environment,
   with `R2_BACKUP_ENABLED=true` for the CLI. Set `RESTORE_DATABASE_URL` to the
   new database's private URL. The old source database can be offline.
4. Restore the newest completed batch:

   ```powershell
   go run ./services/shared/cmd/backup -action restore
   ```

   `-manifest latest` is the default. The command prints the chosen batch,
   verifies the archive, refuses a nonempty target, and restores in a single
   transaction without destructive replacement. Failure rolls back the restore.
   It does not silently fall back if the latest batch is corrupt. To explicitly
   recover the older batch, choose its exact key from `-action list`:

   ```powershell
   go run ./services/shared/cmd/backup -action restore -manifest "halomail/production/manifests/EXACT-KEY.json"
   ```

5. Check restored records and application behavior. Reconcile queued email,
   webhooks and calendar work: some external actions may already have happened
   after the snapshot, so replay could duplicate them. Preserve the original
   JWT and calendar encryption secrets.
6. Update backend `DATABASE_URL` everywhere to the replacement Neon connection.
   Update or clear `BACKUP_DATABASE_URL` too, so future backups use the new
   database. Clear `RESTORE_DATABASE_URL` after recovery.
7. Restart the app and its single worker. Confirm `/readyz`, `/backupz` and
   the next hourly batch. All normal operations now use the replacement Neon.

Restore and verification are read-only against R2 and never rotate backups.
Keep other writers disconnected from the target throughout restoration.
A nonempty target is never overwritten.

## Recovery limits

Recovery restores data at the selected snapshot. Later changes are not in that
batch. Normally the possible loss window is about one hour plus backup time;
failed backups or worker downtime make it longer. During a Neon outage, service
waits for Neon recovery or restoration and connection of a replacement database.

Two recent recovery points do not provide days of history. Corruption/deletions
already present in both batches cannot be undone with these backups. A deletion
appears in the next batch; older batches keep earlier data until rotation.

Archives contain database rows, schemas, sequences, functions, indexes,
constraints and PostgreSQL large objects. They exclude global roles/passwords,
ownership/grants and replication infrastructure. They do not back up Redis,
in-memory OTP/rate limits, `.env`, files outside PostgreSQL, remote foreign-table
data or data held by Google Calendar/Resend/Razorpay. Preserve deployment and
encryption secrets separately in your secret manager.

References:
- [PostgreSQL pg_dump](https://www.postgresql.org/docs/current/app-pgdump.html)
- [PostgreSQL pg_restore](https://www.postgresql.org/docs/current/app-pgrestore.html)
- [R2 credentials](https://developers.cloudflare.com/r2/api/tokens/)
