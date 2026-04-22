(function () {
  const krokiImgs = Array.from(document.querySelectorAll(".kroki-diagram img"));
  const mermaidSvgs = Array.from(document.querySelectorAll(".mermaid-diagram svg"));
  if (!krokiImgs.length && !mermaidSvgs.length) return;

  const overlay = document.createElement("div");
  overlay.className = "kroki-lightbox";
  overlay.setAttribute("aria-hidden", "true");

  const closeBtn = document.createElement("button");
  closeBtn.type = "button";
  closeBtn.className = "kroki-lightbox-close";
  closeBtn.setAttribute("aria-label", "Close preview");
  closeBtn.textContent = "×";

  const big = document.createElement("img");
  big.alt = "diagram preview";

  overlay.appendChild(closeBtn);
  overlay.appendChild(big);
  document.body.appendChild(overlay);

  function open(src, alt) {
    big.src = src;
    big.alt = alt || "diagram preview";
    overlay.classList.add("open");
    overlay.setAttribute("aria-hidden", "false");
    document.body.style.overflow = "hidden";
  }

  function svgToDataURL(svgEl) {
    const cloned = svgEl.cloneNode(true);
    if (!cloned.getAttribute("xmlns")) {
      cloned.setAttribute("xmlns", "http://www.w3.org/2000/svg");
    }
    const xml = new XMLSerializer().serializeToString(cloned);
    return "data:image/svg+xml;charset=utf-8," + encodeURIComponent(xml);
  }

  function close() {
    overlay.classList.remove("open");
    overlay.setAttribute("aria-hidden", "true");
    big.removeAttribute("src");
    document.body.style.overflow = "";
  }

  krokiImgs.forEach((img) => {
    img.addEventListener("click", () => open(img.src, img.alt));
  });

  mermaidSvgs.forEach((svg) => {
    svg.style.cursor = "zoom-in";
    svg.addEventListener("click", () => {
      open(svgToDataURL(svg), "mermaid diagram");
    });
  });

  closeBtn.addEventListener("click", close);
  overlay.addEventListener("click", (evt) => {
    if (evt.target === overlay) close();
  });
  window.addEventListener("keydown", (evt) => {
    if (evt.key === "Escape" && overlay.classList.contains("open")) {
      close();
    }
  });
})();
