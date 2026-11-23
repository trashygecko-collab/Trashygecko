/*
 * Offline-Speicher für T.Spiele (iPhone/Web-App).
 *
 * Liegt eine Ebene über ios/, damit er auch die Spielseiten unter
 * /spiele/brettspiele/… bedienen darf. Ein Dienst direkt in ios/ hätte nur
 * diesen einen Ordner im Griff und Mahjong liefe offline nicht.
 *
 * Die Dateiliste und VERSION schreibt ../arbeit/werkzeuge/pwa-cache-liste.mjs.
 * VERSION ist ein Fingerabdruck der mitgespeicherten Dateien: ändert sich
 * nichts, bleibt sie gleich und die App behält ihre Kopie.
 */

const VERSION = "f2a2373c";
const CACHE = "tspiele-" + VERSION;
const START = "ios/ios.html";

// >>> DATEIEN (erzeugt von werkzeuge/pwa-cache-liste.mjs)
const DATEIEN = [
    "brettspiele/mahjong/assets/stein (1).webp",
    "brettspiele/mahjong/assets/stein (10).webp",
    "brettspiele/mahjong/assets/stein (11).webp",
    "brettspiele/mahjong/assets/stein (12).webp",
    "brettspiele/mahjong/assets/stein (13).webp",
    "brettspiele/mahjong/assets/stein (14).webp",
    "brettspiele/mahjong/assets/stein (15).webp",
    "brettspiele/mahjong/assets/stein (16).webp",
    "brettspiele/mahjong/assets/stein (17).webp",
    "brettspiele/mahjong/assets/stein (18).webp",
    "brettspiele/mahjong/assets/stein (19).webp",
    "brettspiele/mahjong/assets/stein (2).webp",
    "brettspiele/mahjong/assets/stein (20).webp",
    "brettspiele/mahjong/assets/stein (21).webp",
    "brettspiele/mahjong/assets/stein (22).webp",
    "brettspiele/mahjong/assets/stein (23).webp",
    "brettspiele/mahjong/assets/stein (24).webp",
    "brettspiele/mahjong/assets/stein (25).webp",
    "brettspiele/mahjong/assets/stein (26).webp",
    "brettspiele/mahjong/assets/stein (27).webp",
    "brettspiele/mahjong/assets/stein (28).webp",
    "brettspiele/mahjong/assets/stein (29).webp",
    "brettspiele/mahjong/assets/stein (3).webp",
    "brettspiele/mahjong/assets/stein (30).webp",
    "brettspiele/mahjong/assets/stein (31).webp",
    "brettspiele/mahjong/assets/stein (32).webp",
    "brettspiele/mahjong/assets/stein (33).webp",
    "brettspiele/mahjong/assets/stein (34).webp",
    "brettspiele/mahjong/assets/stein (35).webp",
    "brettspiele/mahjong/assets/stein (36).webp",
    "brettspiele/mahjong/assets/stein (37).webp",
    "brettspiele/mahjong/assets/stein (38).webp",
    "brettspiele/mahjong/assets/stein (39).webp",
    "brettspiele/mahjong/assets/stein (4).webp",
    "brettspiele/mahjong/assets/stein (40).webp",
    "brettspiele/mahjong/assets/stein (41).webp",
    "brettspiele/mahjong/assets/stein (42).webp",
    "brettspiele/mahjong/assets/stein (43).webp",
    "brettspiele/mahjong/assets/stein (44).webp",
    "brettspiele/mahjong/assets/stein (45).webp",
    "brettspiele/mahjong/assets/stein (46).webp",
    "brettspiele/mahjong/assets/stein (47).webp",
    "brettspiele/mahjong/assets/stein (48).webp",
    "brettspiele/mahjong/assets/stein (49).webp",
    "brettspiele/mahjong/assets/stein (5).webp",
    "brettspiele/mahjong/assets/stein (50).webp",
    "brettspiele/mahjong/assets/stein (51).webp",
    "brettspiele/mahjong/assets/stein (52).webp",
    "brettspiele/mahjong/assets/stein (53).webp",
    "brettspiele/mahjong/assets/stein (54).webp",
    "brettspiele/mahjong/assets/stein (55).webp",
    "brettspiele/mahjong/assets/stein (56).webp",
    "brettspiele/mahjong/assets/stein (57).webp",
    "brettspiele/mahjong/assets/stein (58).webp",
    "brettspiele/mahjong/assets/stein (59).webp",
    "brettspiele/mahjong/assets/stein (6).webp",
    "brettspiele/mahjong/assets/stein (60).webp",
    "brettspiele/mahjong/assets/stein (61).webp",
    "brettspiele/mahjong/assets/stein (62).webp",
    "brettspiele/mahjong/assets/stein (63).webp",
    "brettspiele/mahjong/assets/stein (64).webp",
    "brettspiele/mahjong/assets/stein (65).webp",
    "brettspiele/mahjong/assets/stein (66).webp",
    "brettspiele/mahjong/assets/stein (67).webp",
    "brettspiele/mahjong/assets/stein (68).webp",
    "brettspiele/mahjong/assets/stein (69).webp",
    "brettspiele/mahjong/assets/stein (7).webp",
    "brettspiele/mahjong/assets/stein (70).webp",
    "brettspiele/mahjong/assets/stein (71).webp",
    "brettspiele/mahjong/assets/stein (72).webp",
    "brettspiele/mahjong/assets/stein (8).webp",
    "brettspiele/mahjong/assets/stein (9).webp",
    "brettspiele/mahjong/assets/stein-alt-dunkel.webp",
    "brettspiele/mahjong/assets/stein-alt-hell.webp",
    "brettspiele/mahjong/assets/stein-alt-kanten-hell.webp",
    "brettspiele/mahjong/assets/stein-alt-kanten-kuehl.webp",
    "brettspiele/mahjong/assets/stein-alt-kanten.webp",
    "brettspiele/mahjong/assets/stein-alt-korn.png",
    "brettspiele/mahjong/css/fonts/atkinson-hyperlegible-400.woff2",
    "brettspiele/mahjong/css/fonts/atkinson-hyperlegible-700.woff2",
    "brettspiele/mahjong/css/mahjong.css",
    "brettspiele/mahjong/js/board-layout-100.js",
    "brettspiele/mahjong/js/board-layout-2.js",
    "brettspiele/mahjong/js/board-layout-schwer.js",
    "brettspiele/mahjong/js/board-layout.js",
    "brettspiele/mahjong/js/board.js",
    "brettspiele/mahjong/js/dialoge.js",
    "brettspiele/mahjong/js/generator.js",
    "brettspiele/mahjong/js/gewonnen.js",
    "brettspiele/mahjong/js/groesse.js",
    "brettspiele/mahjong/js/hud.js",
    "brettspiele/mahjong/js/main.js",
    "brettspiele/mahjong/js/pech.js",
    "brettspiele/mahjong/js/renderer.js",
    "brettspiele/mahjong/js/storage.js",
    "brettspiele/mahjong/mahjong.html",
    "brettspiele/risiko/risiko-fern.js",
    "brettspiele/risiko/risiko.html",
    "ios/app.webmanifest",
    "ios/icons/icon-180.png",
    "ios/icons/icon-192.png",
    "ios/icons/icon-512.png",
    "ios/ios.html",
    "lob.js"
];
// <<< DATEIEN

