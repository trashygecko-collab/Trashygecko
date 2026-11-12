/* Schwaches WLAN: Keine Anfrage darf eine Seite ewig hängen lassen.
   Ohne Frist bleibt eine Seite bei einem abreißenden Funknetz im Ladezustand
   stecken – Bedienung tot, bis der Browser irgendwann aufgibt. Jede Anfrage
   bekommt deshalb hier eine Frist; danach gilt sie als fehlgeschlagen und die
   Seite versucht es beim nächsten Takt einfach neu. */
(function () {
  var original = window.fetch;
  if (!original || typeof AbortController !== "function" || window.fetchMitFrist) {
    return;
  }

  window.fetchMitFrist = function (url, options) {
    var opts = {};
    if (options) {
      for (var key in options) {
        if (Object.prototype.hasOwnProperty.call(options, key)) {
          opts[key] = options[key];
        }
      }
    }
    if (opts.signal) {
      return original(url, opts);
    }
    var controller = new AbortController();
    var timer = setTimeout(function () { controller.abort(); }, 15000);
    opts.signal = controller.signal;
    return original(url, opts).finally(function () { clearTimeout(timer); });
  };

  window.fetch = function (url, options) {
    return window.fetchMitFrist(url, options);
  };
})();

/* --- Ein Spiel verlassen ---------------------------------------------------
   Jede Spielseite bekommt einen kleinen Knopf "Spiel verlassen". Der trägt die
   eigene Figur beziehungsweise die Karten auf dem Server aus, damit niemand auf
   einen wartet, und geht danach zurück in den Raum. Die Mitspieler erfahren das
   sofort im Chat.

   Zusätzlich fragt die Seite im Hintergrund, ob man noch in einer Partie sitzt.
   Ist die eigene Partie vorbei (weil jemand gegangen ist), landet man von selbst
   wieder im Raum – ohne langes Warten. */
(function () {
  var params = new URLSearchParams(location.search);
  var room = params.get("room") || "";
  var id = params.get("id") || "";

  function zurueckInDenRaum() {
    location.href = "/connector" + (room ? "?room=" + encodeURIComponent(room) : "");
  }
  window.tgRaumZurueck = zurueckInDenRaum;

  window.tgVerlassen = function () {
    if (!room || !id) {
      zurueckInDenRaum();
      return;
    }
    var weg = function () { zurueckInDenRaum(); };
    var notbremse = setTimeout(weg, 1500); // schwaches WLAN darf nicht blockieren
    fetch("/api/leave?room=" + encodeURIComponent(room), {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ id: id })
    }).catch(function () {}).then(function () {
      clearTimeout(notbremse);
      weg();
    });
  };

  if (!room || !id) {
    return; // Connector-Seite oder lokales Spiel: kein Wächter nötig
  }

  // Seiten mit eigenem Platz für den Knopf (Reversi, Arschloch) zeigen ihn hier.
  function zeigeVerlassen() {
    var knoepfe = document.querySelectorAll("[data-tg-verlassen]");
    for (var i = 0; i < knoepfe.length; i++) {
      knoepfe[i].removeAttribute("hidden");
      knoepfe[i].style.display = "";
    }
  }
  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", zeigeVerlassen);
  } else {
    zeigeVerlassen();
  }

  var warDrin = false;
  function wachen() {
    // Auch im Hintergrund fragen: sonst gilt man beim Wechsel in eine andere
    // App als weg und der Raum räumt die eigene Partie weg.
    fetch("/api/imspiel?room=" + encodeURIComponent(room) + "&id=" + encodeURIComponent(id))
      .then(function (r) { return r.ok ? r.json() : null; })
      .then(function (d) {
        if (!d) return;
        if (d.inGame) {
          warDrin = true;
        } else if (warDrin) {
          zurueckInDenRaum(); // Partie vorbei – zurück in den Raum
        }
      })
      .catch(function () {});
  }
  setInterval(wachen, 2500);
  setTimeout(wachen, 1200);
})();
