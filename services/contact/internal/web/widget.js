/* HaloMail contact widget — progressive enhancement for contact forms.
 *
 *   <script src="https://<api-host>/widget.js" defer></script>
 *   <form data-halomail="your-forms-access-key">
 *     <input name="name"><input name="email"><textarea name="message"></textarea>
 *     <input name="_hl_hp" tabindex="-1" autocomplete="off" style="display:none">
 *     <button type="submit">Send</button>
 *   </form>
 *
 * Or call directly: HaloMail.contact("slug", { name, email, data: {...} })
 */
(function () {
  var ENDPOINT = "/halomail.contact.v1.MessageService/SubmitMessage";

  function apiBase(el) {
    try { return new URL(el.src).origin; } catch (e) { return ""; }
  }

  function submit(base, slug, payload) {
    return fetch(base + ENDPOINT, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        accessKey: slug,
        senderName: payload.name || "",
        senderEmail: payload.email || "",
        data: payload.data || {},
        honeypot: payload.honeypot || ""
      })
    }).then(function (response) {
      return response.json().then(function (result) {
        if (!response.ok || !result.accepted) throw new Error(result.message || "Submission failed");
        return result;
      });
    });
  }

  function bind(base) {
    document.querySelectorAll("form[data-halomail]").forEach(function (form) {
      if (form.__halomail) return;
      form.__halomail = true;
      form.addEventListener("submit", function (e) {
        e.preventDefault();
        if (form.__sending) return;
        form.__sending = true;
        var slug = form.getAttribute("data-halomail");
        var data = {}, name = "", email = "", honeypot = "";
        new FormData(form).forEach(function (v, k) {
          if (k === "_hl_hp") { honeypot = v; return; }
          if (k === "name") name = v;
          if (k === "email") email = v;
          data[k] = v;
        });
        submit(base, slug, { name: name, email: email, data: data, honeypot: honeypot })
          .then(function (res) {
            if (res && res.redirectUrl) { window.location.href = res.redirectUrl; return; }
            form.dispatchEvent(new CustomEvent("halomail:sent", { detail: res }));
            form.reset();
          })
          .catch(function (err) {
            form.dispatchEvent(new CustomEvent("halomail:error", { detail: err }));
          }).finally(function () { form.__sending = false; });
      });
    });
  }

  var me = document.currentScript;
  var base = apiBase(me);
  window.HaloMail = window.HaloMail || {};
  window.HaloMail.contact = function (slug, payload) { return submit(base, slug, payload || {}); };

  if (document.readyState !== "loading") bind(base);
  else document.addEventListener("DOMContentLoaded", function () { bind(base); });
})();
