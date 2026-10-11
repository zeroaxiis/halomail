import { NextRequest, NextResponse } from "next/server";
import { ListObjectsV2Command } from "@aws-sdk/client-s3";
import { billingError, billingUser, invoicePrefixes, invoiceStorage, ownsInvoice } from "@/lib/billing-server";
export async function GET(req: NextRequest) {
  try {
    const user = await billingUser(req);
    const { client, bucket } = invoiceStorage();
    const invoices = new Map<string, { key: string; id: string; date: string }>();
    for (const prefix of invoicePrefixes(user)) {
      let token: string | undefined;
      do {
        const data = await client.send(new ListObjectsV2Command({ Bucket: bucket, Prefix: prefix, ContinuationToken: token }));
        for (const object of data.Contents || []) {
          if (!object.Key || !ownsInvoice(object.Key, user)) continue;
          const id = object.Key.split("/").pop()!.replace(".html", "");
          if (!invoices.has(id)) invoices.set(id, { key: object.Key, id, date: object.LastModified?.toISOString() || "" });
        }
        token = data.IsTruncated ? data.NextContinuationToken : undefined;
      } while (token);
    }
    return NextResponse.json({ invoices: [...invoices.values()].sort((a, b) => b.date.localeCompare(a.date)) }, { headers: { "Cache-Control": "no-store" } });
  } catch (error) { return billingError(error); }
}
