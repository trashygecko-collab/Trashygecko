/* ===========================================================================
   ZENTRALER GRAFIKSTIL

   Eine Wahl, alle Spiele. Der Stil gehört dem Spieler, nicht dem Raum:
   er liegt ausschließlich im localStorage dieses Browsers (Schlüssel
   "tg-stil") und wird von keinem Server, keiner Raumkonfiguration und
   keiner Admin-Einstellung überschrieben.

   Drei Werte, überall dieselben:
     klassisch · maedchen · buben

   Ein Spiel liest ihn so:
       <script src="/spiele/stil.js"></script>
       var stil = TGStil.hole();          // "klassisch" | "maedchen" | "buben"

   Hat ein Spiel eigene Namen für die Ansichten (z. B. "normal" oder
   "jungs"), bildet es sie einmal ab:
       var stil = TGStil.lokal({ klassisch: "normal", maedchen: "maedchen", buben: "buben" });

   Ein Spiel ohne Alternativen braucht nichts zu tun und bleibt klassisch.

   Die zentrale Auswahl selbst erscheint überall dort, wo im HTML
       <div data-tg-stil></div>
   steht (Connector, Spieleübersicht). Das Skript füllt die Stelle mit den
   drei Knöpfen und hält sie mit der gespeicherten Wahl synchron.
   =========================================================================== */
(function () {
  "use strict";

  var SCHLUESSEL = "tg-stil";
  var WERTE = ["klassisch", "maedchen", "buben"];
  var TEXTE = { klassisch: "Klassisch", maedchen: "Mädchen", buben: "Buben" };

  function gueltig(wert) {
    return WERTE.indexOf(wert) >= 0 ? wert : "klassisch";
  }

  /* Die persönliche Wahl. Unbekanntes oder fehlender Speicher heißt klassisch. */
  function hole() {
    try { return gueltig(localStorage.getItem(SCHLUESSEL)); } catch (e) { return "klassisch"; }
  }

  /* Dieselbe Wahl im Vokabular eines Spiels. */
  function lokal(abbild) {
    var wert = hole();
    if (!abbild) return wert;
    return abbild[wert] || "klassisch";
  }

  function zeigeKnopfZustand(gruppe, wert) {
    var knoepfe = gruppe.querySelectorAll("button[data-stil]");
    for (var i = 0; i < knoepfe.length; i++) {
      var an = knoepfe[i].dataset.stil === wert;
      knoepfe[i].classList.toggle("aktiv", an);
      knoepfe[i].setAttribute("aria-pressed", an ? "true" : "false");
    }
  }

  function anwenden(wert) {
    document.documentElement.dataset.stil = wert;
    /* Natives Erscheinungsbild mitführen: Auswahlfelder, Kontrollkästchen und
       Bildlaufleisten zeichnet der Browser sonst nach dem Betriebssystem.
       Im System-Dunkelmodus würden so schwarze Felder auf den hellen
       Mädchen- und Klassisch-Flächen entstehen. */
    document.documentElement.style.colorScheme = wert === "buben" ? "dark" : "light";
    var gruppen = document.querySelectorAll("[data-tg-stil]");
    for (var i = 0; i < gruppen.length; i++) zeigeKnopfZustand(gruppen[i], wert);
  }

  /* Wahl setzen: speichern, anzeigen, alle offenen Seiten und Spiele rufen. */
  function setze(wert) {
    wert = gueltig(wert);
    try { localStorage.setItem(SCHLUESSEL, wert); } catch (e) { /* ohne Speicher gilt die Wahl für diese Seite */ }
    anwenden(wert);
    window.dispatchEvent(new CustomEvent("tg-stil", { detail: wert }));
    return wert;
  }

  /* Ein Spiel meldet sich hier an und wird bei einer neuen Wahl neu gezeichnet. */
  function beiWechsel(cb) {
    if (typeof cb !== "function") return;
    window.addEventListener("tg-stil", function (e) { cb(e.detail); });
    window.addEventListener("storage", function (e) {
      if (e.key !== SCHLUESSEL) return;
      var wert = hole();
      anwenden(wert);
      cb(wert);
    });
  }

  function style() {
    if (document.getElementById("tg-stil-css")) return;
    var css = document.createElement("style");
    css.id = "tg-stil-css";
    css.textContent =
      ".tg-stil{display:flex;flex-wrap:wrap;align-items:center;gap:6px;font-size:12px;color:inherit}" +
      ".tg-stil .tg-titel{opacity:.75;letter-spacing:.04em;margin-right:2px}" +
      ".tg-stil button{font:inherit;font-size:12.5px;line-height:1.2;padding:5px 11px;border-radius:999px;" +
      "cursor:pointer;border:1px solid currentColor;background:transparent;color:inherit;opacity:.7}" +
      ".tg-stil button:hover,.tg-stil button:focus-visible{opacity:1}" +
      ".tg-stil button.aktiv{opacity:1;font-weight:700;border-color:transparent;background:#e8cd8e;color:#0b241d}" +
      ".tg-stil button[data-stil=\"maedchen\"].aktiv{background:linear-gradient(#ffd0e4,#ff9ec4);color:#43182c}" +
      ".tg-stil button[data-stil=\"buben\"].aktiv{background:#e0242f;color:#ffffff}";
    document.head.appendChild(css);
  }

  /* Baut die drei Knöpfe in jede vorbereitete Stelle. */
  function steuerung(ziel) {
    if (!ziel) return null;
    style();
    ziel.classList.add("tg-stil");
    ziel.setAttribute("role", "group");
    ziel.setAttribute("aria-label", "Grafikstil");
    ziel.innerHTML = "";

    var titel = document.createElement("span");
    titel.className = "tg-titel";
    titel.textContent = "Grafikstil";
    ziel.appendChild(titel);

    WERTE.forEach(function (wert) {
      var knopf = document.createElement("button");
      knopf.type = "button";
      knopf.dataset.stil = wert;
      knopf.textContent = TEXTE[wert];
      knopf.title = "Gilt für alle Spiele – nur für dich";
      knopf.addEventListener("click", function () { setze(wert); });
      ziel.appendChild(knopf);
    });

    zeigeKnopfZustand(ziel, hole());
    return ziel;
  }

  function montiere() {
    var stellen = document.querySelectorAll("[data-tg-stil]");
    for (var i = 0; i < stellen.length; i++) steuerung(stellen[i]);
  }

  window.TGStil = {
    schluessel: SCHLUESSEL,
    werte: WERTE.slice(),
    hole: hole,
    lokal: lokal,
    setze: setze,
    beiWechsel: beiWechsel,
    steuerung: steuerung
  };

  anwenden(hole());
  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", function () { montiere(); anwenden(hole()); });
  } else {
    montiere();
  }
  window.addEventListener("storage", function (e) {
    if (e.key === SCHLUESSEL) anwenden(hole());
  });
})();
