/* ===========================================================================
   SPIELHALLE – ONLINE-TEIL
   ---------------------------------------------------------------------------
   Macht den vorhandenen Connector zum Bestandteil der Spielhalle, ohne eine
   zweite Architektur zu bauen. Verwendet werden ausschließlich bestehende
   Endpunkte:

     POST /api/rooms        Raum holen (Hauptraum) oder anlegen (neuerRaum)
     POST /api/presence     Platz betreten, Lebenszeichen, Umbenennen
     GET  /api/players      Spieler und Anwesenheit
     GET/POST /api/messages Raumchat (auch Einladungszeilen)
     GET/POST /api/angebot  offene Runde anbieten, annehmen, ablehnen
     GET  /api/<spiel>      Spielstand (Spieler, Bereitschaft, Status)
     POST /api/<spiel>      create/join/ready/start macht die Spielseite selbst

   Die Spielseite im Tisch-Rahmen ist dieselbe wie im Connector; sie bekommt
   Raum, Name und Kennung in der Adresse. Bereitschaft und Start bedient sie
   selbst (dort steht die Lobby), die Halle zeigt den Stand und öffnet den
   Tisch. Kennung, Name und Token teilen sich Halle und Connector
   (tg-player-id, tg-player-name, token.js) – es gibt keine zweite Identität.
   ========================================================================= */
