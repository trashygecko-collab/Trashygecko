/* ===========================================================================
   HALMA – Spielsteuerung

   Verbindet Menü, Regelwerk, KI, Rätsel, Lernsystem und – wenn die Seite mit
   ?room=…&id=… geöffnet wurde – den bestehenden Connector (/api/halma).
   Die Regeln selbst stehen in js/regeln.js; hier wird nur angezeigt und
   bedient. Im Online-Spiel entscheidet ausschließlich der Server.
   =========================================================================== */
(function () {
  "use strict";

  var $ = function (id) { return document.getElementById(id); };
  var PARAMS = new URLSearchParams(location.search);
  var ONLINE = PARAMS.has("room") && PARAMS.has("id");
  var API = "/api/halma";

  var S = {
    modus: "menue",          // menue | lokal | raetsel | online
    variante: "klassisch",
    anzahl: 2,
    gegner: "mensch",        // mensch | computer
    stufe: "mittel",
    lernen: true,
    brett: null, belegung: null, plaetze: [], farben: [],
    spieler: [],             // {name, art, stufe}
    current: 0, auswahl: -1, ziele: [],
    letzterZug: null, verlauf: [], sieger: -1, zugNr: 1,
    raetsel: null, raetselZuege: 0, zugLimit: 0,
    online: false, room: PARAMS.get("room") || "", ich: PARAMS.get("id") || "",
    game: null, busy: false, lernUhr: null, kiUhr: null, pollUhr: null,
    lupeWahl: null, lupeMitte: -1, lupePunkt: { x: 0, y: 0 },
    seite: 1
  };

  /* ------------------------------------------------------------- Menü -- */
  var VARIANTEN = [
    { id: "klassisch", titel: "Klassisches Halma",
      text: "16×16-Feld, 19 Figuren zu zweit, 13 zu viert – Ziel ist das gegenüberliegende Eckhaus.",
      spielerzahlen: [2, 4], art: "multi" },
    { id: "stern", titel: "Sternhalma",
      text: "Sechs Zacken mit 121 Feldern. Ziel ist die gegenüberliegende Zacke.",
      spielerzahlen: [2, 3, 4, 6], art: "multi" },
    { id: "sprung", titel: "Sprunghalma",
      text: "Weite Sprünge über eine Figur – auch über mehrere Felder Abstand.",
      spielerzahlen: [2, 4], art: "multi" },
    { id: "solo", titel: "Solo gegen Computer",
      text: "Du gegen den Computer: klassisch oder Stern, drei Schwierigkeitsstufen.",
      spielerzahlen: [2], art: "solo" },
    { id: "knobel", titel: "Solo – Knobelmodus",
      text: "Vorgegebene Stellungen, Zuglimit, Zugzähler und Bestwert je Aufgabe.",
      spielerzahlen: [1], art: "raetsel" }
  ];

  var VORSCHAU = {
    klassisch: '<svg viewBox="0 0 44 44" aria-hidden="true"><rect x="2" y="2" width="40" height="40" rx="4" fill="none" stroke="currentColor" stroke-opacity=".45"/><path d="M2 12 12 2" stroke="currentColor" stroke-opacity=".25"/><g fill="currentColor"><circle cx="7" cy="7" r="2"/><circle cx="13" cy="7" r="2"/><circle cx="7" cy="13" r="2"/><circle cx="37" cy="37" r="2"/><circle cx="31" cy="37" r="2"/><circle cx="37" cy="31" r="2"/></g></svg>',
    stern: '<svg viewBox="0 0 44 44" aria-hidden="true"><path d="M22 3 27 14 38 14 30 22 33 34 22 27 11 34 14 22 6 14 17 14Z" fill="none" stroke="currentColor" stroke-opacity=".45"/><g fill="currentColor"><circle cx="22" cy="22" r="2.2"/><circle cx="22" cy="6" r="1.8"/><circle cx="38" cy="22" r="1.8"/><circle cx="6" cy="22" r="1.8"/></g></svg>',
    sprung: '<svg viewBox="0 0 44 44" aria-hidden="true"><rect x="2" y="2" width="40" height="40" rx="4" fill="none" stroke="currentColor" stroke-opacity=".45"/><path d="M8 36 20 24 36 8" fill="none" stroke="currentColor" stroke-width="1.6" stroke-dasharray="3 2"/><circle cx="20" cy="24" r="3" fill="currentColor" fill-opacity=".55"/><circle cx="8" cy="36" r="3" fill="currentColor"/><circle cx="36" cy="8" r="3" fill="currentColor"/></svg>',
    solo: '<svg viewBox="0 0 44 44" aria-hidden="true"><rect x="2" y="2" width="40" height="40" rx="4" fill="none" stroke="currentColor" stroke-opacity=".45"/><circle cx="12" cy="32" r="3.4" fill="currentColor"/><rect x="26" y="10" width="10" height="13" rx="2.5" fill="none" stroke="currentColor" stroke-width="1.6"/><circle cx="28.5" cy="14" r="1.2" fill="currentColor"/><circle cx="33.5" cy="14" r="1.2" fill="currentColor"/><circle cx="31" cy="19.5" r="1.2" fill="currentColor"/></svg>',
    knobel: '<svg viewBox="0 0 44 44" aria-hidden="true"><rect x="2" y="2" width="40" height="40" rx="4" fill="none" stroke="currentColor" stroke-opacity=".45"/><text x="22" y="30" font-size="20" text-anchor="middle" fill="currentColor" font-family="Georgia,serif">?</text></svg>'
  };

  function variantenBauen() {
    var box = $("varianten");
    box.innerHTML = "";
    VARIANTEN.forEach(function (v) {
      var b = document.createElement("button");
      b.type = "button";
      b.className = "variante" + (v.id === S.variante ? " aktiv" : "");
      b.dataset.variante = v.id;
      b.innerHTML = VORSCHAU[v.id] + "<span><b>" + v.titel + "</b><span>" + v.text + "</span></span>";
      b.addEventListener("click", function () { varianteWaehlen(v.id); });
      box.appendChild(b);
    });
  }

  function varianteWaehlen(id) {
    S.variante = id;
    var v = varianteById(id);
    if (v.art === "raetsel") S.anzahl = 1;
    else if (v.spielerzahlen.indexOf(S.anzahl) < 0) S.anzahl = v.spielerzahlen[0];
    variantenBauen();
    optionenBauen();
  }

  function varianteById(id) {
    for (var i = 0; i < VARIANTEN.length; i++) if (VARIANTEN[i].id === id) return VARIANTEN[i];
    return VARIANTEN[0];
  }

  function optionenBauen() {
    var v = varianteById(S.variante);
    var box = $("optionen");
    box.innerHTML = "";
    $("aufgabenblock").hidden = v.art !== "raetsel";

    if (v.art === "raetsel") {
      aufgabenBauen();
      $("startKnopf").textContent = "Aufgabe starten";
      $("menueHinweis").textContent = "Jede Aufgabe hat eine geprüfte Lösung und ein Zuglimit. " +
        "Dein Bestwert wird gespeichert.";
      $("startKnopf").hidden = false;
      $("onlineKnopf").hidden = true;
      return;
    }

    if (v.art === "solo") {
      auswahlFeld(box, "Spielart", "soloart", [["klassisch", "Klassisches Halma"], ["stern", "Sternhalma"]],
        S.soloArt || "klassisch", function (wert) { S.soloArt = wert; });
      auswahlFeld(box, "Schwierigkeit", "stufe", [["leicht", "Leicht"], ["mittel", "Mittel"], ["schwer", "Schwer"]],
        S.stufe, function (wert) { S.stufe = wert; });
      schalterFeld(box, "Lernhilfe", "lernen", S.lernen, function (an) { S.lernen = an; });
      $("startKnopf").textContent = "Gegen Computer spielen";
      $("menueHinweis").textContent = "Der Computer nutzt dieselben Regeln wie du – " +
        "die Stufe ändert nur, wie gründlich er überlegt.";
      $("startKnopf").hidden = false;
      $("onlineKnopf").hidden = true;
      return;
    }

    /* Mehrspieler-Varianten */
    auswahlFeld(box, "Spieler", "anzahl",
      v.spielerzahlen.map(function (n) { return [String(n), n + " Spieler"]; }),
      String(S.anzahl), function (wert) { S.anzahl = Number(wert); optionenBauen(); });
    auswahlFeld(box, "Mitspieler", "gegner",
      [["mensch", "alle am Gerät / im Raum"], ["computer", "Computer füllt auf"]],
      S.gegner, function (wert) { S.gegner = wert; optionenBauen(); });
    if (S.gegner === "computer") {
      auswahlFeld(box, "Schwierigkeit", "stufe",
        [["leicht", "Leicht"], ["mittel", "Mittel"], ["schwer", "Schwer"]],
        S.stufe, function (wert) { S.stufe = wert; });
    }
    schalterFeld(box, "Lernhilfe", "lernen", S.lernen, function (an) { S.lernen = an; });

    $("startKnopf").textContent = "Spiel starten";
    $("startKnopf").hidden = false;
    $("onlineKnopf").hidden = !ONLINE;
    $("menueHinweis").textContent = ONLINE
      ? "Raum " + S.room + ": hier eröffnen oder mitspielen. Gewürfelt wird nicht – " +
        "alle Züge prüft der Server."
      : "Zwei Menschen am selben Gerät, ein Mensch gegen Computer, oder im Connector-Raum " +
        "gegen andere. «Regeln» erklärt jede Variante.";
  }

  function auswahlFeld(box, titel, name, werte, wert, aendern) {
    var d = document.createElement("div");
    d.className = "feldgruppe";
    var l = document.createElement("label");
    l.textContent = titel;
    l.setAttribute("for", "opt-" + name);
    var s = document.createElement("select");
    s.id = "opt-" + name;
    werte.forEach(function (p) {
      var o = document.createElement("option");
      o.value = p[0];
      o.textContent = p[1];
      if (p[0] === wert) o.selected = true;
      s.appendChild(o);
    });
    s.addEventListener("change", function () { aendern(s.value); });
    d.append(l, s);
    box.appendChild(d);
    return s;
  }

  function schalterFeld(box, titel, name, wert, aendern) {
    var d = document.createElement("div");
    d.className = "feldgruppe";
    var l = document.createElement("label");
    l.textContent = "Hilfen";
    var w = document.createElement("label");
    w.className = "schalter";
    var c = document.createElement("input");
    c.type = "checkbox";
    c.checked = !!wert;
    c.id = "opt-" + name;
    c.addEventListener("change", function () { aendern(c.checked); });
    var t = document.createElement("span");
    t.textContent = titel;
    w.append(c, t);
    d.append(l, w);
    box.appendChild(d);
  }

  function aufgabenBauen() {
    var box = $("aufgaben");
    box.innerHTML = "";
    HB.Raetsel.AUFGABEN.forEach(function (a) {
      var b = document.createElement("button");
      b.type = "button";
      b.className = "aufgabe" + (a.id === S.raetselId ? " aktiv" : "");
      var best = bestwert(a.id);
      b.innerHTML = "<b>" + a.titel + "</b><span>" + a.text.slice(0, 78) + "…</span>" +
        '<span class="best">' + (best ? "Bestwert: " + zuegeText(best) : "noch nicht gelöst") +
        " · höchstens " + a.maxZuege + " Züge</span>";
      b.addEventListener("click", function () { S.raetselId = a.id; aufgabenBauen(); });
      box.appendChild(b);
    });
    if (!S.raetselId) S.raetselId = HB.Raetsel.AUFGABEN[0].id;
  }

  function bestwert(id) {
    try { return Number(localStorage.getItem("tg-halma-raetsel-" + id)) || 0; } catch (e) { return 0; }
  }
  function bestwertSetzen(id, zuege) {
    var alt = bestwert(id);
    if (!alt || zuege < alt) {
      try { localStorage.setItem("tg-halma-raetsel-" + id, String(zuege)); } catch (e) { /* egal */ }
    }
  }

  /* ------------------------------------------------------------ Start -- */
  /* Auswahl, letzter Zug, Verlauf und Sieger einer Runde – für jeden Start. */
  function zugStandZuruecksetzen() {
    S.auswahl = -1;
    S.ziele = [];
    S.letzterZug = null;
    S.verlauf = [];
    S.sieger = -1;
  }

  function spielStarten() {
    var v = varianteById(S.variante);
    if (v.art === "raetsel") return raetselStarten(S.raetselId);
    if (v.art === "solo") return lokalStarten(S.soloArt || "klassisch", 2, "computer", true);
    return lokalStarten(S.variante, S.anzahl, S.gegner, false);
  }

  function lokalStarten(varianteId, anzahl, gegner, solo) {
    S.modus = solo ? "solo" : "lokal";
    S.raetsel = null;
    S.brett = varianteId === "stern" ? HB.Brett.stern() : HB.Brett.klassisch(anzahl);
    S.plaetze = HB.Regeln.plaetze(varianteId, anzahl);
    S.belegung = HB.Regeln.startAufstellung(S.brett, varianteId, anzahl);
    S.variante = varianteId;
    S.anzahl = anzahl;
    S.farben = HB.Design.aktuell().figuren;
    S.spieler = S.plaetze.map(function (p, k) {
      var computer = (gegner === "computer" && k > 0) || (solo && k === 1);
      return { name: computer ? "Computer" + (k > 1 ? " " + k : "") : "Spieler " + (k + 1),
               art: computer ? "computer" : "mensch", stufe: S.stufe };
    });
    S.current = 0;
    zugStandZuruecksetzen();
    S.zugNr = 1;
    spielZeigen();
    zeichne();
    naechsterZugFallsComputer();
  }

  function raetselStarten(id) {
    var a = HB.Raetsel.nachId(id || S.raetselId) || HB.Raetsel.AUFGABEN[0];
    S.raetsel = a;
    S.modus = "raetsel";
    S.variante = a.variante;
    S.brett = HB.Raetsel.brettFuer(a);
    S.belegung = HB.Raetsel.aufstellung(a, S.brett);
    S.plaetze = [0];
    S.farben = HB.Design.aktuell().figuren;
    S.spieler = [{ name: "Du", art: "mensch" }];
    S.current = 0;
    zugStandZuruecksetzen();
    S.raetselZuege = 0;
    S.zugLimit = a.maxZuege;
    spielZeigen();
    $("tippKnopf").hidden = false;
    $("loesungKnopf").hidden = false;
    zeichne();
    lern("<b>" + a.titel + "</b> – " + a.text + " Höchstens " + a.maxZuege + " Züge.", 9000);
  }

  function spielZeigen() {
    $("menue").hidden = true;
    $("spielbereich").hidden = false;
    $("spielerbereich").hidden = false;
    $("varianteMarke").textContent = S.raetsel ? "Knobeln" :
      (HB.Regeln.VARIANTEN[S.variante] || {}).name || "Halma";
    $("tippKnopf").hidden = !S.raetsel;
    $("loesungKnopf").hidden = !S.raetsel;
    $("lernKnopf").textContent = S.lernen ? "Lernhilfe an" : "Lernhilfe aus";
    brettBauen();
    passeBrett();
    /* Der Rahmen bekommt seine Höhe erst nach dem Einblenden – deshalb
       zweimal nachfassen (rAF läuft in Hintergrund-Tabs nicht). */
    setTimeout(passeBrett, 40);
    setTimeout(passeBrett, 220);
  }

  /* ----------------------------------------------------------- Brett -- */
  function brettBauen() {
    var b = S.brett;
    var brett = $("brett");
    brett.innerHTML = "";
    brett.className = "brett" + (b.art === "klassisch" ? " quadrat" : "");
    var r = HB.Brett.rahmen(b);
    var verhaeltnis = r.breite / r.hoehe;
    brett.dataset.verhaeltnis = String(verhaeltnis);
    var feldBreite = 100 / r.breite;
    var feldHoehe = 100 / r.hoehe;
    b.felder.forEach(function (f, i) {
      var d = document.createElement("div");
      d.className = "feld";
      d.dataset.i = String(i);
      var w = Math.min(feldBreite, feldHoehe * verhaeltnis) * 0.92;
      d.style.width = w + "%";
      d.style.aspectRatio = "1";
      d.style.left = ((f.px - r.minX + 0.5) * feldBreite - w / 2) + "%";
      d.style.top = ((f.py - r.minY + 0.5) * feldHoehe - w / (2 * verhaeltnis)) + "%";
      brett.appendChild(d);
    });
    brett.onclick = function (ev) {
      if (lupeAktiv()) return;                    // erledigt der Rahmen
      var f = ev.target.closest(".feld");
      if (f) feldKlick(Number(f.dataset.i));
    };
    $("brettrahmen").onclick = function (ev) {
      if (!lupeAktiv() || ev.target.closest(".lupe")) return;
      var f = ev.target.closest(".feld");
      var i = f ? Number(f.dataset.i) : naechstesFeld(ev.clientX, ev.clientY);
      if (i >= 0) lupeOeffnen(i, ev.clientX, ev.clientY);
    };
  }

  function passeBrett() {
    var rahmen = $("brettrahmen");
    if (!rahmen || !S.brett) return;
    var box = rahmen.getBoundingClientRect();
    var b = S.brett;
    var r = HB.Brett.rahmen(b);
    var verhaeltnis = r.breite / r.hoehe;
    var hoehe = Math.min(box.height, box.width / verhaeltnis);
    var breite = hoehe * verhaeltnis;
    var el = $("brett");
    el.style.width = Math.max(120, Math.floor(breite)) + "px";
    el.style.height = Math.max(120, Math.floor(hoehe)) + "px";
    lupeKnopfAktualisieren();
  }

  /* ---------------------------------------------------------- Feldlupe -- */
  /* Am Handy sind die 16×16-Felder nur gut 20 px breit – ein Daumen trifft
     dann leicht das Nachbarfeld. Die Lupe zeigt den 5×5-Ausschnitt um das
     angetippte Feld vergrößert (60 px je Feld); erst der zweite Tipp wählt.
     Sie erscheint nur, wo die Felder wirklich schmal sind (unter 26 px). */
  var LUPE_WEITE = 5;

  function feldBreite() {
    var el = $("brett");
    if (!el || !S.brett) return 99;
    return el.clientWidth / HB.Brett.rahmen(S.brett).breite;
  }

  function lupeSinnvoll() {
    if (!S.brett || S.modus === "menue") return false;
    return HB.Brett.rahmen(S.brett).breite >= 12 && feldBreite() < 26;
  }

  function lupeAktiv() {
    return S.lupeWahl !== false && lupeSinnvoll();
  }

  function lupeKnopfAktualisieren() {
    var k = $("lupeKnopf");
    if (!k) return;
    k.hidden = !lupeSinnvoll();
    k.textContent = lupeAktiv() ? "Lupe an" : "Lupe aus";
    k.setAttribute("aria-pressed", lupeAktiv() ? "true" : "false");
  }

  function lupeOffen() {
    var l = $("lupe");
    return !!l && !l.hidden;
  }

  function lupeZu() {
    if (!lupeOffen()) return;
    $("lupe").hidden = true;
    $("lupeGitter").innerHTML = "";
  }

  /* Baut die Lupe neu – aus dem Brett, damit Figuren, Ziele und Marken
     genau so aussehen wie unten. */
  function lupeBauen() {
    var l = $("lupe"), g = $("lupeGitter"), quelle = $("brett");
    if (!l || !g || !quelle || !S.brett || S.lupeMitte < 0) return;
    var r = HB.Brett.rahmen(S.brett);
    var spalten = r.breite, reihen = r.hoehe;
    var i = S.lupeMitte;
    var halb = (LUPE_WEITE - 1) / 2;
    var x0 = Math.min(Math.max((i % spalten) - halb, 0), Math.max(spalten - LUPE_WEITE, 0));
    var y0 = Math.min(Math.max(Math.floor(i / spalten) - halb, 0), Math.max(reihen - LUPE_WEITE, 0));
    g.innerHTML = "";
    for (var yy = 0; yy < LUPE_WEITE; yy++) {
      for (var xx = 0; xx < LUPE_WEITE; xx++) {
        var idx = (y0 + yy) * spalten + (x0 + xx);
        var zelle = document.createElement("div");
        zelle.className = "lupe-zelle" + (idx === i ? " mitte" : "");
        zelle.dataset.i = String(idx);
        zelle.setAttribute("role", "button");
        zelle.tabIndex = 0;
        var original = quelle.children[idx];
        if (original) zelle.appendChild(original.cloneNode(true));
        g.appendChild(zelle);
      }
    }
    g.setAttribute("aria-label", "Vergrößerung um " + feldName(i));
    if (!lupeOffen()) l.hidden = false;
    var box = l.getBoundingClientRect();
    var py = S.lupePunkt.y;
    /* Waagerecht immer mittig: so liegt die Lupe bei jedem Finger weit
       genug vom Bildrand entfernt und das Feld bleibt ganz sichtbar. */
    var links = Math.max((window.innerWidth - box.width) / 2, 8);
    var oben = py - box.height - 18;              // über dem Finger
    if (oben < 8) oben = Math.min(py + 26, window.innerHeight - box.height - 8);
    if (oben < 8) oben = 8;
    l.style.left = Math.round(links) + "px";
    l.style.top = Math.round(oben) + "px";
  }

  /* Nächstes Feld zur Tippstelle – ein Tipper knapp neben dem Brett landet
     damit im Randfeld statt im Nichts. */
  function naechstesFeld(x, y) {
    var el = $("brett");
    if (!el) return -1;
    var best = -1, dist = Infinity;
    for (var i = 0; i < el.children.length; i++) {
      var b = el.children[i].getBoundingClientRect();
      var dx = x - (b.left + b.width / 2), dy = y - (b.top + b.height / 2);
      var d = dx * dx + dy * dy;
      if (d < dist) { dist = d; best = i; }
    }
    return best;
  }

  function lupeOeffnen(i, x, y) {
    S.lupeMitte = i;
    S.lupePunkt = { x: x, y: y };
    if (!x && !y) {                               // Tastatur: Feldmitte
      var el = $("brett").children[i];
      if (el) {
        var b = el.getBoundingClientRect();
        S.lupePunkt = { x: b.left + b.width / 2, y: b.top + b.height / 2 };
      }
    }
    lupeBauen();
  }

  function lupeGitterKlick(ev) {
    var z = ev.target.closest(".lupe-zelle");
    if (!z) return;
    lupeZu();
    feldKlick(Number(z.dataset.i));
  }

  function zielFelder() {
    var set = {};
    S.plaetze.forEach(function (p, k) {
      HB.Regeln.zielZone(S.brett, p).forEach(function (i) { set[i] = k; });
    });
    return set;
  }

  function zeichne() {
    if (!S.brett) return;
    var ziele = zielFelder();
    var startFelder = {};
    S.plaetze.forEach(function (p) {
      S.brett.zonen[p].forEach(function (i) { startFelder[i] = true; });
    });
    var marken = {};
    S.ziele.forEach(function (z) { marken[z.feld] = z.art; });
    var letzteFelder = {};
    if (S.letzterZug) {
      letzteFelder[S.letzterZug.von] = true;
      letzteFelder[S.letzterZug.nach] = true;
      (S.letzterZug.weg || []).forEach(function (w) { letzteFelder[w] = true; });
    }

    var felder = $("brett").children;
    for (var i = 0; i < felder.length; i++) {
      var el = felder[i];
      el.className = "feld";
      el.innerHTML = "";
      if (ziele[i] !== undefined || startFelder[i]) el.classList.add("zone");
      if (startFelder[i]) el.classList.add("start");
      if (marken[i] === "sprung") el.classList.add("mark", "sprung");
      else if (marken[i]) el.classList.add("mark");
      if (letzteFelder[i]) el.classList.add("letzter");
      var wer = S.belegung[i];
      if (wer >= 0) {
        var werK = spielerZuPlatz(wer);
        if (werK < 0) werK = wer;
        var fig = document.createElement("div");
        fig.className = "figur" + (werK === S.current ? " eigene" : " fremd") +
          (S.auswahl === i ? " gewaehlt" : "");
        fig.style.cssText = HB.Design.figurStil(S.farben[werK % S.farben.length].farbe);
        fig.style.width = "82%";
        fig.style.aspectRatio = "1";
        fig.style.left = "9%";
        fig.style.top = "9%";
        fig.dataset.figur = String(i);
        fig.title = SBeg(S.brett, wer, i);
        el.appendChild(fig);
      }
    }

    /* Kopfzeile */
    var dran = S.spieler[S.current];
    $("zugMarke").textContent = S.modus === "raetsel"
      ? "Zug " + S.raetselZuege + " / " + S.zugLimit
      : "Zug " + S.zugNr;
    $("dranMarke").textContent = S.sieger >= 0 ? "Partie beendet"
      : (S.modus === "raetsel" ? (S.raetselZuege + " von " + S.zugLimit + " Zügen")
        : (dran ? (S.online && S.current === serverPlatz()
          ? "Du bist dran" : nameVon(S.current) + " ist dran") : "–"));
    $("dranMarke").classList.toggle("dran", S.sieger < 0);

    letzterText();
    spielerZeichnen();
    passeBrett();
    if (lupeOffen()) lupeBauen();
  }

  /* Figur trägt ihren Platz (die Ecke); der Name gehört dem Spieler an
     diesem Platz – die Position in S.spieler ist nur die Zugreihenfolge. */
  function spielerZuPlatz(platz) {
    for (var k = 0; k < S.plaetze.length; k++) {
      if (S.plaetze[k] === platz) return k;
    }
    return -1;
  }

  function SBeg(brett, platz, feld) {
    var k = spielerZuPlatz(platz);
    var name = k >= 0 && S.spieler[k] ? S.spieler[k].name : "Figur";
    var drin = HB.Regeln.zielAbstand(brett, feld, platz) === 0;
    return name + (drin ? " (im Ziel)" : "");
  }

  function nameVon(platz) {
    return S.spieler[platz] ? S.spieler[platz].name : "Spieler " + (platz + 1);
  }

  function letzterText() {
    var z = S.letzterZug;
    if (!z) {
      $("letzterZug").textContent = S.modus === "raetsel"
        ? S.raetsel.text
        : "Ziel: alle eigenen Figuren in den gegenüberliegenden Bereich.";
      return;
    }
    var t = nameVon(z.spieler) + ": ";
    if (z.art === "sprung") {
      t += "Sprung" + (z.weg && z.weg.length > 1 ? "kette über " + z.weg.length + " Figuren" : "") +
        " auf Feld " + feldName(z.nach);
    } else {
      t += "Schritt auf Feld " + feldName(z.nach);
    }
    if (S.modus === "raetsel") {
      t = "Zug " + S.raetselZuege + " von " + S.zugLimit + " · " + t.replace(/^Du: /, "");
    }
    $("letzterZug").textContent = t;
  }

  function feldName(i) {
    if (!S.brett || i < 0 || !S.brett.felder[i]) return "–";
    var f = S.brett.felder[i];
    if (S.brett.art === "stern") return "(" + f.q + "/" + f.r + ")";
    return String.fromCharCode(65 + f.x) + (f.y + 1);
  }

  function spielerZeichnen() {
    var liste = $("spielerliste");
    liste.innerHTML = "";
    S.spieler.forEach(function (p, k) {
      var li = document.createElement("li");
      if (S.sieger < 0 && k === S.current) li.classList.add("dran");
      var punkt = document.createElement("span");
      punkt.className = "punkt";
      punkt.style.background = S.farben[k % S.farben.length].farbe;
      var name = document.createElement("span");
      name.className = "name";
      name.textContent = p.name + (p.art === "computer" ? " (Computer)" : "");
      var stand = document.createElement("span");
      stand.className = "stand";
      var imZiel = 0, gesamt = 0;
      if (S.plaetze[k] !== undefined) {
        imZiel = HB.Regeln.imZiel(S.brett, S.belegung, S.plaetze[k]);
        gesamt = HB.Regeln.zielZone(S.brett, S.plaetze[k]).length;
      }
      stand.textContent = S.modus === "raetsel" ? (k === 0 ? "du knobelst" : "")
        : imZiel + " / " + gesamt + " im Ziel";
      var balken = document.createElement("span");
      balken.className = "balken";
      var i = document.createElement("i");
      i.style.width = gesamt ? Math.round(imZiel / gesamt * 100) + "%" : "0%";
      balken.appendChild(i);
      li.append(punkt, name, stand, balken);
      liste.appendChild(li);
    });
    var vl = $("verlauf");
    vl.innerHTML = "";
    S.verlauf.slice(-12).reverse().forEach(function (t) {
      var li = document.createElement("li");
      li.textContent = t;
      vl.appendChild(li);
    });
  }

  /* ------------------------------------------------------------ Züge -- */
  function feldKlick(i) {
    if (S.sieger >= 0 || S.online) {
      if (S.online) return onlineKlick(i);
      return;
    }
    if (S.belegung[i] === S.plaetze[S.current]) {
      waehle(i);
      return;
    }
    if (S.auswahl >= 0) {
      var ziel = S.ziele.filter(function (z) { return z.feld === i; })[0];
      if (ziel) { macheZug(S.auswahl, i); return; }
      if (S.belegung[i] >= 0) { lern(HB.Lernen.fehlerText("besetzt", S.variante)); return; }
      var erlaubt = HB.Regeln.zuegeVon(S.brett, S.belegung, S.variante, S.auswahl);
      var doppelt = erlaubt.length === 0 && S.brett.sprung && S.brett.sprung[S.auswahl].some(function (p) {
        return S.belegung[p[0]] >= 0 && S.belegung[p[1]] >= 0;
      });
      lern(HB.Lernen.fehlerText(doppelt ? "doppelt" : "zuweit", S.variante), 4200);
      return;
    }
    if (S.belegung[i] >= 0) {
      lern(HB.Lernen.fehlerText("fremd", S.variante));
      return;
    }
    lern("Wähle zuerst eine eigene Figur – die erlaubten Felder leuchten dann.", 3600);
  }

  function waehle(i) {
    S.auswahl = i;
    S.ziele = HB.Regeln.zuegeVon(S.brett, S.belegung, S.variante, i)
      .map(function (z) { return { feld: z.nach, art: z.art }; });
    zeichne();
    if (S.lernen) lern(HB.Lernen.auswahlText(S.brett, S.belegung, S.variante, i, S.plaetze[S.current]), 5200);
    if (!S.ziele.length) {
      var el = $("brett").children[i];
      if (el) el.classList.add("gerade");
    }
  }

  function macheZug(von, nach) {
    var zuege = HB.Regeln.zuegeVon(S.brett, S.belegung, S.variante, von);
    var zug = null;
    for (var i = 0; i < zuege.length; i++) if (zuege[i].nach === nach) { zug = zuege[i]; break; }
    if (!zug) return;

    S.belegung = HB.Regeln.ziehe(S.belegung, zug);
    S.letzterZug = { von: von, nach: nach, weg: zug.weg, art: zug.art, spieler: S.current };
    S.auswahl = -1;
    S.ziele = [];
    S.zugNr++;
    if (S.modus === "raetsel") S.raetselZuege++;

    var text = nameVon(S.current) + ": " + HB.Lernen.zugText(S.brett, zug) ;
    if (S.modus === "raetsel") text = "Zug " + S.raetselZuege + ": " + HB.Lernen.zugText(S.brett, zug);
    S.verlauf.push(text);

    zeichne();
    huepfAnimation(nach);

    if (S.modus === "raetsel") { raetselPruefen(); return; }
    if (istFertig(S.current)) { sieg(S.current); return; }
    S.current = (S.current + 1) % S.spieler.length;
    zeichne();
    if (S.lernen) lern(HB.Lernen.lageText(S.brett, S.belegung, S.plaetze[S.current], HB.Regeln.zielZone(S.brett, S.plaetze[S.current]).length), 3800);
    naechsterZugFallsComputer();
  }

  function huepfAnimation(i) {
    var el = $("brett").children[i];
    if (!el) return;
    var fig = el.querySelector(".figur");
    if (!fig) return;
    fig.classList.add("huepft");
    setTimeout(function () { fig.classList.remove("huepft"); }, 460);
  }

  function istFertig(platz) {
    return HB.Regeln.sieg(S.brett, S.belegung, S.plaetze[platz]);
  }

  function sieg(platz) {
    S.sieger = platz;
    zeichne();
    var imZiel = HB.Regeln.imZiel(S.brett, S.belegung, S.plaetze[platz]);
    dialog("Partie beendet",
      "<p><b>" + nameVon(platz) + "</b> hat alle " + imZiel + " Figuren im Zielbereich.</p>" +
      "<p>Züge insgesamt: " + (S.zugNr - 1) + ". " +
      "Auswertung: " + auswertungText() + "</p>",
      [["Nochmal", function () { dialogZu(); spielStarten(); }],
       ["Zum Menü", function () { dialogZu(); zumMenue(); }]]);
  }

  function auswertungText() {
    return S.spieler.map(function (p, k) {
      var gesamt = HB.Regeln.zielZone(S.brett, S.plaetze[k]).length;
      return p.name + " " + HB.Regeln.imZiel(S.brett, S.belegung, S.plaetze[k]) + "/" + gesamt;
    }).join(" · ");
  }

  function naechsterZugFallsComputer() {
    clearTimeout(S.kiUhr);
    if (S.sieger >= 0) return;
    var p = S.spieler[S.current];
    if (!p || p.art !== "computer") return;
    S.kiUhr = setTimeout(function () {
      var zug = HB.KI.waehleZug(S.brett, S.belegung, S.variante, S.plaetze[S.current], p.stufe || S.stufe, S.zugNr + 7);
      if (!zug) {
        /* Kein Zug möglich: aussetzen und weitergeben. */
        S.verlauf.push(p.name + " kann nicht ziehen und setzt aus.");
        S.current = (S.current + 1) % S.spieler.length;
        zeichne();
        naechsterZugFallsComputer();
        return;
      }
      macheZug(zug.von, zug.nach);
    }, HB.KI.tempo(p.stufe || S.stufe));
  }

  /* Ein Zug, zwei Züge – kleine Hilfe gegen „1 Züge“. */
  function zuegeText(n) { return n + (n === 1 ? " Zug" : " Züge"); }

  /* ---------------------------------------------------------- Rätsel -- */
  function raetselPruefen() {
    var a = S.raetsel;
    if (HB.Raetsel.erfuellt(a, S.brett, S.belegung)) {
      S.sieger = 0;
      bestwertSetzen(a.id, S.raetselZuege);
      var best = bestwert(a.id);
      var lob = S.raetselZuege <= (a.min || a.maxZuege) ? "Genau die Mindestzahl!" : "Geschafft!";
      dialog("Aufgabe gelöst",
        "<p>" + lob + " Du hast <b>" + zuegeText(S.raetselZuege) + "</b> gebraucht " +
        (a.min ? "(kürzest möglich: " + zuegeText(a.min) + ")" : "(erlaubt: " + zuegeText(a.maxZuege) + ")") + ".</p>" +
        "<p>Bestwert: " + zuegeText(best) + ".</p>",
        [["Nächste Aufgabe", function () { dialogZu(); naechsteAufgabe(); }],
         ["Nochmal", function () { dialogZu(); raetselStarten(a.id); }],
         ["Zum Menü", function () { dialogZu(); zumMenue(); }]]);
      return;
    }
    if (S.raetselZuege >= S.zugLimit) {
      S.sieger = 0;
      dialog("Zuglimit erreicht",
        "<p>Die Aufgabe war in " + S.zugLimit + " Zügen nicht geschafft. " +
        "Versuch es noch einmal – oder schau dir den Tipp an.</p>",
        [["Nochmal", function () { dialogZu(); raetselStarten(a.id); }],
         ["Zum Menü", function () { dialogZu(); zumMenue(); }]]);
    }
  }

  function naechsteAufgabe() {
    var liste = HB.Raetsel.AUFGABEN;
    var idx = liste.indexOf(S.raetsel);
    raetselStarten(liste[(idx + 1) % liste.length].id);
  }

  function tippZeigen() {
    if (!S.raetsel) return;
    var zug = HB.Raetsel.tippZug(S.raetsel);
    if (!zug) { lern("Für diese Stellung finde ich gerade keinen Zug.", 3000); return; }
    S.auswahl = zug.von;
    S.ziele = [{ feld: zug.nach, art: "sprung" }];
    zeichne();
    lern("Ein sinnvoller Zug: die markierte Figur auf das grüne Feld. " +
      "Der Zug ist geprüft und regelkonform.", 5200);
  }

  function loesungZeigen() {
    if (!S.raetsel) return;
    var r = HB.Raetsel.solver(S.raetsel, S.raetsel.maxZuege);
    if (!r) {
      dialog("Keine Lösung gefunden", "<p>Für diese Stellung wurde im Zuglimit keine Lösung gefunden.</p>",
        [["Schließen", dialogZu]]);
      return;
    }
    var zeilen = r.loesung.map(function (z, k) {
      return "<li>Zug " + (k + 1) + ": " + feldName(z.von) + " → " + feldName(z.nach) +
        (z.art === "sprung" ? " (Sprung)" : "") + "</li>";
    }).join("");
    dialog("Lösung (" + zuegeText(r.zuege) + ")",
      "<p>Die kürzeste gefundene Lösung:</p><ul>" + zeilen + "</ul>",
      [["Schließen", dialogZu]]);
  }

  /* ----------------------------------------------------------- Online -- */
  function serverPlatz() {
    if (!S.game) return -1;
    for (var i = 0; i < (S.game.spieler || []).length; i++) {
      if (S.game.spieler[i].id === S.ich) return i;
    }
    return -1;
  }

  async function onlineFetch(url, optionen) {
    var steuer = new AbortController();
    var lauf = Object.assign({}, optionen || {}, { signal: steuer.signal });
    var uhr = setTimeout(function () { steuer.abort(); }, 12000);
    try {
      var antwort = await fetch(url, lauf);
      if (!antwort.ok) throw Error((await antwort.text()).trim() || "Die Aktion ist fehlgeschlagen.");
      return antwort.json();
    } finally {
      clearTimeout(uhr);
    }
  }

  function onlineStarten() {
    S.online = true;
    S.modus = "online";
    $("menue").hidden = true;
    $("spielbereich").hidden = false;
    $("spielerbereich").hidden = true;
    clearInterval(S.pollUhr);
    S.pollUhr = setInterval(hole, 1500);
    hole();                        // erster Stand; ohne Partie zurück ins Menü
    aktion("rejoin", null, true);  // Platz sichern (Seite neu geladen) – ohne Meldung
  }

  /* Noch keine Partie im Raum: zurück ins Menü, dort eröffnet der Gastgeber. */
  function warteAufEroeffnung() {
    S.game = null;
    $("spielbereich").hidden = true;
    $("spielerbereich").hidden = true;
    $("menue").hidden = false;
    $("onlineKnopf").hidden = false;
    $("menueHinweis").textContent = "Raum " + S.room + " – hier läuft noch keine Halma-Partie. " +
      "Eröffne sie, die anderen kommen über den Raum dazu.";
  }

  /* Lobby vor dem Start: Bereit-Knopf für alle, Start nur für den Gastgeber. */
  function lobbyZeigen(game) {
    var leiste = $("onlineLeiste");
    var offen = game.status === "invited";
    leiste.hidden = !offen;
    if (!offen) return;
    var ich = null;
    game.spieler.forEach(function (p) { if (p.id === S.ich) ich = p; });
    var bereit = game.spieler.filter(function (p) { return p.bereit; }).length;
    $("onlineText").textContent = game.spieler.length + " von " + game.anzahl +
      " am Tisch · bereit: " + bereit;
    $("bereitKnopf").hidden = !ich;
    $("bereitKnopf").textContent = ich && ich.bereit ? "Doch nicht bereit" : "Bereit";
    $("startOnlineKnopf").hidden = game.hostId !== S.ich;
  }

  async function aktion(name, zusatz, still) {
    if (S.busy) {
      if (!still) meldung("Einen Moment – die letzte Aktion läuft noch.", 2000);
      return null;
    }
    S.busy = true;
    try {
      var d = await onlineFetch(API + "?room=" + encodeURIComponent(S.room) + "&action=" + name, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(Object.assign({ id: S.ich }, zusatz || {}))
      });
      if (d.game) uebernehmen(d.game);
      return d;
    } catch (e) {
      if (!still) lern(e.message, 5000);
      return null;
    } finally {
      S.busy = false;
    }
  }

  async function hole() {
    try {
      var d = await onlineFetch(API + "?room=" + encodeURIComponent(S.room) + "&id=" + encodeURIComponent(S.ich),
        { cache: "no-store" });
      if (d.game) {
        uebernehmen(d.game);
        return d.game;
      }
      if (S.online) warteAufEroeffnung();
    } catch (e) {
      var text = (e && e.message) || "";
      /* Nur ein endgültiges Nein des Servers beendet das Abfragen: Raum weg
         oder dieser Platz gehört einem anderen Gerät. „Bitte erst beitreten“
         ist dagegen normal (die Anmeldung läuft noch) und ein Netzfehler
         ebenso – beides probiert der nächste Takt erneut. */
      if (/Spielraum nicht gefunden|anderen Gerät/.test(text)) {
        clearInterval(S.pollUhr);
        S.pollUhr = null;
        S.game = null;
        warteAufEroeffnung();
        $("menueHinweis").textContent = "Der Spielraum ist weg – „" + text +
          "“. Bitte im Chat neu beitreten.";
        meldung("Verbindung zum Spielraum verloren – bitte im Chat neu beitreten.", 6000);
      }
    }
    return null;
  }

  function uebernehmen(game) {
    if (!game) return;
    S.game = game;
    S.variante = game.variante;
    S.anzahl = game.anzahl;
    S.brett = game.brett === "stern" ? HB.Brett.stern() : HB.Brett.klassisch(game.anzahl);
    S.belegung = game.belegung.slice();
    S.plaetze = game.plaetze.slice();
    S.farben = HB.Design.aktuell().figuren;
    S.spieler = game.spieler.map(function (p) { return { name: p.name, art: p.art || "mensch" }; });
    S.current = game.current;
    S.zugNr = game.zugNr;
    S.sieger = game.status === "finished" ? game.sieger : -1;
    S.letzterZug = game.letzter || null;
    S.verlauf = (game.verlauf || []).slice();
    if (!S.brettGebaut || S.brett.art !== (game.brett || "klassisch") ||
        S.gebautFuer !== game.variante + game.anzahl) {
      brettBauen();
      S.brettGebaut = true;
      S.gebautFuer = game.variante + game.anzahl;
    }
    $("menue").hidden = true;
    $("spielbereich").hidden = false;
    $("spielerbereich").hidden = false;
    $("varianteMarke").textContent = (HB.Regeln.VARIANTEN[game.variante] || {}).name || "Halma";
    $("dranMarke").textContent = game.status === "finished" ? "Partie beendet"
      : (game.current === serverPlatz() ? "Du bist dran" : nameVon(game.current) + " ist dran");
    lobbyZeigen(game);
    zeichne();
    if (game.status === "finished" && !S.endeGezeigt) {
      S.endeGezeigt = true;
      dialog("Partie beendet",
        "<p>" + (game.sieger >= 0 ? "<b>" + nameVon(game.sieger) + "</b> hat gewonnen." : "Kein Sieger.") + "</p>",
        [["Nochmal", function () { dialogZu(); S.endeGezeigt = false; aktion("reset"); }],
         ["Zum Menü", function () { dialogZu(); zumMenue(); }]]);
    }
  }

  function onlineKlick(i) {
    var platz = serverPlatz();
    if (platz < 0) { meldung("Du bist noch nicht am Tisch – bitte über den Raum beitreten.", 4000); return; }
    if (S.game && S.game.status !== "active") { meldung("Die Partie läuft noch nicht.", 3000); return; }
    if (S.current !== platz) { meldung("Jetzt ist " + nameVon(S.current) + " am Zug.", 3000); return; }
    if (S.belegung[i] === S.plaetze[platz]) {
      waehle(i);   // dieselbe Auswahl wie lokal (erlaubte Felder, Lernhinweis)
      return;
    }
    if (S.auswahl >= 0 && S.ziele.some(function (z) { return z.feld === i; })) {
      onlineZug(S.auswahl, i);
      return;
    }
    if (S.belegung[i] >= 0) { meldung(HB.Lernen.fehlerText("besetzt", S.variante)); return; }
    if (S.auswahl >= 0) { lern(HB.Lernen.fehlerText("zuweit", S.variante), 4200); return; }
    meldung("Wähle zuerst eine eigene Figur – die erlaubten Felder leuchten dann.", 3600);
  }

  function onlineZug(von, nach) {
    S.auswahl = -1;
    S.ziele = [];
    aktion("zug", { von: von, nach: nach }).then(function (d) {
      if (d && d.game) {
        var letzter = d.game.letzter;
        if (letzter) huepfAnimation(letzter.nach);
      }
    });
  }

  /* ----------------------------------------------------------- Dialog -- */
  function dialog(titel, html, knoepfe) {
    var box = $("dialog");
    box.innerHTML = "";
    var h = document.createElement("h2");
    h.textContent = titel;
    var p = document.createElement("div");
    p.innerHTML = html;
    var reihe = document.createElement("div");
    reihe.className = "knopfreihe";
    (knoepfe || [["Schließen", dialogZu]]).forEach(function (k) {
      var b = document.createElement("button");
      b.type = "button";
      b.className = "knopf";
      b.textContent = k[0];
      b.addEventListener("click", k[1]);
      reihe.appendChild(b);
    });
    box.append(h, p, reihe);
    $("dialogSchleier").hidden = false;
  }

  function dialogZu() { $("dialogSchleier").hidden = true; }

  /* Lernhilfe: nur bei eingeschalteter Hilfe (im Knobelmodus immer). */
  function lern(text, ms) {
    if (!S.lernen && S.modus !== "raetsel") return;
    meldung(text, ms);
  }

  /* Meldung, die immer erscheint – auch bei ausgeblendeter Lernhilfe.
     Fehler des Servers dürfen nicht verschwinden, nur weil jemand die
     Lernhilfe abgeschaltet hat. */
  function meldung(text, ms) {
    if (!text) return;
    var karte = $("lernkarte");
    karte.innerHTML = text;
    karte.classList.add("sichtbar");
    clearTimeout(S.lernUhr);
    S.lernUhr = setTimeout(function () { karte.classList.remove("sichtbar"); }, ms || 4200);
  }

  function zumMenue() {
    lupeZu();
    clearTimeout(S.kiUhr);
    clearInterval(S.pollUhr);
    S.modus = "menue";
    S.sieger = -1;
    S.online = false;
    S.endeGezeigt = false;
    $("menue").hidden = false;
    $("spielbereich").hidden = true;
    $("spielerbereich").hidden = true;
    $("lernkarte").classList.remove("sichtbar");
    optionenBauen();
  }

  /* ----------------------------------------------------------- Regeln -- */
  function regelnZeigen() {
    var v = S.variante;
    var titel = "Regeln – " + ((HB.Regeln.VARIANTEN[v] || {}).name || "Halma");
    var text = "";
    if (v === "stern") {
      text = "<ul>" +
        "<li>121 Felder: ein Mittelhexagon und sechs Zacken mit je zehn Feldern.</li>" +
        "<li>Zehn Figuren stehen in einer Zacke, Ziel ist die gegenüberliegende Zacke.</li>" +
        "<li>2, 3, 4 oder 6 Spieler. Bei 5 Spielern bliebe einem die eigene Zielzacke besetzt – deshalb nicht angeboten.</li>" +
        "<li>Schritt auf ein angrenzendes freies Feld (sechs Richtungen).</li>" +
        "<li>Sprung über genau eine benachbarte Figur auf das freie Feld dahinter, beliebig oft hintereinander.</li>" +
        "<li>Übersprungene Figuren bleiben stehen. Landen nur auf freien Feldern.</li>" +
        "<li>Figuren dürfen aus dem Zielbereich wieder herausziehen.</li>" +
        "</ul>";
    } else if (v === "sprung") {
      text = "<ul>" +
        "<li>16×16-Feld wie beim klassischen Halma, 19 Figuren zu zweit, 13 zu viert.</li>" +
        "<li>Schritt auf ein angrenzendes freies Feld (acht Richtungen).</li>" +
        "<li><b>Weiter Sprung:</b> Auf der geraden Linie liegt genau <i>eine</i> Figur. " +
        "Das Zielfeld liegt im selben Abstand hinter ihr wie sie vor dir – das ist der Spiegelpunkt.</li>" +
        "<li>Jede Weite und alle acht Richtungen sind erlaubt, auch mehrere Sprünge hintereinander.</li>" +
        "<li>Unterschied zum klassischen Halma: Dort geht der Sprung nur über eine <i>direkt benachbarte</i> " +
        "Figur zwei Felder weit. Hier darf die übersprungene Figur beliebig weit weg stehen.</li>" +
        "<li>Über zwei Figuren hintereinander wird nicht gesprungen.</li>" +
        "</ul>";
    } else {
      text = "<ul>" +
        "<li>16×16-Feld mit 256 Feldern. Jedes Eckhaus hat 13 Felder (Reihen 1+2+3+4+3).</li>" +
        "<li>Zu zweit steht in zwei gegenüberliegenden Häusern eine Reihe mehr zur Verfügung: 19 Felder (1+2+3+4+5+4) und damit 19 Figuren.</li>" +
        "<li>Zu viert spielen alle vier Ecken mit je 13 Figuren. Ziel ist jeweils das diagonale Eckhaus.</li>" +
        "<li>Schritt auf ein angrenzendes freies Feld, auch diagonal (acht Richtungen).</li>" +
        "<li>Sprung über genau eine benachbarte Figur auf das freie Feld dahinter.</li>" +
        "<li>Mehrere Sprünge hintereinander mit derselben Figur sind ein einziger Zug.</li>" +
        "<li>Übersprungene Figuren bleiben stehen; Landen nur auf freien Feldern.</li>" +
        "<li>Figuren dürfen aus dem Zielbereich wieder herausziehen (kein Sperrfeld).</li>" +
        "<li>Gewonnen hat, wer alle eigenen Figuren (19 bzw. 13) im gegenüberliegenden Haus hat.</li>" +
        "</ul>";
    }
    dialog(titel, text + '<p style="font-size:12px;color:var(--muted)">Standardregel ohne Hausregeln. ' +
      'Der Computer nutzt exakt dieselben Regeln. Am Handy zeigt <b>Lupe an</b> das angetippte Feld ' +
      'vergrößert – der zweite Tipp wählt es aus.</p>', [["Alles klar", dialogZu]]);
  }

  /* ------------------------------------------------------------- Start -- */
  function initialisieren() {
    HB.Design.stilSetzen(HB.Design.gespeichert());
    $("hilfeKnopf").addEventListener("click", regelnZeigen);
    $("startKnopf").addEventListener("click", spielStarten);
    $("neuKnopf").addEventListener("click", function () {
      if (S.modus === "raetsel") raetselStarten(S.raetsel.id);
      else if (S.modus === "online") aktion("reset");
      else spielStarten();
    });
    $("menueKnopf").addEventListener("click", zumMenue);
    $("lernKnopf").addEventListener("click", function () {
      S.lernen = !S.lernen;
      $("lernKnopf").textContent = S.lernen ? "Lernhilfe an" : "Lernhilfe aus";
      if (!S.lernen) $("lernkarte").classList.remove("sichtbar");
      else lern("Lernhilfe ist an: erlaubte Felder leuchten, kurze Hinweise erklären jeden Schritt.", 3600);
    });
    $("lupeKnopf").addEventListener("click", function () {
      S.lupeWahl = !lupeAktiv();
      lupeZu();
      lupeKnopfAktualisieren();
    });
    $("lupeZu").addEventListener("click", lupeZu);
    $("lupeGitter").addEventListener("click", lupeGitterKlick);
    $("lupeGitter").addEventListener("keydown", function (ev) {
      if (ev.key !== "Enter" && ev.key !== " ") return;
      var z = ev.target.closest(".lupe-zelle");
      if (!z) return;
      ev.preventDefault();
      lupeZu();
      feldKlick(Number(z.dataset.i));
    });
    document.addEventListener("keydown", function (ev) {
      if (ev.key === "Escape") lupeZu();
    });
    $("tippKnopf").addEventListener("click", tippZeigen);
    $("loesungKnopf").addEventListener("click", loesungZeigen);
    $("onlineKnopf").addEventListener("click", function () {
      if (!S.room) { meldung("Für eine Online-Partie bitte über einen Spielraum starten.", 4200); return; }
      aktion("create", { variante: S.variante, anzahl: S.anzahl });
    });
    $("bereitKnopf").addEventListener("click", function () { aktion("ready"); });
    $("startOnlineKnopf").addEventListener("click", function () { aktion("start"); });
    $("dialogSchleier").addEventListener("click", function (e) {
      if (e.target === $("dialogSchleier")) dialogZu();
    });
    window.addEventListener("resize", passeBrett);
    if (window.ResizeObserver) new ResizeObserver(passeBrett).observe($("brettrahmen"));

    variantenBauen();
    optionenBauen();

    if (ONLINE) {
      $("menueHinweis").textContent = "Raum " + S.room + " – bitte die Verbindung abwarten.";
      $("startKnopf").hidden = true;
      onlineStarten();
    } else {
      /* Vorschau: ein leeres Brett zeigt schon die Form der Variante. */
      S.brett = HB.Brett.klassisch(S.anzahl);
    }
  }

  HB.Main = {
    stilGewechselt: function () {
      S.farben = HB.Design.aktuell().figuren;
      if (S.brett && S.belegung) zeichne();
    },
    zustand: function () { return S; }
  };

  if (document.readyState === "loading") document.addEventListener("DOMContentLoaded", initialisieren);
  else initialisieren();
})();