// Beim ersten Start alles herunterladen, was offline gebraucht wird.
self.addEventListener("install", (ereignis) => {
    ereignis.waitUntil(
        caches.open(CACHE)
            .then((speicher) => speicher.addAll(DATEIEN))
            .then(() => self.skipWaiting())
    );
});

// Alte Fassungen wegräumen, sobald die neue übernimmt.
self.addEventListener("activate", (ereignis) => {
    ereignis.waitUntil(
        caches.keys()
            .then((namen) => Promise.all(
                namen.filter((name) => name !== CACHE).map((name) => caches.delete(name))
            ))
            .then(() => self.clients.claim())
    );
});

self.addEventListener("fetch", (ereignis) => {
    const anfrage = ereignis.request;

    // Nur eigene Seiten, nur lesend. Fremde Adressen (Connector, Homepage)
    // gehen unverändert ins Netz.
    if (anfrage.method !== "GET") return;
    if (new URL(anfrage.url).origin !== self.location.origin) return;

    ereignis.respondWith(
        caches.match(anfrage, { ignoreSearch: true }).then((treffer) => {
            if (treffer) return treffer;

            return fetch(anfrage).catch(() =>
                anfrage.mode === "navigate"
                    ? caches.match(START).then((start) => start || new Response(
                        "<!DOCTYPE html><meta charset=\"utf-8\"><title>Offline</title>" +
                        "<body style=\"font-family:Georgia,serif;background:#0a211b;color:#efe6d2;" +
                        "display:flex;align-items:center;justify-content:center;height:100vh;margin:0\">" +
                        "Diese Seite ist nicht offline gespeichert.</body>",
                        { status: 200, headers: { "Content-Type": "text/html; charset=utf-8" } }
                    ))
                    : new Response("", { status: 504, statusText: "offline" })
            );
        })
    );
});
