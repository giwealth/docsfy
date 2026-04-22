(function () {
  const toggles = Array.from(document.querySelectorAll(".nav-accordion-toggle"));
  const panels = Array.from(document.querySelectorAll(".nav-accordion-panel"));
  const links = Array.from(document.querySelectorAll(".nav-link"));
  const subAccordionLinks = Array.from(
    document.querySelectorAll("[data-sub-accordion-link]")
  );
  let skipRefreshOnce = false;

  /**
   * 在侧栏内滚动，保证菜单标题行可见。
   * pinToAsideTop：把该行贴到侧栏可视区顶部（适合底部项展开大量子菜单时，避免标题被顶出视野）。
   */
  function scrollNavRowIntoView(el, opts) {
    if (!el) return;
    const aside = el.closest("aside");
    const pad = opts && opts.pad != null ? opts.pad : 12;
    const pinToAsideTop = opts && opts.pinToAsideTop;

    const run = () => {
      if (!aside) {
        el.scrollIntoView({ block: "nearest", inline: "nearest", behavior: "smooth" });
        return;
      }

      const asideRect = aside.getBoundingClientRect();
      const rowRect = el.getBoundingClientRect();
      const rowScrollY = aside.scrollTop + (rowRect.top - asideRect.top);
      const maxScroll = Math.max(0, aside.scrollHeight - aside.clientHeight);

      if (pinToAsideTop) {
        aside.scrollTop = Math.max(0, Math.min(maxScroll, rowScrollY - pad));
        return;
      }

      if (rowRect.top < asideRect.top + pad) {
        aside.scrollTop = Math.max(0, Math.min(maxScroll, rowScrollY - pad));
      } else if (rowRect.bottom > asideRect.bottom - pad) {
        aside.scrollTop = Math.max(
          0,
          Math.min(maxScroll, aside.scrollTop + (rowRect.bottom - asideRect.bottom + pad))
        );
      }

      const r2 = el.getBoundingClientRect();
      const v2 = aside.getBoundingClientRect();
      if (r2.top < v2.top + pad) {
        const y = aside.scrollTop + (r2.top - v2.top);
        aside.scrollTop = Math.max(0, Math.min(maxScroll, y - pad));
      }
    };

    requestAnimationFrame(() => requestAnimationFrame(run));
  }

  function collapseAll() {
    toggles.forEach((btn) => {
      btn.setAttribute("aria-expanded", "false");
      const icon = btn.querySelector("span:last-child");
      if (icon) {
        icon.textContent = "▾";
      }
    });
    panels.forEach((panel) => {
      panel.classList.add("hidden");
    });
  }

  function collapseSubByParent(parentId) {
    subAccordionLinks.forEach((link) => {
      if (link.getAttribute("data-sub-accordion-parent") !== parentId) return;
      link.setAttribute("aria-expanded", "false");
      const icon = link.querySelector(".nav-sub-accordion-icon");
      if (icon) {
        icon.textContent = "▾";
      }
      const targetId = link.getAttribute("data-sub-accordion-link");
      if (!targetId) return;
      const panel = document.getElementById(targetId);
      if (panel) {
        panel.classList.add("hidden");
      }
    });
  }

  toggles.forEach((btn) => {
    btn.addEventListener("click", function () {
      const targetId = btn.getAttribute("data-accordion-target");
      if (!targetId) return;
      const panel = document.getElementById(targetId);
      if (!panel) return;

      const isExpanded = btn.getAttribute("aria-expanded") === "true";
      collapseAll();
      subAccordionLinks.forEach((subLink) => {
        const parentId = subLink.getAttribute("data-sub-accordion-parent");
        if (parentId) {
          collapseSubByParent(parentId);
        }
      });

      if (!isExpanded) {
        btn.setAttribute("aria-expanded", "true");
        const icon = btn.querySelector("span:last-child");
        if (icon) {
          icon.textContent = "▴";
        }
        panel.classList.remove("hidden");
        scrollNavRowIntoView(btn);
      }
    });
  });

  subAccordionLinks.forEach((link) => {
    link.addEventListener("click", function (evt) {
      const targetId = link.getAttribute("data-sub-accordion-link");
      const parentId = link.getAttribute("data-sub-accordion-parent");
      if (!targetId || !parentId) return;
      const panel = document.getElementById(targetId);
      if (!panel) return;
      const isExpanded = !panel.classList.contains("hidden");

      // First click expands grouped submenu; second click navigates normally.
      if (isExpanded) {
        return;
      }

      const icon = link.querySelector(".nav-sub-accordion-icon");
      if (!parentId) return;

      evt.preventDefault();
      skipRefreshOnce = true;
      link.setAttribute("aria-expanded", "true");
      if (icon) {
        icon.textContent = "▴";
      }
      panel.classList.remove("hidden");
      setImmediateActiveWithinLevel(link);
      // 底部项展开大量子菜单时，把父级标题滚到侧栏顶部附近，避免名称被顶出可视区
      scrollNavRowIntoView(link, { pinToAsideTop: true });
    });
  });

  function normalizePath(pathname) {
    if (!pathname.endsWith("/")) {
      return pathname + "/";
    }
    return pathname;
  }

  function isActiveLink(link) {
    const url = new URL(link.href, window.location.origin);
    const currentPath = normalizePath(window.location.pathname);
    const linkPath = normalizePath(url.pathname);
    if (linkPath !== currentPath) {
      return false;
    }
    const currentHash = decodeURIComponent(window.location.hash || "");
    const linkHash = decodeURIComponent(url.hash || "");
    if (url.hash) {
      return linkHash === currentHash;
    }
    return currentHash === "";
  }

  function clearAccordionButtonStyles() {
    toggles.forEach((btn) => {
      btn.classList.remove("bg-brand-50", "text-brand-600", "font-semibold");
    });
  }

  function resetLinkStyle(link) {
    const inactiveColor = link.dataset.inactiveColor;
    link.classList.remove("bg-brand-50", "text-brand-600", "font-semibold");
    if (inactiveColor) {
      link.classList.add(inactiveColor);
    }
  }

  function activateLinkStyle(link) {
    const inactiveColor = link.dataset.inactiveColor;
    if (inactiveColor) {
      link.classList.remove(inactiveColor);
    }
    link.classList.add("bg-brand-50", "text-brand-600", "font-semibold");
  }

  function setImmediateActiveWithinLevel(clickedLink) {
    const subPanel = clickedLink.closest(".nav-sub-accordion-panel");
    if (subPanel) {
      const levelLinks = Array.from(subPanel.querySelectorAll(":scope > .nav-link"));
      levelLinks.forEach((link) => resetLinkStyle(link));
      activateLinkStyle(clickedLink);
      return;
    }

    const accordionPanel = clickedLink.closest(".nav-accordion-panel");
    if (accordionPanel) {
      const levelLinks = Array.from(
        accordionPanel.querySelectorAll(":scope > a.nav-link, :scope > .space-y-1 > a.nav-link")
      );
      levelLinks.forEach((link) => resetLinkStyle(link));
      activateLinkStyle(clickedLink);
      return;
    }

    const navRoot = clickedLink.closest("nav");
    if (!navRoot) return;
    const levelLinks = Array.from(navRoot.querySelectorAll(":scope > a.nav-link"));
    levelLinks.forEach((link) => resetLinkStyle(link));
    activateLinkStyle(clickedLink);
  }

  function applyActiveStyles() {
    let activeLink = null;
    links.forEach((link) => {
      resetLinkStyle(link);
      if (isActiveLink(link)) {
        activateLinkStyle(link);
        activeLink = link;
      }
    });
    return activeLink;
  }

  function expandForActiveLink(activeLink) {
    collapseAll();
    clearAccordionButtonStyles();
    if (!activeLink) {
      return;
    }
    const panel = activeLink.closest(".nav-accordion-panel");
    if (!panel || !panel.id) {
      return;
    }
    const btn = document.querySelector(
      `.nav-accordion-toggle[data-accordion-target="${panel.id}"]`
    );
    if (!btn) {
      return;
    }
    btn.setAttribute("aria-expanded", "true");
    const icon = btn.querySelector("span:last-child");
    if (icon) {
      icon.textContent = "▴";
    }
    btn.classList.add("bg-brand-50", "text-brand-600", "font-semibold");
    panel.classList.remove("hidden");

    const subPanel = activeLink.closest(".nav-sub-accordion-panel");
    if (subPanel && subPanel.id) {
      const subLink = document.querySelector(
        `[data-sub-accordion-link="${subPanel.id}"]`
      );
      if (subLink) {
        subLink.setAttribute("aria-expanded", "true");
        const icon = subLink.querySelector(".nav-sub-accordion-icon");
        if (icon) {
          icon.textContent = "▴";
        }
        subPanel.classList.remove("hidden");
        scrollNavRowIntoView(subLink, { pinToAsideTop: true });
      }
    } else {
      scrollNavRowIntoView(activeLink);
    }
  }

  function refreshActiveState() {
    const activeLink = applyActiveStyles();
    expandForActiveLink(activeLink);
  }

  refreshActiveState();
  window.addEventListener("hashchange", refreshActiveState);
  window.addEventListener("popstate", refreshActiveState);
  document.addEventListener("click", (evt) => {
    const target = evt.target;
    if (!(target instanceof Element)) return;
    const clickedLink = target.closest(".nav-link");
    if (clickedLink) {
      setImmediateActiveWithinLevel(clickedLink);
      if (skipRefreshOnce) {
        skipRefreshOnce = false;
        return;
      }
      window.setTimeout(refreshActiveState, 0);
    }
  });
})();
