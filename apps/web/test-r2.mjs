import { S3Client, PutObjectCommand, ListObjectsV2Command } from "@aws-sdk/client-s3";
import dotenv from "dotenv";
dotenv.config({ path: "../../.env" });

async function testR2() {
  console.log("Testing R2 Connection...");
  try {
    const client = new S3Client({
      region: process.env.R2_REGION || "auto",
      endpoint: process.env.R2_ENDPOINT,
      credentials: {
        accessKeyId: process.env.R2_ACCESS_KEY_ID || "",
        secretAccessKey: process.env.R2_SECRET_ACCESS_KEY || "",
      },
    });

    const bucketName = process.env.R2_BUCKET_NAME || "halomail";
    console.log("Bucket:", bucketName);

    // Test Upload
    const testKey = "test_upload.txt";
    console.log(`Uploading ${testKey}...`);
    await client.send(new PutObjectCommand({
      Bucket: bucketName,
      Key: testKey,
      Body: "Hello from R2 Test Script!",
      ContentType: "text/plain",
    }));
    console.log("✅ Upload successful!");

    // Test List
    console.log("Listing objects...");
    const data = await client.send(new ListObjectsV2Command({
      Bucket: bucketName,
    }));
    
    console.log("✅ Objects found:", data.Contents?.length || 0);
    if (data.Contents) {
      data.Contents.slice(0, 5).forEach(obj => {
        console.log(` - ${obj.Key} (Last Modified: ${obj.LastModified})`);
      });
    }

  } catch (error) {
    console.error("❌ R2 Test Failed:", error);
  }
}

testR2();
