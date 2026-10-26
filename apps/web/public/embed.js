/**
 * HaloMail Meetings Embed — Cal.com Style Button & Modal Widget
 *
 * Usage:
 *   <script src="https://halomail.zeroaxiis.tech/embed.js" async></script>
 *
 *   <button data-halomail-book="drdead" data-theme="obsidian">
 *     Schedule a Consultation
 *   </button>
 *
 * Or Programmatically:
 *   HaloMail.book("drdead", { theme: "obsidian" });
 */
(function () {
  if (typeof window === "undefined") return;

  var SCRIPT_SRC = document.currentScript ? document.currentScript.src : "";
  var BASE_URL = "https://halomail.zeroaxiis.tech";
  try {
    if (SCRIPT_SRC && SCRIPT_SRC.startsWith("http")) {
      BASE_URL = new URL(SCRIPT_SRC).origin;
    }
  } catch (e) {}

  var activeModal = null;

  function closeModal() {
    if (activeModal && activeModal.parentNode) {
      activeModal.style.opacity = "0";
      setTimeout(function () {
        if (activeModal && activeModal.parentNode) {
          activeModal.parentNode.removeChild(activeModal);
        }
        activeModal = null;
        document.body.style.overflow = "";
      }, 200);
    }
  }

  function openModal(handle, options) {
    if (!handle) return;
    options = options || {};
    var theme = options.theme || "obsidian";
    var key = options.key || "";

    closeModal();

    // Backdrop
    var overlay = document.createElement("div");
    overlay.id = "halomail-booking-modal";
    overlay.style.position = "fixed";
    overlay.style.top = "0";
    overlay.style.left = "0";
    overlay.style.width = "100vw";
    overlay.style.height = "100vh";
    overlay.style.backgroundColor = "rgba(0, 0, 0, 0.78)";
    overlay.style.backdropFilter = "blur(8px)";
    overlay.style.webkitBackdropFilter = "blur(8px)";
    overlay.style.zIndex = "9999999";
    overlay.style.display = "flex";
    overlay.style.alignItems = "center";
    overlay.style.justifyContent = "center";
    overlay.style.padding = "16px";
    overlay.style.opacity = "0";
    overlay.style.transition = "opacity 0.2s ease";
    overlay.style.boxSizing = "border-box";

    // Dialog Card
    var dialog = document.createElement("div");
    dialog.style.width = "100%";
    dialog.style.maxWidth = "760px";
    dialog.style.height = "85vh";
    dialog.style.maxHeight = "720px";
    dialog.style.backgroundColor = theme === "minimal" ? "#18181b" : theme === "classic" ? "#0d1527" : "#0d0b14";
    dialog.style.borderRadius = "16px";
    dialog.style.border = "1px solid rgba(255, 255, 255, 0.12)";
    dialog.style.boxShadow = "0 30px 60px rgba(0, 0, 0, 0.8), 0 0 0 1px rgba(255, 255, 255, 0.05)";
    dialog.style.position = "relative";
    dialog.style.overflow = "hidden";
    dialog.style.display = "flex";
    dialog.style.flexDirection = "column";

    // Close button
    var closeBtn = document.createElement("button");
    closeBtn.setAttribute("type", "button");
    closeBtn.setAttribute("aria-label", "Close");
    closeBtn.innerHTML = "&times;";
    closeBtn.style.position = "absolute";
    closeBtn.style.top = "12px";
    closeBtn.style.right = "14px";
    closeBtn.style.width = "32px";
    closeBtn.style.height = "32px";
    closeBtn.style.borderRadius = "50%";
    closeBtn.style.border = "1px solid rgba(255, 255, 255, 0.1)";
    closeBtn.style.background = "rgba(0, 0, 0, 0.4)";
    closeBtn.style.color = "#fff";
    closeBtn.style.fontSize = "20px";
    closeBtn.style.lineHeight = "1";
    closeBtn.style.cursor = "pointer";
    closeBtn.style.zIndex = "10";
    closeBtn.style.display = "flex";
    closeBtn.style.alignItems = "center";
    closeBtn.style.justifyContent = "center";
    closeBtn.style.transition = "background 0.15s ease";

    closeBtn.onmouseenter = function () { closeBtn.style.background = "rgba(255, 255, 255, 0.15)"; };
    closeBtn.onmouseleave = function () { closeBtn.style.background = "rgba(0, 0, 0, 0.4)"; };
    closeBtn.onclick = closeModal;

    // Iframe container
    var iframe = document.createElement("iframe");
    var params = new URLSearchParams();
    params.set("embed", "true");
    if (theme) params.set("theme", theme);
    if (key) params.set("key", key);

    iframe.src = BASE_URL + "/book/" + encodeURIComponent(handle) + "?" + params.toString();
    iframe.style.width = "100%";
    iframe.style.height = "100%";
    iframe.style.border = "none";
    iframe.style.display = "block";
    iframe.style.flex = "1";
    iframe.title = "Schedule a meeting with " + handle;

    dialog.appendChild(closeBtn);
    dialog.appendChild(iframe);
    overlay.appendChild(dialog);

    overlay.onclick = function (e) {
      if (e.target === overlay) closeModal();
    };

    document.body.appendChild(overlay);
    document.body.style.overflow = "hidden";
    activeModal = overlay;

    requestAnimationFrame(function () {
      overlay.style.opacity = "1";
    });
  }

  // Handle ESC key
  window.addEventListener("keydown", function (e) {
    if (e.key === "Escape" && activeModal) {
      closeModal();
    }
  });

  // Auto-bind to buttons with data-halomail-book
  function bindElements() {
    var buttons = document.querySelectorAll("[data-halomail-book], [data-halomail-meeting]");
    buttons.forEach(function (btn) {
      if (btn.__halomail_bound) return;
      btn.__halomail_bound = true;
      btn.addEventListener("click", function (e) {
        e.preventDefault();
        var handle = btn.getAttribute("data-halomail-book") || btn.getAttribute("data-halomail-meeting");
        var theme = btn.getAttribute("data-theme") || "obsidian";
        var key = btn.getAttribute("data-key") || "";
        openModal(handle, { theme: theme, key: key });
      });
    });
  }

  window.HaloMail = window.HaloMail || {};
  window.HaloMail.book = openModal;
  window.HaloMail.openMeetingModal = openModal;
  window.HaloMail.closeModal = closeModal;

  if (document.readyState !== "loading") {
    bindElements();
  } else {
    document.addEventListener("DOMContentLoaded", bindElements);
  }
})();
