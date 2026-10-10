import { NextRequest, NextResponse } from "next/server";
import nodemailer from "nodemailer";
import crypto from "crypto";
import { S3Client, PutObjectCommand } from "@aws-sdk/client-s3";

export async function POST(req: NextRequest) {
  try {
    const rawBody = await req.text();
    const signature = req.headers.get("x-razorpay-signature");
    const secret = process.env.RAZORPAY_WEBHOOK_SECRET;

    // Verify webhook signature if a secret is configured
    if (secret && signature) {
      const expectedSignature = crypto
        .createHmac("sha256", secret)
        .update(rawBody)
        .digest("hex");

      if (expectedSignature !== signature) {
        console.error("Invalid Razorpay webhook signature");
        return NextResponse.json({ error: "Invalid signature" }, { status: 400 });
      }
    }

    const body = JSON.parse(rawBody);

    // Listen for payment captured or order paid events
    if (body.event === "payment.captured" || body.event === "order.paid") {
      const payment = body.payload.payment.entity;
      
      const { email, name, product, tier, account_id, user_id } = payment.notes || {};
      const amount = (payment.amount / 100).toFixed(2);
      const currency = payment.currency;

      if (!email) {
        console.warn("Webhook received payment without email in notes:", payment.id);
        return NextResponse.json({ status: "skipped", reason: "no email in notes" });
      }

      // Configure nodemailer using SMTP. Defaults to local Mailpit (port 1025) for dev
      const transporter = nodemailer.createTransport({
        host: process.env.SMTP_HOST || "localhost",
        port: parseInt(process.env.SMTP_PORT || "1025", 10),
        secure: process.env.SMTP_PORT === "465",
        auth: process.env.SMTP_USER ? {
          user: process.env.SMTP_USER,
          pass: process.env.SMTP_PASS,
        } : undefined,
      });

      const invoiceHTML = `
        <div style="font-family: 'Inter', Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 24px; border: 1px solid #e5e7eb; border-radius: 8px;">
          <h2 style="color: #111827; margin-bottom: 24px;">Payment Receipt / Invoice</h2>
          <p style="color: #374151;">Hi ${name || "Customer"},</p>
          <p style="color: #374151;">Thank you for your payment to <strong>zeroaxiis</strong>.</p>
          
          <div style="background-color: #f9fafb; padding: 16px; border-radius: 6px; margin: 24px 0;">
            <table style="width: 100%; text-align: left; border-collapse: collapse;">
              <tr>
                <th style="padding: 8px 0; color: #6b7280; font-weight: 500; font-size: 14px;">Product:</th>
                <td style="padding: 8px 0; color: #111827; font-weight: 500;">${product || "halomail"} - ${tier ? tier.toUpperCase() : "Subscription"}</td>
              </tr>
              <tr>
                <th style="padding: 8px 0; color: #6b7280; font-weight: 500; font-size: 14px;">Amount Paid:</th>
                <td style="padding: 8px 0; color: #111827; font-weight: 500;">${currency} ${amount}</td>
              </tr>
              <tr>
                <th style="padding: 8px 0; color: #6b7280; font-weight: 500; font-size: 14px;">Payment ID:</th>
                <td style="padding: 8px 0; color: #111827; font-size: 14px;">${payment.id}</td>
              </tr>
              <tr>
                <th style="padding: 8px 0; color: #6b7280; font-weight: 500; font-size: 14px;">Account ID:</th>
                <td style="padding: 8px 0; color: #111827; font-size: 14px;">${account_id || "N/A"}</td>
              </tr>
              <tr>
                <th style="padding: 8px 0; color: #6b7280; font-weight: 500; font-size: 14px;">User ID:</th>
                <td style="padding: 8px 0; color: #111827; font-size: 14px;">${user_id || "N/A"}</td>
              </tr>
            </table>
          </div>
          
          <p style="font-size: 12px; color: #9ca3af; margin-top: 32px; border-top: 1px solid #e5e7eb; padding-top: 16px;">
            This is an automated receipt generated upon successful payment.<br />
            If you have any questions, please contact <a href="mailto:support@zeroaxiis.tech" style="color: #2563eb;">support@zeroaxiis.tech</a>.
          </p>
        </div>
      `;

      await transporter.sendMail({
        from: '"zeroaxiis Billing" <noreply@nts.email.zeroaxiis.tech>',
        to: email,
        subject: `Your receipt for ${product || "halomail"} [${payment.id}]`,
        html: invoiceHTML,
      });

      // Upload Invoice to R2
      if (account_id && process.env.R2_ACCESS_KEY_ID) {
        try {
          const s3 = new S3Client({
            region: process.env.R2_REGION || "auto",
            endpoint: process.env.R2_ENDPOINT,
            credentials: {
              accessKeyId: process.env.R2_ACCESS_KEY_ID,
              secretAccessKey: process.env.R2_SECRET_ACCESS_KEY || "",
            },
          });
          const key = `invoices/${account_id}/invoice_${payment.id}.html`;
          await s3.send(new PutObjectCommand({
            Bucket: process.env.R2_BUCKET_NAME || "halomail",
            Key: key,
            Body: invoiceHTML,
            ContentType: "text/html",
          }));
        } catch (err) {
          console.error("Failed to upload invoice to R2:", err);
        }
      }

      // Forward the exact raw webhook to the Go backend so it updates the account tier in the database
      const backendUrl = process.env.PUBLIC_API_URL || "http://localhost:8080";
      try {
        await fetch(`${backendUrl}/v1/billing/razorpay/webhook`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "x-razorpay-signature": signature || ""
          },
          body: rawBody
        });
      } catch (err) {
        console.error("Failed to forward webhook to Go backend:", err);
      }

      return NextResponse.json({ status: "success", message: "Invoice sent and backend updated" });
    }

    return NextResponse.json({ status: "ignored" });
  } catch (error: any) {
    console.error("Razorpay Webhook Error:", error);
    return NextResponse.json(
      { error: "Webhook processing failed" },
      { status: 500 }
    );
  }
}
