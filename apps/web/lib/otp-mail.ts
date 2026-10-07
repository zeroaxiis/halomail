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
  <meta name="supported-color-schemes" content="light">
  <title>${subject}</title>
  <style>
    @media screen and (max-width: 600px) {
      .email-shell { padding: 16px 8px !important; }
      .email-header { padding: 28px 26px 0 !important; }
      .email-content { padding: 36px 26px 32px !important; }
      .email-heading { font-size: 34px !important; }
      .code-content { padding: 22px !important; }
      .email-code { font-size: 40px !important; letter-spacing: 5px !important; }
      .email-footer { padding: 24px 26px !important; }
    }
  </style>
</head>
<body style="margin:0;padding:0;background:#eeede9;color:#1a1c20;font-family:Helvetica,Arial,sans-serif;">
  <div style="display:none;max-height:0;overflow:hidden;font-size:1px;line-height:1px;opacity:0;">${sample ? "A preview of your HaloMail sign-in email." : "Your HaloMail sign-in code is ready. It expires in 10 minutes."}</div>
  <table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0" bgcolor="#eeede9" style="background:#eeede9;">
    <tr><td class="email-shell" align="center" style="padding:48px 16px;">
      <table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0" bgcolor="#f5f1e8" style="max-width:560px;background:#f5f1e8;border:1px solid #dcd8ce;">
        <tr><td class="email-header" style="padding:36px 42px 0;">
          <table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0">
            <tr>
              <td valign="middle"><img src="${logoSrc}" alt="ZeroAxiis" width="52" height="52" style="display:block;width:52px;height:52px;border:0;" /></td>
              <td align="right" valign="middle" style="font-family:Helvetica,Arial,sans-serif;font-size:17px;font-weight:600;color:#1a1c20;">HaloMail</td>
            </tr>
            <tr><td colspan="2" style="height:28px;border-bottom:1px solid #d4cfc4;font-size:1px;line-height:1px;">&nbsp;</td></tr>
          </table>
        </td></tr>
        <tr><td class="email-content" style="padding:42px 42px 36px;">
          <p style="margin:0 0 14px;color:#656157;font-size:13px;line-height:1.5;">Account access</p>
          <h1 class="email-heading" style="margin:0 0 18px;color:#1a1c20;font-family:Helvetica,Arial,sans-serif;font-size:42px;font-weight:700;line-height:1.08;letter-spacing:-1.4px;">Your sign-in code.</h1>
          <p style="margin:0 0 30px;color:#55534e;font-size:16px;line-height:1.6;">${intro}</p>
          <table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0" bgcolor="#e9e4da" style="background:#e9e4da;">
            <tr><td class="code-content" style="padding:26px 30px;">
              <p style="margin:0 0 10px;color:#656157;font-size:12px;line-height:1.5;">Verification code</p>
              <p class="email-code" style="margin:0;color:#1a1c20;font-family:Helvetica,Arial,sans-serif;font-size:50px;font-weight:600;line-height:1.15;letter-spacing:7px;font-variant-numeric:tabular-nums;">${code}</p>
            </td></tr>
            <tr><td bgcolor="#ffcc00" style="height:3px;background:#ffcc00;font-size:1px;line-height:1px;">&nbsp;</td></tr>
          </table>
          <p style="margin:20px 0 0;color:#656157;font-size:13px;line-height:1.65;">${note}</p>
        </td></tr>
        <tr><td class="email-footer" style="padding:24px 42px;border-top:1px solid #d4cfc4;color:#777268;font-size:12px;line-height:1.65;">If you didn't request this email, you can ignore it.</td></tr>
      </table>
      <p style="margin:20px 0 0;color:#706d65;font-family:Helvetica,Arial,sans-serif;font-size:12px;line-height:1.6;">HaloMail by <a href="https://zeroaxiis.tech" style="color:#48463f;text-decoration:underline;">ZeroAxiis</a></p>
    </td></tr>
  </table>
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
