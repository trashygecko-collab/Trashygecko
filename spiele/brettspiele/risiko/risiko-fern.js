/* ===========================================================================
   Risiko online (Phase 2, 22.11.2026)
   ---------------------------------------------------------------------------
   Geladen von spiele/brettspiele/risiko/risiko.html, wenn die Adresse
   ?room=…&id=…&name=… trägt. Der Server (/api/risiko) führt den Stand und
   prüft jede Aktion; diese Datei baut nur die Lobby und schickt Absichten.
   Ohne Raumangabe bleibt das Spiel wie vorher im Einzelbetrieb.
   ========================================================================= */
(function () {
  "use strict";

  var frage = new URLSearchParams(location.search);
  var RAUM = (frage.get("room") || "").toUpperCase();
  if (!RAUM) return;

  /* Ein Einladungslink trägt nur den Raum. Wer ihn öffnet, bekommt hier eine
     eigene Kennung und einen eigenen Namen – nie die des Einladenden. */
  function gespeichert(schluessel) {
    try { return localStorage.getItem(schluessel) || ""; } catch (e) { return ""; }
  }
  function merken(schluessel, wert) {
    try { localStorage.setItem(schluessel, wert); } catch (e) {}
  }
  var ICH = frage.get("id") || gespeichert("tg-risiko-id:" + RAUM);
  if (!ICH) {
    ICH = "r" + Date.now().toString(36) + Math.floor(Math.random() * 1e6).toString(36);
    merken("tg-risiko-id:" + RAUM, ICH);
  }
  var MEIN_NAME = (frage.get("name") || gespeichert("tg-risiko-name") || "").trim() || "Spieler";
  merken("tg-risiko-name", MEIN_NAME);

  // Adresse mit eigener Kennung: token.js hängt das Token an jeden Ruf.
  function adresseAuffrischen() {
    try {
      var u = new URL(location.href);
      u.searchParams.set("room", RAUM);
      u.searchParams.set("id", ICH);
      u.searchParams.set("name", MEIN_NAME);
      history.replaceState(null, "", u.toString());
    } catch (e) {}
  }
  adresseAuffrischen();

  var SP = window.Risiko;
  if (!SP) return;

  var API = "/api/risiko?room=" + encodeURIComponent(RAUM);
  var spiel = null;          // letzter Serverstand
  var meinPlatz = -1;
  var laeuft = false;        // eine Anfrage ist unterwegs
  var gestartet = false;     // Spielfeld schon aufgebaut
  var letzterFehler = "";
  var takt = 0;
  var abrufLaeuft = false;
  var laeuftSeit = 0;
  var aufgebenBereit = false;
  var letzteZeichnung = "";

  /* ------------------------------------------------------------------ Leiste */
  var leiste = document.createElement("div");
  leiste.id = "fernLeiste";
  leiste.style.cssText = [
    "position:sticky", "top:0", "z-index:60", "display:flex", "flex-wrap:wrap",
    "gap:8px", "align-items:center", "padding:10px 12px", "margin:0 0 10px",
    "background:#0b1a22f2", "color:#e6eef2", "border:1px solid #ffffff26",
    "border-radius:12px", "font:14px/1.45 system-ui,-apple-system,'Segoe UI',sans-serif"
  ].join(";");
  var standText = document.createElement("span");
  standText.style.cssText = "flex:1 1 240px;min-width:0";
  leiste.appendChild(standText);
  var steuerung = document.createElement("span");
  steuerung.style.cssText = "display:flex;flex-wrap:wrap;gap:8px;align-items:center";
  leiste.appendChild(steuerung);

  function knopf(text, aktion, art) {
    var b = document.createElement("button");
    b.type = "button";
    b.textContent = text;
    b.style.cssText = [
      "font:inherit", "padding:7px 13px", "border-radius:9px", "cursor:pointer",
      "border:1px solid #ffffff33",
      art === "haupt" ? "background:#2c6b4f;color:#ffffff" : "background:#17362b;color:#e6eef2"
    ].join(";");
    b.addEventListener("click", function () { aktion(); });
    steuerung.appendChild(b);
    return b;
  }

  function auswahl(werte, beschriftung) {
    var s = document.createElement("select");
    s.setAttribute("aria-label", beschriftung);
    s.style.cssText = "font:inherit;padding:6px 8px;border-radius:8px;background:#0b1a22;color:#e6eef2;border:1px solid #ffffff33";
    werte.forEach(function (w) {
      var o = document.createElement("option");
      o.value = w[0];
      o.textContent = w[1];
      s.appendChild(o);
    });
    steuerung.appendChild(s);
    return s;
  }

  /* ------------------------------------------------------------------ Server */
  function ruf(aktion, daten) {
    var ziel = API;
    var opt = { headers: { "Content-Type": "application/json" } };
    if (aktion) {
      ziel += "&action=" + encodeURIComponent(aktion);
      opt.method = "POST";
      opt.body = JSON.stringify(Object.assign({ id: ICH }, daten || {}));
    } else {
      ziel += "&id=" + encodeURIComponent(ICH);
      opt.method = "GET";
    }
    return fetch(ziel, opt).then(function (r) {
      return r.text().then(function (text) {
        var d = null;
        try { d = JSON.parse(text); } catch (e) {}
        if (!r.ok) throw new Error((d && d.fehler) || text || ("Fehler " + r.status));
        return d;
      });
    });
  }

  // Aktionen werden nie verworfen: sie warten in einer Schlange, bis der
  // laufende Ruf fertig ist. Der Abgleich hat dafür einen eigenen Haken.
  var warteschlange = [];
  var sendetGerade = false;

  function sende(aktion, daten) {
    warteschlange.push([aktion, daten || {}]);
    sendeNaechste();
  }

  function sendeNaechste() {
    if (sendetGerade || !warteschlange.length) return;
    sendetGerade = true;
    var auftrag = warteschlange.shift();
    ruf(auftrag[0], auftrag[1]).then(function (d) {
      letzterFehler = "";
      if (d && d.game) uebernehmen(d.game);
    }).catch(function (e) {
      letzterFehler = e.message;
    }).then(function () {
      sendetGerade = false;
      zeichneLeiste();
      sendeNaechste();
    });
  }

  /* ------------------------------------------------------- Name und Einladung */
  function nameFeld() {
    var f = document.createElement("input");
    f.type = "text";
    f.value = MEIN_NAME;
    f.maxLength = 14;
    f.setAttribute("aria-label", "Dein Name");
    f.title = "Dein Name";
    f.style.cssText = "font:inherit;padding:6px 8px;width:110px;border-radius:8px;" +
      "background:#0b1a22;color:#e6eef2;border:1px solid #ffffff33";
    f.addEventListener("change", function () {
      var neu = (f.value || "").trim().slice(0, 14) || "Spieler";
      MEIN_NAME = neu;
      merken("tg-risiko-name", neu);
      adresseAuffrischen();
      fetch("/api/presence", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ room: RAUM, id: ICH, name: neu, umbenennen: true, active: true })
      }).then(function () { abholen(); }).catch(function () {});
      zeichneLeiste();
    });
    steuerung.appendChild(f);
  }

  function einladungsLink() {
    return location.origin + location.pathname + "?room=" + encodeURIComponent(RAUM);
  }

  function linkKnopf() {
    knopf("Link kopieren", function () {
      var adresse = einladungsLink();
      var fertig = function () {
        letzterFehler = "";
        standText.textContent = "Einladungslink kopiert: " + adresse;
      };
      if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(adresse).then(fertig, function () {
          window.prompt("Einladungslink:", adresse);
        });
      } else {
        window.prompt("Einladungslink:", adresse);
      }
    });
  }

  /* ---------------------------------------------------------------- Anzeigen */
  function namen() {
    return (spiel.spieler || []).map(function (p) {
      return p.name + (p.bot ? " (KI)" : "") + (p.bereit ? " ✓" : "");
    }).join(" · ");
  }

  function zeichneLeiste() {
    steuerung.innerHTML = "";
    if (!spiel) {
      nameFeld();
      standText.textContent = "Raum " + RAUM + " – noch keine Partie." + (letzterFehler ? " " + letzterFehler : "");
      var zahl = auswahl([[2, "2 Spieler"], [3, "3 Spieler"], [4, "4 Spieler"],
        [5, "5 Spieler"], [6, "6 Spieler"]], "Spielerzahl");
      zahl.value = "4";
      knopf("Partie eröffnen", function () {
        sende("create", { anzahl: parseInt(zahl.value, 10), modus: "welt" });
      }, "haupt");
      return;
    }
    if (spiel.status === "invited") {
      nameFeld();
      linkKnopf();
      standText.textContent = "Lobby " + spiel.spieler.length + "/" + spiel.anzahl + ": " + namen() +
        (meinPlatz < 0 ? " – du sitzt noch nicht" : "") + (letzterFehler ? " · " + letzterFehler : "");
      if (meinPlatz < 0) {
        knopf("Platz nehmen", function () { sende("join"); }, "haupt");
        return;
      }
      if (meinPlatz !== 0) {
        knopf(spiel.spieler[meinPlatz].bereit ? "Doch nicht bereit" : "Bereit", function () { sende("ready"); });
        return;
      }
      // Gastgeber: Modus, Tempo, Bots, Start.
      var modus = auswahl([["welt", "Weltherrschaft"], ["schnell", "Schnellspiel"],
        ["kontinent", "Kontinentkrieg"], ["mission", "Geheime Mission"],
        ["hauptstadt", "Hauptstadtkrieg"], ["team", "Teamkrieg"]], "Spielmodus");
      modus.value = spiel.modus || "welt";
      modus.addEventListener("change", function () { sende("modus", { modus: modus.value }); });
      var tempo = auswahl([["schnell", "Tempo schnell"], ["normal", "Tempo normal"], ["langsam", "Tempo langsam"]], "Tempo");
      tempo.value = spiel.pause || "normal";
      tempo.addEventListener("change", function () { sende("pause", { pause: tempo.value }); });
      var stufe = auswahl([["leicht", "Bot leicht"], ["mittel", "Bot mittel"],
        ["schwer", "Bot schwer"], ["experte", "Bot experte"]], "Bot-Stärke");
      var wesen = auswahl([["ausgewogen", "ausgewogen"], ["aggressiv", "aggressiv"],
        ["defensiv", "defensiv"], ["expansion", "Expansion"], ["kontinent", "Kontinentjäger"],
        ["taktiker", "Taktiker"], ["risikofreudig", "risikofreudig"]], "Bot-Persönlichkeit");
      if (spiel.spieler.length < spiel.anzahl) {
        knopf("Bot hinzufügen", function () {
          sende("botadd", { stufe: stufe.value, wesen: wesen.value });
        });
      }
      var botPlatz = -1;
      (spiel.spieler || []).forEach(function (p, i) { if (p.bot) botPlatz = i; });
      if (botPlatz >= 0) {
        knopf("Bot entfernen", function () { sende("botremove", { platz: botPlatz }); });
      }
      knopf("Partie starten", function () { sende("start"); }, "haupt");
      return;
    }

    if (spiel.status === "finished") {
      var sieger = spiel.sieger >= 0 && spiel.spieler[spiel.sieger] ? spiel.spieler[spiel.sieger].name : "–";
      standText.textContent = "Beendet – " + sieger + " gewinnt. " + (spiel.siegGrund || "") +
        (letzterFehler ? " · " + letzterFehler : "");
      knopf("Neue Runde", function () { sende("neu"); }, "haupt");
      return;
    }

    var dran = spiel.spieler[spiel.current] || spiel.spieler[spiel.amZug] || {};
    var ichDran = meinPlatz === spiel.amZug;
    standText.textContent = (ichDran ? "Du bist am Zug" : (dran.name || "?") + " ist am Zug") +
      " · Runde " + spiel.runde + " · " + (spiel.phase || "") +
      (letzterFehler ? " · " + letzterFehler : "");
    if (meinPlatz >= 0) {
      // Zwei Klicks statt Dialog: der erste fragt nach, der zweite gibt auf.
      knopf(aufgebenBereit ? "Wirklich aufgeben?" : "Aufgeben", function () {
        if (!aufgebenBereit) {
          aufgebenBereit = true;
          letzteZeichnung = "";
          zeichneLeiste();
          window.setTimeout(function () {
            if (aufgebenBereit) { aufgebenBereit = false; letzteZeichnung = ""; zeichneLeiste(); }
          }, 4000);
          return;
        }
        aufgebenBereit = false;
        sende("aufgeben");
      });
    }
  }

  function uebernehmen(d) {
    spiel = d;
    meinPlatz = -1;
    (spiel.spieler || []).forEach(function (p, i) { if (p.id === ICH) meinPlatz = i; });
    if (spiel.status === "active" || spiel.status === "finished") {
      if (!gestartet) {
        gestartet = true;
        try {
          SP.fern(schnittstelle);
        } catch (e) {
          // Das Feld wird auch ohne Schmuck aufgebaut; der Stand kommt gleich.
          console.error("Spielfeld:", e);
        }
        // Auf dem Handy ist die Welt in Originalgröße zu klein: gleich die
        // große Karte zeigen (seitlich schieben).
        if (window.innerWidth < 620 && SP.zoom) {
          try { SP.zoom(true); } catch (e) {}
        }
      }
      SP.fernStand(spiel);
    } else {
      document.getElementById("spiel").hidden = true;
      document.getElementById("einrichtung").hidden = true;
    }
    // Nur neu bauen, wenn sich wirklich etwas gezeigt hat: sonst verliert das
    // Namensfeld bei jedem Abgleich den Fokus und Knopfklicks gehen ins Leere.
    var zeichen = JSON.stringify([spiel.status, spiel.phase, spiel.runde, spiel.amZug,
      spiel.pool, spiel.anzahl, meinPlatz, spiel.sieger, spiel.grund, letzterFehler,
      (spiel.spieler || []).map(function (p) {
        return [p.id, p.name, p.land, p.armeen, p.bereit];
      })]);
    if (zeichen === letzteZeichnung) return;
    letzteZeichnung = zeichen;
    zeichneLeiste();
  }

  var schnittstelle = {
    sende: function (aktion, daten) { sende(aktion, daten); },
    dran: function () { return !!spiel && spiel.status === "active" && meinPlatz === spiel.amZug; },
    bereit: function () {}
  };

  /* --------------------------------------------------------------- Abgleich */
  function abholen() {
    // Ein hängender Ruf darf den Takt nicht für immer blockieren.
    if (abrufLaeuft) {
      if (Date.now() - laeuftSeit < 6000) return;
      abrufLaeuft = false;
    }
    abrufLaeuft = true;
    laeuftSeit = Date.now();
    ruf().then(function (d) {
      if (d && d.game) uebernehmen(d.game);
    }).catch(function (e) {
      letzterFehler = e.message;
      zeichneLeiste();
    }).then(function () { abrufLaeuft = false; },
      function () { abrufLaeuft = false; });
  }

  /* ------------------------------------------------------------------- Start */
  document.body.insertBefore(leiste, document.body.firstChild);
  // Im Online-Betrieb führt immer der Server; Einrichtung und Spielfeld
  // bleiben zu, bis der Server eine laufende Partie meldet.
  document.getElementById("einrichtung").hidden = true;
  document.getElementById("spiel").hidden = true;
  zeichneLeiste();

  fetch("/api/presence", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ room: RAUM, id: ICH, name: MEIN_NAME, neu: true, active: true })
  }).then(function () { return ruf("rejoin"); }).then(function (d) {
    if (d && d.game) uebernehmen(d.game);
  }).catch(function () {
    // Noch keine Partie im Raum: Lobby zeigen, der Gastgeber eröffnet.
    spiel = null;
    zeichneLeiste();
  });

  // Lebenszeichen: Wer nur den Stand abruft, bleibt sonst nach 150 s nicht
  // als Spieler erkannt, und der Raum räumt seinen Platz weg.
  window.setInterval(function () {
    fetch("/api/presence", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ room: RAUM, id: ICH, name: MEIN_NAME, active: true })
    }).catch(function () {});
  }, 45000);

  takt = window.setInterval(abholen, 600);
  // Ein Hintergrund-Tab darf seinen Takt drosseln: beim Zurückkommen sofort
  // nachsehen und sich neu anmelden, statt auf den nächsten Schlag zu warten.
  ["visibilitychange", "focus", "pageshow", "online"].forEach(function (ereignis) {
    window.addEventListener(ereignis, function () {
      if (document.hidden) return;
      fetch("/api/presence", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ room: RAUM, id: ICH, name: MEIN_NAME, active: true })
      }).catch(function () {});
      abrufLaeuft = false;
      abholen();
    });
  });
  window.addEventListener("beforeunload", function () { window.clearInterval(takt); });
})();
