/* ===========================================================================
   HALMA – Lernsystem

   Kurze Hinweise, die zur gewählten Variante passen. Sie erklären nur, was
   nach dem Regelwerk dieser Variante wirklich gilt.
   =========================================================================== */
(function (HB) {
  "use strict";

  function name(varianteId) {
    var v = HB.Regeln.VARIANTEN[varianteId];
    return v ? v.name : "Halma";
  }

  /* Was kann die Figur auf diesem Feld? */
  function auswahlText(brett, belegung, varianteId, feld, platz) {
    var zuege = HB.Regeln.zuegeVon(brett, belegung, varianteId, feld);
    var schritte = 0, spruenge = 0, ketten = 0, weit = 0;
    zuege.forEach(function (z) {
      if (z.art === "sprung") {
        spruenge++;
        if (z.weg.length > 1) ketten++;
        if (z.weg.length && brett.art !== "stern") {
          var a = brett.felder[z.von], b = brett.felder[z.nach];
          if (Math.max(Math.abs(a.x - b.x), Math.abs(a.y - b.y)) > 2) weit++;
        }
      } else {
        schritte++;
      }
    });
    if (!schritte && !spruenge) {
      return "Diese Figur ist eingesperrt – sie kann in diesem Zug nicht ziehen.";
    }
    var teile = [];
    if (schritte) teile.push(schritte + (schritte === 1 ? " Schritt" : " Schritte"));
    if (spruenge) teile.push(spruenge + (spruenge === 1 ? " Sprungziel" : " Sprungziele"));
    var text = "Diese Figur kann " + teile.join(" und ") + " erreichen.";
    if (ketten) text += " Eine Kette mit " + (ketten === 1 ? "mehreren Sprüngen" : "mehreren Sprüngen") +
      " ist erlaubt – dieselbe Figur springt mehrfach.";
    if (weit) text += " Auch ein weiter Sprung über eine Figur mit Abstand ist möglich.";
    if (platz >= 0 && HB.Regeln.imZiel(brett, belegung, platz) > 0) {
      text += " Figuren dürfen aus dem Zielbereich wieder herausziehen.";
    }
    return text;
  }

  /* Warum ging dieser Klick nicht? */
  function fehlerText(grund, varianteId, brett) {
    switch (grund) {
      case "besetzt":
        return "Dieses Feld ist besetzt – Figuren dürfen nur auf freien Feldern landen.";
      case "fremd":
        return "Das ist keine eigene Figur.";
      case "zuweit":
        return varianteId === "sprung"
          ? "So weit geht kein Zug: Beim weiten Sprung muss genau eine Figur auf der geraden " +
            "Linie liegen, und das Zielfeld liegt im selben Abstand dahinter."
          : "So weit kommt eine Figur nicht. Springen geht nur über eine direkt benachbarte Figur, " +
            "und dahinter muss ein freies Feld sein.";
      case "doppelt":
        return "Über zwei Figuren direkt hintereinander wird nicht gesprungen – " +
          "dazwischen muss ein freies Feld liegen.";
      case "dran":
        return "Jetzt ist ein anderer Spieler am Zug.";
      case "besetztstart":
        return "Auf dem Startfeld steht deine Figur – wähle sie zuerst aus.";
      default:
        return "";
    }
  }

  /* Erklärung der Lage (nach einem Zug, dezent). */
  function lageText(brett, belegung, platz, figurZahl) {
    var drin = HB.Regeln.imZiel(brett, belegung, platz);
    var fehlt = figurZahl - drin;
    if (fehlt <= 0) return "Alle Figuren sind im Ziel.";
    var text = drin + " von " + figurZahl + " Figuren im Ziel";
    if (drin === figurZahl - 1) text += " – nur noch eine fehlt.";
    else text += " – " + fehlt + " fehlen noch.";
    return text;
  }

  function zugText(brett, zug) {
    if (!zug) return "";
    if (zug.art === "sprung") {
      return "Sprung mit " + (zug.weg.length + 1) + " Feld(ern) Weite" +
        (zug.weg.length > 1 ? " – Kette über " + zug.weg.length + " Figuren" : "") + ".";
    }
    return "Ein Schritt auf das Nachbarfeld.";
  }

  HB.Lernen = {
    name: name,
    auswahlText: auswahlText,
    fehlerText: fehlerText,
    lageText: lageText,
    zugText: zugText
  };
})(window.HB = window.HB || {});
