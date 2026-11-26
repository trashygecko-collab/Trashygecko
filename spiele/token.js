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

  // Rueckfall im Speicher: im privaten Modus oder bei gesperrtem Speicher
  // liefert localStorage nichts. Ohne diesen Merker erzeugte jeder Ruf ein
  // neues Token - der Server sah dann bei jedem Lebenszeichen ein "anderes
  // Geraet" und antwortete mit 403.
  var merker = {};

  function tokenFuer(raum, id) {
    if (!raum || !id) {
      return "";
    }
    var schluesselWert = schluessel(raum, id);
    try {
      var gespeichert = localStorage.getItem(schluesselWert);
      if (gespeichert) {
        merker[schluesselWert] = gespeichert;
        return gespeichert;
      }
    } catch (e) {
      /* kein Speicher - dann gilt der Merker */
    }
    return merker[schluesselWert] || "";
  }

  function merkeToken(raum, id, token) {
    if (!raum || !id || !token) {
      return;
    }
    var schluesselWert = schluessel(raum, id);
    merker[schluesselWert] = token;
    try {
      localStorage.setItem(schluesselWert, token);
    } catch (e) {
      /* privater Modus: der Merker haelt das Token fuer diese Sitzung */
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

  /* Gerätekennung: bleibt im Browser und überlebt ein verlorenes Token. Der
     Server gibt damit denselben Platz zurück (zehn Minuten), statt einen
     neuen mit Nummernanhang anzulegen. Sie ist ein eigenes Geheimnis je
     Browser – ein geratener Name gibt damit keinen fremden Platz her. */
  var geraeteMerker = "";
  function geraetekennung() {
    if (geraeteMerker) {
      return geraeteMerker;
    }
    try {
      var gespeichert = localStorage.getItem("tg-geraet");
      if (gespeichert && gespeichert.length >= 16) {
        geraeteMerker = gespeichert;
        return geraeteMerker;
      }
    } catch (e) {
      /* kein Speicher - dann gilt der Merker dieser Sitzung */
    }
    geraeteMerker = neuesToken();
    try {
      localStorage.setItem("tg-geraet", geraeteMerker);
    } catch (e) {
      /* privater Modus: die Kennung gilt nur für diese Sitzung */
    }
    return geraeteMerker;
  }

  /* Diese Endpunkte brauchen die Gerätekennung: Sie holt eine bestehende
     Sitzung zurück, wenn das Token verloren ging (/api/ich), und sie gehört
     zur Anmeldung (/api/sitzung). Das Geheimnis bleibt im Browser. */
  function brauchtGeraet(adresse) {
    return (
      adresse.indexOf("/api/presence") >= 0 ||
      adresse.indexOf("/api/ich") >= 0 ||
      adresse.indexOf("/api/sitzung") >= 0 ||
      // Der zentrale Statusruf ist zugleich das Lebenszeichen – er braucht
      // dasselbe Token und dieselbe Gerätekennung wie /api/presence.
      adresse.indexOf("/api/zustand") >= 0
    );
  }

  function mitToken(adresse, optionen) {
    var teile = ausAnfrage(adresse, optionen);
    var token = tokenFuer(teile.raum, teile.id);
    var frisch = false;
    if (!token && adresse.indexOf("/api/presence") >= 0 && teile.raum && teile.id) {
      // Vor dem ersten Ruf erzeugen und merken: beide Rufe schicken dasselbe.
      token = neuesToken();
      merkeToken(teile.raum, teile.id, token);
      frisch = true;
    }
    var ziel = adresse;
    if (token && !/[?&]token=/.test(ziel)) {
      ziel += (ziel.indexOf("?") >= 0 ? "&" : "?") + "token=" + encodeURIComponent(token);
    }
    if (brauchtGeraet(adresse)) {
      // Der Beitritt, die Identitätsabfrage und die Sitzungsverwaltung kennen
      // die Gerätekennung – sie holt eine Sitzung zurück, deren Token
      // verloren ging, ohne eine neue Identität anzulegen.
      ziel += (ziel.indexOf("?") >= 0 ? "&" : "?") + "geraet=" + encodeURIComponent(geraetekennung());
    }
    return { ziel: ziel, frisch: frisch };
  }

  /* Ein unbekannter Platz entsteht nur noch mit ausdrücklichem "neu". Das setzt
     hier nur, wer ihn wirklich anlegen will: das erste Lebenszeichen eines
     Geräts und der Wiederholungsversuch nach einer 409 (der Server kennt den
     Platz nach einem Neustart nicht mehr). */
  function mitNeumarke(optionen) {
    if (!optionen || typeof optionen.body !== "string") {
      return optionen;
    }
    var rumpf;
    try {
      rumpf = JSON.parse(optionen.body);
    } catch (e) {
      return optionen;
    }
    if (!rumpf || typeof rumpf !== "object" || rumpf.neu) {
      return optionen;
    }
    rumpf.neu = true;
    var kopie = {};
    for (var name in optionen) {
      if (Object.prototype.hasOwnProperty.call(optionen, name)) {
        kopie[name] = optionen[name];
      }
    }
    kopie.body = JSON.stringify(rumpf);
    return kopie;
  }

  /* Doppelte Spielzüge: Jeder POST bekommt eine eigene Kennung. Wiederholt die
     Seite denselben Aufruf (weil die Antwort unterwegs verloren ging), reist
     dieselbe Kennung mit – der Server führt ihn dann nicht noch einmal aus.
     Der Merker hängt am Optionen-Objekt: Der vorhandene Wiederholungsweg ruft
     fetch mit genau demselben Objekt erneut auf. */
  var aktionsMerker = typeof WeakMap === "function" ? new WeakMap() : null;

  function neueAktionsmarke() {
    try {
      var puffer = new Uint8Array(12);
      (window.crypto || window.msCrypto).getRandomValues(puffer);
      var text = "";
      for (var i = 0; i < puffer.length; i++) {
        text += ("0" + puffer[i].toString(16)).slice(-2);
      }
      return text;
    } catch (e) {
      return String(Date.now()) + Math.random().toString(36).slice(2, 10);
    }
  }

  function mitAktionsmarke(adresse, optionen) {
    if (!optionen || String(optionen.method || "GET").toUpperCase() !== "POST") {
      return optionen;
    }
    if (adresse.indexOf("/api/") < 0) {
      return optionen;
    }
    var marke = "";
    if (aktionsMerker && typeof optionen === "object" && aktionsMerker.has(optionen)) {
      marke = aktionsMerker.get(optionen);
    }
    if (!marke) {
      marke = neueAktionsmarke();
      if (aktionsMerker && typeof optionen === "object") {
        aktionsMerker.set(optionen, marke);
      }
    }
    var kopie = {};
    for (var name in optionen) {
      if (Object.prototype.hasOwnProperty.call(optionen, name)) {
        kopie[name] = optionen[name];
      }
    }
    var kopf = {};
    var alt = optionen.headers;
    if (alt && typeof alt === "object" && !alt.forEach) {
      for (var k in alt) {
        if (Object.prototype.hasOwnProperty.call(alt, k)) {
          kopf[k] = alt[k];
        }
      }
    }
    kopf["X-Aktion"] = marke;
    kopie.headers = kopf;
    return kopie;
  }

  /* Ein Platz, den der Server nicht mehr kennt (Neustart) oder ein verlorenes
     Lebenszeichen: einmal neu anmelden und dieselbe Anfrage wiederholen. Die
     Marke bleibt gleich, ein schon ausgeführter Zug wird also nicht doppelt
     gespielt. Das gilt für jede Spielseite, die token.js einbindet. */
  function beitrittErneuern(adresse, optionen) {
    var teile = ausAnfrage(adresse, optionen);
    if (!teile.raum || !teile.id) {
      return Promise.resolve(false);
    }
    var name = "";
    try {
      name = new URL(adresse, location.href).searchParams.get("name") || "";
    } catch (e) {
      name = "";
    }
    if (!name && optionen && typeof optionen.body === "string") {
      try {
        name = JSON.parse(optionen.body).name || "";
      } catch (e) {
        name = "";
      }
    }
    // Wichtig: mit dem eigenen Token und der Gerätekennung. Ohne beides legte
    // der Server den Platz mit einem Zufallstoken an – der Browser käme danach
    // nie wieder hinein (jeder Ruf 403 „gehört zu einem anderen Gerät“).
    var ziel = "/api/presence";
    var eigenes = tokenFuer(teile.raum, teile.id);
    if (eigenes) {
      ziel += "?token=" + encodeURIComponent(eigenes);
    }
    ziel += (ziel.indexOf("?") >= 0 ? "&" : "?") + "geraet=" + encodeURIComponent(geraetekennung());
    return original(ziel, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        room: teile.raum,
        id: teile.id,
        name: name,
        active: true,
        neu: true
      })
    })
      .then(function (antwort) {
        if (!antwort || !antwort.ok) {
          return false;
        }
        // Ein neu vergebenes Token sofort merken, sonst passt der Platz beim
        // nächsten Ruf nicht mehr zu diesem Gerät.
        antwort
          .clone()
          .json()
          .then(function (daten) {
            if (daten && daten.token) {
              merkeToken(daten.room || teile.raum, teile.id, daten.token);
            }
          })
          .catch(function () {});
        return true;
      })
      .catch(function () {
        return false;
      });
  }

  function brauchtBeitritt(text) {
    return /beitreten/i.test(text || "");
  }

  window.fetchMitToken = function (adresse, optionen) {
    if (typeof adresse !== "string" || adresse.indexOf("/api/") < 0) {
      return original(adresse, optionen);
    }
    var teile = mitToken(adresse, optionen);
    var ziel = teile.ziel;
    var gefragt = optionen;
    if (teile.frisch && ziel.indexOf("/api/presence") >= 0) {
      // Erster Ruf dieses Geräts: der Platz ist neu, das sagen wir auch.
      gefragt = mitNeumarke(optionen);
    }
    gefragt = mitAktionsmarke(ziel, gefragt);
    var laeuft = original(ziel, gefragt);
    // Verlorener Platz (Server neu gestartet, Lebenszeichen weg): einmal
    // beitreten und dieselbe Anfrage wiederholen. GET ist wiederholbar, POST
    // trägt die Marke und wird vom Server nicht doppelt ausgeführt.
    if (ziel.indexOf("/api/presence") < 0 && laeuft && laeuft.then) {
      laeuft = laeuft.then(function (antwort) {
        if (!antwort || antwort.status !== 403) {
          return antwort;
        }
        return antwort
          .clone()
          .text()
          .then(function (text) {
            if (!brauchtBeitritt(text)) {
              return antwort;
            }
            return beitrittErneuern(adresse, optionen).then(function (ok) {
              if (!ok) {
                return antwort;
              }
              return original(ziel, gefragt);
            });
          })
          .catch(function () {
            return antwort;
          });
      });
    }
    if (ziel.indexOf("/api/presence") >= 0 && laeuft && laeuft.then) {
      // 409: Der Server kennt den Platz nicht mehr (Neustart). Denselben Platz
      // einmal ausdrücklich neu anlegen – gleiche Kennung, gleiches Token.
      laeuft = laeuft.then(function (antwort) {
        if (!antwort || antwort.status !== 409) {
          return antwort;
        }
        var neu = mitNeumarke(optionen);
        if (neu === optionen) {
          return antwort;
        }
        return original(ziel, neu);
      });
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
