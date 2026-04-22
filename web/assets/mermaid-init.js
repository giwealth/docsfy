(function () {
  if (typeof window.mermaid === "undefined") return;

  function convertMermaidBlocks() {
    document.querySelectorAll(".markdown-body pre > code.language-mermaid").forEach((code) => {
      const pre = code.parentElement;
      if (!pre || pre.tagName !== "PRE") return;
      const graph = code.textContent || "";
      if (!graph.trim()) return;

      const wrap = document.createElement("div");
      wrap.className = "mermaid-diagram";
      const el = document.createElement("div");
      el.className = "mermaid";
      el.textContent = graph;
      wrap.appendChild(el);
      pre.replaceWith(wrap);
    });
  }

  convertMermaidBlocks();
  if (!document.querySelector(".markdown-body .mermaid")) return;

  window.mermaid.initialize({
    startOnLoad: false,
    theme: "neutral",
    securityLevel: "strict",
    flowchart: { useMaxWidth: true },
  });
  window.mermaid.run({ querySelector: ".markdown-body .mermaid" });
})();
