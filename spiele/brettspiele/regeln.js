/* Regelblatt für die Kurzzeile „So geht's“ (regeln.css).

   Der Text steht nur einmal in der Seite: das Blatt zeigt die vorhandene
   Regelliste (<details class="regeln">) im Seitenbereich. Es liest sie beim
   Öffnen und kopiert sie in ein Fenster – kein zweiter Regeltext, keine
   Änderung am Spielstand.

   Eine Seite braucht nur:
     <link rel="stylesheet" href="/spiele/brettspiele/regeln.css">
     <script src="/spiele/brettspiele/regeln.js"></script>
     <p class="so-gehts">
       <b>So geht's</b><span class="so-text">…ein Satz…</span>
       <button type="button" class="so-mehr">Spielregeln</button>
     </p>
*/
(function () {
  "use strict";

  var schleier = null;
  var ausloeser = null;

  function quelle() {
    return document.querySelector("details.regeln") || document.querySelector(".regeln");
  }

  function text(el) {
    return el ? (el.textContent || "").replace(/\s+/g, " ").trim() : "";
  }

  function schliessen() {
    if (!schleier || schleier.hidden) return;
    schleier.hidden = true;
    if (ausloeser && ausloeser.focus) ausloeser.focus();
  }

  function baue() {
    schleier = document.createElement("div");
    schleier.className = "regeln-schleier";
    schleier.hidden = true;
    schleier.setAttribute("role", "dialog");
    schleier.setAttribute("aria-modal", "true");
    schleier.setAttribute("aria-labelledby", "regelnBlattTitel");
    schleier.addEventListener("click", function (e) {
      if (e.target === schleier) schliessen();
    });

    var blatt = document.createElement("div");
    blatt.className = "regeln-blatt";

    var kopf = document.createElement("div");
    kopf.className = "regeln-kopf";
    var texte = document.createElement("div");
    var marke = document.createElement("span");
    marke.className = "regeln-marke";
    marke.textContent = "Spielregeln";
    var titel = document.createElement("h2");
    titel.id = "regelnBlattTitel";
    texte.append(marke, titel);
    var zu = document.createElement("button");
    zu.type = "button";
    zu.className = "regeln-zu";
    zu.textContent = "Schließen";
    zu.addEventListener("click", schliessen);
    kopf.append(texte, zu);

    var inhalt = document.createElement("div");
    inhalt.className = "regeln-inhalt";

    var fuss = document.createElement("div");
    fuss.className = "regeln-fuss";
    var weiter = document.createElement("button");
    weiter.type = "button";
    weiter.textContent = "Weiterspielen";
    weiter.addEventListener("click", schliessen);
    fuss.appendChild(weiter);

    blatt.append(kopf, inhalt, fuss);
    schleier.appendChild(blatt);
    document.body.appendChild(schleier);
  }

  function fuellen() {
    var q = quelle();
    if (!q || !schleier) return false;
    var liste = q.querySelector("ul");
    if (!liste) return false;
    var kopfTitel = document.querySelector("h1");
    schleier.querySelector(".regeln-kopf h2").textContent =
      text(kopfTitel) || text(q.querySelector("summary")) || "Spielregeln";
    var inhalt = schleier.querySelector(".regeln-inhalt");
    inhalt.innerHTML = "";
    inhalt.appendChild(liste.cloneNode(true));
    return true;
  }

  function oeffnen(knopf) {
    if (!quelle()) return;
    if (!schleier) baue();
    if (!fuellen()) return;
    ausloeser = knopf || null;
    schleier.hidden = false;
    var zu = schleier.querySelector(".regeln-zu");
    if (zu) zu.focus();
  }

  function verdrahten() {
    var stellen = document.querySelectorAll(".so-gehts");
    for (var i = 0; i < stellen.length; i++) {
      (function (stelle) {
        var knopf = stelle.querySelector(".so-mehr");
        if (!knopf) return;
        knopf.setAttribute("aria-label", "Spielregeln ausführlich anzeigen");
        knopf.addEventListener("click", function () { oeffnen(knopf); });
      })(stellen[i]);
    }
  }

  /* Solange das Blatt offen ist, gehören Leertaste und Eingabetaste dem
     Lesen, nicht dem Spiel (10.000 würfelt sonst mit der Leertaste). */
  document.addEventListener("keydown", function (e) {
    if (!schleier || schleier.hidden) return;
    if (e.key === "Escape" || e.code === "Escape") {
      e.preventDefault();
      e.stopPropagation();
      schliessen();
      return;
    }
    if (e.key === " " || e.key === "Spacebar" || e.key === "Enter") {
      if (!schleier.contains(e.target)) {
        e.preventDefault();
        e.stopPropagation();
      }
    }
  }, true);

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", verdrahten);
  } else {
    verdrahten();
  }
})();
