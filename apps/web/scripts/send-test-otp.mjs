import { randomInt } from "node:crypto";
import { sendLoginCode } from "../lib/otp-mail.ts";

process.loadEnvFile(new URL("../.env", import.meta.url));

const recipient = process.argv[2];
if (!recipient || !/^[^@\s]+@[^@\s]+\.[^@\s]+$/.test(recipient)) {
  console.error("Usage: pnpm --filter @halomail/web sample:otp recipient@example.com");
  process.exit(1);
}

const code = String(randomInt(0, 1_000_000)).padStart(6, "0");
try {
  const id = await sendLoginCode(recipient, code, { sample: true });
  console.log(`Resend accepted sample OTP email for ${recipient}; message ID: ${id}`);
} catch (error) {
  console.error(error instanceof Error ? error.message : "Could not send sample OTP email");
  process.exitCode = 1;
}
