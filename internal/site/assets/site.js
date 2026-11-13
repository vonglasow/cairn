(function () {
  var root = document.documentElement;
  var pop = document.getElementById("themePop");
  var themeBtn = document.getElementById("themeBtn");
  function apply(slug, keep) {
    root.setAttribute("data-scheme", slug);
    var picked = pop.querySelector('[data-pick="' + slug + '"]');
    document.getElementById("themeName").textContent = picked ? picked.textContent : slug;
    document.querySelectorAll("[data-pick]").forEach(function (b) { b.setAttribute("aria-pressed", String(b.dataset.pick === slug)); });
    document.querySelectorAll("img[data-shot]").forEach(function (img) { img.src = "shots/window-" + slug + ".png"; });
    if (keep) { try { localStorage.setItem("cairn-theme", slug); } catch (e) {} }
  }
  apply(root.getAttribute("data-scheme"), false);
  themeBtn.addEventListener("click", function (e) {
    e.stopPropagation();
    var open = !pop.classList.contains("open");
    pop.classList.toggle("open", open);
    themeBtn.setAttribute("aria-expanded", String(open));
  });
  document.addEventListener("click", function (e) {
    var b = e.target.closest("[data-pick]");
    if (b) { apply(b.dataset.pick, true); }
    if (!e.target.closest("#themePop") && !e.target.closest("#themeBtn")) {
      pop.classList.remove("open");
      themeBtn.setAttribute("aria-expanded", "false");
    }
  });

  var side = document.getElementById("side");
  document.getElementById("menuBtn").addEventListener("click", function () { if (side) { side.classList.toggle("open"); } });

  var search = document.getElementById("search"), q = document.getElementById("q"), results = document.getElementById("results");
  var pages = window.cairnSearch || [];
  function escaped(s) { return s.replace(/[&<>"]/g, function (c) { return { "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;" }[c]; }); }
  function find(said) {
    var words = said.toLowerCase().split(/\s+/).filter(Boolean);
    if (!words.length) { return []; }
    return pages.map(function (p) {
      var title = p.t.toLowerCase(), heads = (p.h || []).join(" ").toLowerCase(), text = p.x.toLowerCase(), score = 0;
      for (var i = 0; i < words.length; i++) {
        var w = words[i];
        if (title.indexOf(w) >= 0) { score += 10; } else if (heads.indexOf(w) >= 0) { score += 4; } else if (text.indexOf(w) >= 0) { score += 1; } else { return null; }
      }
      var at = text.indexOf(words[0]);
      return { p: p, score: score, said: at >= 0 ? p.x.substr(Math.max(0, at - 40), 120) : "" };
    }).filter(Boolean).sort(function (a, b) { return b.score - a.score; }).slice(0, 12);
  }
  function show() {
    results.innerHTML = find(q.value).map(function (r, i) {
      return '<li><a href="' + r.p.u + '"' + (i === 0 ? ' class="sel"' : "") + ">" + escaped(r.p.t) + "<small>" + escaped(r.said) + "</small></a></li>";
    }).join("");
  }
  function open() { search.classList.add("open"); q.value = ""; results.innerHTML = ""; q.focus(); }
  function close() { search.classList.remove("open"); }
  document.getElementById("searchBtn").addEventListener("click", open);
  q.addEventListener("input", show);
  search.addEventListener("click", function (e) { if (e.target === search) { close(); } });
  document.addEventListener("keydown", function (e) {
    if (e.key === "/" && !search.classList.contains("open") && document.activeElement.tagName !== "INPUT") { e.preventDefault(); open(); }
    else if (e.key === "Escape") { close(); }
    else if (e.key === "Enter" && search.classList.contains("open")) { var first = results.querySelector("a"); if (first) { location.href = first.href; } }
  });
})();
