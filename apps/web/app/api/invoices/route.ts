import { NextRequest, NextResponse } from "next/server";
import { S3Client, ListObjectsV2Command } from "@aws-sdk/client-s3";

export async function GET(req: NextRequest) {
  const orgId = req.nextUrl.searchParams.get("orgId");
  if (!orgId) return NextResponse.json({ error: "orgId required" }, { status: 400 });

  try {
    const client = new S3Client({
      region: process.env.R2_REGION || "auto",
      endpoint: process.env.R2_ENDPOINT,
      credentials: {
        accessKeyId: process.env.R2_ACCESS_KEY_ID || "",
        secretAccessKey: process.env.R2_SECRET_ACCESS_KEY || "",
      },
    });
    
    const data = await client.send(new ListObjectsV2Command({
      Bucket: process.env.R2_BUCKET_NAME || "halomail",
      Prefix: `invoices/${orgId}/`,
    }));

    const invoices = (data.Contents || []).map(obj => ({
      key: obj.Key,
      id: obj.Key?.split("/").pop()?.replace(".html", ""),
      date: obj.LastModified,
    }));

    return NextResponse.json({ invoices: invoices.sort((a, b) => new Date(b.date!).getTime() - new Date(a.date!).getTime()) });
  } catch (error) {
    console.error("Failed to list invoices:", error);
    return NextResponse.json({ error: "Failed to fetch invoices" }, { status: 500 });
  }
}
