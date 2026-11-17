/* „Spiel verlassen“: gibt den Platz auf dem Server frei, damit niemand auf
   einen wartet, und geht zurück in den Raum. Ohne den Knopf blieb die Partie
   stehen, wenn man nur den Reiter geschlossen oder den Zurück-Link benutzt
   hat. Fest unten rechts, damit er auch auf langen Spielseiten erreichbar ist.

   Wird von Spielseiten geladen, die keinen eigenen Platz für den Knopf haben:
   <script src="/spiele/verlassen.js"></script>
   Ohne room+id in der Adresse (Einzelspiel) passiert nichts. */
(function () {
  var p = new URLSearchParams(location.search);
  var raum = p.get("room") || "";
  var id = p.get("id") || "";
  if (!raum || !id) return;
  /* Seiten mit eigenem Verlassen-Knopf (Tetris, Sim, Käsekästchen,
     Mensch ärgere dich) brauchen keinen zweiten, der ihn überdeckt. */
  var eigene = document.querySelectorAll("button, a");
  for (var i = 0; i < eigene.length; i++) {
    if (/spiel\s*verlassen/i.test((eigene[i].textContent || "").replace(/\s+/g, " "))) return;
  }
  var box = document.createElement("div");
  box.setAttribute("data-spiel-verlassen", "");
  box.style.cssText = "position:fixed;right:10px;bottom:10px;z-index:9999";
  var b = document.createElement("button");
  b.type = "button";
  b.textContent = "Spiel verlassen";
  b.style.cssText = "font:inherit;font-size:14px;padding:9px 13px;border-radius:9px;" +
    "cursor:pointer;border:1px solid #d9bc7c;background:#12332ae8;color:#e8cd8e;" +
    "box-shadow:0 4px 12px #0006";
  b.onclick = function () {
    if (typeof window.tgVerlassen === "function") {
      window.tgVerlassen();
      return;
    }
    b.disabled = true;
    fetch("/api/leave?room=" + encodeURIComponent(raum), {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ id: id })
    }).catch(function () {}).then(function () {
      location.href = "/connector?room=" + encodeURIComponent(raum);
    });
  };
  box.appendChild(b);
  document.body.appendChild(box);
  /* Auf Seiten, die scrollen, unten Platz lassen – sonst liegt der Knopf
     dauerhaft über dem letzten Inhalt. */
  requestAnimationFrame(function () {
    if (document.documentElement.scrollHeight > window.innerHeight + 40) {
      var unten = parseFloat(getComputedStyle(document.body).paddingBottom) || 0;
      document.body.style.paddingBottom = (unten + 56) + "px";
    }
  });
})();
