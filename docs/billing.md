# Billing operations

Checkout uses published plan pricing from `apps/web/lib/billing-pricing.ts` for Pro and Business, monthly and yearly. Both display and server-side order amounts use `apps/web/lib/billing-pricing.ts`.

## Deployment

1. Apply identity migrations through `00005_billing_refunds.sql` using the existing migration workflow, then deploy the Go backend and Next.js web app together. Payment activation now reads the refund ledger, so apply the migration before deploying the backend.
2. On the web app set `API_BASE_URL` to the Go backend URL, `RAZORPAY_KEY_ID`, `RAZORPAY_KEY_SECRET`, and `RAZORPAY_WEBHOOK_SECRET`. Existing `RAZERPAY_LIVE_API_KEY` and `RAZERPAY_LIVE_KEY_SECRET` spellings remain supported. The checkout uses the public key returned with the order.
3. Set the same `RAZORPAY_WEBHOOK_SECRET` in the backend and in the Razorpay webhook configuration. The OTP secret is not a billing signature.
4. On the **web app**, also set `R2_ENDPOINT`, `R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY`, `R2_BUCKET_NAME`, `R2_REGION=auto`, and `R2_INVOICE_PREFIX=halomail/invice`. Credentials need object read/write access. Backend backup credentials alone do not configure the web app.
5. Keep the existing webhook URL `https://halomail.zeroaxiis.tech/api/razorpay/webhook` and enable `payment.captured`, `order.paid`, and `refund.processed`. Payments must reach captured status. Refunds created in the Razorpay dashboard use the same webhook.

Invoices are PDF downloads at `halomail/invice/<org-id>/<user-id>/invoice_<payment-id>.pdf`, with an archived HTML copy alongside each PDF and the existing `halomail/production` backup prefix. If `halomail` is only the bucket name and you want `invice` at its root, set `R2_INVOICE_PREFIX=invice`. The bucket remains private. History prefers PDF over HTML, includes existing invoices under `invoices/<org-id>/`, and reads all R2 pages. A retry repairs a missing PDF even when an HTML archive already exists.

## Processing and recovery

Invoice emails must always come from `noreply@nts.zeroaxiis.tech`, using the verified Resend domain `nts.zeroaxiis.tech`. Set `NTS_RESEND_API_KEY` on the web app with the key authorized for that domain, and in the root `.env` for the sample invoice script. Invoice sending uses this dedicated key rather than the shared `RESEND_API_KEY`. The billing sender and sample invoice script use `BILLING_EMAIL_FROM` in `apps/web/lib/billing-mail.ts`; the shared `EMAIL_FROM` setting cannot override it.

The invoice identifies the seller as HaloMail, Gorakhpur - 273004, with the ZeroAxiis logo. No GSTIN, assumed GST rate, or digital-signature claim is included. The captured amount is passed explicitly in minor units; the PDF and HTML use the same invoice details. Emails contain a compact receipt summary, plain-text alternative, PDF attachment, and an authenticated link to the saved PDF. The full printable HTML is never embedded in the email.

Checkout verifies the Razorpay signature, fetches payment and order from Razorpay, checks captured/paid state and matching amount/currency, updates the account through the backend, then archives its invoice. Webhooks use the same flow. Database or R2 failures return non-2xx so delivery can retry. A payment ledger and stable invoice key prevent duplicate entitlement and invoice creation.

For an existing payment, sign in to its original HaloMail account, open **Billing**, enter its `pay_...` payment ID and click **Sync payment**. This verifies the existing purchase; it does not charge again. Missing legacy `org_id` is resolved from the recorded `user_id` in the database. Orders missing a valid user or plan need support investigation. Old orders did not record monthly/yearly selection, so recovery defaults those to monthly.

Pro allowances are 1,000 submissions and 100 meetings; Business allowances are 2,500 submissions and 250 meetings. Existing usage is retained. Access expires at the recorded end of the purchased monthly/yearly term. Replaying an old payment cannot extend its end or replace a newer plan.

This integration purchases a fixed term through Razorpay Orders; it does not set up automatic recurring debits.

## Refunds

After approving a refund under the existing refund policy, issue it in Razorpay. HaloMail waits for Razorpay's `refund.processed` event, verifies its signature, and fetches the refund, payment, order, and all refund-history pages from Razorpay before updating the account. This integration does not initiate refunds or change refund eligibility.

Only a **full completed refund** ends access for the refunded purchase. A partial refund keeps the plan; multiple completed partial refunds count as full once their total equals the original payment amount. Pending and failed refunds do not count. See [Razorpay's refund webhook events](https://razorpay.com/docs/webhooks/refunds/).

When the refunded purchase is the account's active purchase, the account switches to Free and its paid expiry is cleared. Stored data and usage counts remain intact, and Free limits apply. Refunding an older purchase does not remove a newer subscription. The dashboard refreshes plan and usage when the user returns to the browser window or reloads the page.

The refund ledger records each refund ID once. Duplicate, out-of-order, and concurrent payment/refund callbacks cannot double-count refunds or reactivate the fully refunded purchase. Failed account updates return an error so Razorpay can retry. To recover a missed refund event, use **Sync payment** for the original payment ID; it reconciles completed refunds before attempting activation and does not issue a new invoice for a fully refunded purchase. Refunds completed before enabling the event require this recovery step or a webhook replay.

## Verification

Run `npm run test:billing` and `npx tsc --noEmit` in `apps/web`, and `go test ./services/identity/... ./services/shared/usage/...` from the root. Database integration tests require an isolated `TEST_DATABASE_URL`.

For a local design preview, run `node --experimental-strip-types scripts/test-invoice-gen.mjs` in `apps/web`. This writes the PDF, printable HTML, and email preview to `output/pdf/`. After reviewing those files, `npm run sample:invoice` sends that saved PDF to the recipient in `scripts/sample-invoice-data.mjs`. The sample is explicitly marked as a preview and does not represent a recorded payment.

After deployment, use a test-mode or staging purchase to confirm the plan and quota change, download the invoice, and resend both payment webhook events to verify one invoice and unchanged expiry. Verify that a processed partial refund keeps the plan and a completed full refund switches the active purchase to Free. Replay the refund and original payment callbacks to confirm the account stays Free; also verify that refunding an older purchase preserves a newer plan. Simulate an unavailable R2 store only in staging and retry after restoring it.
