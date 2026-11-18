/* --- Spielerkennung: Token mitschicken -------------------------------------
   Beim Beitritt gibt der Server dem Gerät ein zufälliges Token. Es steht im
   lokalen Speicher des Browsers (je Raum und Kennung) und wird hier an jeden
   API-Aufruf angehängt. Ohne das Token lehnt der Server Änderungen an einem
   fremden Platz ab – eine ausgelesene Kennung allein genügt niemandem mehr.

   Das Skript kapselt window.fetch, damit keine der vielen Spielseiten einzeln
   angepasst werden muss. Es läuft vor allen anderen Skripten der Seite. */
(function () {
  var original = window.fetch;
  if (!original || window.fetchMitToken) {
    return;
  }

  function schluessel(raum, id) {
    return "tg-token:" + raum + ":" + id;
  }

  function tokenFuer(raum, id) {
    if (!raum || !id) {
      return "";
    }
    try {
      return localStorage.getItem(schluessel(raum, id)) || "";
    } catch (e) {
      return "";
    }
  }

  function merkeToken(raum, id, token) {
    if (!raum || !id || !token) {
      return;
    }
    try {
      localStorage.setItem(schluessel(raum, id), token);
    } catch (e) {
      /* privater Modus: dann eben ohne Token – der Platz bleibt gesperrt */
    }
  }

  // Raum und Kennung stehen je nach Aufruf in der Adresse oder im Rumpf.
  function ausAnfrage(adresse, optionen) {
    var raum = "";
    var id = "";
    var frage = adresse.indexOf("?") >= 0 ? adresse.slice(adresse.indexOf("?") + 1) : "";
    if (frage) {
      var teile = frage.split("&");
      for (var i = 0; i < teile.length; i++) {
        var paar = teile[i].split("=");
        var name = decodeURIComponent(paar[0] || "");
        var wert = decodeURIComponent((paar[1] || "").replace(/\+/g, " "));
        if (name === "room") {
          raum = wert;
        } else if (name === "id") {
          id = wert;
        }
      }
    }
    if (optionen && typeof optionen.body === "string") {
      try {
        var rumpf = JSON.parse(optionen.body);
        if (rumpf && typeof rumpf === "object") {
          raum = rumpf.room || raum;
          id = rumpf.id || id;
        }
      } catch (e) {
        /* kein JSON-Rumpf (etwa Formulardaten) – dann bleibt es bei der Adresse */
      }
    }
    return { raum: raum, id: id };
  }

  /* Ein frisches Geheimnis, bevor der erste Beitritt hinausgeht. Ohne das
     legte der Server den Platz mit einem eigenen Token an; ein zweiter,
     fast gleichzeitiger Ruf (Lebenszeichen direkt nach dem Beitritt) kannte
     es noch nicht und lief in eine 403. */
  function neuesToken() {
    var text = "";
    try {
      var puffer = new Uint8Array(16);
      (window.crypto || window.msCrypto).getRandomValues(puffer);
      for (var i = 0; i < puffer.length; i++) {
        text += ("0" + puffer[i].toString(16)).slice(-2);
      }
    } catch (e) {
      text = "";
    }
    if (text.length < 16) {
      text = String(Date.now()) + Math.random().toString(36).slice(2);
    }
    return text;
  }

  function mitToken(adresse, optionen) {
    var teile = ausAnfrage(adresse, optionen);
    var token = tokenFuer(teile.raum, teile.id);
    if (!token && adresse.indexOf("/api/presence") >= 0 && teile.raum && teile.id) {
      // Vor dem ersten Ruf erzeugen und merken: beide Rufe schicken dasselbe.
      token = neuesToken();
      merkeToken(teile.raum, teile.id, token);
    }
    if (!token || /[?&]token=/.test(adresse)) {
      return adresse;
    }
    return adresse + (adresse.indexOf("?") >= 0 ? "&" : "?") + "token=" + encodeURIComponent(token);
  }

  window.fetchMitToken = function (adresse, optionen) {
    if (typeof adresse !== "string" || adresse.indexOf("/api/") < 0) {
      return original(adresse, optionen);
    }
    var ziel = mitToken(adresse, optionen);
    var laeuft = original(ziel, optionen);
    if (ziel.indexOf("/api/presence") >= 0 && laeuft && laeuft.then) {
      // Der Beitritt liefert das Token genau einmal für dieses Gerät.
      laeuft
        .then(function (antwort) {
          if (!antwort || !antwort.ok || !antwort.clone) {
            return;
          }
          antwort
            .clone()
            .json()
            .then(function (daten) {
              if (!daten || !daten.token) {
                return;
              }
              var teile = ausAnfrage(adresse, optionen);
              merkeToken(teile.raum || daten.room, teile.id, daten.token);
            })
            .catch(function () {});
        })
        .catch(function () {});
    }
    return laeuft;
  };

  window.fetch = function (adresse, optionen) {
    return window.fetchMitToken(adresse, optionen);
  };
})();
