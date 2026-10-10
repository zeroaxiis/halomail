import { NextRequest, NextResponse } from "next/server";
import { S3Client, GetObjectCommand } from "@aws-sdk/client-s3";

export async function GET(req: NextRequest) {
  const key = req.nextUrl.searchParams.get("key");
  if (!key) return NextResponse.json({ error: "key required" }, { status: 400 });

  try {
    const client = new S3Client({
      region: process.env.R2_REGION || "auto",
      endpoint: process.env.R2_ENDPOINT,
      credentials: {
        accessKeyId: process.env.R2_ACCESS_KEY_ID || "",
        secretAccessKey: process.env.R2_SECRET_ACCESS_KEY || "",
      },
    });

    const data = await client.send(new GetObjectCommand({
      Bucket: process.env.R2_BUCKET_NAME || "halomail",
      Key: key,
    }));

    return new NextResponse(data.Body as any, {
      headers: {
        "Content-Type": "text/html",
        "Content-Disposition": `attachment; filename="${key.split("/").pop()}"`
      }
    });
  } catch (error) {
    console.error("Download failed:", error);
    return NextResponse.json({ error: "Failed to download invoice" }, { status: 500 });
  }
}
