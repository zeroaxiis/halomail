/**
 * HaloMail Meetings Embed — Button & Modal Booking Widget
 *
 * Usage:
 *   <script src="https://halomail.zeroaxiis.tech/embed.js" async></script>
 *
 *   <button data-halomail-book="drdead">
 *     Schedule a Consultation
 *   </button>
 *
 * Programmatic Trigger:
 *   HaloMail.book("drdead", { key: "..." });
 */
(function () {
  if (typeof window === "undefined") return;

  var SCRIPT_SRC = document.currentScript ? document.currentScript.src : "";
  var BASE_URL = "https://halomail.zeroaxiis.tech";
  try {
    if (SCRIPT_SRC && SCRIPT_SRC.startsWith("http")) {
      BASE_URL = new URL(SCRIPT_SRC).origin;
    } else if (typeof window !== "undefined" && window.location.origin) {
      if (window.location.hostname === "localhost" || window.location.hostname === "127.0.0.1") {
        BASE_URL = window.location.origin;
      }
    }
  } catch (e) {}

  var activeModal = null;

  function closeModal() {
    if (activeModal && activeModal.parentNode) {
      activeModal.style.opacity = "0";
      var dialog = activeModal.querySelector(".halomail-dialog");
      if (dialog) dialog.style.transform = "scale(0.97)";
      setTimeout(function () {
        if (activeModal && activeModal.parentNode) {
          activeModal.parentNode.removeChild(activeModal);
        }
        activeModal = null;
        document.body.style.overflow = "";
      }, 180);
    }
  }

  function openModal(handle, options) {
    if (!handle) return;
    options = options || {};
    var key = options.key || "";

    closeModal();

    // Backdrop overlay
    var overlay = document.createElement("div");
    overlay.id = "halomail-booking-modal";
    overlay.style.position = "fixed";
    overlay.style.top = "0";
    overlay.style.left = "0";
    overlay.style.width = "100vw";
    overlay.style.height = "100vh";
    overlay.style.backgroundColor = "rgba(0, 0, 0, 0.75)";
    overlay.style.backdropFilter = "blur(10px)";
    overlay.style.webkitBackdropFilter = "blur(10px)";
    overlay.style.zIndex = "9999999";
    overlay.style.display = "flex";
    overlay.style.alignItems = "center";
    overlay.style.justifyContent = "center";
    overlay.style.padding = "16px";
    overlay.style.opacity = "0";
    overlay.style.transition = "opacity 0.2s cubic-bezier(0.16, 1, 0.3, 1)";
    overlay.style.boxSizing = "border-box";

    // Dialog card matching HaloMail website dark theme
    var dialog = document.createElement("div");
    dialog.className = "halomail-dialog";
    dialog.style.width = "100%";
    dialog.style.maxWidth = "840px";
    dialog.style.height = "86vh";
    dialog.style.maxHeight = "680px";
    dialog.style.backgroundColor = "#09090b";
    dialog.style.borderRadius = "16px";
    dialog.style.border = "1px solid rgba(255, 255, 255, 0.1)";
    dialog.style.boxShadow = "0 30px 60px -12px rgba(0, 0, 0, 0.85), 0 0 0 1px rgba(255, 255, 255, 0.06)";
    dialog.style.position = "relative";
    dialog.style.overflow = "hidden";
    dialog.style.display = "flex";
    dialog.style.flexDirection = "column";
    dialog.style.transform = "scale(0.97)";
    dialog.style.transition = "transform 0.2s cubic-bezier(0.16, 1, 0.3, 1)";

    // Clean SVG close button (no emojis or crude text)
    var closeBtn = document.createElement("button");
    closeBtn.setAttribute("type", "button");
    closeBtn.setAttribute("aria-label", "Close");
    closeBtn.innerHTML = '<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>';
    closeBtn.style.position = "absolute";
    closeBtn.style.top = "12px";
    closeBtn.style.right = "12px";
    closeBtn.style.width = "28px";
    closeBtn.style.height = "28px";
    closeBtn.style.borderRadius = "8px";
    closeBtn.style.border = "1px solid rgba(255, 255, 255, 0.08)";
    closeBtn.style.background = "rgba(255, 255, 255, 0.04)";
    closeBtn.style.color = "#a1a1aa";
    closeBtn.style.cursor = "pointer";
    closeBtn.style.zIndex = "20";
    closeBtn.style.display = "flex";
    closeBtn.style.alignItems = "center";
    closeBtn.style.justifyContent = "center";
    closeBtn.style.transition = "all 0.15s ease";

    closeBtn.onmouseenter = function () {
      closeBtn.style.background = "rgba(255, 255, 255, 0.1)";
      closeBtn.style.color = "#ffffff";
    };
    closeBtn.onmouseleave = function () {
      closeBtn.style.background = "rgba(255, 255, 255, 0.04)";
      closeBtn.style.color = "#a1a1aa";
    };
    closeBtn.onclick = closeModal;

    // Iframe container
    var iframe = document.createElement("iframe");
    var params = new URLSearchParams();
    params.set("embed", "true");
    if (key) params.set("key", key);

    iframe.src = BASE_URL + "/book/" + encodeURIComponent(handle) + "?" + params.toString();
    iframe.style.width = "100%";
    iframe.style.height = "100%";
    iframe.style.border = "none";
    iframe.style.display = "block";
    iframe.style.flex = "1";
    iframe.style.backgroundColor = "transparent";
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
      dialog.style.transform = "scale(1)";
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
        var key = btn.getAttribute("data-key") || "";
        openModal(handle, { key: key });
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
