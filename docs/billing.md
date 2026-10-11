# Billing operations

The current checkout intentionally charges **1 unit of the selected currency** for Pro and Business, monthly and yearly, while testing. Both display and server-side order amounts use `apps/web/lib/billing-pricing.ts`.

## Deployment

1. Apply identity migration `00004_billing_payments.sql` using the existing migration workflow, then deploy the Go backend and Next.js web app together.
2. On the web app set `API_BASE_URL` to the Go backend URL, `RAZORPAY_KEY_ID`, `RAZORPAY_KEY_SECRET`, and `RAZORPAY_WEBHOOK_SECRET`. Existing `RAZERPAY_LIVE_API_KEY` and `RAZERPAY_LIVE_KEY_SECRET` spellings remain supported. The checkout uses the public key returned with the order.
3. Set the same `RAZORPAY_WEBHOOK_SECRET` in the backend and in the Razorpay webhook configuration. The OTP secret is not a billing signature.
4. On the **web app**, also set `R2_ENDPOINT`, `R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY`, `R2_BUCKET_NAME`, `R2_REGION=auto`, and `R2_INVOICE_PREFIX=halomail/invice`. Credentials need object read/write access. Backend backup credentials alone do not configure the web app.
5. Keep the existing webhook URL `https://halomail.zeroaxiis.tech/api/razorpay/webhook` and both events `payment.captured` and `order.paid`. Payments must reach captured status.

Invoices are HTML downloads at `halomail/invice/<org-id>/<user-id>/invoice_<payment-id>.html`, alongside the existing `halomail/production` backup prefix. If `halomail` is only the bucket name and you want `invice` at its root, set `R2_INVOICE_PREFIX=invice`. The bucket remains private. History also includes existing invoices under `invoices/<org-id>/` and reads all R2 pages.

## Processing and recovery

Checkout verifies the Razorpay signature, fetches payment and order from Razorpay, checks captured/paid state and matching amount/currency, updates the account through the backend, then archives its invoice. Webhooks use the same flow. Database or R2 failures return non-2xx so delivery can retry. A payment ledger and stable invoice key prevent duplicate entitlement and invoice creation.

For an existing payment, sign in to its original HaloMail account, open **Billing**, enter its `pay_...` payment ID and click **Sync payment**. This verifies the existing purchase; it does not charge again. Missing legacy `org_id` is resolved from the recorded `user_id` in the database. Orders missing a valid user or plan need support investigation. Old orders did not record monthly/yearly selection, so recovery defaults those to monthly.

Pro allowances are 1,000 submissions and 100 meetings; Business allowances are 2,500 submissions and 250 meetings. Existing usage is retained. Access expires at the recorded end of the purchased monthly/yearly term. Replaying an old payment cannot extend its end or replace a newer plan.

This integration purchases a fixed term through Razorpay Orders; it does not set up automatic recurring debits. Subscription lifecycle and refund automation are outside this order-payment flow.

## Verification

Run `npm run test:billing` and `npx tsc --noEmit` in `apps/web`, and `go test ./services/identity/... ./services/shared/usage/...` from the root. Database integration tests require an isolated `TEST_DATABASE_URL`.

After deployment, make one 1-unit purchase, confirm the plan and quota change, download the invoice, and resend both webhook events to verify one invoice and unchanged expiry. Simulate an unavailable R2 store only in staging and retry after restoring it.
