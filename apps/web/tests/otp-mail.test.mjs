import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { createRequire } from "node:module";
import test from "node:test";
import vm from "node:vm";

const require = createRequire(import.meta.url);
const mailRequire = createRequire(new URL("../lib/otp-mail.ts", import.meta.url));
const ts = require("typescript");
const source = readFileSync(new URL("../lib/otp-mail.ts", import.meta.url), "utf8");
const compiled = ts.transpileModule(source, {
  compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022, esModuleInterop: true },
}).outputText;

function mailer({ env = {}, response = () => Response.json({ id: "email-id" }) } = {}) {
  const requests = [];
  const smtpMessages = [];
  const exports = {};
  vm.runInNewContext(compiled, {
    exports, process: { env }, AbortSignal,
    require: name => name === "nodemailer" ? {
      createTransport: () => ({ sendMail: async message => { smtpMessages.push(message); return { accepted: [message.to] }; } }),
    } : mailRequire(name),
    fetch: async (url, init) => { requests.push({ url, init }); return response(); },
  });
  return { send: exports.sendLoginCode, render: exports.renderLoginEmail, requests, smtpMessages };
}

test("Resend sends the code from the configured domain to the login recipient", async () => {
  const from = "HaloMail <no-reply@email.zeroaxiis.tech>";
  const sender = mailer({ env: { RESEND_API_KEY: "test-key", EMAIL_FROM: from, NODE_ENV: "production" } });
  await sender.send("user@example.com", "012345");
  assert.equal(sender.requests.length, 1);
  const { url, init } = sender.requests[0];
  assert.equal(url, "https://api.resend.com/emails");
  assert.equal(init.headers.Authorization, "Bearer test-key");
  const body = JSON.parse(init.body);
  assert.equal(body.from, from);
  assert.deepEqual(body.to, ["user@example.com"]);
  assert.match(body.text, /012345/);
  assert.match(body.text, /10 minutes/);
  assert.match(body.html, /012345/);
  assert.match(body.html, /Your sign-in code/);
  assert.match(body.html, /ZeroAxiis/);
  assert.match(body.html, /src="cid:zeroaxiis-logo"/);
  assert.equal(body.attachments[0].content_id, "zeroaxiis-logo");
  assert.equal(Buffer.from(body.attachments[0].content, "base64").subarray(1, 4).toString(), "PNG");
  assert.doesNotMatch(body.html, /border-radius|linear-gradient|Courier|monospace|Georgia|01 \/|NOT VALID/);
  assert.match(body.html, /expires in 10 minutes/);
  assert.equal(sender.smtpMessages.length, 0);
});

test("sample email contains a visible code and states it cannot sign in", async () => {
  const sender = mailer({ env: { RESEND_API_KEY: "test-key" } });
  await sender.send("user@example.com", "630927", { sample: true });
  const body = JSON.parse(sender.requests[0].init.body);
  assert.match(body.subject, /sample/i);
  assert.match(body.html, /630927/);
  assert.match(body.html, /cannot sign you in/);
  assert.match(body.text, /630927/);
  assert.match(body.text, /cannot sign you in/);
  assert.throws(() => sender.render("<script>"), /six-digit/);
});

test("Resend failures reject delivery without SMTP fallback", async () => {
  for (const status of [401, 403, 429, 500]) {
    const sender = mailer({ env: { RESEND_API_KEY: "test-key" }, response: () => Response.json({ message: "rejected" }, { status }) });
    await assert.rejects(sender.send("user@example.com", "012345"), new RegExp(String(status)));
    assert.equal(sender.smtpMessages.length, 0);
  }
  const sender = mailer({ env: { RESEND_API_KEY: "test-key" }, response: () => Response.json({}) });
  await assert.rejects(sender.send("user@example.com", "012345"), /did not accept/);
});

test("production requires Resend while local development can use Mailpit", async () => {
  const production = mailer({ env: { NODE_ENV: "production" } });
  await assert.rejects(production.send("user@example.com", "012345"), /RESEND_API_KEY is required/);
  assert.equal(production.smtpMessages.length, 0);
  const local = mailer({ env: { NODE_ENV: "development" } });
  await local.send("user@example.com", "012345");
  assert.equal(local.requests.length, 0);
  assert.equal(local.smtpMessages.length, 1);
  assert.match(local.smtpMessages[0].html, /012345/);
  assert.equal(local.smtpMessages[0].attachments[0].cid, "zeroaxiis-logo");
});
