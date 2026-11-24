/* Online-Spielhalle: Die Multiplayer-Zentrale ist der Connector.
 *
 * Früher baute diese Datei auf der Spieleübersicht eine eigene Halle auf
 * (Räume, Chat, Angebote, Tischfenster). Damit gab es zwei halbe Hallen
 * nebeneinander. Jetzt bleibt die Übersicht eine Übersicht: Ein Online-Spiel
 * schickt den Spieler in den Connector, und zwar mit der Kennung des Spiels
 * in der Adresse (?spiel=…). Dort stehen Räume, Einladungen, Chat, Namen,
 * Grafikstil und alle Partien.
 *
 * Örtlichkeiten:
 *   GitHub Pages  …/Trashygecko/spiele/connector/connector.html
 *   Render        …/spiele/connector/connector.html (der Server liefert die
 *                 Adresse selbst aus, obwohl der Ordner im Abbild fehlt)
 *   APK / file:// kein Connector erreichbar → das Spiel öffnet sich direkt.
 */
(function () {
  "use strict";

  var CONNECTOR = "connector/connector.html";

  // Im APK (file://) gibt es keinen Server: dort bleibt der direkte Weg.
  function hatConnector() {
    return location.protocol === "http:" || location.protocol === "https:";
  }

  function kennungFuer(pfad) {
    var liste = (window.TGMultiplayer && window.TGMultiplayer.spiele) || [];
    for (var i = 0; i < liste.length; i++) {
      if (liste[i] && liste[i].pfad === pfad) return liste[i].id;
    }
    // Zweiter Versuch über den Ordnernamen (…/risiko/risiko.html → risiko).
    var teile = String(pfad || "").split("/");
    return teile.length > 1 ? teile[teile.length - 2] : "";
  }

  function ziel(pfad) {
    if (!hatConnector()) return "";
    var id = kennungFuer(pfad);
    return CONNECTOR + (id ? "?spiel=" + encodeURIComponent(id) : "");
  }

  window.TGHalle = {
    bereit: true,
    ziel: ziel,

    /* Wird von einer Spielekachel gerufen, sobald das Spiel auch online geht. */
    auswahl: function (pfad) {
      var adresse = ziel(pfad);
      if (!adresse) {
        // Ohne Connector (App) direkt ins Spiel – nie in eine leere Halle.
        location.href = pfad;
        return;
      }
      location.href = adresse;
    },

    /* Für Verweise von außen. */
    oeffnen: function (pfad) {
      var adresse = ziel(pfad || "");
      if (adresse) location.href = adresse;
    }
  };

  // Kopfzeile: der Weg in die Spielzentrale bleibt einen Klick entfernt.
  function knopfEinhaengen() {
    if (!hatConnector()) return; // Im APK gibt es keinen Connector-Link.
    var aktionen = document.querySelector(".kopf-aktionen");
    if (!aktionen || document.getElementById("online-knopf")) return;
    var a = document.createElement("a");
    a.id = "online-knopf";
    a.className = "online-knopf";
    a.href = CONNECTOR;
    a.setAttribute("title", "Räume, Einladungen, Chat und alle Partien");
    var punkt = document.createElement("span");
    punkt.className = "punkt";
    var wort = document.createElement("span");
    wort.className = "wort";
    wort.textContent = "Online spielen";
    a.appendChild(punkt);
    a.appendChild(wort);
    aktionen.insertBefore(a, aktionen.firstChild);
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", knopfEinhaengen);
  } else {
    knopfEinhaengen();
  }
})();
