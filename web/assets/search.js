(async function () {
  const searchInput = document.getElementById("searchInput");
  const searchResult = document.getElementById("searchResultTop");
  let index = [];
  let ready = false;
  let loading = false;
  let lastMatches = [];
  let currentQuery = "";
  let es = null;
  let lastReloadAt = 0;

  async function ensureIndexLoaded() {
    if (ready || loading) return ready;
    loading = true;
    try {
      const res = await fetch("/search-index.json", { cache: "no-store" });
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      index = await res.json();
      ready = true;
      return true;
    } catch (e) {
      console.warn("load search index failed:", e);
      if (searchResult) {
        searchResult.textContent = "搜索索引加载失败，请刷新页面重试。";
      }
      return false;
    } finally {
      loading = false;
    }
  }

  function appendHighlightedText(parent, text, query) {
    const source = text || "";
    const q = (query || "").trim();
    if (!q) {
      parent.textContent = source;
      return;
    }
    const lower = source.toLowerCase();
    const needle = q.toLowerCase();
    let start = 0;
    while (start < source.length) {
      const idx = lower.indexOf(needle, start);
      if (idx === -1) {
        parent.appendChild(document.createTextNode(source.slice(start)));
        break;
      }
      if (idx > start) {
        parent.appendChild(document.createTextNode(source.slice(start, idx)));
      }
      const mark = document.createElement("mark");
      mark.textContent = source.slice(idx, idx + needle.length);
      parent.appendChild(mark);
      start = idx + needle.length;
    }
  }

  function render(items) {
    if (!searchResult) return;
    searchResult.innerHTML = "";
    if (!ready) return;
    if (!items.length) {
      const empty = document.createElement("div");
      empty.className = "search-empty";
      empty.textContent = "无匹配结果";
      searchResult.appendChild(empty);
      return;
    }
    items.slice(0, 8).forEach((it) => {
      const a = document.createElement("a");
      a.href = it.route;

      const title = document.createElement("strong");
      appendHighlightedText(title, it.title || "(无标题)", currentQuery);

      const br = document.createElement("br");
      const desc = document.createElement("small");
      appendHighlightedText(desc, it.content || "", currentQuery);

      a.appendChild(title);
      a.appendChild(br);
      a.appendChild(desc);
      searchResult.appendChild(a);
    });
  }

  if (searchInput) {
    searchInput.addEventListener("focus", async () => {
      await ensureIndexLoaded();
    });

    searchInput.addEventListener("input", async () => {
      if (!ready) {
        const ok = await ensureIndexLoaded();
        if (!ok) return;
      }
      const q = searchInput.value.trim().toLowerCase();
      if (!q) {
        currentQuery = "";
        lastMatches = [];
        searchResult.innerHTML = "";
        return;
      }
      currentQuery = searchInput.value.trim();
      lastMatches = index.filter(
        (it) =>
          (it.title || "").toLowerCase().includes(q) ||
          (it.content || "").toLowerCase().includes(q)
      );
      render(lastMatches);
    });

    searchInput.addEventListener("keydown", (evt) => {
      if (evt.key === "Enter" && lastMatches.length > 0) {
        window.location.href = lastMatches[0].route;
      }
    });

    searchInput.addEventListener("blur", () => {
      window.setTimeout(() => {
        if (document.activeElement !== searchInput) {
          searchResult.innerHTML = "";
        }
      }, 150);
    });
  }

  document.addEventListener("click", (evt) => {
    if (!searchResult || !searchInput) return;
    const target = evt.target;
    if (!(target instanceof Node)) return;
    if (!searchResult.contains(target) && target !== searchInput) {
      searchResult.innerHTML = "";
    }
  });

  if (!!window.EventSource) {
    if (window.__docsfyEventSource) {
      window.__docsfyEventSource.close();
    }
    es = new EventSource("/__livereload");
    window.__docsfyEventSource = es;
    es.onmessage = function (evt) {
      if (evt.data === "reload") {
        const now = Date.now();
        if (now-lastReloadAt < 1500) {
          return;
        }
        lastReloadAt = now;
        window.location.reload();
      }
    };
    window.addEventListener("pagehide", () => {
      if (es) {
        es.close();
      }
    });
  }
})();
