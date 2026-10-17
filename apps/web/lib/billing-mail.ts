import nodemailer from "nodemailer";

// Invoice mail always uses this sender and its dedicated Resend key.
export const BILLING_EMAIL_FROM = "noreply@nts.zeroaxiis.tech";

export interface BillingEmailDetails {
  invoiceNumber: string;
  customerName: string;
  total: string;
  dateLabel: string;
  sample?: boolean;
  downloadURL?: string;
}

const escapeHTML = (value: string) => value.replace(/[&<>"']/g, c => ({
  "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;",
}[c]!));

/** Small table-based email; the complete invoice lives in the PDF attachment. */
export function renderBillingEmail(details: BillingEmailDetails) {
  const e = escapeHTML;
  let downloadURL = "";
  if (details.downloadURL) {
    const url = new URL(details.downloadURL);
    if (!["http:", "https:"].includes(url.protocol)) throw new Error("Invalid invoice download URL");
    downloadURL = url.href;
  }
  const subject = `${details.sample ? "Preview: " : ""}Your HaloMail invoice - ${details.invoiceNumber}`;
  const heading = details.sample ? "Your invoice preview" : "Your invoice is ready";
  const intro = details.sample
    ? "Here is the revised invoice design. The PDF is attached for you to open or download."
    : "Thank you for your payment. Your invoice is attached as a PDF for your records.";
  const note = details.sample
    ? "This is a sample preview. No payment has been recorded."
    : "You can open or download the PDF using the attachment below.";
  const text = `Hello ${details.customerName},\n\n${intro}\n\nInvoice: ${details.invoiceNumber}\nDate: ${details.dateLabel}\n${details.sample ? "Preview total" : "Amount paid"}: ${details.total}\n\n${note}${downloadURL ? `\n\nDownload invoice (sign in to HaloMail): ${downloadURL}` : ""}\n\nHaloMail by ZeroAxiis\nGorakhpur - 273004`;
  const html = `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>${e(subject)}</title></head>
<body style="margin:0;padding:0;background-color:#f4f4f4;font-family:Arial,Helvetica,sans-serif;color:#202124;">
<div style="display:none;max-height:0;overflow:hidden;mso-hide:all;">Your HaloMail invoice PDF is attached. ${e(details.total)}.</div>
<table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0" style="background-color:#f4f4f4;"><tr><td align="center" style="padding:32px 12px;">
<table role="presentation" width="560" cellspacing="0" cellpadding="0" border="0" style="width:100%;max-width:560px;background-color:#ffffff;border:1px solid #dedede;">
<tr><td style="padding:28px 28px 22px;border-bottom:1px solid #dedede;"><span style="font-size:24px;font-weight:bold;">HaloMail</span><br><span style="font-size:12px;color:#62666b;">by ZeroAxiis</span></td></tr>
<tr><td style="padding:28px;"><h1 style="margin:0 0 20px;font-size:24px;line-height:32px;">${heading}</h1><p style="margin:0 0 14px;font-size:14px;line-height:22px;">Hello ${e(details.customerName)},</p><p style="margin:0 0 22px;font-size:14px;line-height:22px;">${intro}</p>
<table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0" style="font-size:13px;line-height:20px;border-top:1px solid #dedede;border-bottom:1px solid #dedede;"><tr><td style="padding:16px 0 5px;color:#62666b;vertical-align:top;width:105px;">Invoice</td><td style="padding:16px 0 5px;word-break:break-all;">${e(details.invoiceNumber)}</td></tr><tr><td style="padding:5px 0;color:#62666b;">Date</td><td style="padding:5px 0;">${e(details.dateLabel)}</td></tr><tr><td style="padding:5px 0 16px;color:#62666b;">${details.sample ? "Preview total" : "Amount paid"}</td><td style="padding:5px 0 16px;font-weight:bold;">${e(details.total)}</td></tr></table>
<p style="margin:22px 0 0;font-size:14px;line-height:22px;">${note}</p>
${downloadURL ? `<table role="presentation" cellspacing="0" cellpadding="0" border="0" style="margin-top:22px;"><tr><td bgcolor="#202124" style="padding:13px 22px;"><a href="${e(downloadURL)}" style="color:#ffffff;text-decoration:none;font-size:14px;font-weight:bold;display:inline-block;">Download invoice</a></td></tr></table><p style="margin:12px 0 0;font-size:12px;line-height:18px;color:#62666b;">Sign in to your HaloMail account to download your saved copy.</p>` : ""}
</td></tr><tr><td style="padding:20px 28px;border-top:1px solid #dedede;font-size:12px;line-height:19px;color:#62666b;">HaloMail by ZeroAxiis<br>Gorakhpur - 273004</td></tr></table>
</td></tr></table></body></html>`;
  return { subject, html, text };
}

// Receipt delivery must never prevent account activation or invoice storage.
export async function sendBillingReceipt(to: string, paymentId: string, pdfBuffer: Buffer, details: BillingEmailDetails): Promise<string | undefined> {
  const apiKey = process.env.NTS_RESEND_API_KEY?.trim();
  const smtpHost = process.env.SMTP_HOST;
  if (!to || (!apiKey && !smtpHost)) return;
  const { subject, html, text } = renderBillingEmail(details);
  const filename = `HaloMail_Invoice_${paymentId}.pdf`;
  if (apiKey) {
    try {
      const response = await fetch("https://api.resend.com/emails", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${apiKey}`, "Content-Type": "application/json",
          "Idempotency-Key": `invoice-v2/${paymentId}`,
        },
        body: JSON.stringify({ from: BILLING_EMAIL_FROM, to: [to], subject, html, text,
          attachments: [{ filename, content: pdfBuffer.toString("base64"), content_type: "application/pdf" }],
        }),
        signal: AbortSignal.timeout(15000), cache: "no-store",
      });
      if (!response.ok) throw new Error(`Resend rejected invoice email (${response.status})`);
      const result = await response.json();
      if (typeof result.id !== "string" || !result.id) throw new Error("Resend did not confirm invoice acceptance");
      return result.id;
    } catch (error) {
      // Avoid duplicate sends after an ambiguous API response.
      console.error("Invoice email failed; the PDF remains available in Billing", error);
      return;
    }
  }
  const transporter = nodemailer.createTransport({
    host: smtpHost, port: Number(process.env.SMTP_PORT || "1025"), secure: process.env.SMTP_PORT === "465",
    connectionTimeout: 5000, greetingTimeout: 5000, socketTimeout: 5000,
    auth: process.env.SMTP_USER ? { user: process.env.SMTP_USER, pass: process.env.SMTP_PASS } : undefined,
  });
  try {
    const result = await transporter.sendMail({
      from: BILLING_EMAIL_FROM, to, subject, html, text,
      attachments: [{ filename, content: pdfBuffer, contentType: "application/pdf", contentDisposition: "attachment" }],
    });
    return result.messageId;
  } catch (error) { console.error("Invoice email failed; the PDF remains available in Billing", error); }
  finally { transporter.close(); }
}
