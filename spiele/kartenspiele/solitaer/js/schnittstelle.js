/* =========================================================================
   SOLITÄR – Öffentliche Spielschnittstelle 2.0

   Ein globaler Name: window.Solitaer. Alles andere bleibt intern (TG.Shell,
   TG.MODI, TG.Karten sind gemeinsame Helfer des Spiels).

   Aktionen für Konnektor, Sammlung, Bots und Tests – alle über dieselben
   internen Funktionen wie Knopf, Taste und Finger:

     start      {}          beginnt eine Partie mit der aktuellen Wahl
     modus      {modus}     Variante am Startbildschirm wählen und beginnen
     neu        {}          neue Partie, frisch gemischt (Alias: mischen)
     mischen    {}          dasselbe wie neu
     pause      {}          anhalten (nur im Lauf)
     weiter     {}          fortsetzen (nur in der Pause)
     tipp       {}          einen Tipp zeigen (nur im Lauf)
     zurueck    {}          letzten Zug zurücknehmen (nur mit Verlauf)
     regeln     {}          Regelhilfe zeigen
     statistik  {}          Statistik zeigen

   start({stil}) stellt AUSSCHLIESSLICH die Welt ein und lässt den
   Startbildschirm stehen – so übergibt der Konnektor die Stilwahl beim
   Öffnen und bei jedem Wechsel, ohne die Partie neu zu beginnen. Der Stil
   von außen hat Vorrang: Die eigene Stilwahl verschwindet und bleibt auch
   nach startbildschirm() verborgen.

   Abweichung von Abschnitt 3.2 (dokumentiert in SPIELE-SCHNITTSTELLE.md):
   Eine unbekannte Variante antwortet mit `unbekannter-modus` statt
   `unbekannte-stufe` – Solitär hat keine Schwierigkeitsstufen.
   ========================================================================= */
