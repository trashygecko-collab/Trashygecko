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
    auswahl: null
  };

  var MAX_VERLAUF = 200;

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
    Shell.modus = id;
    Shell.optionen = opt || Shell.optionen || {};
    Shell.verlauf = [];
    Shell.auswahl = null;
    Shell.spiel = m;
    var nun = optionenVon(id, Shell.optionen);
    Shell.optionen = nun;

    $("modusAuswahl").hidden = true;
    $("spielbereich").hidden = false;
    $("kopfTitel").textContent = m.name;
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
  function uhrStarten() {
    uhrStoppen();
    Shell.sekunden = 0;
    Shell.laeuft = true;
    zeigeZeit();
    Shell.uhr = setInterval(function () {
      if (!Shell.laeuft) return;
      Shell.sekunden++;
      zeigeZeit();
    }, 1000);
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
    if (!Shell.spiel || !Shell.spiel.tipp) return;
    var r = Shell.spiel.tipp();
    Shell.hinweis(r || "");
    Shell.stand(Shell.spiel.info());
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
    var m = TG.MODI[Shell.modus];
    var id = Shell.modus;
    var stat = statModus(id);
    stat.gewonnen++;
    var alle = statLaden();
    alle[id] = stat;
    statSpeichern(alle);
    var info = Shell.spiel.info();
    dialog("Gewonnen – " + m.name,
      "<p><b>Alle Karten sind weg.</b></p>" +
      "<p>Zeit " + zeitText() + " · Züge " + info.zuege + " · Punkte " + info.punkte + "</p>" +
      (zusatz ? "<p>" + zusatz + "</p>" : ""),
      [{ text: "Nochmal", haupt: true, beiKlick: function () { dialogZu(); Shell.neu(); } },
      { text: "Modusauswahl", beiKlick: function () { dialogZu(); Shell.menue(); } }]);
  };

  Shell.verloren = function (text) {
    uhrAnhalten();
    dialog("Keine Züge mehr", "<p>" + text + "</p>",
      [{ text: "Neues Spiel", haupt: true, beiKlick: function () { dialogZu(); Shell.neu(); } },
      { text: "Zurücknehmen", beiKlick: function () { dialogZu(); Shell.zurueck(); } }]);
  };

  /* ---------------------------------------------------------- Menüwechsel */
  Shell.menue = function () {
    if (Shell.spiel && Shell.sekunden > 3 && !Shell.verlauf.length) {
      // nichts Besonderes – die Partie wird nur verlassen
    }
    uhrAnhalten();
    if (Shell.spiel) Shell.spiel = Shell.spiel;   // Zustand bleibt für ein späteres "Neues Spiel"
    $("spielbereich").hidden = true;
    $("modusAuswahl").hidden = false;
    $("kopfTitel").textContent = "Solitär";
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

  function brettKlick(ev) {
    if (!Shell.spiel) return;
    if (Shell.gezogen) { Shell.gezogen = false; return; }
    var karte = ev.target.closest("[data-art]");
    if (karte) {
      Shell.auswahl = wasVon(karte);
      Shell.spiel.waehlen(Shell.auswahl);
      Shell.stand(Shell.spiel.info());
      return;
    }
    var zone = zoneVon(ev.target);
    if (zone) {
      if (Shell.auswahl) Shell.spiel.ablegen(Shell.auswahl, zone);
      else Shell.spiel.zoneKlick(zone);
      Shell.stand(Shell.spiel.info());
      return;
    }
    Shell.spiel.brettKlick && Shell.spiel.brettKlick(ev);
  }

  /* Ziehen: Karte greifen, Geist folgt dem Finger, beim Loslassen entscheidet
     die Zone darunter. Ohne Bewegung bleibt es ein Tippen. */
  var zug = null;
  function ziehenStart(ev) {
    if (!Shell.spiel || (ev.button !== undefined && ev.button > 0)) return;
    var karte = ev.target.closest("[data-art]");
    if (!karte) return;
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
    if (zone) Shell.spiel.ablegen(z.was, zone);
    Shell.stand(Shell.spiel.info());
  }

  /* -------------------------------------------------------------- Aufbau */
  function tasten(ev) {
    if (!Shell.spiel) return;
    var t = ev.key.toLowerCase();
    if (ev.ctrlKey && t === "z") { ev.preventDefault(); Shell.zurueck(); return; }
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

    var brett = $("brett");
    brett.addEventListener("click", brettKlick);
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

    if (TG.MODI[Shell.modus || "klondike"]) menueBauen();
    $("modusAuswahl").hidden = false;
    $("spielbereich").hidden = true;
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
