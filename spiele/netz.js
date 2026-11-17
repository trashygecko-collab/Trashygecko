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
   "Spiel verlassen" trägt die eigene Figur beziehungsweise die Karten auf dem
   Server aus, damit niemand auf einen wartet, und geht zurück in den Raum.

   Zusätzlich fragt die Seite im Hintergrund, ob man noch in einer Partie
   sitzt. Ist die Partie vorbei (weil jemand gegangen ist oder die Verbindung
   weg war), landet man von selbst wieder im Raum – mit dem Grund als Hinweis. */
(function () {
  var params = new URLSearchParams(location.search);
  var room = params.get("room") || "";
  var id = params.get("id") || "";

  var unterwegs = false; // Rückleitung läuft

  // Steckt die Seite im Spielrahmen des Connectors (PC)? Dann darf sie auf
  // keinen Fall selbst einen Connector laden – sie meldet sich beim Rahmen und
  // der schließt sie.
  var imRahmen = false;
  try {
    imRahmen = !!(window.parent && window.parent !== window &&
      window.parent.location && window.parent.location.origin === location.origin);
  } catch (e) {
    imRahmen = false;
  }

  function rahmenMelden(art, hinweis) {
    try {
      window.parent.postMessage({
        tg: "spiel", art: art, hinweis: hinweis || "", room: room, id: id
      }, location.origin);
    } catch (e) {}
  }

  // Zurück in den Raum: im Rahmen den Rahmen schließen lassen, im eigenen Tab
  // erst versuchen, den Tab zu schließen (er wurde vom Connector geöffnet) und
  // sonst wie bisher auf den Connector umleiten.
  function weg(hinweis) {
    if (imRahmen) {
      rahmenMelden("verlassen", hinweis);
      return;
    }
    if (window.opener) {
      try { window.close(); } catch (e) {}
      setTimeout(function () {
        if (!window.closed) zurueckInDenRaum(hinweis);
      }, 400);
      return;
    }
    zurueckInDenRaum(hinweis);
  }
  var laeuft = false; // eine Abfrage ist offen
  var takt = 0; // Hintergrundabfrage

  function zurueckInDenRaum(hinweis) {
    if (unterwegs) {
      return;
    }
    unterwegs = true;
    clearInterval(takt);
    if (imRahmen) {
      // Kein zweiter Connector im Rahmen: dort genügt die Meldung.
      rahmenMelden("ende", hinweis);
      return;
    }
    var ziel = "/connector" + (room ? "?room=" + encodeURIComponent(room) : "");
    if (hinweis) {
      ziel += (room ? "&" : "?") + "hinweis=" + encodeURIComponent(hinweis);
    }
    location.href = ziel;
  }
  window.tgRaumZurueck = function () { zurueckInDenRaum(""); };

  window.tgVerlassen = function () {
    if (!room || !id) {
      weg("");
      return;
    }
    var jetztWeg = function () { weg(""); };
    var notbremse = setTimeout(jetztWeg, 1500); // schwaches WLAN darf nicht blockieren
    fetch("/api/leave?room=" + encodeURIComponent(room), {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ id: id })
    }).catch(function () {}).then(function () {
      clearTimeout(notbremse);
      jetztWeg();
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

  // Kurze, deutliche Meldung über der Partie, danach zurück in den Raum.
  function zeigeHinweis(text) {
    if (imRahmen) {
      // Der Rahmen zeigt die Meldung in der Statuszeile, der Tab bleibt nicht
      // mit einer halbfertigen Partie zurück.
      clearInterval(takt);
      rahmenMelden("ende", text);
      return;
    }
    var box = document.getElementById("tg-hinweis");
    if (!box) {
      box = document.createElement("div");
      box.id = "tg-hinweis";
      box.setAttribute("role", "status");
      box.style.cssText =
        "position:fixed;left:50%;top:16px;transform:translateX(-50%);z-index:2147483647;" +
        "max-width:min(92vw,440px);padding:12px 18px;border-radius:10px;text-align:center;" +
        "background:#14392f;color:#f4e6bd;border:1px solid #e0c37e;font:15px/1.45 Georgia,serif;" +
        "box-shadow:0 8px 24px rgba(0,0,0,.35)";
      document.body.appendChild(box);
    }
    box.textContent = text;
    setTimeout(function () { zurueckInDenRaum(text); }, 1600);
  }

  var gesehen = false; // eigene Partie auf dieser Seite schon gesehen
  var ohneSpiel = 0; // Abfragen ohne Partie, ohne je eine gesehen zu haben

  // Im Hintergrund meldet sich die Seite mit seite=versteckt: Der Server hält
  // die Partie dann stehen, statt sie nach 90 s zu beenden – ein eingefrorener
  // Tab ist von einem geschlossenen nicht zu unterscheiden. Sichtbar wird das
  // Signal beim nächsten Ruf wieder gelöscht.
  function seite() {
    return document.visibilityState === "hidden" ? "&seite=versteckt" : "";
  }

  function wachen() {
    // Auch im Hintergrund fragen: sonst gilt man beim Wechsel in eine andere
    // App als weg und der Raum räumt die eigene Partie weg.
    if (laeuft || unterwegs) {
      return;
    }
    laeuft = true;
    fetch("/api/imspiel?room=" + encodeURIComponent(room) + "&id=" + encodeURIComponent(id) + seite())
      .then(function (r) { return r.ok ? r.json() : null; })
      .then(function (d) {
        laeuft = false;
        if (!d || unterwegs) {
          return;
        }
        if (d.inGame) {
          gesehen = true;
          ohneSpiel = 0;
          return;
        }
        if (d.hinweis) {
          // Der Server hat die Partie für uns beendet – der Grund kommt mit.
          clearInterval(takt);
          zeigeHinweis(d.hinweis);
          return;
        }
        if (gesehen) {
          zurueckInDenRaum(""); // Partie vorbei – zurück in den Raum
          return;
        }
        // Seite ohne Partie geöffnet (etwa nach einem Abbruch): Wer nie in
        // einer Partie saß, wird nicht sofort verjagt – erst nach 15 s.
        ohneSpiel += 1;
        if (ohneSpiel >= 6) {
          zurueckInDenRaum("");
        }
      })
      .catch(function () { laeuft = false; });
  }
  takt = setInterval(wachen, 2500);
  setTimeout(wachen, 1200);

  // Zurückgeholter Tab: sofort nachsehen statt bis zum nächsten Takt zu warten.
  // Weggedrückter Tab: sofort versteckt melden, bevor das System ihn drosselt
  // oder einfriert – sonst kommt das Signal nie an.
  document.addEventListener("visibilitychange", function () {
    wachen();
  });
  window.addEventListener("focus", wachen);
})();
