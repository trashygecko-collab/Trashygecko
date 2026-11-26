/* ===========================================================================
   Online-Mehrspieler für Brettspiele mit eigenem Spielstand
   ---------------------------------------------------------------------------
   Geladen von:
     spiele/brettspiele/risiko/risiko.html
     spiele/brettspiele/verruecktes-labyrinth/verruecktes-labyrinth.html

   Die Regeln bleiben im Spiel (dort liegen Karten, Armeen, Kacheln). Der
   Server (/api/risiko bzw. /api/labyrinth) führt den verbindlichen Stand und
   die Zugreihenfolge:
     - wer am Tisch sitzt,
     - wer am Zug ist (nur dieser darf einen Stand schicken),
     - der letzte Stand mit Versionsnummer.
   Jeder Zug geht als vollständiger Stand an den Server und kommt bei allen
   anderen beim nächsten Abruf an. Wiederverbinden heißt nur: Stand abholen.

   Ohne ?room=…&id=… in der Adresse passiert hier nichts – das Spiel bleibt
   genau wie vorher im Einzelbetrieb.
   ========================================================================= */
(function () {
  "use strict";

  var frage = new URLSearchParams(location.search);
  var RAUM = (frage.get("room") || "").toUpperCase();
  var ICH = frage.get("id") || "";
  var MEIN_NAME = (frage.get("name") || "Spieler").trim() || "Spieler";
  var SPIEL = window.TG_ONLINE_SPIEL || "";
  if (!RAUM || !ICH || !SPIEL) return;

  var API = "/api/" + SPIEL;
  var RISIKO = SPIEL === "risiko";
  // Mensch ärgere dich nicht heißt im Server maensync: dort liegt nur der
  // Stand, gewürfelt und gezogen wird im Spiel.
  var MAEN = SPIEL === "maensync";
  var SP = window[MAEN ? "MenschAergereDichNicht" : (RISIKO ? "Risiko" : "Labyrinth")];
  if (!SP) return;

  /* ---------------------------------------------------------- Spiel-Anbindung */
  var anbindung = {
    stand: function () { return RISIKO ? SP.exportieren() : SP.stand(); },
    lade: function (d) { return RISIKO ? SP.laden(d) : SP.laden(d); },
    neu: function (anzahl, namen) {
      if (MAEN) {
        SP.neu({ anzahl: anzahl, namen: namen });
      } else if (RISIKO) {
        var ki = [];
        for (var i = 0; i < anzahl; i++) ki.push(false);
        SP.neu({ modus: "welt", anzahl: anzahl, namen: namen, ki: ki });
      } else {
        SP.neu({ modus: "mehrspieler", spielerzahl: (namen.length || anzahl), namen: namen, groesse: 7, kreuze: true, aufgaben: 5 });
      }
    },
    /* Risiko braucht den Aufbau der Karte (start); das Labyrinth zeigt das
       Brett schon beim Übernehmen des Stands (laden) – ein zweiter Start
       würde dort ein neues, zufälliges Brett bauen. */
    zeige: function () { if (MAEN) SP.zeige(); else if (RISIKO) SP.start(); },
    dran: function (d) { return Number(d && d.amZug) || 0; },
    vorbei: function (d) { return !!(d && (MAEN ? d.beendet : d.vorbei)); }
  };

  /* ------------------------------------------------------------------ Zustand */
  var spiel = null;          // Spielstand des Servers
  var meinPlatz = -1;
  var letzterStand = "";     // zuletzt bekannter Stand als JSON-Text
  var gezeigt = -1;          // Version, die im Spiel angezeigt wird
  var gestartet = false;     // Spielansicht schon aufgebaut?
  var laeuft = false;        // ein Abruf ist unterwegs
  var hinweisTimer = null;

  /* -------------------------------------------------------------------- Leiste */
  var leiste = document.createElement("div");
  leiste.id = "onlineLeiste";
  leiste.setAttribute("aria-live", "polite");
  leiste.style.cssText = [
    "position:sticky", "top:0", "z-index:60", "display:flex", "flex-wrap:wrap",
    "gap:8px", "align-items:center", "padding:8px 10px", "margin:0 0 8px",
    "background:#0b1a22ee", "color:#e6eef2", "border:1px solid #ffffff24",
    "border-radius:12px", "font:14px/1.4 system-ui,-apple-system,'Segoe UI',sans-serif"
  ].join(";");
  var leisteText = document.createElement("span");
  leisteText.style.cssText = "flex:1 1 220px;min-width:0";
  leiste.appendChild(leisteText);

  function knopf(text, aktion) {
    var b = document.createElement("button");
    b.type = "button";
    b.textContent = text;
    b.style.cssText = [
      "font:inherit", "padding:7px 14px", "border-radius:9px", "cursor:pointer",
      "border:1px solid #ffffff33", "background:#17362b", "color:#e6eef2"
    ].join(";");
    b.addEventListener("click", aktion);
    leiste.appendChild(b);
    return b;
  }

  /* ---------------------------------------------------------------- Sperrschicht */
  var schleier = document.createElement("div");
  schleier.id = "onlineSchleier";
  schleier.style.cssText = [
    "position:fixed", "inset:0", "z-index:55", "display:none",
    "align-items:center", "justify-content:center", "pointer-events:auto",
    "background:#06131a80", "color:#f2f7fa", "text-align:center", "padding:24px",
    "font:16px/1.6 system-ui,-apple-system,'Segoe UI',sans-serif"
  ].join(";");
  var schleierText = document.createElement("div");
  schleierText.style.cssText = "background:#0b1a22f2;border:1px solid #ffffff2e;border-radius:14px;padding:18px 22px;max-width:420px";
  schleier.appendChild(schleierText);

  function sperren(text) {
    if (!text) { schleier.style.display = "none"; return; }
    schleierText.textContent = text;
    schleier.style.display = "flex";
  }

  /* ------------------------------------------------------------------- Server */
  function ruf(aktion, daten) {
    var ziel = API + "?room=" + encodeURIComponent(RAUM) + "&id=" + encodeURIComponent(ICH);
    if (aktion) ziel += "&action=" + encodeURIComponent(aktion);
    var opt = { method: aktion ? "POST" : "GET", headers: { "Content-Type": "application/json" } };
    if (aktion) opt.body = JSON.stringify(Object.assign({ id: ICH }, daten || {}));
    return fetch(ziel, opt).then(function (r) {
      return r.text().then(function (text) {
        var d = null;
        try { d = JSON.parse(text); } catch (e) {}
        if (!r.ok) throw new Error((d && d.fehler) || text || ("Fehler " + r.status));
        return d;
      });
    });
  }

  function melde(text) {
    leisteText.textContent = text;
    clearTimeout(hinweisTimer);
    hinweisTimer = setTimeout(zeichneLeiste, 4000);
  }

  /* ------------------------------------------------------------------ Anzeige */
  function namenListe() {
    return (spiel.spieler || []).map(function (p) {
      return p.name + (p.bot ? " (Computer)" : "") + (p.bereit ? " ✓" : "");
    }).join(" · ");
  }

  function zeichneLeiste() {
    if (!spiel) { leisteText.textContent = "Verbinde mit Spielraum " + RAUM + " …"; return; }
    Array.prototype.slice.call(leiste.querySelectorAll("button")).forEach(function (b) { b.remove(); });

    if (spiel.status === "invited") {
      var anzahl = spiel.anzahl;
      leisteText.textContent = "Lobby " + spiel.spieler.length + "/" + anzahl + ": " + namenListe() +
        (meinPlatz < 0 ? " – du sitzt noch nicht" : "");
      if (meinPlatz >= 0) knopf("Bereit", function () { aktion("ready"); });
      if (meinPlatz === 0) {
        // Computer am Tisch gibt es nur, wo der Server sie auch spielt. Beim
        // Mensch-ärgere-dich-nicht sitzt der Bot im Spiel (Einzelpartie).
        if (!MAEN && spiel.spieler.length < spiel.anzahl) {
          knopf("+ Computer", function () { aktion("botadd", { stufe: "mittel" }); });
        }
        knopf("Start", function () { aktion("start"); });
      }
      sperren("Warten auf den Start – " + spiel.spieler.length + " von " + anzahl + " sitzen am Tisch.");
      return;
    }

    if (spiel.status === "finished") {
      var sieger = spiel.sieger >= 0 && spiel.spieler[spiel.sieger] ? spiel.spieler[spiel.sieger].name : "–";
      leisteText.textContent = "Partie beendet – " + sieger + " gewinnt. " + (spiel.note || "");
      knopf("Nochmal", function () { aktion("neu"); });
      return;
    }

    var dran = spiel.spieler[spiel.current];
    var ichDran = meinPlatz >= 0 && spiel.current === meinPlatz;
    leisteText.textContent = (ichDran ? "Du bist am Zug" : (dran ? dran.name + " ist am Zug" : "Warte …")) +
      " · Zug " + spiel.zugNr;
    sperren(ichDran ? "" : (dran ? dran.name + " ist am Zug …" : "Warte auf den Spielstand …"));
  }

  /* ------------------------------------------------------------------ Aktionen */
  function aktion(name, daten) {
    return ruf(name, daten).then(function (d) {
      if (d && d.game) uebernehmen(d.game);
    }).catch(function (e) { melde(e.message); });
  }

  function uebernehmen(neu) {
    spiel = neu;
    meinPlatz = -1;
    (spiel.spieler || []).forEach(function (p, i) { if (p.id === ICH) meinPlatz = i; });
    // Das Spiel prüft selbst, wer ziehen darf: dafür bekommt es seinen Platz.
    if (MAEN && SP.platz) { try { SP.platz(meinPlatz >= 0 ? meinPlatz : null); } catch (e) {} }
    zeichneLeiste();
    if (spiel.status !== "active") { gestartet = false; return; }

    if (!spiel.stand) {
      /* Der Gastgeber baut die Aufstellung und schickt sie als ersten Stand. */
      if (meinPlatz === 0 && !laeuft) {
        laeuft = true;
        var namen = spiel.spieler.map(function (p) { return p.name; });
        anbindung.neu(spiel.anzahl, namen);
        anbindung.zeige();
        var stand = anbindung.stand();
        letzterStand = JSON.stringify(stand);
        gezeigt = spiel.version;
        gestartet = true;
        ruf("stand", { stand: stand, weiter: anbindung.dran(stand) }).then(function (d) {
          if (d && d.game) spiel = d.game;
        }).catch(function (e) { melde(e.message); }).then(function () { laeuft = false; });
      } else if (!gestartet) {
        melde("Aufstellung wird gebaut …");
      }
      return;
    }

    if (gezeigt !== spiel.version) {
      var text = JSON.stringify(spiel.stand);
      if (text !== letzterStand) {
        anbindung.lade(spiel.stand);
        letzterStand = text;
      }
      if (!gestartet) { anbindung.zeige(); gestartet = true; }
      gezeigt = spiel.version;
    }
  }

  /* --------------------------------------------------------------- Abgleich */
  function abholen() {
    ruf().then(function (d) { uebernehmen(d && d.game); })
      .catch(function (e) { melde("Verbindung: " + e.message); });
  }

  /* Eigenen Zug schicken, sobald sich der örtliche Stand geändert hat und ich
     am Zug bin. Der Stand ist vollständig, deshalb braucht der Server keine
     Regelkenntnis. */
  function senden() {
    if (!spiel || spiel.status !== "active" || !spiel.stand) return;
    if (meinPlatz < 0 || spiel.current !== meinPlatz) return;
    /* Erst den Serverstand übernehmen: nach dem Neuladen steht im Spiel noch
       der alte/eigene Aufbau – der darf nicht als Zug zurückgehen. */
    if (gezeigt !== spiel.version) return;
    var stand;
    try { stand = anbindung.stand(); } catch (e) { return; }
    var text = JSON.stringify(stand);
    if (text === letzterStand) return;
    letzterStand = text;
    ruf("zug", {
      stand: stand,
      weiter: anbindung.dran(stand),
      ende: anbindung.vorbei(stand)
    }).then(function (d) {
      if (d && d.game) {
        spiel = d.game;
        gezeigt = spiel.version;
        zeichneLeiste();
      }
    }).catch(function (e) {
      letzterStand = "";   // beim nächsten Takt noch einmal versuchen
      melde(e.message);
    });
  }

  /* ---------------------------------------------------------------- Start */
  document.body.insertBefore(leiste, document.body.firstChild);
  document.body.appendChild(schleier);
  zeichneLeiste();

  /* Erst den Platz im Raum sichern: spiele/token.js hängt sein Token an den
     Ruf, der Server legt den Sitz an (neu) und gibt das Token zurück. Danach
     trägt jede Anfrage das richtige Token. */
  fetch("/api/presence", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ room: RAUM, id: ICH, name: MEIN_NAME, neu: true, active: true })
  }).then(function () { return ruf("rejoin"); }).then(function (d) {
    if (d && d.game) { uebernehmen(d.game); return; }
    /* Noch keine Partie im Raum: dann eröffnen wir sie. */
    return ruf("create", { anzahl: 2 }).then(function (e) {
      if (e && e.game) uebernehmen(e.game);
    });
  }).catch(function (e) {
    melde(e.message);
    /* Ohne Partie im Raum legt der erste Spieler sie an. */
    ruf("create", { anzahl: 2 }).then(function (e) {
      if (e && e.game) uebernehmen(e.game);
    }).catch(function (e2) { melde(e2.message); });
  });

  setInterval(abholen, 1200);
  setInterval(senden, 450);
})();
