/* ===========================================================================
   ROTLICHT – Oberfläche

   Zeichnet den Spielstand, nimmt Berührung und Maus, baut die Kulisse und
   speichert im Browser. Alle Wirtschaftsregeln stehen in rotlicht-kern.js.
   =========================================================================== */
(function () {
  "use strict";

  var RL = window.RLKern;
  if (!RL) return;

  var SCHLUESSEL = "tg-rotlicht-1";
  var TAB_NAMEN = { personal: "Personal", ausbau: "Ausbau", erfolge: "Erfolge", chronik: "Chronik" };

  /* Kulisse je Stufe: Schild, Hausbreite/-höhe, Etagen, Leute auf der Straße */
  var SZENEN = [
    { schild: "24/7", b: "26%", h: "34%", etagen: 1, leute: 1 },
    { schild: "ZIMMER", b: "29%", h: "37%", etagen: 1, leute: 2 },
    { schild: "STUDIO", b: "33%", h: "41%", etagen: 2, leute: 3 },
    { schild: "HAUS ROSA", b: "37%", h: "45%", etagen: 3, leute: 4 },
    { schild: "ETABLISSEMENT", b: "41%", h: "49%", etagen: 4, leute: 5 },
    { schild: "ROTLICHT", b: "45%", h: "53%", etagen: 5, leute: 6 },
    { schild: "MAISON ROUGE", b: "49%", h: "57%", etagen: 6, leute: 7 },
    { schild: "STUNDENHOTEL", b: "52%", h: "61%", etagen: 7, leute: 8 },
    { schild: "GRAND HOTEL", b: "55%", h: "65%", etagen: 8, leute: 9 },
    { schild: "IMPERIUM", b: "58%", h: "68%", etagen: 9, leute: 10 }
  ];

  var s = null;
  var tab = "personal";
  var e = {};
  var letzteAnsicht = { tab: "", stand: "" };

  /* --------------------------------------------------------------- Start */
  function start() {
    e = {
      geld: document.getElementById("geld"),
      proStunde: document.getElementById("proStunde"),
      stufeNr: document.getElementById("stufeNr"),
      stufeName: document.getElementById("stufeName"),
      szene: document.getElementById("szene"),
      haus: document.getElementById("haus"),
      schild: document.getElementById("schild"),
      fassade: document.getElementById("fassade"),
      volk: document.getElementById("volk"),
      skyline: document.getElementById("skyline"),
      sterne: document.getElementById("sterne"),
      stufeChipNr: document.getElementById("stufeChipNr"),
      stufeKurz: document.getElementById("stufeKurz"),
      fortschrittBalken: document.getElementById("fortschrittBalken"),
      fortschrittText: document.getElementById("fortschrittText"),
      aktionHand: document.getElementById("aktionHand"),
      aktionExtra: document.getElementById("aktionExtra"),
      betragHand: document.getElementById("betragHand"),
      betragExtra: document.getElementById("betragExtra"),
      statusHand: document.getElementById("statusHand"),
      statusHandText: document.getElementById("statusHandText"),
      statusExtra: document.getElementById("statusExtra"),
      statusExtraText: document.getElementById("statusExtraText"),
      balkenHand: document.getElementById("balkenHand"),
      balkenExtra: document.getElementById("balkenExtra"),
      kondome: document.getElementById("kondome"),
      kondomeKaufen: document.getElementById("kondomeKaufen"),
      reiter: document.getElementById("reiter"),
      inhalt: document.getElementById("inhalt"),
      flieger: document.getElementById("flieger"),
      toasts: document.getElementById("toasts"),
      schleier: document.getElementById("berichtSchleier"),
      berichtZeit: document.getElementById("berichtZeit"),
      berichtSumme: document.getElementById("berichtSumme"),
      berichtHinweis: document.getElementById("berichtHinweis"),
      berichtKnopf: document.getElementById("berichtKnopf"),
      neuAnfangen: document.getElementById("neuAnfangen")
    };

    s = lade();

    baueSterne();
    baueSzene();
    binde();
    zeigeTab("personal");
    zeichneKopf();
    zeichneAktionen();
    pruefeBericht();

    setInterval(takt, 250);
    setInterval(speichern, 5000);
    document.addEventListener("visibilitychange", function () {
      if (document.hidden) speichern();
      else { takt(); pruefeBericht(); }
    });
    window.addEventListener("pagehide", speichern);
  }

  /* ------------------------------------------------------------- Speichern */
  function lade() {
    var roh = null;
    try { roh = localStorage.getItem(SCHLUESSEL); } catch (fehler) { roh = null; }
    var st = RL.laden(roh);
    RL.abrechnen(st, Date.now());
    return st;
  }

  function speichern() {
    if (!s) return;
    try { localStorage.setItem(SCHLUESSEL, RL.speicherText(s)); } catch (fehler) { /* privater Modus */ }
  }

  /* --------------------------------------------------------------- Takt */
  function takt() {
    var jetzt = Date.now();
    RL.abrechnen(s, jetzt);

    var neueErfolge = RL.erfolgePruefen(s, jetzt);
    if (neueErfolge.length) {
      for (var i = 0; i < neueErfolge.length; i++) {
        toast("Erfolg: " + neueErfolge[i].name + " · +" + Math.round(neueErfolge[i].bonus * 100) + " % Einnahmen", "gold");
      }
      neuZeichnen();
    }

    var ereignis = RL.ereignis(s, jetzt);
    if (ereignis) {
      toast(ereignis.geld
        ? ereignis.text + " · +" + RL.formatGeld(ereignis.geld)
        : ereignis.text, ereignis.geld ? "gut" : "");
      if (tab === "chronik" || tab === "erfolge") neuZeichnen();
    }

    zeichneKopf();
    zeichneAktionen(jetzt);
    zeichneKaufbarkeit();
    pruefeBericht();
  }

  function zeichneKopf() {
    var w = RL.werte(s);
    e.geld.textContent = RL.formatGeld(s.geld);
    e.proStunde.textContent = RL.formatGeld(w.proStunde) + "/h";
    e.stufeNr.textContent = s.stufe;
    e.stufeName.textContent = w.stufe.name;
    e.stufeChipNr.textContent = s.stufe;
    e.stufeKurz.textContent = w.stufe.name;

    var naechste = RL.naechsteStufe(s);
    if (!naechste) {
      e.fortschrittBalken.style.transform = "scaleX(1)";
      e.fortschrittText.textContent = "Höchste Stufe – jetzt geht es um die Welt";
    } else {
      var anteil = Math.max(0, Math.min(1, s.geld / naechste.kosten));
      e.fortschrittBalken.style.transform = "scaleX(" + anteil.toFixed(3) + ")";
      e.fortschrittText.textContent = naechste.name + " · " + RL.formatGeld(naechste.kosten) +
        " (" + Math.round(anteil * 100) + " %)";
    }
  }

  function zeichneAktionen(jetzt) {
    if (!jetzt) jetzt = Date.now();
    var w = RL.werte(s);
    var cd = w.cooldown * 1000;
    zeichneAktion({
      art: "hand", knoten: e.aktionHand, status: e.statusHand, text: e.statusHandText,
      balken: e.balkenHand, betrag: e.betragHand, jetzt: jetzt, cd: cd
    });
    zeichneAktion({
      art: "extra", knoten: e.aktionExtra, status: e.statusExtra, text: e.statusExtraText,
      balken: e.balkenExtra, betrag: e.betragExtra, jetzt: jetzt, cd: cd
    });
    e.kondome.textContent = RL.formatZahl(s.kondome);
    e.aktionExtra.classList.toggle("verbraucht", s.kondome < 1);
  }

  function zeichneAktion(a) {
    var rest = Math.max(0, (s.bereit[a.art] || 0) - a.jetzt);
    var bereit = rest <= 0;
    var anteil = bereit ? 0 : Math.min(1, rest / a.cd);
    a.balken.style.transform = "scaleX(" + anteil.toFixed(3) + ")";
    a.knoten.classList.toggle("bereit", bereit);
    a.status.classList.toggle("frei", bereit);
    a.status.classList.toggle("wartet", !bereit);
    a.betrag.textContent = RL.formatGeld(RL.aktionsBetrag(s, a.art));
    if (bereit) {
      if (a.art === "extra" && s.kondome < 1) a.text.textContent = "Kondom fehlt – erst nachkaufen";
      else a.text.textContent = "Bereit";
    } else {
      a.text.textContent = "Bereit in " + Math.ceil(rest / 1000) + " s";
    }
  }

  function zeichneKaufbarkeit() {
    var knoepfe = e.inhalt.querySelectorAll("[data-kosten]");
    for (var i = 0; i < knoepfe.length; i++) {
      var kosten = parseFloat(knoepfe[i].getAttribute("data-kosten"));
      var reicht = isFinite(kosten) ? s.geld + 1e-9 >= kosten : true;
      knoepfe[i].classList.toggle("zu-teuer", !reicht);
    }
  }

  /* --------------------------------------------------------------- Reiter */
  function binde() {
    e.reiter.addEventListener("click", function (ev) {
      var knopf = ev.target.closest("button[data-tab]");
      if (knopf) zeigeTab(knopf.getAttribute("data-tab"));
    });

    e.aktionHand.addEventListener("click", function (ev) { tueKlick("hand", ev); });
    e.aktionExtra.addEventListener("click", function (ev) { tueKlick("extra", ev); });
    e.kondomeKaufen.addEventListener("click", kaufeKondome);
    e.neuAnfangen.addEventListener("click", neuAnfangen);
    e.berichtKnopf.addEventListener("click", function () {
      e.schleier.hidden = true;
      s.bericht = null;
      speichern();
    });

    e.inhalt.addEventListener("click", function (ev) {
      var knopf = ev.target.closest("button");
      if (!knopf || knopf.disabled) return;
      var person = knopf.getAttribute("data-personal");
      if (person) {
        var anzahl = knopf.getAttribute("data-anzahl");
        var p = RL.personalDef(person);
        var n = anzahl === "max" ? RL.maxPersonal(s, p) : parseInt(anzahl, 10) || 1;
        if (n < 1) { toast("Dafür reicht das Geld gerade nicht.", ""); return; }
        var r = RL.personalKaufen(s, person, n, Date.now());
        if (!r.ok) { kaufFehler(r); return; }
        toast(n + " × " + p.name + " eingestellt · −" + RL.formatGeld(r.preis), "gut");
        speichern(); zeichneKopf(); neuZeichnen();
        return;
      }
      var upgrade = knopf.getAttribute("data-upgrade");
      if (upgrade) {
        var ru = RL.upgradeKaufen(s, upgrade, Date.now());
        if (!ru.ok) { kaufFehler(ru); return; }
        toast(RL.upgradeDef(upgrade).name + " gekauft · −" + RL.formatGeld(ru.preis), "gut");
        speichern(); zeichneKopf(); neuZeichnen();
        return;
      }
      if (knopf.hasAttribute("data-stufe")) {
        var rs = RL.stufeKaufen(s, Date.now());
        if (!rs.ok) { kaufFehler(rs); return; }
        toast("Ausgebaut: " + rs.name, "gold");
        baueSzene();
        speichern(); zeichneKopf(); neuZeichnen();
        return;
      }
      if (knopf.hasAttribute("data-standort")) {
        var rz = RL.standortKaufen(s, Date.now());
        if (!rz.ok) { kaufFehler(rz); return; }
        toast("Standort " + rz.standorte + " eröffnet · −" + RL.formatGeld(rz.preis), "gut");
        speichern(); zeichneKopf(); neuZeichnen();
      }
    });
  }

  function kaufFehler(r) {
    if (!r) return;
    if (r.grund === "geld") toast("Dafür fehlt Geld: " + RL.formatGeld(r.preis), "");
    else if (r.grund === "gesperrt") toast("Erst ab Stufe " + r.ab + ".", "");
    else if (r.grund === "max") toast("Weiter geht es nicht – du bist ganz oben.", "gold");
    else if (r.grund === "gekauft") toast("Das ist schon gekauft.", "");
  }

  function neuZeichnen() {
    zeigeTab(tab, true);
  }

  function zeigeTab(name, ohneScroll) {
    if (!TAB_NAMEN[name]) name = "personal";
    tab = name;
    var knoepfe = e.reiter.querySelectorAll("button[data-tab]");
    for (var i = 0; i < knoepfe.length; i++) {
      var aktiv = knoepfe[i].getAttribute("data-tab") === name;
      knoepfe[i].classList.toggle("aktiv", aktiv);
      knoepfe[i].setAttribute("aria-selected", aktiv ? "true" : "false");
    }
    if (name === "personal") e.inhalt.innerHTML = bauePersonal();
    else if (name === "ausbau") e.inhalt.innerHTML = baueAusbau();
    else if (name === "erfolge") e.inhalt.innerHTML = baueErfolge();
    else e.inhalt.innerHTML = baueChronik();
    zeichneKaufbarkeit();
  }

  /* --------------------------------------------------------------- Personal */
  function bauePersonal() {
    var w = RL.werte(s);
    var teile = [];
    teile.push("<div class='abschnitt'><h2>Mitarbeiter</h2><span>" +
      RL.anzahlPersonal(s) + " angestellt · " + RL.formatGeld(w.personalStunde) +
      "/h Grundlohn · ×" + zahlKurz(w.gesamtFaktor) + " Geschäftsfaktor</span></div>");
    teile.push("<div class='karten'>");

    for (var i = 0; i < RL.PERSONAL.length; i++) {
      var p = RL.PERSONAL[i];
      var anzahl = s.personal[p.id] || 0;
      var gesperrt = s.stufe < p.ab;
      var preis1 = RL.personalKosten(s, p, 1);
      var preis10 = RL.personalKosten(s, p, 10);
      var max = gesperrt ? 0 : RL.maxPersonal(s, p);

      teile.push("<article class='karte" + (gesperrt ? " gesperrt" : "") + "' data-karte='" + p.id + "'>");
      teile.push("<div class='k-kopf'><div><div class='k-name'>" + p.name +
        "</div><span class='k-rolle'>" + p.rolle + "</span></div>" +
        "<span class='k-badge" + (anzahl ? "" : " aus") + "'>" +
        (gesperrt ? "Stufe " + p.ab : "×" + RL.formatZahl(anzahl)) + "</span></div>");
      teile.push("<p class='k-text'>" + p.text + "</p>");
      teile.push("<div class='k-werte'><span>" + RL.formatGeld(p.stunde) + "/h je Kopf</span>" +
        "<span><b>" + RL.formatGeld(p.stunde * anzahl) + "/h</b></span></div>");

      if (gesperrt) {
        teile.push("<div class='k-sperre'>Gesperrt bis Stufe " + p.ab + " · " +
          RL.STUFEN[p.ab - 1].name + "</div>");
      } else {
        teile.push("<div class='k-kauf'>");
        teile.push("<button class='knopf kauf' data-personal='" + p.id + "' data-anzahl='1' data-kosten='" +
          preis1 + "'>1 × · " + RL.formatGeld(preis1) + "</button>");
        teile.push("<button class='knopf' data-personal='" + p.id + "' data-anzahl='10' data-kosten='" +
          preis10 + "'>10 × · " + RL.formatGeld(preis10) + "</button>");
        teile.push("<button class='knopf' data-personal='" + p.id + "' data-anzahl='max' data-kosten='" +
          (max > 0 ? preis1 : 0) + "'>Max (" + max + ")</button>");
        teile.push("</div>");
      }
      teile.push("</article>");
    }
    teile.push("</div>");
    return teile.join("");
  }

  /* ----------------------------------------------------------------- Ausbau */
  function baueAusbau() {
    var w = RL.werte(s);
    var naechste = RL.naechsteStufe(s);
    var teile = [];

    teile.push("<div class='stufe-karte'>");
    teile.push("<h3>Stufe " + s.stufe + " · " + w.stufe.name + "</h3>");
    teile.push("<p class='stufen-text'>" + w.stufe.text + "</p>");
    if (naechste) {
      teile.push("<div class='stufe-reihe'><span>Nächster Ausbau: <b>" + naechste.name + "</b> · " +
        RL.formatGeld(naechste.kosten) + " · Einnahmen ×" + naechste.faktor + " statt ×" + w.stufe.faktor + "</span>");
      teile.push("<button class='knopf kauf' data-stufe='1' data-kosten='" + naechste.kosten + "'>" +
        "Ausbauen · " + RL.formatGeld(naechste.kosten) + "</button></div>");
    } else {
      teile.push("<div class='stufe-reihe'><span><b>Höchste Stufe erreicht.</b> " +
        "Jetzt zählt nur noch, wie viele Standorte dazukommen.</span></div>");
    }
    teile.push("<div class='stufen-liste'>");
    for (var k = 0; k < RL.STUFEN.length; k++) {
      var st = RL.STUFEN[k];
      teile.push("<div class='stufen-punkt" + (s.stufe >= st.nr ? " erreicht" : "") + "'><i></i>" +
        st.nr + ". " + st.name + (s.stufe >= st.nr ? "" : " · " + RL.formatGeld(st.kosten)) + "</div>");
    }
    teile.push("</div></div>");

    /* Standorte */
    var standortPreis = RL.standortKosten(s);
    teile.push("<div class='abschnitt'><h2>Standorte</h2><span>" + s.standorte +
      " · Faktor ×" + zahlKurz(w.standortFaktor) + "</span></div>");
    teile.push("<div class='karten'><article class='karte'>");
    teile.push("<div class='k-kopf'><div><div class='k-name'>Neuer Standort</div>" +
      "<span class='k-rolle'>Zweigstelle in einer anderen Stadt</span></div>" +
      "<span class='k-badge'>×" + s.standorte + "</span></div>");
    teile.push("<p class='k-text'>Jeder Standort erhöht alle Einnahmen dauerhaft um 22 %. " +
      "Die Miete steigt mit jedem Haus, das dazukommt.</p>");
    teile.push("<div class='k-werte'><span>Nächster Standort</span><span><b>" +
      RL.formatGeld(standortPreis) + "</b></span></div>");
    teile.push("<div class='k-kauf'><button class='knopf kauf' data-standort='1' data-kosten='" +
      standortPreis + "'>Eröffnen · " + RL.formatGeld(standortPreis) + "</button></div>");
    teile.push("</article></div>");

    /* Upgrades */
    teile.push("<div class='abschnitt'><h2>Verbesserungen</h2><span>" +
      RL.anzahlUpgrades(s) + " von " + RL.UPGRADES.length + " gekauft</span></div>");
    teile.push("<div class='karten'>");
    for (var i = 0; i < RL.UPGRADES.length; i++) {
      var u = RL.UPGRADES[i];
      var gekauft = !!s.upgrades[u.id];
      var gesperrt = s.stufe < u.ab;
      teile.push("<article class='karte" + (gekauft ? " fertig" : (gesperrt ? " gesperrt" : "")) + "'>");
      teile.push("<div class='k-kopf'><div><div class='k-name'>" + u.name + "</div>" +
        "<span class='k-rolle'>" + wirkungText(u) + "</span></div>" +
        "<span class='k-badge" + (gekauft ? "" : " aus") + "'>" +
        (gekauft ? "✓" : gesperrt ? "Stufe " + u.ab : "neu") + "</span></div>");
      teile.push("<p class='k-text'>" + u.text + "</p>");
      if (!gekauft && !gesperrt) {
        teile.push("<div class='k-kauf'><button class='knopf kauf' data-upgrade='" + u.id +
          "' data-kosten='" + u.preis + "'>Kaufen · " + RL.formatGeld(u.preis) + "</button></div>");
      } else if (gekauft) {
        teile.push("<div class='k-werte'><span>Im Betrieb</span><span><b>aktiv</b></span></div>");
      } else {
        teile.push("<div class='k-sperre'>Gesperrt bis Stufe " + u.ab + "</div>");
      }
      teile.push("</article>");
    }
    teile.push("</div>");
    return teile.join("");
  }

  function wirkungText(u) {
    if (u.art === "einnahme") return "+" + Math.round((u.faktor - 1) * 100) + " % Einnahmen";
    if (u.art === "tempo") return "−" + Math.round((1 - u.faktor) * 100) + " % Wartezeit";
    return "+" + Math.round((u.faktor - 1) * 100) + " % Offline-Ertrag · +" + u.stunden + " h";
  }

  /* ---------------------------------------------------------------- Erfolge */
  function baueErfolge() {
    var teile = [];
    var hat = 0;
    for (var i = 0; i < RL.ERFOLGE.length; i++) if (s.erfolge[RL.ERFOLGE[i].id]) hat++;
    teile.push("<div class='abschnitt'><h2>Erfolge</h2><span>" + hat + " von " +
      RL.ERFOLGE.length + " · jeder Erfolg erhöht die Einnahmen dauerhaft</span></div>");
    teile.push("<div class='erfolge'>");
    for (i = 0; i < RL.ERFOLGE.length; i++) {
      var a = RL.ERFOLGE[i];
      var geschafft = !!s.erfolge[a.id];
      teile.push("<div class='erfolg" + (geschafft ? " hat" : "") + "'>" +
        "<span class='medaille'>" + (geschafft ? "★" : "☆") + "</span>" +
        "<div><div class='e-name'>" + a.name + "</div>" +
        "<div class='e-text'>" + a.text + "</div>" +
        "<div class='e-bonus'>" + (geschafft ? "geschafft" : "offen") + " · +" +
        Math.round(a.bonus * 100) + " % Einnahmen</div></div></div>");
    }
    teile.push("</div>");
    return teile.join("");
  }

  /* ---------------------------------------------------------------- Chronik */
  function baueChronik() {
    var w = RL.werte(s);
    var teile = ["<div class='werte-liste'>"];
    teile.push(wert("Kontostand", RL.formatGeld(s.geld)));
    teile.push(wert("Einnahmen je Stunde", RL.formatGeld(w.proStunde)));
    teile.push(wert("Gesamteinnahmen", RL.formatGeld(s.gesamtEinnahmen)));
    teile.push(wert("Gesamtausgaben", RL.formatGeld(s.gesamtAusgaben)));
    teile.push(wert("Mitarbeiter", RL.anzahlPersonal(s)));
    teile.push(wert("Standorte", s.standorte));
    teile.push(wert("Geschäftsfaktor", "×" + zahlKurz(w.gesamtFaktor)));
    teile.push(wert("Spielzeit", RL.formatDauer(s.gespielt)));
    teile.push(wert("Eigene Handgriffe", RL.formatZahl(s.klicks)));
    teile.push(wert("Wartezeit je Aktion", w.cooldown + " s"));
    teile.push("</div>");

    teile.push("<div class='abschnitt'><h2>Chronik</h2><span>die letzten Ereignisse</span></div>");
    if (!s.chronik || !s.chronik.length) {
      teile.push("<div class='leer'>Noch nichts passiert. Die Straße wartet.</div>");
    } else {
      teile.push("<ul class='chronik'>");
      for (var i = 0; i < s.chronik.length; i++) {
        var z = s.chronik[i];
        teile.push("<li><time>" + uhrzeit(z.t) + "</time><span>" + z.text + "</span></li>");
      }
      teile.push("</ul>");
    }
    return teile.join("");
  }

  function wert(name, inhalt) {
    return "<div class='wert'><span>" + name + "</span><b>" + inhalt + "</b></div>";
  }

  function uhrzeit(t) {
    try {
      return new Date(t).toLocaleTimeString("de-DE", {
        hour: "2-digit", minute: "2-digit", timeZone: "Europe/Berlin"
      });
    } catch (fehler) {
      return "—";
    }
  }

  function zahlKurz(x) {
    if (!isFinite(x)) return "∞";
    if (x < 1000) return (Math.round(x * 10) / 10).toString().replace(".", ",");
    return RL.formatGeld(x).replace(" €", "");
  }

  /* --------------------------------------------------------------- Aktionen */
  function tueKlick(art, ev) {
    var r = RL.klick(s, art, Date.now());
    if (!r.ok) {
      if (r.grund === "keine-kondome") {
        toast("Ohne Kondom geht das nicht – vier Stück kosten 5 €.", "");
        e.aktionExtra.classList.add("verbraucht");
      } else if (r.grund === "warte") {
        toast("Nächste Spielstunde in " + r.rest + " s.", "");
      }
      return r;
    }
    var x = 200, y = 300;
    if (ev && typeof ev.clientX === "number" && (ev.clientX || ev.clientY)) {
      x = ev.clientX; y = ev.clientY;
    } else {
      var b = (art === "hand" ? e.aktionHand : e.aktionExtra).getBoundingClientRect();
      x = b.left + b.width / 2; y = b.top + b.height / 2;
    }
    flieger(x, y, "+" + RL.formatGeld(r.betrag));
    e.geld.classList.remove("puls");
    void e.geld.offsetWidth;
    e.geld.classList.add("puls");
    setTimeout(function () { e.geld.classList.remove("puls"); }, 160);
    zeichneKopf();
    zeichneAktionen(Date.now());
    if (tab === "personal" || tab === "ausbau") neuZeichnen();
    speichern();
    return r;
  }

  function kaufeKondome() {
    var r = RL.kondomeKaufen(s, 1, Date.now());
    if (!r.ok) { kaufFehler(r); return; }
    toast("Vier Kondome gekauft · −" + RL.formatGeld(r.preis), "gut");
    zeichneAktionen(Date.now());
    speichern();
  }

  function neuAnfangen() {
    if (!window.confirm("Wirklich neu anfangen? Der Spielstand in diesem Browser wird gelöscht.")) return;
    try { localStorage.removeItem(SCHLUESSEL); } catch (fehler) { /* egal */ }
    s = RL.neuerStand(Date.now());
    baueSzene();
    zeigeTab("personal");
    zeichneKopf();
    zeichneAktionen();
    speichern();
    toast("Neue Straße, neues Glück. Viel Erfolg.", "gold");
  }

  /* ------------------------------------------------------------- Rückkehr */
  function pruefeBericht() {
    if (!s || !s.bericht) return;
    var b = s.bericht;
    e.berichtZeit.textContent = "Du warst " + RL.formatDauer(b.sek) + " weg.";
    e.berichtSumme.textContent = "+" + RL.formatGeld(b.geld);
    var hinweis = "Das Personal hat ohne dich weitergearbeitet.";
    if (b.gekappt) hinweis = "Mehr als " + RL.werte(s).offlineStunden + " Stunden werden nicht vergütet.";
    else if (b.bonus > 1) hinweis += " Dank Buchhaltung " + Math.round((b.bonus - 1) * 100) + " % mehr.";
    e.berichtHinweis.textContent = hinweis;
    e.schleier.hidden = false;
    s.bericht = null;
  }

  /* --------------------------------------------------------------- Kulisse */
  function baueSterne() {
    var stueck = [];
    for (var i = 0; i < 26; i++) {
      var x = (i * 37 + 11) % 97;
      var y = (i * 53 + 7) % 52;
      stueck.push("<i class='stern' style='left:" + x + "%;top:" + y + "%;animation-delay:" +
        ((i % 7) * 0.6).toFixed(1) + "s'></i>");
    }
    e.sterne.innerHTML = stueck.join("");
  }

  function baueSzene() {
    var d = SZENEN[Math.min(SZENEN.length, Math.max(1, s.stufe)) - 1];
    e.szene.setAttribute("data-s", s.stufe);
    e.haus.style.setProperty("--b", d.b);
    e.haus.style.setProperty("--h", d.h);
    e.schild.textContent = d.schild;

    var reihen = [];
    for (var et = 0; et < d.etagen; et++) {
      var zellen = [];
      for (var f = 0; f < 3; f++) {
        var nr = et * 3 + f;
        var klasse = "fenster";
        if (nr % 5 === 0) klasse += " aus";
        else if (nr % 4 === 1) klasse += " rot";
        zellen.push("<i class='" + klasse + "' style='animation-delay:" + (nr * 0.41).toFixed(2) + "s'></i>");
      }
      reihen.push("<div class='reihe'>" + zellen.join("") + "</div>");
    }
    e.fassade.innerHTML = reihen.join("");

    var hoch = [];
    for (var i = 0; i < 9; i++) {
      var h = 22 + ((i * 17 + s.stufe * 11) % 34) + s.stufe * 3;
      var lichter = "";
      for (var k = 0; k < Math.max(1, Math.round(h / 22)); k++) lichter += "<i></i>";
      hoch.push("<div class='sk' style='height:" + h + "px'>" + lichter + "</div>");
    }
    e.skyline.innerHTML = hoch.join("");

    var leute = [];
    var anzahl = Math.min(6, Math.max(1, Math.round(d.leute / 2)));
    for (i = 0; i < anzahl; i++) {
      leute.push("<span class='mensch' style='animation-delay:" + (i * 0.7).toFixed(1) + "s'></span>");
    }
    e.volk.innerHTML = leute.join("");
  }

  /* --------------------------------------------------------------- Anzeigen */
  function flieger(x, y, text) {
    var span = document.createElement("span");
    span.className = "flug";
    span.textContent = text;
    span.style.left = Math.round(x) + "px";
    span.style.top = Math.round(y) + "px";
    e.flieger.appendChild(span);
    setTimeout(function () {
      if (span.parentNode) span.parentNode.removeChild(span);
    }, 1050);
  }

  function toast(text, art) {
    var div = document.createElement("div");
    div.className = "toast" + (art ? " " + art : "");
    div.textContent = text;
    e.toasts.appendChild(div);
    while (e.toasts.children.length > 3) e.toasts.removeChild(e.toasts.firstChild);
    setTimeout(function () {
      if (div.parentNode) div.parentNode.removeChild(div);
    }, 3800);
  }

  /* ------------------------------------------------------------ Prüfhaken */
  window.RLSpiel = {
    stand: function () { return s; },
    werte: function () { return RL.werte(s); },
    takt: takt,
    klick: tueKlick,
    kaufePersonal: function (id, n) { var r = RL.personalKaufen(s, id, n || 1, Date.now()); speichern(); zeichneKopf(); neuZeichnen(); return r; },
    kaufeUpgrade: function (id) { var r = RL.upgradeKaufen(s, id, Date.now()); speichern(); zeichneKopf(); neuZeichnen(); return r; },
    kaufeStufe: function () { var r = RL.stufeKaufen(s, Date.now()); if (r.ok) baueSzene(); speichern(); zeichneKopf(); neuZeichnen(); return r; },
    kaufeStandort: function () { var r = RL.standortKaufen(s, Date.now()); speichern(); zeichneKopf(); neuZeichnen(); return r; },
    setzeGeld: function (n) { s.geld = Math.max(0, Number(n) || 0); zeichneKopf(); neuZeichnen(); },
    zeigeTab: zeigeTab,
    speichern: speichern,
    standort: function () { return location.origin + location.pathname; }
  };

  if (document.readyState === "loading") document.addEventListener("DOMContentLoaded", start);
  else start();
})();
