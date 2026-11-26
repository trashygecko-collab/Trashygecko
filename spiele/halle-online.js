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

  /* Auf GitHub Pages liegt nur die statische Seite – Räume, Chat und Partien
     brauchen den Go-Server. Dort gilt deshalb die laufende Adresse. */
  var CONNECTOR_WEB = "https://trashygecko.onrender.com/spiele/connector/connector.html";

  function connectorBasis() {
    return /\.github\.io$/.test(location.hostname) ? CONNECTOR_WEB : CONNECTOR;
  }

  // Im APK (file://) gibt es keinen Server: dort bleibt der direkte Weg.
  function hatConnector() {
    return location.protocol === "http:" || location.protocol === "https:";
  }

  // Pfade vergleichen: die Spielhalle schreibt "brettspiele/risiko/risiko.html",
  // das Register "/spiele/brettspiele/risiko/risiko.html". Beides wird auf
  // dieselbe Form gebracht, sonst findet keine Kachel ihr Spiel.
  function norm(pfad) {
    return String(pfad || "").replace(/^\.?\//, "").replace(/^spiele\//, "");
  }

  function eintragFuer(pfad) {
    var gesucht = norm(pfad);
    if (!gesucht) return null;
    var liste = (window.TGMultiplayer && window.TGMultiplayer.spiele) || [];
    for (var i = 0; i < liste.length; i++) {
      if (liste[i] && norm(liste[i].pfad) === gesucht) return liste[i];
    }
    return null;
  }

  function kennungFuer(pfad) {
    var eintrag = eintragFuer(pfad);
    return eintrag ? eintrag.id : "";
  }

  function ziel(pfad) {
    if (!hatConnector()) return "";
    var id = kennungFuer(pfad);
    return connectorBasis() + (id ? "?spiel=" + encodeURIComponent(id) : "");
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

  /* Jede Kachel, deren Spiel auch online geht, führt in den Connector: dort
     stehen Raum, Einladung, Chat und alle Spielmodi. Die vorgerenderte
     Spielhalle hatte diesen Weg verloren – jeder Klick landete direkt im
     Spiel, ohne Tisch und ohne Einladung. Einzelspieler-Spiele (nicht im
     Register) bleiben beim direkten Weg, in der App (file://) ebenfalls. */
  function kachelKlick(e) {
    if (e.defaultPrevented || e.button !== 0) return;
    // Neuer Tab / neues Fenster bleibt dem Browser überlassen.
    if (e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) return;
    var a = e.target && e.target.closest ? e.target.closest("a.hl, a.karte") : null;
    if (!a) return;
    var pfad = a.getAttribute("href") || "";
    if (!pfad || pfad.charAt(0) === "#") return;
    if (!eintragFuer(pfad)) return;
    var adresse = ziel(pfad);
    if (!adresse) return;
    e.preventDefault();
    location.href = adresse;
  }
  document.addEventListener("click", kachelKlick);

  /* Auch das Ziel der Kachel selbst führt in den Connector. So stimmt schon
     die Adresse in der Statuszeile, und „in neuem Tab öffnen“ landet nicht
     wieder in einer Partie ohne Tisch. */
  function kachelnRichten() {
    var kacheln = document.querySelectorAll("a.hl, a.karte");
    for (var i = 0; i < kacheln.length; i++) {
      var pfad = kacheln[i].getAttribute("href") || "";
      if (!pfad || pfad.charAt(0) === "#") continue;
      if (!eintragFuer(pfad)) continue;
      var adresse = ziel(pfad);
      if (adresse) kacheln[i].setAttribute("href", adresse);
    }
  }

  // Kopfzeile: der Weg in die Spielzentrale bleibt einen Klick entfernt.
  function knopfEinhaengen() {
    if (!hatConnector()) return; // Im APK gibt es keinen Connector-Link.
    var aktionen = document.querySelector(".kopf-aktionen");
    if (!aktionen || document.getElementById("online-knopf")) return;
    var a = document.createElement("a");
    a.id = "online-knopf";
    a.className = "online-knopf";
    a.href = connectorBasis();
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
    document.addEventListener("DOMContentLoaded", function () { kachelnRichten(); knopfEinhaengen(); });
  } else {
    kachelnRichten();
    knopfEinhaengen();
  }
})();
