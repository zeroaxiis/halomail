import nodemailer from "nodemailer";

// Delivery is optional and must never prevent account activation or invoice storage.
export async function sendBillingReceipt(to: string, paymentId: string, html: string) {
  if (!to || (!process.env.RESEND_API_KEY && !process.env.SMTP_HOST)) return;
  const useResend = !!process.env.RESEND_API_KEY;
  const transporter = nodemailer.createTransport({
    host: useResend ? "smtp.resend.com" : process.env.SMTP_HOST,
    port: useResend ? 465 : Number(process.env.SMTP_PORT || "1025"),
    secure: useResend || process.env.SMTP_PORT === "465",
    connectionTimeout: 5000, greetingTimeout: 5000, socketTimeout: 5000,
    auth: useResend ? { user: "resend", pass: process.env.RESEND_API_KEY } :
      process.env.SMTP_USER ? { user: process.env.SMTP_USER, pass: process.env.SMTP_PASS } : undefined,
  });
  try {
    await transporter.sendMail({
      from: process.env.EMAIL_FROM || '"HaloMail Billing" <no-reply@email.zeroaxiis.tech>',
      to, subject: `Your HaloMail receipt [${paymentId}]`, html,
    });
  } catch (error) { console.error("Receipt email failed; invoice remains available in Billing", error); }
  finally { transporter.close(); }
}