(function () {
  "use strict";

  var VERSION = "2.0.0";
  var STILE = ["normal", "maedchen", "buben"];
  var bereit = false;
  var vonAussen = false;          // Stil kam vom Konnektor
  var warten = [];

  function shell() {
    return window.TG && window.TG.Shell ? window.TG.Shell : null;
  }

  function eigeneWahl() {
    var wert = "";
    try { wert = String(localStorage.getItem("solitaer-stil") || ""); } catch (e) { wert = ""; }
    if (STILE.indexOf(wert) >= 0) return wert;
    if (window.TGStil && typeof window.TGStil.hole === "function") return window.TGStil.hole();
    return "normal";
  }

  function fehler(code, feld) {
    var a = { ok: false, fehler: code };
    if (feld) a.feld = feld;
    return a;
  }

  function hat(o, k) {
    return Object.prototype.hasOwnProperty.call(o, k);
  }

  function istObjekt(w) {
    return w !== null && typeof w === "object" && !Array.isArray(w);
  }

  function zuege() {
    var s = shell();
    if (!s || !s.spiel || !s.spiel.info) return 0;
    try { return s.spiel.info().zuege || 0; } catch (e) { return 0; }
  }

  function punkte() {
    var s = shell();
    if (!s || !s.spiel || !s.spiel.info) return 0;
    try { return s.spiel.info().punkte || 0; } catch (e) { return 0; }
  }

  function ereignis(was, detail) {
    try {
      window.dispatchEvent(new CustomEvent("solitaer:" + was, { detail: detail }));
    } catch (e) { /* sehr alter Browser: Ereignis entfällt, Aufrufe bleiben */ }
  }

  function stilApi() {
    var s = shell();
    return s ? s.stilApi() : "normal";
  }

  function spielOffen() {
    var s = shell();
    return !!(s && s.spiel && !s.beendet && !s.amStart());
  }

  /* --------------------------------------------------------------- start() */
  function start(konfig) {
    if (!bereit) return fehler("spiel-nicht-bereit");
    if (konfig === undefined || konfig === null) konfig = {};
    else if (!istObjekt(konfig)) return fehler("konfiguration-muss-objekt-sein");

    var erlaubt = ["stil", "modus"];
    for (var f in konfig) {
      if (hat(konfig, f) && erlaubt.indexOf(f) < 0) return fehler("unbekanntes-feld", f);
    }

    var mitStil = hat(konfig, "stil") && konfig.stil !== null && konfig.stil !== undefined;
    if (mitStil) {
      vonAussen = true;                    // auch bei unbekanntem Wert: keine Doppelwahl
      shell().stilSetzen(typeof konfig.stil === "string" ? konfig.stil : "normal", false);
      shell().stilWahlZeigen(false);
    }

    var mitModus = hat(konfig, "modus") && konfig.modus !== null && konfig.modus !== undefined;
    if (!mitModus && mitStil) {
      /* Reine Stilübergabe: Welt einrichten, Partie und Startbildschirm bleiben. */
      return { ok: true, stil: stilApi(), modus: shell().modus || null };
    }

    var modus = mitModus ? konfig.modus : (shell().modus || Object.keys(window.TG.MODI)[0]);
    if (typeof modus !== "string" || !Object.prototype.hasOwnProperty.call(window.TG.MODI, modus)) {
      return fehler("unbekannter-modus", "modus");
    }

    if (spielOffen()) return fehler("spiel-laeuft");

    shell().starten(modus, {});
    return { ok: true, modus: modus, stil: stilApi(), modi: Object.keys(window.TG.MODI) };
  }

  /* ----------------------------------------------------------- pause/resume */
  function pause() {
    if (!bereit) return fehler("spiel-nicht-bereit");
    if (!spielOffen()) return fehler("spiel-laeuft-nicht");
    var vorher = shell().pausiert;
    shell().pausieren();
    if (!vorher && shell().pausiert) ereignis("pausiert", { tick: zuege() });
    return { ok: true, pausiert: !!shell().pausiert };
  }

  function resume() {
    if (!bereit) return fehler("spiel-nicht-bereit");
    if (!spielOffen()) return fehler("spiel-laeuft-nicht");
    var vorher = shell().pausiert;
    shell().fortsetzen();
    if (vorher && !shell().pausiert) ereignis("weiter", { tick: zuege() });
    return { ok: true, pausiert: !!shell().pausiert };
  }

  /* ------------------------------------------------------------- Aktionen */
  function aktion(name, daten) {
    if (!bereit) return fehler("spiel-nicht-bereit");
    if (typeof name !== "string") return fehler("unbekannte-aktion");
    if (daten !== undefined && daten !== null && !istObjekt(daten)) return fehler("daten-muessen-objekt-sein");
    var d = daten || {};
    var s = shell();

    switch (name) {
      case "start":
        return start(d);
      case "modus":
        if (typeof d.modus !== "string") return fehler("unbekannter-modus", "modus");
        return start({ modus: d.modus });
      case "pause":
        return pause();
      case "weiter":
        return resume();
      case "neu":
      case "mischen":
        if (!spielOffen() || s.pausiert) return fehler("aktion-nicht-erlaubt");
        s.neu();
        return { ok: true };
      case "tipp":
        if (!spielOffen() || s.pausiert) return fehler("aktion-nicht-erlaubt");
        s.tippen();
        return { ok: true };
      case "zurueck":
        if (!spielOffen() || s.pausiert) return fehler("aktion-nicht-erlaubt");
        if (!s.verlauf.length) return fehler("aktion-nicht-erlaubt");
        s.zurueck();
        return { ok: true };
      case "regeln":
        if (!s.modus || s.amStart()) return fehler("aktion-nicht-erlaubt");
        s.regelnZeigen();
        return { ok: true };
      case "statistik":
        s.statistikZeigen();
        return { ok: true };
      default:
        return fehler("unbekannte-aktion");
    }
  }

  /* --------------------------------------------------------------- Zustand */
  function zustand() {
    var s = shell();
    var offen = spielOffen();
    /* Auch nach dem Ergebnis bleiben Züge, Punkte und Zeit stehen – der
       Vergleichswert einer ausgespielten Partie geht nicht verloren. */
    var imSpiel = !!(s && s.spiel && !s.amStart());
    return {
      laeuft: offen,
      pausiert: !!(s && s.pausiert),
      tick: imSpiel ? zuege() : 0,
      richtung: { x: 0, y: 0 },
      level: 1,
      punkte: imSpiel ? punkte() : 0,
      extras: [],
      startbildschirm: !!(s && s.amStart()),
      modus: s ? s.modus : null,
      modi: Object.keys(window.TG.MODI || {}),
      zuege: imSpiel ? zuege() : 0,
      zeit: s ? s.sekunden : 0,
      stil: stilApi(),
      beendet: !!(s && s.beendet)
    };
  }

  function einstellungen() {
    var s = shell();
    return {
      modi: Object.keys(window.TG.MODI || {}),
      modus: s ? s.modus : null,
      stile: STILE.slice(),
      stil: stilApi(),
      stilWahl: !vonAussen
    };
  }

  /* ------------------------------------------------------------ Fertig/Frei */
  function fertig() {
    if (bereit) return;
    bereit = true;
    API.bereit = true;
    ereignis("bereit", { version: VERSION });
    var liste = warten.splice(0);
    liste.forEach(function (fn) { try { fn(); } catch (e) { } });
    Object.freeze(API);
  }

  var API = {
    version: VERSION,
    bereit: false,
    whenReady: function () {
      return new Promise(function (ok) {
        if (bereit) { ok({ ok: true, version: VERSION }); return; }
        warten.push(function () { ok({ ok: true, version: VERSION }); });
      });
    },
    start: start,
    einstellungen: einstellungen,
    zustand: zustand,
    startbildschirm: function () {
      if (!bereit) return fehler("spiel-nicht-bereit");
      shell().menue();
      return { ok: true };
    },
    stop: function () {
      if (!bereit) return fehler("spiel-nicht-bereit");
      shell().menue();
      return { ok: true };
    },
    pause: pause,
    resume: resume,
    aktion: aktion
  };

  window.Solitaer = API;

  /* ------------------------------------------------------------- Anschluss */
  function anmelden() {
    var s = shell();
    if (!s) { fertig(); return; }

    /* Ergebnis nach außen, bevor das Ergebnisfenster aufgeht. */
    s.beiEnde = function (grund, info) {
      ereignis("ende", {
        grund: grund,
        punkte: (info && info.punkte) || 0,
        level: 1,
        levelAnzahl: 1,
        tick: (info && info.zuege) || 0,
        modus: s.modus,
        zuege: (info && info.zuege) || 0,
        zeit: s.sekunden,
        stil: s.stilApi()
      });
    };

    /* Die eigene Wahl gilt beim direkten Start; ein Stil von außen kommt
       gleich danach über start({stil}) und hat Vorrang. */
    s.stilSetzen(eigeneWahl(), false);
    s.stilWahlZeigen(true);

    fertig();
  }

  if (document.readyState === "loading") document.addEventListener("DOMContentLoaded", anmelden);
  else anmelden();
})();
