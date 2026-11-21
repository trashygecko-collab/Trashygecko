/* online.js – gemeinsame Runde für alle Würfelspiele.
 *
 * Die Seite bekommt vom Connector ?room=…&id=… . Nur dann mischt sich dieses
 * Skript ein; ohne Raum bleibt das Spiel die lokale Partie wie bisher.
 *
 * Gezogen wird immer auf dem Server: Ein Klick geht als Aktion dorthin, der
 * Server führt das Regelmodul aus (also auch die Würfel) und schickt den
 * Spielstand zurück. Gezeichnet wird er vom selben Rahmen wie lokal
 * (TGW.setzeStand), es gibt hier keine zweite Spieloberfläche.
 */
(function () {
  var frage = new URLSearchParams(location.search);
  var raum = (frage.get("room") || "").trim().toUpperCase();
  if (!raum) {
    return; // lokale Partie
  }
  var kennung = (frage.get("id") || "").trim();
  var spielId = window.TGW_SPIEL || "";
  if (!kennung || !spielId) {
    return;
  }

  var stand = null;      // letzte Antwort des Servers
  var gefragt = false;   // läuft gerade ein Ruf?
  var ersterRuf = true;  // der erste Abruf geht auch im Hintergrund-Tab raus
  var takt = null;
  var kasten = null;
  var zustandZeile = null;
  var namenZeile = null;
  var tastenZeile = null;

  /* --- Server ---------------------------------------------------------- */

  function ruf(aktion, felder) {
    var adresse = "/api/wuerfel?room=" + encodeURIComponent(raum) +
      "&id=" + encodeURIComponent(kennung);
    var optionen = null;
    if (!aktion || aktion === "stand") {
      adresse += "&action=stand";
    } else {
      adresse += "&action=" + encodeURIComponent(aktion);
      var koerper = { room: raum, id: kennung, spiel: spielId };
      Object.keys(felder || {}).forEach(function (k) { koerper[k] = felder[k]; });
      optionen = {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(koerper)
      };
    }
    return fetch(adresse, optionen)
      .then(function (antwort) {
        if (!antwort.ok) {
          return antwort.text().then(function (text) {
            return { fehler: (text || "").trim() || "Fehler " + antwort.status };
          });
        }
        return antwort.json();
      })
      .then(function (daten) {
        zeige(daten);
        return daten;
      })
      .catch(function () {
        /* Netz weg: der nächste Durchlauf holt den Stand nach. */
      });
  }

  function hole() {
    if (gefragt) {
      return;
    }
    // Im Hintergrund-Tab wird nicht gepollt – aber der erste Stand muss auch
    // dort ankommen, sonst bliebe die Seite bei „Verbinde …“ stehen.
    if (document.hidden && !ersterRuf) {
      return;
    }
    ersterRuf = false;
    gefragt = true;
    ruf("stand").then(function () { gefragt = false; }, function () { gefragt = false; });
  }

  function tue(aktion, felder) {
    return ruf(aktion, felder);
  }

  /* --- Anzeige --------------------------------------------------------- */

  function zeige(daten) {
    if (!daten) {
      return;
    }
    if (daten.fehler) {
      zustandZeile.textContent = daten.fehler;
      zustandZeile.className = "online-stand fehler";
      return;
    }
    stand = daten.game || null;
    if (stand && stand.zustand && TGW.setzeStand) {
      TGW.setzeStand(stand.zustand);
    }
    zeichne();
  }

  function knopf(text, art, felder, haupt) {
    var b = document.createElement("button");
    b.type = "button";
    b.className = "knopf" + (haupt ? " haupt" : "");
    b.textContent = text;
    b.addEventListener("click", function () {
      b.disabled = true;
      tue(art, felder).then(function () { b.disabled = false; }, function () { b.disabled = false; });
    });
    return b;
  }

  function zeichne() {
    tastenZeile.innerHTML = "";
    namenZeile.innerHTML = "";

    if (!stand) {
      zustandZeile.className = "online-stand";
      zustandZeile.textContent = "Noch keine Runde in Raum " + raum + ".";
      namenZeile.textContent = "Der Gastgeber eröffnet die Runde – oder du fängst an.";
      tastenZeile.appendChild(knopf("Runde eröffnen", "create", {}, true));
      return;
    }

    var ich = stand.du;
    var dran = ich >= 0 && stand.dran === ich && stand.status === "active";
    zustandZeile.className = "online-stand" + (dran ? " dran" : "");

    if (stand.status === "invited") {
      zustandZeile.textContent = "Warte auf Mitspieler (" + stand.spieler.length + "/" + stand.max + ")";
    } else if (stand.status === "finished") {
      zustandZeile.textContent = stand.ende || "Die Runde ist zu Ende.";
    } else if (dran) {
      zustandZeile.textContent = "Du bist am Zug";
    } else {
      var wer = stand.spieler[stand.dran];
      zustandZeile.textContent = (wer ? wer.name : "Mitspieler") + " ist am Zug";
    }

    stand.spieler.forEach(function (s, i) {
      var zeile = document.createElement("span");
      zeile.className = "online-name" + (stand.status === "active" && i === stand.dran ? " aktiv" : "");
      zeile.textContent = s.name + (i === ich ? " (du)" : "");
      namenZeile.appendChild(zeile);
    });

    if (stand.status === "invited") {
      if (ich < 0) {
        tastenZeile.appendChild(knopf("Platz nehmen", "join", {}, true));
      } else if (stand.hostId === kennung) {
        var los = knopf("Runde starten", "start", {}, true);
        los.disabled = stand.spieler.length < stand.min;
        if (los.disabled) {
          los.title = "Mindestens " + stand.min + " Spieler";
        }
        tastenZeile.appendChild(los);
      }
    }
    if (stand.status === "finished" && stand.hostId === kennung && stand.spieler.length >= stand.min) {
      tastenZeile.appendChild(knopf("Neue Runde", "neu", {}, true));
    }
    if (ich >= 0) {
      tastenZeile.appendChild(knopf("Verlassen", "verlassen", {}, false));
    }
  }

  /* --- Aufbau ---------------------------------------------------------- */

  function baue() {
    var rahmen = document.querySelector(".rahmen");
    if (!rahmen) {
      return;
    }
    kasten = document.createElement("section");
    kasten.className = "tafel online";

    var kopf = document.createElement("div");
    kopf.className = "online-kopf";
    var marke = document.createElement("span");
    marke.className = "online-marke";
    marke.textContent = "Online · Raum " + raum;
    var titel = document.createElement("span");
    titel.className = "online-titel";
    titel.textContent = window.TGW && TGW.spiele && TGW.spiele[spielId] ? TGW.spiele[spielId].titel : "";
    kopf.append(marke, titel);

    zustandZeile = document.createElement("div");
    zustandZeile.className = "online-stand";
    zustandZeile.textContent = "Verbinde …";

    namenZeile = document.createElement("div");
    namenZeile.className = "online-namen";

    tastenZeile = document.createElement("div");
    tastenZeile.className = "tasten";

    kasten.append(kopf, zustandZeile, namenZeile, tastenZeile);

    var einrichten = rahmen.querySelector(".einrichten");
    if (einrichten) {
      rahmen.insertBefore(kasten, einrichten);
    } else {
      rahmen.insertBefore(kasten, rahmen.firstChild);
    }

    // Lebenszeichen: hält Platz und Partie im Raum am Leben (dieselbe
    // Anbindung wie bei den anderen Online-Würfelspielen).
    if (window.WG && WG.leben) {
      WG.leben({ room: raum, id: kennung, name: (frage.get("name") || "").trim() });
    }

    hole();
    takt = setInterval(hole, 900);
    document.addEventListener("visibilitychange", function () {
      if (!document.hidden) {
        hole();
      }
    });
  }

  // Der Rahmen baut seine Seite bei DOMContentLoaded; dieser Aufsatz läuft
  // danach, damit Kasten und Spieltafel beide stehen.
  TGW.online = {
    aktion: function (art, daten) {
      if (art === "neu") {
        tue("neu");
        return;
      }
      tue("aktion", { art: art, daten: daten || null });
    }
  };

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", baue);
  } else {
    baue();
  }
})();
