import { NextRequest, NextResponse } from "next/server";
import { GetObjectCommand } from "@aws-sdk/client-s3";
import { BillingError, billingError, billingUser, invoiceStorage, ownsInvoice } from "@/lib/billing-server";
export async function GET(req: NextRequest) {
  try {
    const user = await billingUser(req);
    const key = req.nextUrl.searchParams.get("key") || "";
    if (!ownsInvoice(key, user)) throw new BillingError("Invoice not found.", 404);
    const { client, bucket } = invoiceStorage();
    const data = await client.send(new GetObjectCommand({ Bucket: bucket, Key: key }));
    if (!data.Body) throw new BillingError("Invoice not found.", 404);
    return new NextResponse(new Uint8Array(await data.Body.transformToByteArray()).buffer, { headers: {
      "Content-Type": "text/html; charset=utf-8", "Content-Disposition": `attachment; filename="${key.split("/").pop()}"`,
      "Cache-Control": "private, no-store", "X-Content-Type-Options": "nosniff", "Content-Security-Policy": "sandbox",
    } });
  } catch (error) { return billingError(error); }
}
