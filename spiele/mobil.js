/* --- Zentrale mobile Bedienung für den Connector ---------------------------
   Gehört zur Oberfläche (connector.html), zu keinem Spiel: Sie liest den Chat
   und das Spielfenster, die dort ohnehin stehen, und legt einen kleinen
   schwebenden Knopf darüber.

     · Sprung zum Chat, solange eine Partie läuft – und wieder zurück
     · Zähler für ungelesene Nachrichten
     · Position frei verschiebbar, je Spiel gemerkt
     · im Vollbild und im PC-Layout (ab 1024 px) unsichtbar

   Der Knopf ist 46 px groß und fängt nur seine eigenen Berührungen ab – kein
   Schleier über dem Spielfeld, die Seite bleibt normal scrollbar. */
(function () {
  "use strict";
  if (window.TGMobil) return;

  var GROESSE = 46;   // px, zugleich Mindestabstand zum Bildschirmrand
  var RAND = 8;
  var ZIEH_WEG = 6;   // ab so vielen px gilt eine Berührung als Ziehen

  var knopf = null, zeichen = null, zaehler = null;
  var holeChat = null, holeSpiel = null;
  var spiel = "";                        // Kennung des offenen Spiels
  var ungelesen = 0;
  var position = { x: 0, y: 0 };
  var griff = null;                      // laufende Zeigerbewegung
  var scrollTakt = 0;

  // Hier zählt nur die Breite: das PC-Layout hängt allein an dieser Grenze.
  var breit = window.matchMedia ? window.matchMedia("(min-width: 1024px)") : null;

  function begrenzen(x, y) {
    var b = window.innerWidth || document.documentElement.clientWidth;
    var h = window.innerHeight || document.documentElement.clientHeight;
    return {
      x: Math.max(RAND, Math.min(b - GROESSE - RAND, Math.round(x))),
      y: Math.max(RAND, Math.min(h - GROESSE - RAND, Math.round(y)))
    };
  }

  function setzePosition(x, y, merken) {
    position = begrenzen(x, y);
    knopf.style.left = position.x + "px";
    knopf.style.top = position.y + "px";
    if (!merken) return;
    try { localStorage.setItem("tg-chat-knopf:" + spiel, JSON.stringify(position)); } catch (e) {}
  }

  // Gemerkter Platz dieses Spiels, sonst rechter Rand auf Höhe des Chatrands.
  function lade() {
    try {
      var p = JSON.parse(localStorage.getItem("tg-chat-knopf:" + spiel) || "null");
      if (p && typeof p.x === "number" && typeof p.y === "number") return begrenzen(p.x, p.y);
    } catch (e) { /* gesperrter Speicher: dann der Standardplatz */ }
    var h = window.innerHeight || document.documentElement.clientHeight;
    return begrenzen((window.innerWidth || 0) - GROESSE - RAND, h * 0.42);
  }

  function chatSichtbar() {
    var box = holeChat && holeChat();
    if (!box || box.hidden) return true;
    var r = box.getBoundingClientRect();
    return r.top >= -8 && r.bottom <= (window.innerHeight || 0) + 8;
  }

  function setzeUngelesen(n) {
    ungelesen = Math.max(0, n || 0);
    aktualisiere();
  }

  function aktualisiere() {
    if (!knopf) return;
    var voll = document.fullscreenElement || document.webkitFullscreenElement;
    // Nur das Vollbild des Spielfelds nimmt die ganze Seite ein; im Vollbild
    // der Seite (html) sind Chat und Spiel weiter zu sehen – dort gehört der
    // Knopf hin.
    var offen = !!spiel && !(breit && breit.matches) &&
      !(voll && voll !== document.documentElement);
    knopf.style.display = offen ? "flex" : "none";
    if (!offen) return;
    var imChat = chatSichtbar();
    if (imChat) ungelesen = 0;   // was im Bild steht, ist gelesen
    var text = imChat ? "Zurück zum Spiel" : "Zum Chat springen";
    zeichen.textContent = imChat ? "▾" : "💬";
    knopf.title = text;
    knopf.setAttribute("aria-label", text);
    zaehler.style.display = ungelesen > 0 ? "block" : "none";
    zaehler.textContent = ungelesen > 99 ? "99+" : String(ungelesen);
  }

  // Sprung ohne Animation: eine Animation läuft im Hintergrund-Tab nicht und
  // wird abgebrochen, sobald der Rahmen sein Spiel nachlädt.
  function springe(ziel) {
    if (ziel) window.scrollTo(0, ziel.getBoundingClientRect().top + (window.scrollY || 0));
  }

  function umschalten() {
    springe(chatSichtbar() ? holeSpiel && holeSpiel() : holeChat && holeChat());
    aktualisiere();   // räumt beim Chat den Zähler ab
  }

  function baue() {
    // Aussehen in einem eigenen Stilblock – so bleibt das Aufbauen kurz und die
    // Seite muss nichts über den Knopf wissen.
    var stil = document.createElement("style");
    stil.textContent =
      "#tg-mobil-knopf{position:fixed;left:0;top:0;z-index:2147483000;display:none;" +
      "width:" + GROESSE + "px;height:" + GROESSE + "px;align-items:center;justify-content:center;" +
      "padding:0;margin:0;border-radius:50%;border:1px solid #e8cd8e;background:#144238f2;" +
      "color:#e8cd8e;box-shadow:0 4px 14px rgba(0,0,0,.45);cursor:pointer;touch-action:none;" +
      "-webkit-tap-highlight-color:transparent;user-select:none}" +
      "#tg-mobil-knopf>span{pointer-events:none;font-size:21px;line-height:1}" +
      "#tg-mobil-knopf>b{position:absolute;top:-5px;right:-3px;min-width:20px;height:20px;" +
      "border-radius:10px;background:#d24a4a;color:#fff;font:bold 12px/20px system-ui,sans-serif;" +
      "text-align:center;padding:0 5px;box-sizing:border-box;display:none;pointer-events:none;" +
      "box-shadow:0 1px 4px rgba(0,0,0,.5)}";
    document.head.appendChild(stil);

    knopf = document.createElement("button");
    knopf.type = "button";
    knopf.id = "tg-mobil-knopf";
    zeichen = document.createElement("span");
    zeichen.textContent = "💬";
    zaehler = document.createElement("b");
    knopf.append(zeichen, zaehler);

    // Tipp und Ziehen trennen: unter ZIEH_WEG Bewegung ist es ein Tipp.
    knopf.addEventListener("pointerdown", function (e) {
      var r = knopf.getBoundingClientRect();
      griff = { mx: e.clientX, my: e.clientY, x: r.left, y: r.top, weg: 0 };
      try { knopf.setPointerCapture(e.pointerId); } catch (err) {}
      e.preventDefault();
    });
    knopf.addEventListener("pointermove", function (e) {
      if (!griff) return;
      var dx = e.clientX - griff.mx, dy = e.clientY - griff.my;
      griff.weg = Math.max(griff.weg, Math.abs(dx) + Math.abs(dy));
      if (griff.weg >= ZIEH_WEG) setzePosition(griff.x + dx, griff.y + dy, false);
    });
    knopf.addEventListener("pointerup", function (e) {
      if (!griff) return;
      var getippt = griff.weg < ZIEH_WEG;
      griff = null;
      try { knopf.releasePointerCapture(e.pointerId); } catch (err) {}
      if (getippt) umschalten();
      else setzePosition(position.x, position.y, true);
    });
    knopf.addEventListener("pointercancel", function () {
      if (!griff) return;
      griff = null;
      setzePosition(position.x, position.y, true);
    });
    // Tastatur: Enter und Leertaste lösen den Tipp aus.
    knopf.addEventListener("keydown", function (e) {
      if (e.key === "Enter" || e.key === " ") { e.preventDefault(); umschalten(); }
    });

    document.body.appendChild(knopf);
  }

  window.TGMobil = {
    // Einmal von der Oberfläche aufrufen; chat und spiel liefern die Elemente.
    einrichten: function (opts) {
      holeChat = (opts && opts.chat) || null;
      holeSpiel = (opts && opts.spiel) || null;
      baue();
      ["fullscreenchange", "webkitfullscreenchange"].forEach(function (name) {
        document.addEventListener(name, aktualisiere);
      });
      if (breit && breit.addEventListener) breit.addEventListener("change", aktualisiere);
      window.addEventListener("resize", function () {
        setzePosition(position.x, position.y, false);
        aktualisiere();
      });
      window.addEventListener("scroll", function () {
        clearTimeout(scrollTakt);
        scrollTakt = setTimeout(aktualisiere, 120);
      }, { passive: true });
    },
    // Eine Partie ist offen (nur am Handy aufgerufen): Knopf setzen und das
    // Spielfenster ins Bild holen.
    spielGeoeffnet: function (seite) {
      // Nur der Pfad: derselbe Spielplatz gilt für Chat- und Kachelstart.
      spiel = String(seite || "").replace(/^https?:\/\/[^/]+/, "") || "spiel";
      var platz = lade();
      setzePosition(platz.x, platz.y, false);
      setzeUngelesen(0);
      springe(holeSpiel && holeSpiel());
    },
    spielZu: function () {
      spiel = "";
      setzeUngelesen(0);
    },
    spielOffen: function () { return !!spiel; },
    // Eine neue Nachricht kam an, während der Chat nicht zu sehen war.
    ungesehen: function () { setzeUngelesen(ungelesen + 1); }
  };
})();
