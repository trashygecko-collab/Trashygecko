/* ===========================================================================
   HALMA – Designwelten

   Drei Farbwelten. Das Aussehen (Flächen, Rahmen, Schriften) steht in
   css/halma.css unter body[data-stil]. Hier liegen die Farben, die das Spiel
   selbst braucht: Spielfiguren (sechs klar unterscheidbare Töne) und die
   Markierungen auf dem Brett.
   =========================================================================== */
(function (HB) {
  "use strict";

  var STILE = {
    normal: {
      id: "normal", name: "Normal",
      figuren: [
        { farbe: "#f2e7cd", name: "Elfenbein" },
        { farbe: "#33383a", name: "Anthrazit" },
        { farbe: "#b3352f", name: "Karmin" },
        { farbe: "#2f5c96", name: "Königsblau" },
        { farbe: "#2f7a5c", name: "Smaragd" },
        { farbe: "#c99a3f", name: "Gold" }
      ],
      zone: "rgba(214, 178, 96, .14)",
      zoneRand: "rgba(214, 178, 96, .55)",
      mark: "rgba(214, 178, 96, .85)",
      sprung: "rgba(120, 200, 170, .9)",
      letzter: "rgba(255, 236, 190, .75)"
    },
    maedchen: {
      id: "maedchen", name: "Mädchen",
      figuren: [
        { farbe: "#f2cdd4", name: "Altrosa" },
        { farbe: "#6f2c4b", name: "Pflaume" },
        { farbe: "#a8486b", name: "Beere" },
        { farbe: "#c07f9d", name: "Mauve" },
        { farbe: "#e2a68f", name: "Roségold" },
        { farbe: "#8a5470", name: "Malve" }
      ],
      zone: "rgba(214, 150, 170, .16)",
      zoneRand: "rgba(226, 166, 143, .55)",
      mark: "rgba(232, 180, 189, .9)",
      sprung: "rgba(214, 150, 190, .9)",
      letzter: "rgba(255, 226, 232, .7)"
    },
    buben: {
      id: "buben", name: "Buben",
      figuren: [
        { farbe: "#e0242f", name: "Signalrot" },
        { farbe: "#82868e", name: "Stahl" },
        { farbe: "#cbcdd0", name: "Graphit" },
        { farbe: "#7e1f2b", name: "Bordeaux" },
        { farbe: "#d8a13c", name: "Messing" },
        { farbe: "#8f6fbf", name: "Violett" }
      ],
      zone: "rgb(239, 54, 76, .12)",
      zoneRand: "rgb(239, 54, 76, .5)",
      mark: "rgb(239, 54, 76, .85)",
      sprung: "rgb(244, 123, 137, .8)",
      letzter: "rgb(239, 54, 76, .55)"
    }
  };

  function stilSetzen(id) {
    if (!STILE[id]) id = "normal";
    document.body.dataset.stil = id;
    if (HB.Main && HB.Main.stilGewechselt) HB.Main.stilGewechselt(id);
    return id;
  }

  /* Die Wahl kommt aus der zentralen Einstellung (stil.js). */
  function gespeichert() {
    return TGStil.lokal({ klassisch: "normal", maedchen: "maedchen", buben: "buben" });
  }

  /* Heller/dunkler Ton einer Figur für den 3D-Effekt. */
  function mischen(farbe, ziel, anteil) {
    var f = hex(farbe), z = hex(ziel);
    if (!f || !z) return farbe;
    var out = "#";
    for (var i = 0; i < 3; i++) {
      var v = Math.round(f[i] + (z[i] - f[i]) * anteil);
      out += ("0" + Math.max(0, Math.min(255, v)).toString(16)).slice(-2);
    }
    return out;
  }
  function hex(farbe) {
    var m = /^#?([0-9a-f]{6})$/i.exec(farbe || "");
    if (!m) return null;
    var n = parseInt(m[1], 16);
    return [(n >> 16) & 255, (n >> 8) & 255, n & 255];
  }

  function figurStil(farbe) {
    return "--figur:" + farbe +
      ";--figur-hell:" + mischen(farbe, "#ffffff", 0.55) +
      ";--figur-dunkel:" + mischen(farbe, "#000000", 0.45) +
      ";--figur-rand:" + mischen(farbe, "#000000", 0.3);
  }

  HB.Design = {
    STILE: STILE,
    stilSetzen: stilSetzen,
    gespeichert: gespeichert,
    figurStil: figurStil,
    mischen: mischen,
    aktuell: function () { return STILE[document.body.dataset.stil] || STILE.normal; }
  };

  /* Eine neue zentrale Wahl zeichnet das Brett sofort neu. */
  if (window.TGStil) {
    TGStil.beiWechsel(function () { stilSetzen(gespeichert()); });
  }
})(window.HB = window.HB || {});