(function () {
  "use strict";

  var SPIELE = (window.TGMultiplayer && window.TGMultiplayer.spiele) || [];
  var POLL = 2500;
  var PRESENCE_ALLE = 8;      // × 2,5 s = 20 s
  var SPIELER_ALLE = 3;       // × 2,5 s = 7,5 s
  var ANGEBOT_ALLE = 2;       // × 2,5 s = 5 s

  var s = {
    online: false,
    verbindet: false,
    raum: "",
    name: "",
    ich: "",
    spieler: [],
    angebot: null,
    angebotWeg: "",
    chatOffen: false,
    chatNeu: 0,
    panelOffen: false,
    tisch: null,              // { pfad, name, def, spiel, status }
    letzterTisch: null,       // bleibt nach dem Schließen für „fortsetzen“
    spielstand: null,
    toastZeit: 0,
    letzterFehler: "",
    neuversuchBis: 0
  };

  // ------------------------------------------------------------- Speicher ---

  function lies(schluessel) {
    try { return localStorage.getItem(schluessel) || ""; } catch (e) { return ""; }
  }

  function schreibe(schluessel, wert) {
    try { localStorage.setItem(schluessel, wert); } catch (e) { /* privater Modus */ }
  }

  function sitzung(schluessel, wert) {
    try {
      if (wert === undefined) return sessionStorage.getItem(schluessel) || "";
      sessionStorage.setItem(schluessel, wert);
    } catch (e) { /* privater Modus */ }
    return "";
  }

  function neueKennung() {
    var abc = "abcdefghijklmnopqrstuvwxyz0123456789";
    var text = "";
    try {
      var puffer = new Uint8Array(24);
      crypto.getRandomValues(puffer);
      for (var i = 0; i < puffer.length; i++) text += abc.charAt(puffer[i] % abc.length);
    } catch (e) {
      for (var j = 0; j < 24; j++) text += abc.charAt(Math.floor(Math.random() * abc.length));
    }
    return text;
  }

  function kennung() {
    if (s.ich) return s.ich;
    var id = sitzung("tg-player-id") || lies("tg-player-id");
    if (!id) {
      id = neueKennung();
      sitzung("tg-player-id", id);
      schreibe("tg-player-id", id);
    } else {
      sitzung("tg-player-id", id);
      schreibe("tg-player-id", id);
    }
    s.ich = id;
    return id;
  }

  function neueKennungErzwingen() {
    s.ich = neueKennung();
    sitzung("tg-player-id", s.ich);
    schreibe("tg-player-id", s.ich);
  }

  function eigenerName() {
    if (s.name) return s.name;
    s.name = lies("tg-player-name");
    return s.name;
  }

  function nameMerken(name) {
    s.name = String(name || "").trim();
    schreibe("tg-player-name", s.name);
  }

  function gespeicherterRaum() {
    return lies("tg-halle-raum").toUpperCase();
  }

  function raumMerken(code) {
    s.raum = String(code || "").toUpperCase();
    if (s.raum) schreibe("tg-halle-raum", s.raum);
    else {
      try { localStorage.removeItem("tg-halle-raum"); } catch (e) { /* egal */ }
    }
  }

  function urlRaum() {
    try {
      return (new URLSearchParams(location.search).get("raum") || "").toUpperCase();
    } catch (e) {
      return "";
    }
  }

  function anzeigeName() {
    return eigenerName() || "Namenloser";
  }

  /* Ohne eigenen Namen bekommt jeder einen brauchbaren Gastnamen. Er steht
     im selben Speicher wie im Connector und lässt sich im Panel ändern. */
  function gastName() {
    var code = kennung().slice(-3).toUpperCase();
    return "Gast " + code;
  }

  function defFuer(pfad) {
    if (!pfad) return null;
    var absolut = pfad.charAt(0) === "/" ? pfad : "/spiele/" + pfad;
    for (var i = 0; i < SPIELE.length; i++) {
      if (SPIELE[i].pfad === absolut) return SPIELE[i];
    }
    return null;
  }

  function spielApi(def) {
    return "/api/" + (def && def.api ? def.api : "");
  }

  function defNachId(id) {
    for (var i = 0; i < SPIELE.length; i++) {
      if (SPIELE[i].id === id) return SPIELE[i];
    }
    return null;
  }

  function initialen(name) {
    var teile = String(name || "?").trim().split(/\s+/).filter(function (t) { return t; });
    if (!teile.length) return "?";
    if (teile.length === 1) return teile[0].slice(0, 2).toUpperCase();
    return (teile[0].charAt(0) + teile[1].charAt(0)).toUpperCase();
  }

  // --------------------------------------------------------------- Helfer ---

  function el(tag, klasse, text) {
    var n = document.createElement(tag);
    if (klasse) n.className = klasse;
    if (text !== undefined && text !== null) n.textContent = text;
    return n;
  }

  function $(id) { return document.getElementById(id); }

  function warte(ms) {
    return new Promise(function (fertig) { setTimeout(fertig, ms); });
  }

  async function anfrage(pfad, optionen) {
    var antwort = await fetch(pfad, optionen);
    if (!antwort.ok) {
      var text = "";
      try { text = (await antwort.text()).trim(); } catch (e) { text = ""; }
      var fehler = new Error(text || ("Der Connector antwortet mit " + antwort.status + "."));
      fehler.status = antwort.status;
      throw fehler;
    }
    return antwort;
  }

  function jsonRumpf(daten) {
    return {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(daten)
    };
  }

  function meldung(text) {
    var box = $("halle-toast");
    if (!box) {
      box = el("div", "toast");
      box.id = "halle-toast";
      box.setAttribute("role", "status");
      document.body.appendChild(box);
    }
    box.textContent = text;
    box.classList.add("an");
    clearTimeout(s.toastZeit);
    s.toastZeit = setTimeout(function () { box.classList.remove("an"); }, 3800);
  }

  function fehlerText(fehler) {
    if (!fehler) return "Unbekannter Fehler.";
    if (fehler.name === "AbortError") return "Der Connector antwortet nicht.";
    return fehler.message || "Der Connector ist nicht erreichbar.";
  }

  function kopiere(text) {
    if (navigator.clipboard && navigator.clipboard.writeText) {
      return navigator.clipboard.writeText(text).then(function () {
        meldung("In die Zwischenablage kopiert.");
      }).catch(function () {
        fensterPrompt(text);
      });
    }
    fensterPrompt(text);
  }

  function fensterPrompt(text) {
    window.prompt("Zum Kopieren:", text);
  }

  function einladungsLink() {
    if (!s.raum) return "";
    return location.origin + location.pathname + "?raum=" + encodeURIComponent(s.raum);
  }

  // ------------------------------------------------------------ Oberfläche ---

  var ui = {};

  function baueLeiste() {
    var bereich = el("section", "online-bereich");
    bereich.id = "online-bereich";
    bereich.setAttribute("aria-label", "Online-Spielhalle");

    var kopf = el("div", "ob-kopf");
    var titel = el("div", "ob-titel");
    titel.appendChild(el("span", "punkt"));
    var titelText = el("span", "", "Online-Spielhalle");
    titelText.id = "ob-titel-text";
    titel.appendChild(titelText);
    kopf.appendChild(titel);

    var code = el("div", "ob-code");
    code.appendChild(el("span", "", "Raum"));
    var codeWert = el("strong", "", "—");
    codeWert.id = "ob-code";
    code.appendChild(codeWert);
    kopf.appendChild(code);
    bereich.appendChild(kopf);

    var text = el("p", "ob-text", "");
    text.id = "ob-text";
    bereich.appendChild(text);

    var reihe = el("div", "ob-reihe");
    var spieler = el("div", "ob-spieler");
    spieler.id = "ob-spieler";
    reihe.appendChild(spieler);

    var aktionen = el("div", "ob-aktionen");
    aktionen.id = "ob-aktionen";
    aktionen.style.display = "flex";
    aktionen.style.flexWrap = "wrap";
    aktionen.style.gap = "7px";
    reihe.appendChild(aktionen);
    bereich.appendChild(reihe);

    var hinweis = el("p", "ob-hinweis", "");
    hinweis.id = "ob-hinweis";
    bereich.appendChild(hinweis);

    var suchzeile = document.querySelector(".suchzeile");
    if (suchzeile && suchzeile.parentNode) {
      suchzeile.parentNode.insertBefore(bereich, suchzeile.nextSibling);
    } else {
      var halle = document.querySelector(".halle") || document.body;
      halle.insertBefore(bereich, halle.firstChild);
    }
    ui.bereich = bereich;
    ui.titel = titelText;
    ui.code = codeWert;
    ui.text = text;
    ui.spieler = spieler;
    ui.aktionen = aktionen;
    ui.hinweis = hinweis;

    var kopfAktionen = document.querySelector(".kopf-aktionen");
    if (kopfAktionen) {
      var knopf = el("button", "online-knopf");
      knopf.type = "button";
      knopf.id = "online-knopf";
      knopf.setAttribute("aria-haspopup", "dialog");
      var punkt = el("span", "punkt");
      var wort = el("span", "wort", "Online spielen");
      var codeKurz = el("span", "code", "");
      codeKurz.id = "online-knopf-code";
      knopf.appendChild(punkt);
      knopf.appendChild(wort);
      knopf.appendChild(codeKurz);
      knopf.addEventListener("click", function () { panelOeffnen(); });
      kopfAktionen.insertBefore(knopf, kopfAktionen.firstChild);
      ui.kopfKnopf = knopf;
      ui.kopfWort = wort;
      ui.kopfCode = codeKurz;
    }
  }

  function bauePanel() {
    var schleier = el("div", "panel-schleier");
    schleier.id = "online-panel";
    schleier.hidden = true;
    schleier.addEventListener("click", function (e) {
      if (e.target === schleier) panelSchliessen();
    });

    var panel = el("div", "panel");
    panel.setAttribute("role", "dialog");
    panel.setAttribute("aria-modal", "true");
    panel.setAttribute("aria-labelledby", "panel-titel");

    var kopf = el("div", "panel-kopf");
    var h2 = el("h2", "", "Online-Spielhalle");
    h2.id = "panel-titel";
    kopf.appendChild(h2);
    var zu = el("button", "zu", "×");
    zu.type = "button";
    zu.setAttribute("aria-label", "Schließen");
    zu.addEventListener("click", panelSchliessen);
    kopf.appendChild(zu);
    panel.appendChild(kopf);

    var stand = el("p", "panel-stand", "Noch nicht verbunden.");
    stand.id = "panel-stand";
    panel.appendChild(stand);

    var namensFeld = el("div", "feld");
    var namensLabel = el("label", "", "Dein Name");
    namensLabel.setAttribute("for", "panel-name");
    namensFeld.appendChild(namensLabel);
    var namensZeile = el("div", "zeile");
    var nameInput = el("input");
    nameInput.id = "panel-name";
    nameInput.maxLength = 24;
    nameInput.autocomplete = "nickname";
    nameInput.placeholder = "z. B. Anna";
    nameInput.value = eigenerName();
    namensZeile.appendChild(nameInput);
    var nameSpeichern = el("button", "ob-knopf", "Speichern");
    nameSpeichern.type = "button";
    nameSpeichern.addEventListener("click", function () {
      nameSpeichern.disabled = true;
      nameAendern(nameInput.value).finally(function () { nameSpeichern.disabled = false; });
    });
    namensZeile.appendChild(nameSpeichern);
    namensFeld.appendChild(namensZeile);
    panel.appendChild(namensFeld);

    var raumBlock = el("div", "panel-block");
    raumBlock.appendChild(el("h3", "", "Raum"));
    var codeZeile = el("div", "panel-code");
    var code = el("strong", "", "—");
    code.id = "panel-code";
    codeZeile.appendChild(code);
    var codeKopieren = el("button", "mini", "Code kopieren");
    codeKopieren.type = "button";
    codeKopieren.addEventListener("click", function () {
      if (s.raum) kopiere(s.raum);
    });
    codeZeile.appendChild(codeKopieren);
    var linkKopieren = el("button", "mini", "Einladungslink");
    linkKopieren.type = "button";
    linkKopieren.addEventListener("click", function () {
      if (s.raum) kopiere(einladungsLink());
    });
    codeZeile.appendChild(linkKopieren);
    raumBlock.appendChild(codeZeile);
    var qr = el("div", "panel-qr");
    var qrBild = el("img");
    qrBild.alt = "QR-Code zum Einladungslink";
    qrBild.hidden = true;
    qr.appendChild(qrBild);
    raumBlock.appendChild(qr);
    var raumHinweis = el("p", "ob-hinweis", "Der Einladungslink öffnet die Spielhalle und tritt diesem Raum bei.");
    raumBlock.appendChild(raumHinweis);
    panel.appendChild(raumBlock);

    var spielerBlock = el("div", "panel-block");
    spielerBlock.appendChild(el("h3", "", "Im Raum"));
    var spielerListe = el("div", "panel-spieler");
    spielerListe.id = "panel-spieler";
    spielerBlock.appendChild(spielerListe);
    panel.appendChild(spielerBlock);

    var rundenBlock = el("div", "panel-block");
    rundenBlock.id = "panel-angebot";
    rundenBlock.hidden = true;
    rundenBlock.appendChild(el("h3", "", "Offene Runde"));
    var rundenText = el("p", "ob-hinweis", "");
    rundenText.id = "panel-angebot-text";
    rundenText.style.margin = "0 0 8px";
    rundenBlock.appendChild(rundenText);
    var rundenKnoepfe = el("div", "panel-aktionen");
    rundenKnoepfe.id = "panel-angebot-knoepfe";
    rundenBlock.appendChild(rundenKnoepfe);
    panel.appendChild(rundenBlock);

    var aktionen = el("div", "panel-aktionen");
    var neuerRaum = el("button", "", "Neuen Raum öffnen");
    neuerRaum.type = "button";
    neuerRaum.addEventListener("click", function () {
      raumNeu().catch(function (fehler) { meldung(fehlerText(fehler)); });
    });
    aktionen.appendChild(neuerRaum);

    var verlassen = el("button", "gefahr", "Raum verlassen");
    verlassen.type = "button";
    verlassen.addEventListener("click", function () {
      trennen();
      meldung("Raum verlassen.");
    });
    aktionen.appendChild(verlassen);
    panel.appendChild(aktionen);

    var beitretenBlock = el("div", "panel-block");
    beitretenBlock.appendChild(el("h3", "", "Raum betreten"));
    var form = el("form");
    form.style.display = "flex";
    form.style.gap = "7px";
    var codeInput = el("input");
    codeInput.placeholder = "Raumcode, z. B. TG-AB12";
    codeInput.autocomplete = "off";
    codeInput.style.flex = "1 1 auto";
    codeInput.style.minWidth = "0";
    codeInput.style.minHeight = "40px";
    codeInput.style.padding = "8px 11px";
    codeInput.style.border = "1px solid var(--tg-rahmen)";
    codeInput.style.borderRadius = "9px";
    codeInput.style.background = "var(--tg-bg-3)";
    codeInput.style.color = "var(--tg-text)";
    form.appendChild(codeInput);
    var beitreten = el("button", "ob-knopf", "Beitreten");
    beitreten.type = "submit";
    form.appendChild(beitreten);
    form.addEventListener("submit", function (e) {
      e.preventDefault();
      var wunsch = codeInput.value.trim().toUpperCase();
      if (!wunsch) return;
      beitreten.disabled = true;
      raumBetreten(wunsch).then(function () {
        codeInput.value = "";
      }).catch(function (fehler) {
        meldung(fehlerText(fehler));
      }).finally(function () { beitreten.disabled = false; });
    });
    beitretenBlock.appendChild(form);
    panel.appendChild(beitretenBlock);

    schleier.appendChild(panel);
    document.body.appendChild(schleier);

    ui.panel = schleier;
    ui.panelStand = stand;
    ui.panelCode = code;
    ui.panelQr = qrBild;
    ui.panelSpieler = spielerListe;
    ui.nameInput = nameInput;
  }

  function baueChat() {
    var knopf = el("button", "chat-knopf");
    knopf.type = "button";
    knopf.id = "halle-chat-knopf";
    knopf.hidden = true;
    knopf.setAttribute("aria-controls", "halle-chat");
    knopf.appendChild(el("span", "", "Chat"));
    var zahl = el("span", "zahl", "0");
    knopf.appendChild(zahl);
    knopf.addEventListener("click", function () { chatUmschalten(true); });
    document.body.appendChild(knopf);
    ui.chatKnopf = knopf;
    ui.chatZahl = zahl;

    var chat = el("aside", "chat");
    chat.id = "halle-chat";
    chat.hidden = true;
    chat.setAttribute("aria-label", "Raumchat");

    var kopf = el("div", "chat-kopf");
    kopf.appendChild(el("strong", "", "Raumchat"));
    var stand = el("span", "stand", "");
    kopf.appendChild(stand);
    var zu = el("button", "zu", "×");
    zu.type = "button";
    zu.setAttribute("aria-label", "Chat schließen");
    zu.addEventListener("click", function () { chatUmschalten(false); });
    kopf.appendChild(zu);
    chat.appendChild(kopf);

    var verlauf = el("div", "chat-verlauf");
    verlauf.id = "halle-chat-verlauf";
    chat.appendChild(verlauf);

    var form = el("form", "chat-form");
    var eingabe = el("input");
    eingabe.maxLength = 1000;
    eingabe.placeholder = "Nachricht schreiben …";
    eingabe.setAttribute("aria-label", "Chatnachricht");
    eingabe.autocomplete = "off";
    form.appendChild(eingabe);
    var senden = el("button", "", "Senden");
    senden.type = "submit";
    form.appendChild(senden);
    form.addEventListener("submit", function (e) {
      e.preventDefault();
      var text = eingabe.value.trim();
      if (!text) return;
      senden.disabled = true;
      chatSenden(text).then(function () {
        eingabe.value = "";
      }).catch(function (fehler) {
        meldung(fehlerText(fehler));
      }).finally(function () { senden.disabled = false; });
    });
    chat.appendChild(form);
    document.body.appendChild(chat);

    ui.chat = chat;
    ui.chatVerlauf = verlauf;
    ui.chatStand = stand;
    ui.chatInput = eingabe;
  }

  function baueAngebot() {
    var karte = el("div", "angebot-karte");
    karte.id = "halle-angebot";
    karte.hidden = true;
    karte.setAttribute("role", "status");
    ui.angebot = karte;
    document.body.appendChild(karte);
  }

  function baueTisch() {
    var overlay = el("div", "spiel-overlay");
    overlay.id = "halle-tisch";
    overlay.hidden = true;
    overlay.setAttribute("role", "dialog");
    overlay.setAttribute("aria-modal", "true");
    overlay.setAttribute("aria-labelledby", "tisch-name");

    var kopf = el("header", "ov-kopf");
    var titel = el("div", "titel");
    var name = el("strong", "", "Spiel");
    name.id = "tisch-name";
    titel.appendChild(name);
    var stand = el("span", "stand", "");
    stand.id = "tisch-stand";
    titel.appendChild(stand);
    kopf.appendChild(titel);

    var aktionen = el("div", "tisch-btns");
    aktionen.id = "tisch-aktionen";
    kopf.appendChild(aktionen);

    var raum = el("span", "raum", "");
    raum.id = "tisch-raum";
    kopf.appendChild(raum);

    var chat = el("button", "", "Chat");
    chat.type = "button";
    chat.id = "tisch-chat";
    chat.addEventListener("click", function () { chatUmschalten(!s.chatOffen, true); });
    kopf.appendChild(chat);

    var zu = el("button", "zu", "Zurück zur Halle");
    zu.type = "button";
    zu.addEventListener("click", tischSchliessen);
    kopf.appendChild(zu);
    overlay.appendChild(kopf);

    var koerper = el("div", "ov-koerper");
    var rahmen = el("iframe");
    rahmen.id = "tisch-rahmen";
    rahmen.title = "Spiel";
    rahmen.allow = "autoplay; fullscreen";
    rahmen.setAttribute("allowfullscreen", "");
    koerper.appendChild(rahmen);
    overlay.appendChild(koerper);
    document.body.appendChild(overlay);

    ui.tisch = overlay;
    ui.tischName = name;
    ui.tischStand = stand;
    ui.tischAktionen = aktionen;
    ui.tischRaum = raum;
    ui.tischKoerper = koerper;
    ui.tischRahmen = rahmen;
    ui.tischChat = chat;
  }

  // ---------------------------------------------------------------- Render ---

  function zeichneLeiste() {
    if (!ui.bereich) return;
    ui.bereich.dataset.online = s.online ? "true" : "false";
    if (ui.kopfKnopf) ui.kopfKnopf.dataset.online = s.online ? "true" : "false";
    if (ui.kopfWort) ui.kopfWort.textContent = s.online ? "Online" : "Online spielen";
    if (ui.kopfCode) ui.kopfCode.textContent = s.online && s.raum ? s.raum : "";

    ui.code.textContent = s.raum || "—";
    ui.aktionen.textContent = "";
    ui.spieler.textContent = "";

    if (!s.online) {
      ui.titel.textContent = "Online-Spielhalle";
      ui.text.textContent = s.verbindet
        ? "Verbindung zum Spielraum wird hergestellt …"
        : "Gemeinsam spielen: Raum öffnen, Freunde einladen, an einem Tisch Platz nehmen. " +
          "Der Chat gehört dazu.";
      var start = el("button", "ob-knopf haupt", s.verbindet ? "Verbinde …" : "Mit Freunden spielen");
      start.type = "button";
      start.disabled = s.verbindet;
      start.addEventListener("click", function () {
        onlineGehen().catch(function (fehler) { meldung(fehlerText(fehler)); });
      });
      ui.aktionen.appendChild(start);
      var info = el("button", "ob-knopf leise", "So funktioniert's");
      info.type = "button";
      info.addEventListener("click", panelOeffnen);
      ui.aktionen.appendChild(info);
      ui.hinweis.textContent = "Einzelspieler laufen weiter ohne Verbindung.";
      ui.chatKnopf.hidden = true;
      return;
    }

    ui.titel.textContent = "Online · " + anzeigeName();
    var anwesend = s.spieler.filter(function (p) { return p.anwesend; }).length;
    ui.text.textContent = s.spieler.length
      ? s.spieler.length + (s.spieler.length === 1 ? " Spieler im Raum" : " Spieler im Raum") +
        " · " + anwesend + " jetzt da. Spiel wählen, Runde eröffnen – Mitspieler treten über die Einladung bei."
      : "Raum offen. Lade Freunde ein – sie treten über den Link oder den Raumcode bei.";
    if (s.letzterFehler) ui.text.textContent = s.letzterFehler;

    s.spieler.slice(0, 5).forEach(function (p) {
      var a = el("span", "ob-avatar" + (p.anwesend ? "" : " weg"), initialen(p.name));
      a.title = p.name + (p.anwesend ? " · anwesend" : " · nicht verbunden");
      if (p.id === s.ich) a.classList.add("ich");
      ui.spieler.appendChild(a);
    });
    if (s.spieler.length > 5) ui.spieler.appendChild(el("span", "ob-mehr", "+" + (s.spieler.length - 5)));

    var einladen = el("button", "ob-knopf", "Einladen");
    einladen.type = "button";
    einladen.addEventListener("click", function () { kopiere(einladungsLink()); });
    ui.aktionen.appendChild(einladen);

    var chat = el("button", "ob-knopf", "Chat");
    chat.type = "button";
    chat.addEventListener("click", function () { chatUmschalten(true); });
    ui.aktionen.appendChild(chat);

    var raumKnopf = el("button", "ob-knopf leise", "Raum " + s.raum);
    raumKnopf.type = "button";
    raumKnopf.addEventListener("click", panelOeffnen);
    ui.aktionen.appendChild(raumKnopf);

    /* Eigene offene Runde ist immer erreichbar – auch wenn der Tisch zu ist. */
    var eigeneRunde = s.angebot && s.angebot.vonId === s.ich ? s.angebot : null;
    var eigenerTisch = eigeneRunde ? defNachId(eigeneRunde.game) : null;
    if (eigeneRunde && eigenerTisch) {
      var runde = el("button", "ob-knopf haupt",
        "Runde: " + (eigeneRunde.gameName || eigenerTisch.name));
      runde.type = "button";
      runde.title = "Offene Runde öffnen";
      runde.addEventListener("click", function () { tischOeffnen(eigenerTisch.pfad, eigenerTisch.name); });
      ui.aktionen.appendChild(runde);
    }

    var letzterDef = s.letzterTisch ? defFuer(s.letzterTisch.pfad) : null;
    var gleicheRunde = !!(eigeneRunde && letzterDef && eigeneRunde.game === letzterDef.id);
    if (s.letzterTisch && !gleicheRunde) {
      var weiter = el("button", "ob-knopf", "Partie fortsetzen");
      weiter.type = "button";
      weiter.addEventListener("click", function () {
        tischOeffnen(s.letzterTisch.pfad, s.letzterTisch.name, true);
      });
      ui.aktionen.appendChild(weiter);
    }

    ui.chatKnopf.hidden = false;
    ui.hinweis.textContent = "Chat und Raum liegen im Connector – dieselbe Kennung wie dort.";
  }

  function zeichnePanel() {
    if (!ui.panel) return;
    zeichnePanelRunde();
    ui.panelStand.textContent = s.online
      ? "Verbunden als " + anzeigeName() + (s.raum ? " · Raum " + s.raum : "") + "."
      : (s.letzterFehler || "Noch nicht verbunden. „Mit Freunden spielen“ öffnet einen eigenen Raum.");
    ui.panelStand.dataset.art = s.letzterFehler ? "fehler" : "";
    ui.panelCode.textContent = s.raum || "—";
    if (s.raum) {
      var link = einladungsLink();
      var qrAdresse = "/api/qr?url=" + encodeURIComponent(link);
      ui.panelQr.hidden = false;
      if (ui.panelQr.getAttribute("src") !== qrAdresse) ui.panelQr.src = qrAdresse;
    } else {
      ui.panelQr.hidden = true;
      ui.panelQr.removeAttribute("src");
    }

    ui.panelSpieler.textContent = "";
    if (!s.online) {
      ui.panelSpieler.appendChild(el("div", "chat-leer", "Noch keine Verbindung."));
    } else if (!s.spieler.length) {
      ui.panelSpieler.appendChild(el("div", "chat-leer", "Noch niemand sonst im Raum."));
    } else {
      s.spieler.forEach(function (p) {
        var zeile = el("div", "zeile" + (p.id === s.ich ? " ich" : ""));
        var av = el("span", "ob-avatar", initialen(p.name));
        av.style.marginLeft = "0";
        zeile.appendChild(av);
        zeile.appendChild(el("span", "", p.name + (p.id === s.ich ? " (du)" : "")));
        var status = el("span", "status" + (p.anwesend ? "" : " weg"), p.anwesend ? "anwesend" : "weg");
        zeile.appendChild(status);
        ui.panelSpieler.appendChild(zeile);
      });
    }
  }

  function zeichnePanelRunde() {
    var block = $("panel-angebot");
    if (!block) return;
    var a = s.angebot;
    if (!a) {
      block.hidden = true;
      return;
    }
    var def = defNachId(a.game);
    var text = $("panel-angebot-text");
    var knoepfe = $("panel-angebot-knoepfe");
    knoepfe.textContent = "";
    block.hidden = false;
    if (!def) {
      text.textContent = "Offene Runde „" + (a.gameName || a.game) + "“.";
      return;
    }
    if (a.vonId === s.ich) {
      text.textContent = "Deine Runde „" + (a.gameName || def.name) + "“ wartet im Raum " +
        s.raum + " auf Mitspieler.";
      var oeffnen = el("button", "", "Tisch öffnen");
      oeffnen.type = "button";
      oeffnen.addEventListener("click", function () {
        panelSchliessen();
        tischOeffnen(def.pfad, def.name);
      });
      knoepfe.appendChild(oeffnen);
      var zurueck = el("button", "gefahr", "Zurückziehen");
      zurueck.type = "button";
      zurueck.addEventListener("click", function () {
        zurueck.disabled = true;
        angebotZurueck().then(function () {
          meldung("Runde zurückgezogen.");
          zeichnePanel();
          zeichneLeiste();
        }).catch(function (fehler) {
          zurueck.disabled = false;
          meldung(fehlerText(fehler));
        });
      });
      knoepfe.appendChild(zurueck);
    } else {
      text.textContent = a.vonName + " bietet „" + (a.gameName || def.name) +
        "“ an (" + (def.min || 2) + "–" + (def.max || 6) + " Spieler).";
      var mit = el("button", "", "Mitspielen");
      mit.type = "button";
      mit.addEventListener("click", function () {
        mit.disabled = true;
        panelSchliessen();
        angebotAnnehmen(a).catch(function (fehler) {
          mit.disabled = false;
          meldung(fehlerText(fehler));
        });
      });
      knoepfe.appendChild(mit);
      var nein = el("button", "gefahr", "Ablehnen");
      nein.type = "button";
      nein.addEventListener("click", function () {
        nein.disabled = true;
        angebotAblehnen(a).then(function () {
          zeichnePanel();
          zeichneLeiste();
        });
      });
      knoepfe.appendChild(nein);
    }
  }

  function zeichneAngebot() {
    if (!ui.angebot) return;
    var a = s.angebot;
    if (!s.online || !a || s.angebotWeg === a.id || s.tisch) {
      ui.angebot.hidden = true;
      return;
    }
    if (a.vonId === s.ich) {
      ui.angebot.hidden = true;
      if (s.tisch) return;
      // Eigene Runde: oben in der Leiste als „Partie fortsetzen“ erreichbar.
      return;
    }
    ui.angebot.textContent = "";

    var def = null;
    for (var i = 0; i < SPIELE.length; i++) {
      if (SPIELE[i].id === a.game) { def = SPIELE[i]; break; }
    }

    ui.angebot.appendChild(el("div", "wer", a.vonName + " lädt ein"));
    ui.angebot.appendChild(el("div", "spiel", a.gameName || (def ? def.name : a.game)));
    var text = def ? def.text : "Eine offene Runde im Raum.";
    if (def && def.min) text += " " + def.min + "–" + def.max + " Spieler.";
    ui.angebot.appendChild(el("div", "text", text));

    var knoepfe = el("div", "knoepfe");
    var mit = el("button", "haupt", "Mitspielen");
    mit.type = "button";
    mit.addEventListener("click", function () {
      mit.disabled = true;
      angebotAnnehmen(a).catch(function (fehler) {
        mit.disabled = false;
        meldung(fehlerText(fehler));
      });
    });
    knoepfe.appendChild(mit);

    var spaeter = el("button", "", "Ablehnen");
    spaeter.type = "button";
    spaeter.addEventListener("click", function () {
      spaeter.disabled = true;
      angebotAblehnen(a).finally(function () { spaeter.disabled = false; });
    });
    knoepfe.appendChild(spaeter);
    ui.angebot.appendChild(knoepfe);
    ui.angebot.hidden = false;
  }

  // ------------------------------------------------------------------ Raum ---

  async function raumHolen(neu) {
    var antwort = await anfrage("/api/rooms", jsonRumpf({
      id: kennung(),
      name: anzeigeName(),
      neuerRaum: !!neu
    }));
    var daten = await antwort.json();
    if (!daten || !daten.code) throw new Error("Der Connector hat keinen Raum geliefert.");
    raumMerken(daten.code);
    return daten.code;
  }

  async function raumNeu() {
    var antwort = await anfrage("/api/rooms", jsonRumpf({
      id: kennung(),
      name: anzeigeName(),
      neuerRaum: true
    }));
    var daten = await antwort.json();
    if (!daten || !daten.code) throw new Error("Kein Raumcode erhalten.");
    raumMerken(daten.code);
    s.angebot = null;
    s.angebotWeg = "";
    s.tisch = null;
    tischSchliessen();
    await verbinden(true);
    meldung("Neuer Raum " + s.raum + " – Einladung lässt sich kopieren.");
    return s.raum;
  }

  async function raumBetreten(code) {
    raumMerken(code);
    s.angebot = null;
    s.angebotWeg = "";
    s.tisch = null;
    tischSchliessen();
    await verbinden(true);
    meldung("Raum " + s.raum + " betreten.");
    return s.raum;
  }

  async function verbinden(frisch) {
    if (!eigenerName()) nameMerken(gastName());
    if (!s.raum) {
      var wunsch = urlRaum();
      raumMerken(wunsch || gespeicherterRaum());
      // Ohne Einladung entsteht ein eigener Raum: Eine Spielhalle lädt Freunde
      // ein, statt alle Besucher in einen offenen Hauptraum zu setzen.
      if (!s.raum) s.raum = await raumHolen(true);
    }
    kennung();
    var code = s.raum;
    var antwort;
    try {
      antwort = await anfrage("/api/presence?room=" + encodeURIComponent(code) +
        "&id=" + encodeURIComponent(kennung()), jsonRumpf({
          room: code,
          id: kennung(),
          name: anzeigeName(),
          neu: !!frisch
        }));
    } catch (fehler) {
      if (fehler.status === 403) {
        // Der Platz hängt an einem anderen Gerät (Token verloren): neue
        // Kennung, gleicher Name – wie im Connector, höchstens alle 30 s.
        if (Date.now() >= s.neuversuchBis) {
          s.neuversuchBis = Date.now() + 30000;
          neueKennungErzwingen();
          return verbinden(true);
        }
      }
      if (fehler.status === 404) {
        // Raum weg (Server neu gestartet): neuen Raum holen.
        raumMerken("");
        s.raum = await raumHolen(true);
        return verbinden(true);
      }
      throw fehler;
    }
    var daten = await antwort.json().catch(function () { return {}; });
    if (daten && daten.name) {
      nameMerken(daten.name);
      if (ui.nameInput) ui.nameInput.value = daten.name;
    }
    if (daten && daten.room) raumMerken(daten.room);
    s.online = true;
    s.verbindet = false;
    s.letzterFehler = "";
    await spielerLaden();
    await angebotLaden();
    zeichneLeiste();
    zeichnePanel();
    zeichneAngebot();
  }

  async function onlineGehen() {
    if (s.online) {
      panelOeffnen();
      return;
    }
    if (s.verbindet) return;
    s.verbindet = true;
    zeichneLeiste();
    try {
      await verbinden(false);
      meldung("Online – Raum " + s.raum + ".");
    } finally {
      s.verbindet = false;
      zeichneLeiste();
    }
  }

  async function nameAendern(wert) {
    var neu = String(wert || "").trim();
    if (!neu) {
      meldung("Bitte einen Namen eintragen.");
      return;
    }
    nameMerken(neu);
    if (!s.online) {
      zeichneLeiste();
      zeichnePanel();
      return;
    }
    try {
      var antwort = await anfrage("/api/presence?room=" + encodeURIComponent(s.raum) +
        "&id=" + encodeURIComponent(kennung()), jsonRumpf({
          room: s.raum, id: kennung(), name: neu, umbenennen: true
        }));
      var daten = await antwort.json().catch(function () { return {}; });
      if (daten && daten.name) nameMerken(daten.name);
      await spielerLaden();
      meldung("Name gespeichert.");
      zeichneLeiste();
      zeichnePanel();
    } catch (fehler) {
      meldung(fehlerText(fehler));
    }
  }

  function trennen() {
    s.online = false;
    s.spieler = [];
    s.angebot = null;
    s.angebotWeg = "";
    s.tisch = null;
    raumMerken("");
    tischSchliessen();
    zeichneLeiste();
    zeichnePanel();
    zeichneAngebot();
  }

  async function spielerLaden() {
    if (!s.online || !s.raum) return;
    try {
      var antwort = await anfrage("/api/players?room=" + encodeURIComponent(s.raum), { cache: "no-store" });
      var daten = await antwort.json();
      if (Array.isArray(daten)) {
        s.spieler = daten;
        s.letzterFehler = "";
      }
    } catch (fehler) {
      if (fehler.status === 404) {
        s.online = false;
        s.letzterFehler = "Der Raum ist nicht mehr erreichbar. Bitte neu verbinden.";
      } else {
        s.letzterFehler = "Verbindung gestört – neuer Versuch läuft.";
      }
    }
    zeichneLeiste();
    zeichnePanel();
  }

  async function angebotLaden() {
    if (!s.online || !s.raum) return;
    try {
      var antwort = await anfrage("/api/angebot?room=" + encodeURIComponent(s.raum), { cache: "no-store" });
      var daten = await antwort.json();
      s.angebot = daten && daten.id ? daten : null;
    } catch (fehler) {
      s.angebot = null;
    }
    zeichneAngebot();
  }

  // --------------------------------------------------------- Runde eröffnen ---

  async function angebotBieten(def) {
    if (!s.online) await onlineGehen();
    if (!s.online) throw new Error("Ohne Verbindung lässt sich keine Runde eröffnen.");
    await anfrage("/api/angebot?room=" + encodeURIComponent(s.raum), jsonRumpf({
      id: kennung(),
      action: "bieten",
      game: def.id,
      maxPlayers: def.max || 0
    }));
    s.angebotWeg = "";
    await angebotLaden();
    return true;
  }

  async function angebotZurueck() {
    if (!s.angebot) return;
    var game = s.angebot.game;
    await anfrage("/api/angebot?room=" + encodeURIComponent(s.raum), jsonRumpf({
      id: kennung(), action: "zurueck", game: game
    }));
    s.angebot = null;
    zeichneAngebot();
  }

  async function angebotAnnehmen(a) {
    var def = null;
    for (var i = 0; i < SPIELE.length; i++) {
      if (SPIELE[i].id === a.game) { def = SPIELE[i]; break; }
    }
    if (!def) throw new Error("Dieses Spiel kennt die Halle nicht.");
    await anfrage("/api/angebot?room=" + encodeURIComponent(s.raum), jsonRumpf({
      id: kennung(), action: "annehmen", game: a.game
    }));
    s.angebot = null;
    s.angebotWeg = "";
    await angebotLaden();
    tischOeffnen(def.pfad.replace(/^\/spiele\//, ""), def.name);
    return true;
  }

  async function angebotAblehnen(a) {
    try {
      await anfrage("/api/angebot?room=" + encodeURIComponent(s.raum), jsonRumpf({
        id: kennung(), action: "ablehnen", game: a.game
      }));
    } catch (e) {
      /* Zu mehreren im Raum bleibt das Angebot stehen – für mich ausblenden. */
    }
    s.angebotWeg = a.id;
    zeichneAngebot();
  }

  // ------------------------------------------------------------------ Tisch ---

  function tischAdresse(pfad, name, def) {
    var basis = "/spiele/" + String(pfad).replace(/^\/?spiele\//, "");
    return basis + "?room=" + encodeURIComponent(s.raum) +
      "&name=" + encodeURIComponent(anzeigeName()) +
      "&id=" + encodeURIComponent(kennung()) +
      "&v=" + Date.now();
  }

  function tischOeffnen(pfad, name, nurWieder) {
    if (!s.online) {
      meldung("Bitte zuerst online gehen.");
      return;
    }
    var def = defFuer(pfad);
    if (!def) {
      // Kein Connector-Spiel: ganz normal öffnen.
      location.href = pfad;
      return;
    }
    // Am Handy öffnet der Tisch auf eigener Seitenebene. Grund (am 24.11.2026
    // am Gerät gemessen, Huawei FIG-LX1, Firefox 160 und Chrome, 11 Varianten):
    // Wischgesten, die in einem iframe beginnen, lösen dort kein Scrollen aus –
    // der Rahmen bekommt ein touchstart, danach weder touchmove noch touchend
    // und keinen Scroll; das gilt auch für eine lange Textseite im selben
    // Rahmen und unabhängig von overflow, touch-action, fixiertem Overlay,
    // Grid oder der body-Sperre. Programmatisches Scrollen im Rahmen geht
    // (300 von 300 px), die Seite selbst scrollt normal. Am Rechner scrollt der
    // Rahmen einwandfrei. Deshalb am Handy dieselbe Adresse direkt laden.
    if (window.matchMedia("(max-width: 760px)").matches) {
      schreibe("tg-halle-tisch", JSON.stringify({ pfad: pfad, raum: s.raum }));
      location.href = tischAdresse(pfad, name, def);
      return;
    }
    if (s.tisch && s.tisch.pfad === pfad && !nurWieder) {
      ui.tisch.hidden = false;
      return;
    }
    s.tisch = { pfad: pfad, name: name || def.name, def: def, spiel: null, status: "" };
    s.letzterTisch = { pfad: pfad, name: name || def.name };
    schreibe("tg-halle-tisch", JSON.stringify({ pfad: pfad, raum: s.raum }));
    ui.tischName.textContent = name || def.name;
    ui.tischRaum.textContent = s.raum;
    ui.tischStand.textContent = "Tisch wird geöffnet …";
    ui.tischRahmen.src = tischAdresse(pfad, name, def);
    ui.tisch.hidden = false;
    document.body.style.overflow = "hidden";
    if (s.chatOffen) chatEinhaengen(true);
    spielstandLaden();
    zeichneLeiste();
  }

  function tischSchliessen() {
    if (!ui.tisch) return;
    ui.tisch.hidden = true;
    ui.tischRahmen.removeAttribute("src");
    s.tisch = null;
    s.spielstand = null;
    document.body.style.overflow = "";
    chatEinhaengen(false);
    zeichneLeiste();
  }

  async function spielstandLaden() {
    if (!s.tisch || !s.online || !s.tisch.def) return;
    var def = s.tisch.def;
    try {
      var antwort = await anfrage(spielApi(def) + "?room=" + encodeURIComponent(s.raum) +
        "&id=" + encodeURIComponent(kennung()), { cache: "no-store" });
      var daten = await antwort.json();
      var spiel = daten && daten.game ? daten.game : null;
      s.spielstand = spiel;
      if (spiel) {
        var leute = tischSpieler(spiel);
        var anzahl = leute.length;
        var bereit = leute.filter(function (p) { return p.ready || p.bereit; }).length;
        var status = spiel.status || "";
        var text;
        if (status === "invited") {
          text = anzahl
            ? anzahl + " am Tisch · " + bereit + " bereit – hier oder im Spiel BEREIT drücken."
            : "Einladung offen – im Chat oder im Spiel annehmen.";
        } else if (status === "finished" || status === "ende") {
          text = "Partie beendet.";
        } else if (status === "declined") {
          text = "Einladung abgelehnt.";
        } else {
          // Jeder andere Zustand heißt: es wird gespielt. Die Spiele nennen das
          // unterschiedlich (active, spiel, runde, zug …).
          // Backgammon führt im Stand keine Spielerliste, nur die beiden
          // Seiten (redId/blueId). Gezählt wird deshalb auch dort – aber nur
          // hier, damit Einladungstext und Spiele mit players/spieler/seats
          // unverändert bleiben.
          var laufend = anzahl || tischSeiten(spiel);
          text = laufend ? "Partie läuft · " + laufend + " Spieler." : "Partie läuft.";
        }
        ui.tischStand.textContent = text;
        tischAktionenZeichnen(spiel);
        // Läuft die Runde, ist das eigene Angebot erledigt – sonst stünde im
        // Raum weiter „offene Runde“, obwohl niemand mehr beitreten kann.
        if (status !== "invited" && def.art === "tisch" && s.angebot &&
            s.angebot.vonId === s.ich && s.angebot.game === def.id) {
          angebotZurueck().catch(function () {});
        }
      } else {
        ui.tischStand.textContent = "Warte auf Mitspieler …";
        tischAktionenZeichnen(null);
      }
    } catch (fehler) {
      tischAktionenZeichnen(null);
      if (fehler.status === 404) {
        ui.tischStand.textContent = "Noch keine Runde am Tisch – im Spiel eröffnen.";
      } else {
        ui.tischStand.textContent = "Verbindung gestört – versuche erneut.";
      }
    }
  }

  // Bereitschaft und Start direkt in der Halle – dieselben Endpunkte wie im
  // Spiel. Nur Spiele mit der üblichen Tischliste (Spieler mit ready) bekommen
  // die Knöpfe; alles andere bleibt in der Spielseite.
  function tischSpieler(spiel) {
    if (!spiel) return [];
    if (Array.isArray(spiel.players)) return spiel.players;
    if (Array.isArray(spiel.spieler)) return spiel.spieler;
    // Mau Mau nennt die Plätze „seats“ (id, name, count, ready). Ohne diesen
    // Zweig blieben Bereit und Start in der Halle bei Mau Mau unsichtbar,
    // obwohl die Spielseite sie kennt.
    if (Array.isArray(spiel.seats)) return spiel.seats;
    return [];
  }
  // Spiele wie Backgammon führen keine Spielerliste im Stand, sondern nur
  // zwei Seiten. Gezählt wird, was wirklich besetzt ist.
  function tischSeiten(spiel) {
    if (!spiel) return 0;
    var n = 0;
    if (spiel.redId) n++;
    if (spiel.blueId) n++;
    return n;
  }

  function tischAktionenZeichnen(spiel) {
    var box = ui.tischAktionen;
    if (!box) return;
    box.textContent = "";
    if (!spiel || spiel.status !== "invited") return;
    var leute = tischSpieler(spiel);
    var ich = null;
    for (var i = 0; i < leute.length; i++) {
      if (leute[i].id === kennung()) { ich = leute[i]; break; }
    }
    if (!ich || typeof ich.ready !== "boolean") return;

    var bereit = el("button", ich.ready ? "tb-knopf an" : "tb-knopf", ich.ready ? "Bereit ✓" : "Bereit");
    bereit.type = "button";
    bereit.title = ich.ready ? "Bereitschaft zurücknehmen" : "Bereit zum Spielen";
    bereit.addEventListener("click", function () { tischAktion("ready", bereit); });
    box.appendChild(bereit);

    var alle = leute.length >= 2 && leute.every(function (p) { return p.ready || p.bereit; });
    var gastgeber = !spiel.hostId || spiel.hostId === kennung();
    if (gastgeber) {
      var start = el("button", "tb-knopf haupt", alle ? "Spiel starten" : "Erst alle bereit");
      start.type = "button";
      start.disabled = !alle;
      start.addEventListener("click", function () { tischAktion("start", start); });
      box.appendChild(start);
    }
  }

  async function tischAktion(aktion, knopf) {
    if (!s.tisch || !s.online) return;
    if (knopf) knopf.disabled = true;
    try {
      await anfrage(spielApi(s.tisch.def) + "?room=" + encodeURIComponent(s.raum) +
        "&id=" + encodeURIComponent(kennung()) + "&action=" + encodeURIComponent(aktion),
        jsonRumpf({ id: kennung() }));
      await spielstandLaden();
    } catch (fehler) {
      ui.tischStand.textContent = fehlerText(fehler);
      if (knopf) knopf.disabled = false;
    }
  }

  // ------------------------------------------------------------------- Chat ---

  function chatEinhaengen(imSpiel) {
    if (!ui.chat) return;
    if (imSpiel) {
      ui.chat.classList.add("im-spiel");
      ui.tischKoerper.appendChild(ui.chat);
      ui.tischKoerper.classList.add("mit-chat");
    } else {
      ui.chat.classList.remove("im-spiel");
      if (ui.chat.parentNode !== document.body) document.body.appendChild(ui.chat);
      if (ui.tischKoerper) ui.tischKoerper.classList.remove("mit-chat");
    }
  }

  function chatUmschalten(offen, imSpiel) {
    if (!s.online) {
      meldung("Chat gibt es im Online-Raum.");
      return;
    }
    s.chatOffen = offen;
    if (!ui.chat) return;
    if (offen) {
      ui.chat.hidden = false;
      if (imSpiel || s.tisch) {
        ui.tisch.hidden = false;
        chatEinhaengen(true);
      } else {
        chatEinhaengen(false);
      }
      s.chatNeu = 0;
      zeichneChatZaehler();
      chatLaden(true);
      setTimeout(function () { if (ui.chatInput) ui.chatInput.focus(); }, 30);
    } else {
      ui.chat.hidden = true;
      chatEinhaengen(false);
    }
  }

  function zeichneChatZaehler() {
    if (!ui.chatKnopf) return;
    if (s.chatNeu > 0) {
      ui.chatKnopf.classList.add("hat-neu");
      ui.chatZahl.textContent = s.chatNeu > 9 ? "9+" : String(s.chatNeu);
    } else {
      ui.chatKnopf.classList.remove("hat-neu");
    }
  }

  var letzteChatSignatur = "";

  async function chatLaden(erzwingen) {
    if (!s.online || !s.raum) return;
    try {
      var antwort = await anfrage("/api/messages?room=" + encodeURIComponent(s.raum) +
        "&id=" + encodeURIComponent(kennung()), { cache: "no-store" });
      var daten = await antwort.json();
      if (!Array.isArray(daten)) return;
      var signatur = daten.length + "|" + (daten.length ? (daten[daten.length - 1].time + daten[daten.length - 1].text) : "");
      if (!erzwingen && signatur === letzteChatSignatur) return;
      letzteChatSignatur = signatur;
      if (s.chatOffen) {
        zeichneChat(daten);
      } else {
        s.chatNeu = Math.min(99, s.chatNeu + 1);
        zeichneChatZaehler();
      }
    } catch (fehler) {
      if (ui.chatStand) ui.chatStand.textContent = "nicht verbunden";
    }
  }

  function zeichneChat(nachrichten) {
    if (!ui.chatVerlauf) return;
    var warUnten = ui.chatVerlauf.scrollTop + ui.chatVerlauf.clientHeight >= ui.chatVerlauf.scrollHeight - 24;
    ui.chatVerlauf.textContent = "";
    ui.chatStand.textContent = "";
    if (!nachrichten.length) {
      ui.chatVerlauf.appendChild(el("div", "chat-leer",
        "Noch keine Nachrichten. Schreib etwas – alle im Raum sehen es."));
      return;
    }
    nachrichten.slice(-200).forEach(function (m) {
      var zeile = el("div", "chat-zeile" + (m.system ? " system" : "") + (m.senderId === s.ich ? " ich" : ""));
      if (!m.system) {
        var wer = el("span", "wer", m.name || "?");
        if (m.whisper && m.toId === s.ich) wer.textContent = (m.name || "?") + " flüstert dir";
        zeile.appendChild(wer);
      }
      var was = el("span", "was", m.text || "");
      if (m.bild) {
        var bild = el("img");
        bild.src = "connector/smilies/" + m.bild;
        bild.alt = "";
        bild.style.width = "22px";
        bild.style.height = "22px";
        bild.style.verticalAlign = "middle";
        bild.style.marginLeft = "5px";
        was.appendChild(bild);
      }
      zeile.appendChild(was);
      if (m.time) zeile.appendChild(el("span", "zeit", m.time));

      if (m.gameInvite && m.game && m.senderId !== s.ich && !s.tisch) {
        var wer2 = m.name || "Jemand";
        var box = el("div", "chat-einladung");
        var ja = el("button", "", "Mitspielen");
        ja.type = "button";
        ja.addEventListener("click", function () {
          ja.disabled = true;
          angebotAnnehmen({
            id: m.gameId || "", game: m.game, gameName: m.text || m.game, vonId: m.senderId, vonName: wer2
          }).catch(function (fehler) {
            ja.disabled = false;
            meldung(fehlerText(fehler));
          });
        });
        box.appendChild(ja);
        var nein = el("button", "nein", "Ablehnen");
        nein.type = "button";
        nein.addEventListener("click", function () {
          angebotAblehnen({ id: m.gameId || "", game: m.game }).then(function () {
            box.textContent = "Abgelehnt.";
          });
        });
        box.appendChild(nein);
        zeile.appendChild(box);
      }

      ui.chatVerlauf.appendChild(zeile);
    });
    ui.chatStand.textContent = nachrichten.length + " Beiträge";
    if (warUnten) ui.chatVerlauf.scrollTop = ui.chatVerlauf.scrollHeight;
  }

  async function chatSenden(text) {
    if (!s.online) throw new Error("Chat gibt es im Online-Raum.");
    var msgId = "h" + Date.now().toString(36) + Math.random().toString(36).slice(2, 8);
    await anfrage("/api/messages?room=" + encodeURIComponent(s.raum), jsonRumpf({
      id: kennung(), text: text, msgId: msgId
    }));
    await chatLaden(true);
    s.chatNeu = 0;
    zeichneChatZaehler();
  }

  // --------------------------------------------------------------- Auswahl ---

  function baueAuswahl() {
    var schleier = el("div", "blatt-schleier");
    schleier.id = "halle-auswahl";
    schleier.hidden = true;
    schleier.addEventListener("click", function (e) {
      if (e.target === schleier) auswahlSchliessen();
    });
    var blatt = el("div", "blatt");
    blatt.setAttribute("role", "dialog");
    blatt.setAttribute("aria-modal", "true");
    var h2 = el("h2", "", "Spiel");
    h2.id = "auswahl-name";
    blatt.appendChild(h2);
    var unter = el("p", "unter", "");
    unter.id = "auswahl-text";
    blatt.appendChild(unter);
    var optionen = el("div", "blatt-optionen");
    optionen.id = "auswahl-optionen";
    blatt.appendChild(optionen);
    var zu = el("button", "ob-knopf leise", "Abbrechen");
    zu.type = "button";
    zu.style.marginTop = "10px";
    zu.addEventListener("click", auswahlSchliessen);
    blatt.appendChild(zu);
    schleier.appendChild(blatt);
    document.body.appendChild(schleier);
    ui.auswahl = schleier;
    ui.auswahlName = h2;
    ui.auswahlText = unter;
    ui.auswahlOptionen = optionen;
  }

  function option(zeichen, titel, text, handler) {
    var b = el("button", "blatt-option");
    b.type = "button";
    b.appendChild(el("span", "zeichen", zeichen));
    var box = el("span", "text");
    box.appendChild(el("strong", "", titel));
    if (text) box.appendChild(el("span", "", text));
    b.appendChild(box);
    b.addEventListener("click", handler);
    return b;
  }

  function auswahl(pfad, name, text, modi) {
    var def = defFuer(pfad);
    if (!def) {
      location.href = pfad;
      return;
    }
    // Welche Modi das Spiel laut Spieleliste wirklich kann. Ohne Angabe
    // bleibt der lokale Einstieg erhalten, damit der Aufruf nie blockiert.
    modi = modi || [];
    var lokalMoeglich = !modi.length ||
      modi.indexOf("solo") >= 0 || modi.indexOf("lokal") >= 0 || modi.indexOf("computer") >= 0;
    var onlineMoeglich = !modi.length || modi.indexOf("online") >= 0;

    ui.auswahlName.textContent = name || def.name;
    ui.auswahlText.textContent = text || def.text || "";
    ui.auswahlOptionen.textContent = "";

    if (lokalMoeglich) {
      var lokalText;
      if (modi.indexOf("solo") >= 0 && modi.indexOf("computer") >= 0) {
        lokalText = "Einzelspieler oder gegen Bots – ohne Raum.";
      } else if (modi.indexOf("solo") >= 0) {
        lokalText = "Einzelspieler – ohne Raum, sofort los.";
      } else if (modi.indexOf("computer") >= 0) {
        lokalText = "Gegen Bots – ohne Raum.";
      } else {
        lokalText = "Mehrspieler am selben Gerät – ohne Raum.";
      }
      ui.auswahlOptionen.appendChild(option("▶", "Jetzt spielen", lokalText,
        function () {
          location.href = pfad;
        }));
    }

    if (onlineMoeglich) {
      var laufend = s.angebot && s.angebot.game === def.id && s.angebot.vonId !== s.ich;
      var eigen = s.angebot && s.angebot.game === def.id && s.angebot.vonId === s.ich;
      ui.auswahlOptionen.appendChild(option("◉",
        eigen ? "Eigene Runde öffnen" : (laufend ? "Runde beitreten" : "Online-Runde"),
        s.online
          ? (laufend ? "„" + s.angebot.vonName + "“ wartet am Tisch." :
             "Raum " + s.raum + " · " + (def.min || 2) + "–" + (def.max || 6) + " Spieler, Chat inklusive.")
          : "Öffnet den Online-Raum, lädt Freunde ein und startet am Tisch.",
        function () {
          auswahlSchliessen();
          onlineSpielen(def, name || def.name, laufend, eigen);
        }));
    }

    ui.auswahl.hidden = false;
    setTimeout(function () {
      var erster = ui.auswahlOptionen.querySelector("button");
      if (erster) erster.focus();
    }, 30);
  }

  function auswahlSchliessen() {
    if (ui.auswahl) ui.auswahl.hidden = true;
  }

  async function partielaeuft(def) {
    if (!s.online || !s.raum) return false;
    try {
      var antwort = await anfrage(spielApi(def) + "?room=" + encodeURIComponent(s.raum) +
        "&id=" + encodeURIComponent(kennung()), { cache: "no-store" });
      var daten = await antwort.json();
      var spiel = daten && daten.game;
      if (!spiel) return false;
      if (spiel.status === "finished") return false;
      return Array.isArray(spiel.players) &&
        spiel.players.some(function (p) { return p.id === s.ich; });
    } catch (e) {
      return false;
    }
  }

  async function onlineSpielen(def, name, laufend, eigen) {
    try {
      if (!s.online) {
        await onlineGehen();
        if (!s.online) return;
      }
      if (eigen || (s.angebot && s.angebot.game === def.id && s.angebot.vonId === s.ich)) {
        tischOeffnen(def.pfad, name);
        return;
      }
      if (laufend) {
        await angebotAnnehmen(s.angebot);
        return;
      }
      // Läuft die Partie im Raum schon, führt der Weg zurück an denselben
      // Tisch statt zu einer zweiten Runde.
      if (await partielaeuft(def)) {
        tischOeffnen(def.pfad, name);
        return;
      }
      if (s.angebot && s.angebot.game !== def.id) {
        meldung("Im Raum läuft schon eine Runde „" + (s.angebot.gameName || s.angebot.game) + "“.");
      }
      await angebotBieten(def);
      tischOeffnen(def.pfad, name);
      meldung("Runde eröffnet – Einladen zeigt Link und Raumcode.");
    } catch (fehler) {
      meldung(fehlerText(fehler));
    }
  }

  // --------------------------------------------------------------- Schleifen ---

  var takt = 0;
  var taktLaeuft = false;

  async function taktSchritt() {
    if (taktLaeuft) return;
    taktLaeuft = true;
    takt++;
    try {
      if (!s.online) return;
      if (takt % PRESENCE_ALLE === 0) {
        try {
          await anfrage("/api/presence?room=" + encodeURIComponent(s.raum) +
            "&id=" + encodeURIComponent(kennung()), jsonRumpf({
              room: s.raum, id: kennung(), name: anzeigeName(), active: false
            }));
          s.letzterFehler = "";
        } catch (fehler) {
          if (fehler.status === 403 || fehler.status === 409) {
            s.online = false;
            await verbinden(true).catch(function () {});
          } else {
            s.letzterFehler = "Verbindung gestört – neuer Versuch läuft.";
          }
          zeichneLeiste();
        }
      }
      if (takt % SPIELER_ALLE === 0) await spielerLaden();
      if (takt % ANGEBOT_ALLE === 0) await angebotLaden();
      if (s.chatOffen || s.chatNeu) await chatLaden(false);
      if (s.tisch) await spielstandLaden();
    } finally {
      taktLaeuft = false;
    }
  }

  // ------------------------------------------------------------------ Start ---

  function panelOeffnen() {
    s.panelOffen = true;
    if (ui.panel) ui.panel.hidden = false;
    zeichnePanel();
    if (ui.nameInput) {
      ui.nameInput.value = eigenerName();
      setTimeout(function () { ui.nameInput.focus(); }, 40);
    }
  }

  function panelSchliessen() {
    s.panelOffen = false;
    if (ui.panel) ui.panel.hidden = true;
  }

  function tasten(e) {
    if (e.key !== "Escape") return;
    if (ui.auswahl && !ui.auswahl.hidden) { auswahlSchliessen(); return; }
    if (ui.panel && !ui.panel.hidden) { panelSchliessen(); return; }
    if (ui.chat && !ui.chat.hidden) { chatUmschalten(false); return; }
    if (s.tisch) tischSchliessen();
  }

  function start() {
    baueLeiste();
    bauePanel();
    baueChat();
    baueAngebot();
    baueTisch();
    baueAuswahl();
    zeichneLeiste();
    zeichneChatZaehler();
    document.addEventListener("keydown", tasten);

    // Wiederaufnahme: gespeicherter Raum und ggf. Tisch.
    var raum = urlRaum() || gespeicherterRaum();
    if (raum) {
      raumMerken(raum);
      verbinden(false).catch(function (fehler) {
        s.letzterFehler = fehlerText(fehler);
        zeichneLeiste();
        zeichnePanel();
      });
      try {
        var tisch = JSON.parse(lies("tg-halle-tisch") || "null");
        if (tisch && tisch.pfad && tisch.raum === raum) {
          var def = defFuer(tisch.pfad);
          s.tisch = {
            pfad: tisch.pfad,
            name: (def || {}).name || "Partie",
            def: def,
            spiel: null,
            status: ""
          };
          s.letzterTisch = { pfad: tisch.pfad, name: s.tisch.name };
        }
      } catch (e) { s.tisch = null; }
    }

    setInterval(taktSchritt, POLL);

    document.addEventListener("visibilitychange", function () {
      if (document.visibilityState === "visible") {
        if (s.online) {
          spielerLaden();
          angebotLaden();
          if (s.chatOffen) chatLaden(true);
        }
      }
    });
  }

  window.TGHalle = {
    start: start,
    auswahl: auswahl,
    tischOeffnen: tischOeffnen,
    onlineGehen: onlineGehen,
    chat: function () { chatUmschalten(true); },
    stand: function () { return { online: s.online, raum: s.raum, spieler: s.spieler.length }; }
  };

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", start);
  } else {
    start();
  }
})();
