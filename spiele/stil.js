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

   Die zentrale Auswahl steht nur auf der Connector-Seite. Dort füllt dieses
   Skript jede Stelle
       <div data-tg-stil></div>
   mit den drei Knöpfen und hält sie mit der gespeicherten Wahl synchron.
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

  /* --- Mädchen am PC ---------------------------------------------------
     Am Handy ist die kräftige Rosafläche richtig: sie füllt wenig Raum und
     gibt der Seite ihren Ton. Auf großen Bildschirmen wird dieselbe Fläche
     sonst zur Leuchtwand. Deshalb liegt dort ein ruhiger Verlauf aus Rosé,
     Altrosa und Weiß unter einer sehr leisen Herzkulisse – gleiche
     Farbfamilie, deutlich weniger Leuchtkraft. Nur Anzeige: keine Regel,
     kein Spielstand, und unter 641 px ändert sich nichts. */
  /* Ein Verlauf für die ganze Fläche: html und body tragen denselben, damit
     auch der Bereich unter einem kurzen body (leere Fläche am PC) im Ton
     bleibt und keine Kante entsteht. */
  var HERZKULISSE =
        "background-image:url(data:image/svg+xml;utf8,%3Csvg%20xmlns%3D%27http%3A%2F%2Fwww.w3.org%2F2000%2Fsvg%27%20width%3D%27264%27%20height%3D%27264%27%20viewBox%3D%270%200%20264%20264%27%3E%3Cg%20fill%3D%27%23b0476d%27%3E%3Cg%20transform%3D%27translate%2846%2050%29%20rotate%28-16%29%20scale%281.05%29%27%20fill-opacity%3D%270.048%27%3E%3Cpath%20d%3D%27M12%2021C7%2017.2%202%2013.6%202%208.9%202%205.6%204.5%203.2%207.7%203.2c1.8%200%203.3.9%204.3%202.4%201-1.5%202.5-2.4%204.3-2.4C19.5%203.2%2022%205.6%2022%208.9c0%204.7-5%208.3-10%2012.1z%27%2F%3E%3C%2Fg%3E%3Cg%20transform%3D%27translate%28178%2034%29%20rotate%2811%29%20scale%280.6%29%27%20fill-opacity%3D%270.038%27%3E%3Cpath%20d%3D%27M12%2021C7%2017.2%202%2013.6%202%208.9%202%205.6%204.5%203.2%207.7%203.2c1.8%200%203.3.9%204.3%202.4%201-1.5%202.5-2.4%204.3-2.4C19.5%203.2%2022%205.6%2022%208.9c0%204.7-5%208.3-10%2012.1z%27%2F%3E%3C%2Fg%3E%3Cg%20transform%3D%27translate%28210%20152%29%20rotate%287%29%20scale%280.9%29%27%20fill-opacity%3D%270.034%27%3E%3Cpath%20d%3D%27M12%2021C7%2017.2%202%2013.6%202%208.9%202%205.6%204.5%203.2%207.7%203.2c1.8%200%203.3.9%204.3%202.4%201-1.5%202.5-2.4%204.3-2.4C19.5%203.2%2022%205.6%2022%208.9c0%204.7-5%208.3-10%2012.1z%27%2F%3E%3C%2Fg%3E%3Cg%20transform%3D%27translate%2892%20190%29%20rotate%28-12%29%20scale%280.52%29%27%20fill-opacity%3D%270.042%27%3E%3Cpath%20d%3D%27M12%2021C7%2017.2%202%2013.6%202%208.9%202%205.6%204.5%203.2%207.7%203.2c1.8%200%203.3.9%204.3%202.4%201-1.5%202.5-2.4%204.3-2.4C19.5%203.2%2022%205.6%2022%208.9c0%204.7-5%208.3-10%2012.1z%27%2F%3E%3C%2Fg%3E%3Cg%20transform%3D%27translate%2834%20236%29%20rotate%2813%29%20scale%280.72%29%27%20fill-opacity%3D%270.036%27%3E%3Cpath%20d%3D%27M12%2021C7%2017.2%202%2013.6%202%208.9%202%205.6%204.5%203.2%207.7%203.2c1.8%200%203.3.9%204.3%202.4%201-1.5%202.5-2.4%204.3-2.4C19.5%203.2%2022%205.6%2022%208.9c0%204.7-5%208.3-10%2012.1z%27%2F%3E%3C%2Fg%3E%3C%2Fg%3E%3C%2Fsvg%3E)," +
      "radial-gradient(1100px 760px at 50% -18%,rgba(255,255,255,.92),rgba(255,255,255,0) 62%)," +
      "radial-gradient(720px 540px at 92% 6%,rgba(255,205,227,.45),rgba(255,205,227,0) 72%)," +
      "radial-gradient(700px 600px at 4% 94%,rgba(200,124,158,.26),rgba(200,124,158,0) 74%)," +
      "linear-gradient(155deg,#fdf8fa 0%,#fae9f0 32%,#f4d7e2 66%,#eec9d7 100%);" +
    "background-repeat:repeat,no-repeat,no-repeat,no-repeat,no-repeat;" +
    "background-size:264px 264px,auto,auto,auto,cover;" +
    "background-position:0 0,0 0,0 0,0 0,0 0;" +
    "background-attachment:fixed;" +
    "background-color:#f6e4ec";

  var MAEDCHEN_PC =
    "@media (min-width: 641px) {" +
      'html[data-stil="maedchen"]{' + HERZKULISSE + "}" +
      'html[data-stil="maedchen"] body{' + HERZKULISSE + "}" +
      'html[data-stil="maedchen"] ::selection{background:#f6c3d7;color:#3a0f24}' +
      'html[data-stil="maedchen"] a:focus-visible,' +
      'html[data-stil="maedchen"] button:focus-visible,' +
      'html[data-stil="maedchen"] input:focus-visible,' +
      'html[data-stil="maedchen"] select:focus-visible,' +
      'html[data-stil="maedchen"] summary:focus-visible{outline:2px solid #a11e57;outline-offset:2px}' +
    "}";

  function schmuck() {
    if (document.getElementById("tg-maedchen-pc")) return;
    var st = document.createElement("style");
    st.id = "tg-maedchen-pc";
    st.textContent = MAEDCHEN_PC;
    document.head.appendChild(st);
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
    document.addEventListener("DOMContentLoaded", function () { montiere(); schmuck(); anwenden(hole()); });
  } else {
    montiere();
    schmuck();
  }
  window.addEventListener("storage", function (e) {
    if (e.key === SCHLUESSEL) anwenden(hole());
  });
})();
