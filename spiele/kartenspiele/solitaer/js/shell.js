/* =========================================================================
   SOLITÄR – Rahmen: Modusauswahl, Anzeige, Uhr, Rückgängig, Dialoge,
   Antippen und Ziehen. Die Modi bringen nur ihre Regeln und ihr Brett mit.

   Ein Modus ist ein Objekt:
     { id, name, kurz, regeln:[…], vorschau:"<svg…>",
       optionen:[{key,label,werte:[{wert,text}]}], standard:{…},
       neu(opt), snap(), restore(s), info(),
       waehlen(was), ablegen(was, zone), zoneKlick(zone), tipp() }
   ========================================================================= */
(function () {
  "use strict";

  var TG = window.TG = window.TG || {};
  var K = TG.Karten;
  var $ = function (id) { return document.getElementById(id); };

  var Shell = {
    modus: null,
    spiel: null,
    optionen: {},
    verlauf: [],
    sekunden: 0,
    uhr: null,
    laeuft: false,
    auswahl: null,
    autoInProgress: false,
    pausiert: false,
    beendet: false,
    stil: "klassisch",
    beiEnde: null
  };

  var MAX_VERLAUF = 200;

  /* Der Titel in der Kopfleiste. Fehlt das Element (etwa nach einem Umbau der
     Seite), darf das Spiel daran nicht scheitern. */
  function kopfTitelSetzen(text) {
    var el = $("kopfTitel");
    if (el) el.textContent = text;
  }

  /* ------------------------------------------------------------- Statistik */
  function statLaden() {
    try { return JSON.parse(localStorage.getItem("solitaer-stat") || "{}"); } catch (e) { return {}; }
  }
  function statSpeichern(s) {
    try { localStorage.setItem("solitaer-stat", JSON.stringify(s)); } catch (e) { }
  }
  function statModus(id) {
    var s = statLaden();
    if (!s[id]) s[id] = { gewonnen: 0, gespielt: 0 };
    return s[id];
  }

  /* ------------------------------------------------------------- Grafikstil */
  /* Drei Ansichten, ein Wortschatz nach außen: normal · maedchen · buben.
     Intern heißt die klassische Ansicht weiter "klassisch" – so heißen die
     CSS-Regeln dieser Seite seit dem ersten Tag. */
  var STILE = { normal: "klassisch", klassisch: "klassisch", maedchen: "maedchen", buben: "buben" };

  function eigenerStil() {
    var wert = "";
    try { wert = String(localStorage.getItem("solitaer-stil") || ""); } catch (e) { wert = ""; }
    if (STILE[wert]) return wert;
    if (window.TGStil && typeof window.TGStil.hole === "function") return window.TGStil.hole();
    return "normal";
  }

  /* merken=true schreibt die eigene Wahl; ein Stil von außen wird nie
     gespeichert und überschreibt die lokale Wahl nicht. */
  Shell.stilSetzen = function (wert, merken) {
    var roh = String(wert || "normal").toLowerCase();
    var s = STILE[roh] || "klassisch";
    Shell.stil = s;
    document.documentElement.dataset.stil = s;
    document.documentElement.style.colorScheme = s === "buben" ? "dark" : "light";
    if (merken) {
      try { localStorage.setItem("solitaer-stil", s === "klassisch" ? "normal" : s); } catch (e) { }
    }
    var box = $("stilKnoepfe");
    if (box) {
      var api = s === "klassisch" ? "normal" : s;
      box.querySelectorAll("button[data-stil]").forEach(function (b) {
        var an = b.dataset.stil === api;
        b.setAttribute("aria-pressed", an ? "true" : "false");
        b.classList.toggle("aktiv", an);
      });
    }
    return s;
  };

  Shell.stilApi = function () { return Shell.stil === "klassisch" ? "normal" : Shell.stil; };
  Shell.stilWahlZeigen = function (an) { var b = $("stilWahl"); if (b) b.hidden = !an; };
  Shell.amStart = function () { var a = $("modusAuswahl"); return !!(a && !a.hidden); };

  /* ------------------------------------------------------------------ Pause */
  function pausenschleier() {
    var sch = $("pauseSchleier");
    if (sch) sch.hidden = !Shell.pausiert;
    var b = $("btnPause");
    if (b) {
      b.textContent = Shell.pausiert ? "Weiter" : "Pause";
      b.setAttribute("aria-pressed", Shell.pausiert ? "true" : "false");
    }
    if (Shell.pausiert) dialogZu();
  }

  Shell.pausieren = function () {
    if (!Shell.spiel || Shell.pausiert || Shell.beendet || Shell.amStart()) return false;
    Shell.pausiert = true;
    uhrAnhalten();
    pausenschleier();
    return true;
  };

  Shell.fortsetzen = function () {
    if (!Shell.spiel || !Shell.pausiert) return false;
    Shell.pausiert = false;
    pausenschleier();
    taktStarten();
    Shell.stand(Shell.spiel.info());
    return true;
  };

  Shell.pauseUmschalten = function () { return Shell.pausiert ? Shell.fortsetzen() : Shell.pausieren(); };

  /* -------------------------------------------------------------- Statistik */
  Shell.statistikZeigen = function () {
    var s = statLaden();
    var zeilen = Object.keys(TG.MODI).map(function (id) {
      var m = TG.MODI[id];
      var st = s[id] || { gewonnen: 0, gespielt: 0 };
      var quote = st.gespielt ? Math.round(st.gewonnen * 100 / st.gespielt) + " %" : "–";
      return "<tr><td>" + m.name + "</td><td>" + st.gespielt + "</td><td>" + st.gewonnen + "</td><td>" + quote + "</td></tr>";
    }).join("");
    dialog("Statistik", '<p class="klein">Gezählt wird je begonnener Partie. Gewonnen zählt nur ein vollständig abgelegtes Spiel.</p>' +
      '<table class="stat"><thead><tr><th>Variante</th><th>Gespielt</th><th>Gewonnen</th><th>Quote</th></tr></thead><tbody>' +
      zeilen + "</tbody></table>",
      [{ text: "Schließen", haupt: true }]);
  };

  /* Ein Ergebnis nach außen melden – vor dem Ergebnisfenster. */
  function endeMelden(grund, info) {
    if (typeof Shell.beiEnde === "function") {
      try { Shell.beiEnde(grund, info); } catch (e) { }
    }
  }

  /* ------------------------------------------------------------- Modusliste */
  function optionenVon(id, opt) {
    var m = TG.MODI[id];
    var o = {};
    (m.optionen || []).forEach(function (o1) { o[o1.key] = o1.werte[0].wert; });
    Object.keys(o).forEach(function (k) { if (opt && opt[k] !== undefined) o[k] = opt[k]; });
    return o;
  }

  function menueBauen() {
    var liste = $("modusListe");
    liste.innerHTML = "";
    Object.keys(TG.MODI).forEach(function (id) {
      var m = TG.MODI[id];
      var karte = document.createElement("article");
      karte.className = "modus";
      karte.dataset.modus = id;

      var vorschau = document.createElement("div");
      vorschau.className = "modus-vorschau";
      vorschau.innerHTML = m.vorschau || "";

      var text = document.createElement("div");
      text.className = "modus-text";
      var h = document.createElement("h3");
      h.textContent = m.name;
      var p = document.createElement("p");
      p.textContent = m.kurz;
      text.append(h, p);

      if (m.optionen && m.optionen.length) {
        m.optionen.forEach(function (o) {
          var zeile = document.createElement("label");
          zeile.className = "modus-option";
          zeile.textContent = o.label + " ";
          var sel = document.createElement("select");
          sel.dataset.key = o.key;
          o.werte.forEach(function (w) {
            var op = document.createElement("option");
            op.value = String(w.wert);
            op.textContent = w.text;
            sel.appendChild(op);
          });
          zeile.appendChild(sel);
          zeile.addEventListener("click", function (ev) { ev.stopPropagation(); });
          text.appendChild(zeile);
        });
      }

      var knopf = document.createElement("button");
      knopf.type = "button";
      knopf.className = "knopf start";
      knopf.textContent = "Spielen";
      var stat = statModus(id);
      var quote = document.createElement("span");
      quote.className = "modus-quote";
      quote.textContent = stat.gespielt ? "gewonnen " + stat.gewonnen + " von " + stat.gespielt : "";
      knopf.addEventListener("click", function () {
        var opt = {};
        karte.querySelectorAll("select[data-key]").forEach(function (s) {
          opt[s.dataset.key] = istZahl(s.value) ? Number(s.value) : s.value;
        });
        Shell.starten(id, opt);
      });

      karte.append(vorschau, text, knopf, quote);
      liste.appendChild(karte);
    });
  }

  function istZahl(t) { return /^-?\d+$/.test(String(t)); }

  /* ---------------------------------------------------------------- Start */
  Shell.starten = function (id, opt) {
    var m = TG.MODI[id];
    if (!m) return;
    dialogZu();
    Shell.pausiert = false;
    Shell.beendet = false;
    pausenschleier();
    Shell.modus = id;
    Shell.optionen = opt || Shell.optionen || {};
    Shell.verlauf = [];
    Shell.auswahl = null;
    Shell.spiel = m;
    var nun = optionenVon(id, Shell.optionen);
    Shell.optionen = nun;

    $("modusAuswahl").hidden = true;
    $("spielbereich").hidden = false;
    kopfTitelSetzen(m.name);
    document.body.classList.add("spielt");
    document.title = m.name + " – Solitär";
    optionenZeigen();
    uhrStarten();
    var stat = statModus(id);
    stat.gespielt++;
    var alle = statLaden();
    alle[id] = stat;
    statSpeichern(alle);

    var vorher = K.positionen();
    m.neu(nun);
    K.bewege(vorher);
    Shell.stand(m.info());
    Shell.hinweis("");
  };

  function optionenZeigen() {
    var box = $("optionen");
    box.innerHTML = "";
    var m = TG.MODI[Shell.modus];
    if (!m || !m.optionen) { box.hidden = true; return; }
    box.hidden = false;
    m.optionen.forEach(function (o) {
      var lab = document.createElement("label");
      lab.className = "leiste-option";
      var span = document.createElement("span");
      span.textContent = o.label;
      var sel = document.createElement("select");
      sel.dataset.key = o.key;
      o.werte.forEach(function (w) {
        var op = document.createElement("option");
        op.value = String(w.wert);
        op.textContent = w.text;
        if (Shell.optionen[o.key] === w.wert) op.selected = true;
        sel.appendChild(op);
      });
      sel.addEventListener("change", function () {
        Shell.optionen[o.key] = istZahl(sel.value) ? Number(sel.value) : sel.value;
        Shell.neu();
      });
      lab.append(span, sel);
      box.appendChild(lab);
    });
  }

  Shell.neu = function () {
    if (!Shell.spiel) return;
    Shell.pausiert = false;
    Shell.beendet = false;
    pausenschleier();
    Shell.verlauf = [];
    Shell.auswahl = null;
    Shell.sekunden = 0;
    uhrStarten();
    var vorher = K.positionen();
    Shell.spiel.neu(Object.assign({}, Shell.optionen));
    K.bewege(vorher);
    Shell.stand(Shell.spiel.info());
    Shell.hinweis("Neue Partie.");
  };

  /* ---------------------------------------------------------------- Uhr */
  function taktStarten() {
    if (Shell.uhr) clearInterval(Shell.uhr);
    Shell.laeuft = true;
    Shell.uhr = setInterval(function () {
      if (!Shell.laeuft) return;
      Shell.sekunden++;
      zeigeZeit();
    }, 1000);
  }
  function uhrStarten() {
    uhrStoppen();
    Shell.sekunden = 0;
    zeigeZeit();
    taktStarten();
  }
  function uhrStoppen() {
    if (Shell.uhr) clearInterval(Shell.uhr);
    Shell.uhr = null;
    Shell.laeuft = false;
  }
  function uhrAnhalten() { Shell.laeuft = false; }
  function zeigeZeit() {
    var m = Math.floor(Shell.sekunden / 60), s = Shell.sekunden % 60;
    $("hudZeit").textContent = m + ":" + (s < 10 ? "0" : "") + s;
  }
  function zeitText() { return $("hudZeit").textContent; }

  /* ------------------------------------------------------------- Anzeige */
  Shell.stand = function (info) {
    if (!info) return;
    $("hudPunkte").textContent = info.punkte;
    $("hudZuege").textContent = info.zuege;
    $("hudExtra").textContent = info.extra || "";
    $("hudExtraBox").hidden = !info.extra;
    $("btnUndo").disabled = !Shell.verlauf.length;
    $("btnAuto").disabled = Shell.autoInProgress || !(Shell.spiel && Shell.spiel.autoMoeglich && Shell.spiel.autoMoeglich());
  };

  Shell.hinweis = function (text, warnung) {
    var el = $("hinweis");
    el.textContent = text || "";
    el.classList.toggle("warnung", !!warnung);
  };

  Shell.merken = function () {
    if (!Shell.spiel) return;
    Shell.verlauf.push(JSON.stringify(Shell.spiel.snap()));
    if (Shell.verlauf.length > MAX_VERLAUF) Shell.verlauf.shift();
    $("btnUndo").disabled = false;
  };

  Shell.zurueck = function () {
    if (Shell.pausiert) return;
    if (!Shell.verlauf.length) { Shell.hinweis("Nichts mehr zum Zurücknehmen.", true); return; }
    var stand = JSON.parse(Shell.verlauf.pop());
    Shell.auswahl = null;
    var vorher = K.positionen();
    Shell.spiel.restore(stand);
    K.bewege(vorher);
    Shell.stand(Shell.spiel.info());
    Shell.hinweis("Zug zurückgenommen.");
  };

  Shell.tippen = function () {
    if (!Shell.spiel || Shell.pausiert || !Shell.spiel.tipp) return;
    var r = Shell.spiel.tipp();
    Shell.hinweis(r || "");
    Shell.stand(Shell.spiel.info());
  };

  Shell.autoBeenden = function () {
    if (!Shell.spiel || Shell.pausiert || !Shell.spiel.autoMoeglich || !Shell.spiel.autoMoeglich() || !Shell.spiel.autoVollenden) return;
    Shell.autoInProgress = true;
    Shell.stand(Shell.spiel.info());
    try { Shell.spiel.autoVollenden(); }
    finally { Shell.autoInProgress = false; Shell.stand(Shell.spiel.info()); }
  };

  /* ------------------------------------------------------------ Auswahl */
  /* Ein Modus meldet nach jedem Zug seinen Stand; der Rahmen zählt Züge und
     schreibt die Anzeige. */
  Shell.zugFertig = function (text) {
    if (!Shell.spiel) return;
    var info = Shell.spiel.info();
    Shell.stand(info);
    if (text) Shell.hinweis(text);
  };

  /* ------------------------------------------------------------- Dialoge */
  function dialog(titel, inhaltHtml, knoepfe) {
    var box = $("dialog");
    box.innerHTML = "";
    var kopf = document.createElement("div");
    kopf.className = "dialog-kopf";
    var t = document.createElement("span");
    t.textContent = titel;
    var x = document.createElement("button");
    x.type = "button";
    x.className = "dialog-zu";
    x.setAttribute("aria-label", "Schließen");
    x.textContent = "×";
    x.addEventListener("click", dialogZu);
    kopf.append(t, x);

    var innen = document.createElement("div");
    innen.className = "dialog-inhalt";
    innen.innerHTML = inhaltHtml;

    var reihe = document.createElement("div");
    reihe.className = "dialog-knoepfe";
    (knoepfe || [{ text: "OK", haupt: true }]).forEach(function (k) {
      var b = document.createElement("button");
      b.type = "button";
      b.className = "knopf" + (k.haupt ? "" : " zweit");
      b.textContent = k.text;
      b.addEventListener("click", function () {
        if (k.beiKlick) k.beiKlick();
        else dialogZu();
      });
      reihe.appendChild(b);
    });

    box.append(kopf, innen, reihe);
    $("dialogSchleier").hidden = false;
  }
  function dialogZu() { $("dialogSchleier").hidden = true; }

  Shell.dialog = dialog;
  Shell.dialogZu = dialogZu;

  Shell.regelnZeigen = function () {
    var m = TG.MODI[Shell.modus];
    if (!m) return;
    var text = "<ol>" + m.regeln.map(function (r) { return "<li>" + r + "</li>"; }).join("") + "</ol>" +
      '<p class="klein">Bedienung: Karte antippen oder mit der Maus ziehen. ' +
      "Erlaubte Ziele leuchten, wenn eine Karte gewählt ist.</p>";
    dialog("Regeln – " + m.name, text, [{ text: "Alles klar", haupt: true }]);
  };

  /* --------------------------------------------------------------- Sieg */
  Shell.sieg = function (zusatz) {
    uhrAnhalten();
    Shell.beendet = true;
    Shell.pausiert = false;
    pausenschleier();
    var m = TG.MODI[Shell.modus];
    var id = Shell.modus;
    var stat = statModus(id);
    stat.gewonnen++;
    var alle = statLaden();
    alle[id] = stat;
    statSpeichern(alle);
    var info = Shell.spiel.info();
    endeMelden("gewonnen", info);
    dialog("Gewonnen – " + m.name,
      "<p><b>Alle Karten sind weg.</b></p>" +
      "<p>Zeit " + zeitText() + " · Züge " + info.zuege + " · Punkte " + info.punkte + "</p>" +
      (zusatz ? "<p>" + zusatz + "</p>" : ""),
      [{ text: "Nochmal", haupt: true, beiKlick: function () { dialogZu(); Shell.neu(); } },
      { text: "Modusauswahl", beiKlick: function () { dialogZu(); Shell.menue(); } }]);
  };

  Shell.verloren = function (text) {
    uhrAnhalten();
    Shell.beendet = true;
    Shell.pausiert = false;
    pausenschleier();
    endeMelden("verloren", Shell.spiel && Shell.spiel.info ? Shell.spiel.info() : { zuege: 0, punkte: 0 });
    dialog("Keine Züge mehr", "<p>" + text + "</p>",
      [{ text: "Neues Spiel", haupt: true, beiKlick: function () { dialogZu(); Shell.neu(); } },
      { text: "Zurücknehmen", beiKlick: function () { dialogZu(); Shell.zurueck(); } }]);
  };

  /* ---------------------------------------------------------- Menüwechsel */
  Shell.menue = function () {
    uhrAnhalten();
    Shell.beendet = false;
    Shell.pausiert = false;
    dialogZu();
    pausenschleier();
    $("spielbereich").hidden = true;
    $("modusAuswahl").hidden = false;
    kopfTitelSetzen("Solitär");
    document.body.classList.remove("spielt");
    document.title = "Solitär – fünf Varianten";
    menueBauen();
  };

  /* ---------------------------------------------------- Tippen und Ziehen */
  function wasVon(el) {
    return {
      art: el.dataset.art,
      i: el.dataset.i !== undefined ? Number(el.dataset.i) : undefined,
      index: el.dataset.index !== undefined ? Number(el.dataset.index) : undefined,
      id: el.dataset.id ? Number(el.dataset.id) : undefined
    };
  }

  function zoneVon(el) {
    var z = el && el.closest ? el.closest("[data-zone]") : null;
    return z ? z.dataset.zone : null;
  }

  /* Ein Ablegeversuch für Klick und Ziehen: Der Modus prüft, der Rahmen
     aktualisiert Anzeige und räumt den alten Hinweis nach einem Erfolg weg. */
  function ablegenSauber(was, zone) {
    var ok = Shell.spiel.ablegen(was, zone);
    if (ok) Shell.hinweis("");
    Shell.stand(Shell.spiel.info());
    return ok;
  }

  function brettKlick(ev) {
    if (!Shell.spiel || Shell.pausiert) return;
    if (Shell.gezogen) { Shell.gezogen = false; return; }
    var karte = ev.target.closest("[data-art]");
    if (karte) {
      var zielZone = zoneVon(karte);
      var aus = Shell.auswahl;
      var gleiche = aus && karte.dataset.id && Number(karte.dataset.id) === aus.id;
      /* Liegt schon eine Auswahl und ist die angeklickte Karte Teil eines
         gültigen Ziels, wird abgelegt – auch auf belegte Ablagen und Spalten.
         Sonst würde ein Klick dort nur die obere Karte neu auswählen. */
      if (aus && gleiche) {
        Shell.auswahl = null;
        if (Shell.spiel.zoneKlick) Shell.spiel.zoneKlick("");
        Shell.stand(Shell.spiel.info());
        return;
      }
      if (aus && zielZone && Shell.spiel.zieleFuer &&
          Shell.spiel.zieleFuer(aus).indexOf(zielZone) >= 0) {
        if (ablegenSauber(aus, zielZone)) return;
      }
      Shell.auswahl = wasVon(karte);
      Shell.spiel.waehlen(Shell.auswahl);
      Shell.stand(Shell.spiel.info());
      return;
    }
    var zone = zoneVon(ev.target);
    if (zone) {
      if (Shell.auswahl) { ablegenSauber(Shell.auswahl, zone); return; }
      Shell.spiel.zoneKlick(zone);
      Shell.stand(Shell.spiel.info());
      return;
    }
    Shell.spiel.brettKlick && Shell.spiel.brettKlick(ev);
  }

  /* Ziehen: Karte greifen, Geist folgt dem Finger, beim Loslassen entscheidet
     die Zone darunter. Ohne Bewegung bleibt es ein Tippen. */
  var zug = null;
  function ziehenStart(ev) {
    if (!Shell.spiel || Shell.pausiert || (ev.button !== undefined && ev.button > 0)) return;
    var karte = ev.target.closest("[data-art]");
    if (!karte) return;
    /* Liegt eine Auswahl auf einer anderen Karte und ist die berührte Karte
       ein gültiges Ziel dafür, gehört der Tipp der Auswahl. Sonst würde beim
       Tippen auf das Ass die Ass-Karte selbst gegriffen: Fingerzittern startet
       einen Zug, der Ablegeversuch scheitert und der Klick wird geschluckt. */
    var aus = Shell.auswahl;
    var zielZone = zoneVon(karte);
    if (aus && zielZone && Number(karte.dataset.id) !== aus.id &&
        Shell.spiel.zieleFuer && Shell.spiel.zieleFuer(aus).indexOf(zielZone) >= 0) {
      return;
    }
    zug = {
      startX: ev.clientX, startY: ev.clientY, karte: karte, was: wasVon(karte),
      aktiv: false, geist: null, letzte: null
    };
  }
  function ziehenBewegen(ev) {
    if (!zug) return;
    var dx = ev.clientX - zug.startX, dy = ev.clientY - zug.startY;
    if (!zug.aktiv && Math.abs(dx) + Math.abs(dy) > 7) {
      zug.aktiv = true;
      var r = zug.karte.getBoundingClientRect();
      var g = zug.karte.cloneNode(true);
      g.className += " geist";
      g.style.width = r.width + "px";
      g.style.height = r.height + "px";
      g.style.left = r.left + "px";
      g.style.top = r.top + "px";
      document.body.appendChild(g);
      zug.geist = g;
      document.body.classList.add("zieht");
    }
    if (!zug.aktiv) return;
    zug.geist.style.left = (zug.startX + dx - zug.geist.offsetWidth / 2) + "px";
    zug.geist.style.top = (zug.startY + dy - 18) + "px";
    var ziel = document.elementFromPoint(ev.clientX, ev.clientY);
    if (zug.letzte) zug.letzte.classList.remove("ueber");
    var z = ziel && ziel.closest ? ziel.closest("[data-zone]") : null;
    if (z) { z.classList.add("ueber"); zug.letzte = z; }
    else zug.letzte = null;
  }
  function ziehenEnde(ev) {
    if (!zug) return;
    var z = zug;
    zug = null;
    document.body.classList.remove("zieht");
    if (z.letzte) z.letzte.classList.remove("ueber");
    if (!z.aktiv) return;
    if (z.geist) z.geist.remove();
    Shell.gezogen = true;
    setTimeout(function () { Shell.gezogen = false; }, 60);
    var ziel = document.elementFromPoint(ev.clientX, ev.clientY);
    var zone = zoneVon(ziel);
    if (zone) {
      /* Dieselbe Regel wie beim Klick: Ist die Zone ein gültiges Ziel der
         vorhandenen Auswahl, wird diese abgelegt – nicht die gerade
         berührte Zielkarte. Anzeige und Ablegen prüfen damit gleich. */
      var was = z.was;
      var aus = Shell.auswahl;
      if (aus && z.was.id !== aus.id && Shell.spiel.zieleFuer &&
          Shell.spiel.zieleFuer(aus).indexOf(zone) >= 0) {
        was = aus;
      }
      ablegenSauber(was, zone);
    } else {
      Shell.stand(Shell.spiel.info());
    }
  }

  /* -------------------------------------------------------------- Aufbau */
  function tasten(ev) {
    if (!Shell.spiel) return;
    var t = ev.key.toLowerCase();
    if (Shell.pausiert) {
      if (t === "p" || t === "escape") { ev.preventDefault(); Shell.fortsetzen(); }
      return;
    }
    if (t === "p") { ev.preventDefault(); Shell.pausieren(); return; }
    if (ev.ctrlKey && t === "z") { ev.preventDefault(); Shell.zurueck(); return; }
    /* Auf Knöpfen, Auswahlen und Feldern löst der Browser selbst aus. */
    var fokus = document.activeElement;
    if (fokus && /^(BUTTON|SELECT|INPUT|TEXTAREA|A)$/.test(fokus.tagName)) return;
    if (t === "f2") { ev.preventDefault(); Shell.neu(); return; }
    if (t === "f1") { ev.preventDefault(); Shell.regelnZeigen(); return; }
    if (t === "h" || t === "t") { Shell.tippen(); return; }
    if (t === "escape") { Shell.menue(); }
  }

  function initialisieren() {
    menueBauen();
    $("btnNeu").addEventListener("click", function () { Shell.neu(); });
    $("btnUndo").addEventListener("click", function () { Shell.zurueck(); });
    $("btnTipp").addEventListener("click", function () { Shell.tippen(); });
    $("btnRegeln").addEventListener("click", function () { Shell.regelnZeigen(); });
    $("btnMenue").addEventListener("click", confirmMenue);
    $("dialogSchleier").addEventListener("click", function (ev) {
      if (ev.target === $("dialogSchleier")) dialogZu();
    });

    $("btnAuto").addEventListener("click", Shell.autoBeenden);
    if ($("btnPause")) $("btnPause").addEventListener("click", Shell.pauseUmschalten);
    if ($("btnStat")) $("btnStat").addEventListener("click", Shell.statistikZeigen);
    if ($("btnWeiter")) $("btnWeiter").addEventListener("click", function () { Shell.fortsetzen(); });
    if ($("btnPauseNeu")) $("btnPauseNeu").addEventListener("click", function () { Shell.fortsetzen(); Shell.neu(); });
    if ($("btnPauseMenue")) $("btnPauseMenue").addEventListener("click", function () { Shell.fortsetzen(); Shell.menue(); });
    if ($("btnZurueck")) $("btnZurueck").addEventListener("click", function () {
      if (typeof window.tgZurueck === "function") window.tgZurueck();
      else location.href = "../../spiele.html";
    });

    /* Eigene Stilwahl: nur beim direkten Start sichtbar. Kommt der Stil vom
       Konnektor, versteckt ihn die Schnittstelle. */
    Shell.stilSetzen(eigenerStil(), false);
    var stilBox = $("stilKnoepfe");
    if (stilBox) stilBox.querySelectorAll("button[data-stil]").forEach(function (b) {
      b.addEventListener("click", function () { Shell.stilSetzen(b.dataset.stil, true); });
    });
    var brett = $("brett");
    var einzelklickTimer = null;
    var doppelklickSchonAusgefuehrt = false;
    function doppelklick(ev) {
      var karte = ev.target.closest("[data-art]");
      if (!karte || !Shell.spiel || !Shell.spiel.doppelklick) return;
      ev.preventDefault();
      var ok = Shell.spiel.doppelklick(wasVon(karte));
      if (ok) { Shell.auswahl = null; Shell.hinweis(""); Shell.stand(Shell.spiel.info()); }
    }
    brett.addEventListener("click", function (ev) {
      if (Shell.gezogen) { Shell.gezogen = false; return; }
      if (ev.detail === 0) { brettKlick(ev); return; }
      if (ev.detail >= 2) {
        if (einzelklickTimer) clearTimeout(einzelklickTimer);
        einzelklickTimer = null;
        doppelklickSchonAusgefuehrt = true;
        doppelklick(ev);
        return;
      }
      if (einzelklickTimer) clearTimeout(einzelklickTimer);
      var ziel = ev.target;
      einzelklickTimer = setTimeout(function () {
        einzelklickTimer = null;
        brettKlick({ target: ziel });
      }, 240);
    });
    brett.addEventListener("dblclick", function (ev) {
      ev.preventDefault();
      if (doppelklickSchonAusgefuehrt) { doppelklickSchonAusgefuehrt = false; return; }
      if (einzelklickTimer) clearTimeout(einzelklickTimer);
      einzelklickTimer = null;
      doppelklick(ev);
    });
    brett.addEventListener("pointerdown", ziehenStart);
    document.addEventListener("pointermove", ziehenBewegen);
    document.addEventListener("pointerup", ziehenEnde);
    document.addEventListener("pointercancel", function () {
      if (zug && zug.geist) zug.geist.remove();
      if (zug && zug.letzte) zug.letzte.classList.remove("ueber");
      zug = null;
      document.body.classList.remove("zieht");
    });
    document.addEventListener("keydown", tasten);

    $("modusAuswahl").hidden = false;
    $("spielbereich").hidden = true;
    pausenschleier();
  }

  /* Zurück zur Auswahl nur mit Rückfrage, wenn eine Partie läuft. */
  function confirmMenue() {
    if (Shell.spiel && Shell.laeuft && Shell.sekunden > 5) {
      Shell.dialog("Zur Modusauswahl?", "<p>Die laufende Partie bleibt im Hintergrund erhalten – " +
        "ein neuer Modus beginnt frisch. Fortfahren?</p>",
        [{ text: "Ja, zur Auswahl", haupt: true, beiKlick: function () { dialogZu(); Shell.menue(); } },
        { text: "Weiterspielen" }]);
      return;
    }
    Shell.menue();
  }

  Shell.confirmMenue = confirmMenue;
  TG.Shell = Shell;

  if (document.readyState === "loading") document.addEventListener("DOMContentLoaded", initialisieren);
  else initialisieren();
})();
