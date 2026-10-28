import nodemailer from "nodemailer";
import logo from "./email-logo.json" with { type: "json" };

// The PNG is rendered from the official ZeroAxiis SVG; embed it for email clients.
const logoCID = "zeroaxiis-logo";

export function renderLoginEmail(code: string, sample = false, preview = false) {
  if (!/^\d{6}$/.test(code)) throw new Error("A six-digit login code is required");
  const subject = sample ? "Your HaloMail sample verification code" : "Your HaloMail login code";
  const intro = "Enter this code to finish signing in to HaloMail.";
  const note = sample
    ? "This is a sample email. The code cannot sign you in."
    : "This code expires in 10 minutes and can only be used once. Keep it private.";
  const text = `Your HaloMail sign-in code is ${code}.\n\n${intro}\n\n${note}\n\nIf you didn't request this email, you can ignore it.\n\nHaloMail by ZeroAxiis`;
  const logoSrc = preview ? `data:image/png;base64,${logo.content}` : `cid:${logoCID}`;
  const html = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta name="color-scheme" content="light">
  <title>${subject}</title>
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #f9fafb; padding: 40px 20px; color: #111827; margin: 0;">
  <div style="display:none;max-height:0;overflow:hidden;font-size:1px;line-height:1px;opacity:0;">${sample ? "A preview of your HaloMail sign-in email." : "Your HaloMail sign-in code is ready. It expires in 10 minutes."}</div>
  <div style="max-width: 520px; margin: 0 auto; background: #ffffff; border: 1px solid #e5e7eb; border-radius: 8px; padding: 32px;">
    <h2 style="font-size: 20px; font-weight: 600; margin-top: 0; margin-bottom: 12px; color: #111827;">Your sign-in code</h2>
    <p style="font-size: 15px; line-height: 1.5; color: #374151; margin: 0 0 20px;">${intro}</p>
    <div style="background-color: #f3f4f6; border: 1px solid #e5e7eb; border-radius: 8px; padding: 20px 24px; text-align: center; margin: 24px 0;">
      <span style="font-size: 38px; font-weight: 700; letter-spacing: 8px; color: #111827; display: block; line-height: 1;">${code}</span>
    </div>
    <p style="font-size: 13px; line-height: 1.5; color: #6b7280; margin: 0 0 20px;">${note}</p>
    <p style="font-size: 13px; line-height: 1.5; color: #9ca3af; margin: 0;">If you didn't request this email, you can safely ignore it.</p>
    <hr style="border: none; border-top: 1px solid #e5e7eb; margin: 28px 0 24px;" />
    <div style="text-align: center;">
      <table role="presentation" cellpadding="0" cellspacing="0" border="0" style="margin: 0 auto; text-align: center;">
        <tr>
          <td align="center" style="padding-bottom: 8px;">
            <img src="${logoSrc}" alt="ZeroAxiis" width="30" height="30" style="display: block; width: 30px; height: 30px; border: 0; margin: 0 auto;" />
          </td>
        </tr>
        <tr>
          <td align="center" style="font-size: 13px; font-weight: 500; color: #374151; line-height: 1.5;">
            A product of <a href="https://zeroaxiis.tech" style="color: #111827; font-weight: 600; text-decoration: none;">ZeroAxiis</a>
          </td>
        </tr>
        <tr>
          <td align="center" style="font-size: 11px; color: #9ca3af; font-style: italic; padding-top: 3px;">
            Everything starts from zero.
          </td>
        </tr>
      </table>
    </div>
  </div>
</body>
</html>`;
  return { subject, text, html };
}

export async function sendLoginCode(email: string, code: string, options: { sample?: boolean } = {}) {
  const from = process.env.EMAIL_FROM || "HaloMail <no-reply@email.zeroaxiis.tech>";
  const { subject, text, html } = renderLoginEmail(code, options.sample);
  const apiKey = process.env.RESEND_API_KEY?.trim();
  if (apiKey) {
    const response = await fetch("https://api.resend.com/emails", {
      method: "POST",
      headers: { Authorization: `Bearer ${apiKey}`, "Content-Type": "application/json" },
      body: JSON.stringify({
        from, to: [email], subject, text, html,
        attachments: [{ filename: "zeroaxiis.png", content: logo.content, content_type: "image/png", content_id: logoCID }],
      }),
      signal: AbortSignal.timeout(15000),
      cache: "no-store",
    });
    if (!response.ok) throw new Error(`Resend rejected the login email (${response.status})`);
    const result = await response.json();
    if (typeof result.id !== "string" || !result.id) throw new Error("Resend did not accept the login email");
    return result.id as string;
  }
  if (process.env.NODE_ENV === "production") throw new Error("RESEND_API_KEY is required for login email");
  // Keep Mailpit available for local development without a Resend key.
  const port = Number(process.env.SMTP_PORT || "1025");
  const transporter = nodemailer.createTransport({
    host: process.env.SMTP_HOST || "localhost",
    port,
    secure: process.env.SMTP_SECURE === "true" || port === 465,
    auth: process.env.SMTP_USER ? { user: process.env.SMTP_USER, pass: process.env.SMTP_PASS } : undefined,
    connectionTimeout: 10000,
    greetingTimeout: 10000,
    socketTimeout: 15000,
  });
  const result = await transporter.sendMail({
    from,
    to: email,
    subject,
    text,
    html,
    attachments: [{ filename: "zeroaxiis.png", content: logo.content, encoding: "base64", contentType: "image/png", cid: logoCID }],
  });
  if (!result.accepted.length) throw new Error("Login email was not accepted");
  return result.messageId || "smtp accepted";
}
