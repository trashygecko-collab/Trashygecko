/* ===========================================================================
   HALMA – Knobelmodus

   Jede Aufgabe ist eine geprüfte Stellung: Die Mindestzahl der Züge wurde mit
   dem Tiefensucher (HB.Raetsel.solver) berechnet und steht als `min` in der
   Aufgabe. Werkzeug: arbeit/werkzeuge/halma_pruefen.js prüft beim Lauf, dass
   die Aufgaben lösbar sind und `min` exakt stimmt.
   =========================================================================== */
(function (HB) {
  "use strict";

  /* figuren: [x, y, platz] – platz 0 ist der Spieler, platz 1 sind feste
     Steine (sie bewegen sich nicht), sie dienen als Trittsteine oder Sperren. */
  var AUFGABEN = [
    {
      id: "sprung",
      titel: "Der erste Sprung",
      text: "Deine Figur steht zwei Felder vor dem Zielbereich. Ein fremder Stein " +
            "steht direkt auf dem Weg – springe über ihn hinein.",
      variante: "klassisch", figuren: [[12, 12, 0], [13, 13, 1]],
      maxZuege: 1, stufe: "leicht", min: 1,
      loesung: [{ von: "12:12", nach: "14:14" }],
      tipp: "Springen geht über eine Figur, wenn das Feld dahinter frei ist."
    },
    {
      id: "kette",
      titel: "Die Kette",
      text: "Drei Steine stehen in einer Diagonale. Mit einer einzigen Figur " +
            "kannst du in einem Zug dreimal springen – bis in den Zielbereich.",
      variante: "klassisch",
      figuren: [[8, 8, 0], [9, 9, 1], [11, 11, 1], [13, 13, 1]],
      maxZuege: 1, stufe: "mittel",
      min: 1,
      loesung: [{ von: "8:8", nach: "14:14" }],
      tipp: "Nach jedem Sprung darfst du mit derselben Figur weiterspringen."
    },
    {
      id: "umweg",
      titel: "Zwei Sprünge, ein Zug",
      text: "Zwei Steine stehen in der Diagonalen, dazwischen ist Luft. " +
            "Mit einem einzigen Zug und zwei Sprüngen erreichst du den Zielbereich.",
      variante: "klassisch", figuren: [[10, 10, 0], [11, 11, 1], [13, 13, 1]],
      maxZuege: 3, stufe: "mittel",
      min: 1,
      loesung: null,
      tipp: "Ein Schritt zur Seite öffnet oft den Sprung in die andere Richtung."
    },
    {
      id: "weitsprung",
      titel: "Der weite Sprung",
      text: "Sprunghalma: Der Stein steht weit entfernt auf der Diagonalen. " +
            "Ein Sprung darüber landet im Zielbereich – wenn der Abstand stimmt.",
      variante: "sprung", figuren: [[5, 5, 0], [10, 10, 1]],
      maxZuege: 1, stufe: "mittel", min: 1, loesung: null,
      tipp: "Beim weiten Sprung liegt das Zielfeld genau so weit hinter dem Stein wie er vor dir."
    },
    {
      id: "zwei",
      titel: "Zwei Figuren",
      text: "Bringe beide Figuren in den Zielbereich – in genau drei Zügen. " +
            "Eine Figur kann der anderen als Trittstein dienen.",
      variante: "klassisch",
      figuren: [[11, 11, 0], [11, 12, 0], [10, 10, 1]],
      maxZuege: 3, stufe: "schwer", min: 3, loesung: null,
      tipp: "Die vordere Figur rückt einen Schritt vor, die hintere springt darüber, dann zieht die vordere nach."
    },
    {
      id: "enge",
      titel: "Enge Pforte",
      text: "Zwei Steine lassen nur eine schmale Lücke. Komme mit zwei Zügen " +
            "in den Zielbereich – ohne auf einem besetzten Feld zu landen.",
      variante: "klassisch",
      figuren: [[10, 11, 0], [11, 12, 1], [12, 13, 1], [13, 14, 1]],
      maxZuege: 2, stufe: "schwer", min: 2, loesung: null,
      tipp: "Landen darfst du nur auf freien Feldern – auch nach einem Sprung."
    }
  ];

  /* Klassische Aufgaben spielen im kleinen Eckhaus (13 Felder, wie im
     Viererspiel) – das ist der engste Zielbereich, den die Regeln hergeben.
     Die Texte („zwei Felder vor dem Zielbereich“) passen dazu. */
  function brettFuer(aufgabe) {
    return aufgabe.variante === "stern" ? HB.Brett.stern() : HB.Brett.klassisch(4);
  }

  function feldIndex(brett, x, y) {
    return brett.art === "stern" ? brett.id(x, y) : (y * 16 + x);
  }

  function aufstellung(aufgabe, brett) {
    var b = HB.Regeln.neueBelegung(brett.felder.length);
    aufgabe.figuren.forEach(function (f) {
      var i = feldIndex(brett, f[0], f[1]);
      if (i >= 0) b[i] = f[2];
    });
    return b;
  }

  /* Erfüllt? Alle Figuren von Platz 0 stehen im Zielbereich. */
  function erfuellt(aufgabe, brett, belegung) {
    var zone = HB.Regeln.zielZone(brett, 0);
    var eigene = 0, drin = 0;
    for (var i = 0; i < belegung.length; i++) {
      if (belegung[i] === 0) {
        eigene++;
        if (zone.indexOf(i) >= 0) drin++;
      }
    }
    return eigene > 0 && drin === eigene;
  }

  /* Tiefensuche (breit zuerst) über ganze Züge: kleinste Zugzahl + Lösung.
     Nur Platz 0 zieht; Platz 1 steht fest. Rückgabe { zuege, loesung } oder null. */
  function solver(aufgabe, maxTiefe) {
    var brett = brettFuer(aufgabe);
    var start = aufstellung(aufgabe, brett);
    var grenze = Math.min(maxTiefe || aufgabe.maxZuege || 4, 6);
    var gesehen = {};
    var rand = [{ b: start, weg: [] }];
    gesehen[start.join(",")] = true;
    for (var tiefe = 1; tiefe <= grenze; tiefe++) {
      var naechste = [];
      for (var n = 0; n < rand.length; n++) {
        var knoten = rand[n];
        var zuege = HB.Regeln.alleZuege(brett, knoten.b, aufgabe.variante, 0);
        for (var z = 0; z < zuege.length; z++) {
          var zug = zuege[z];
          var nach = HB.Regeln.ziehe(knoten.b, zug);
          var weg = knoten.weg.concat([zug]);
          if (erfuellt(aufgabe, brett, nach)) return { zuege: tiefe, loesung: weg };
          var k = nach.join(",");
          if (gesehen[k]) continue;
          gesehen[k] = true;
          naechste.push({ b: nach, weg: weg });
        }
      }
      rand = naechste;
      if (rand.length > 400000) break;      // Sicherheitsnetz für sehr offene Stellungen
    }
    return null;
  }

  /* Ein sinnvoller Tipp: erster Zug der gefundenen Lösung (flach gesucht). */
  function tippZug(aufgabe) {
    var r = solver(aufgabe, Math.min(aufgabe.maxZuege || 3, 3));
    return r && r.loesung.length ? r.loesung[0] : null;
  }

  HB.Raetsel = {
    AUFGABEN: AUFGABEN,
    brettFuer: brettFuer,
    aufstellung: aufstellung,
    erfuellt: erfuellt,
    solver: solver,
    tippZug: tippZug,
    nachId: function (id) {
      for (var i = 0; i < AUFGABEN.length; i++) if (AUFGABEN[i].id === id) return AUFGABEN[i];
      return null;
    }
  };
})(window.HB = window.HB || {});
