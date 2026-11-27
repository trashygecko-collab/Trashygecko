/* Pong im Connector – die Brücke zwischen servergeführter Partie und der
   Spiel-API 2.0. Das Spiel selbst ruft kein /api/ auf und liest keine
   Raumkennung: Hier werden Tisch und Stand geholt, hineingeschoben und die
   Schlägerbewegung des Spielers zurückgeschickt. Nur bei modus=online aktiv. */
(function () {
 "use strict";

 const TAKT_SPIEL = 140;
 const TAKT_LOBBY = 850;
 const SENDEN_MIN = 60;

 let api = null;
 let fenster = null;
 let taktTimer = 0;
 let taktLaeuft = false;
 let letzteSende = 0;
 let sendeLaeuft = false;
 let gehoert = new WeakSet();
 let letzteFehlermeldung = "";

 function raumId() {
  try { return activeRoom || ""; } catch (e) { return window.activeRoom || ""; }
}

function spielerId() {
  try { return sessionID || ""; } catch (e) { return window.sessionID || ""; }
}

function rahmen() {
  return document.getElementById("spielRahmen");
 }

 function pongSpiel() {
  const r = rahmen();
  if (!r) return null;
  let pfad = "", suche = "";
  try {
   const w = r.contentWindow;
   if (!w) return null;
   pfad = w.location.pathname || "";
   suche = w.location.search || "";
   if (!/\/spiele\/pong\/pong\.html$/.test(pfad)) return null;
   const p = new URLSearchParams(suche);
   if ((p.get("modus") || "").toLowerCase() !== "online") return null;
   const inhalt = w.Pong;
   if (!inhalt || !inhalt.bereit || typeof inhalt.start !== "function") return null;
   return { api: inhalt, fenster: w };
  } catch (e) {
   return null;
  }
 }

 function thema() {
  try {
   /* Die zentrale Connector-Wahl (TGStil, localStorage "tg-stil") gilt auch
      für Pong. Das Spiel selbst liest sie beim direkten Start; im Connector
      muss sie ausdrücklich übergeben werden, damit sie Vorrang hat. */
   if (window.TGStil && typeof TGStil.lokal === "function") {
    const s = TGStil.lokal({ klassisch: "normal", maedchen: "maedchen", buben: "buben" });
    if (["normal", "maedchen", "buben"].indexOf(s) >= 0) return s;
   }
   if (typeof stilFuerSpiel === "function") return stilFuerSpiel();
  } catch (e) {}
  return "normal";
 }

 async function ruf(aktion, daten) {
  const raum = raumId(), spieler = spielerId();
  if (!raum || !spieler) throw new Error("Kein Raum");
  const url = "/api/pong?room=" + encodeURIComponent(raum) +
   "&id=" + encodeURIComponent(spieler) +
   (aktion ? "&action=" + encodeURIComponent(aktion) : "");
  const opt = { method: aktion ? "POST" : "GET", cache: "no-store" };
  if (aktion) {
   opt.headers = { "Content-Type": "application/json" };
   opt.body = JSON.stringify(Object.assign({ id: spieler }, daten || {}));
  }
  const r = await fetch(url, opt);
  const text = await r.text();
  let d = null;
  try { d = text ? JSON.parse(text) : null; } catch (e) {}
  if (!r.ok) {
   const meldung = (d && (d.fehler || d.error)) || text || ("Fehler " + r.status);
   const fehler = new Error(String(meldung).trim());
   fehler.status = r.status;
   throw fehler;
  }
  return d;
 }

 function melden(text) {
  const t = String(text || "");
  if (t === letzteFehlermeldung) return;
  letzteFehlermeldung = t;
  try { api.aktion("verbindung", { text: t }); } catch (e) {}
 }

 function uebernehmen(game, spiel) {
  if (!game) return;
  const ziel = spiel || pongSpiel();
  if (!ziel) return;
  try {
   const z = ziel.api.zustand();
   if (!z.online) ziel.api.start({ stil: thema(), online: {} });
   ziel.api.aktion("stand", game);
   letzteFehlermeldung = "";
  } catch (e) {
   melden("Verbindung sucht …");
  }
 }

 function starten(spiel) {
  if (!spiel) return false;
  if (api !== spiel.api) {
   api = spiel.api;
   fenster = spiel.fenster;
   try { spiel.api.start({ stil: thema(), online: {} }); } catch (e) {}
  }
  zuhoeren(spiel);
  return true;
 }

 function erzeuge(spiel) {
  if (!spiel) return;
  ruf("create", {}).then(d => {
   if (d && d.game) uebernehmen(d.game, spiel);
   else naechster(TAKT_LOBBY);
  }).catch(e => {
   // Ein zweiter Spieler kann den Tisch im selben Moment eröffnet haben.
   if (e.status === 409 || e.status === 403) naechster(TAKT_LOBBY);
   else { melden("Pong: " + e.message); naechster(1600); }
  });
 }

 function naechster(ms) {
  clearTimeout(taktTimer);
  taktTimer = setTimeout(abgleich, ms);
 }

 async function abgleich() {
  if (taktLaeuft) return;
  const spiel = pongSpiel();
  if (!spiel) return;
  taktLaeuft = true;
  let wartezeit = TAKT_LOBBY;
  try {
   starten(spiel);
   const d = await ruf("");
   if (d && d.game) {
    uebernehmen(d.game, spiel);
    wartezeit = d.game.status === "spiel" ? TAKT_SPIEL : TAKT_LOBBY;
   } else {
    erzeuge(spiel);
    wartezeit = TAKT_LOBBY;
   }
  } catch (e) {
   if (e.status === 404) {
    erzeuge(spiel);
    wartezeit = TAKT_LOBBY;
   } else if (e.status === 403) {
    melden("Du bist nicht mehr im Raum – bitte im Connector neu einsteigen.");
    wartezeit = 1800;
   } else {
    melden("Verbindung sucht …");
    wartezeit = 900;
   }
  } finally {
   taktLaeuft = false;
   naechster(wartezeit);
  }
 }

 function aktion(name, daten, spiel) {
  const ziel = spiel || pongSpiel();
  if (!ziel) return;
  ruf(name, daten).then(d => {
   if (d && d.game) uebernehmen(d.game, ziel);
   else naechster(TAKT_LOBBY);
  }).catch(e => {
   melden("Pong: " + e.message);
   naechster(900);
  });
 }

 function schlaegerSenden(spiel, x) {
  if (!spiel || sendeLaeuft) return;
  const jetzt = performance.now();
  if (jetzt - letzteSende < SENDEN_MIN) return;
  letzteSende = jetzt;
  sendeLaeuft = true;
  ruf("schlaeger", { x: x }).then(d => {
   if (d && d.game) uebernehmen(d.game, spiel);
  }).catch(() => {}).finally(() => { sendeLaeuft = false; });
 }

 function zuhoeren(spiel) {
  if (!spiel || gehoert.has(spiel.fenster)) return;
  gehoert.add(spiel.fenster);
  spiel.fenster.addEventListener("pong:create", () => erzeuge(spiel));
  spiel.fenster.addEventListener("pong:beitreten", () => aktion("join", {}, spiel));
  spiel.fenster.addEventListener("pong:bereit", () => aktion("ready", {}, spiel));
  spiel.fenster.addEventListener("pong:start", () => aktion("start", {}, spiel));
  spiel.fenster.addEventListener("pong:neu", () => aktion("reset", {}, spiel));
  spiel.fenster.addEventListener("pong:schlaeger", ev => {
   const x = ev && ev.detail ? Number(ev.detail.x) : NaN;
   if (Number.isFinite(x)) schlaegerSenden(spiel, x);
  });
 }

 function brueckeAn() {
  const spiel = pongSpiel();
  if (!spiel) return;
  starten(spiel);
  clearTimeout(taktTimer);
  taktTimer = setTimeout(abgleich, 80);
 }

 const r = rahmen();
 if (r) r.addEventListener("load", () => setTimeout(brueckeAn, 80));
 setTimeout(brueckeAn, 250);
})();
