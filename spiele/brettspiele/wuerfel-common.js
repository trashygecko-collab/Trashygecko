/* Gemeinsame Bausteine für die drei Würfelspiele (Schlangen und Leitern, Pig,
   Gänsespiel): Würfelbild, Wurf-Animation, Farben und der Zugriff auf den
   Connector. Die Regeln selbst stehen in den jeweiligen Spielseiten. */
window.WG = (() => {
  "use strict";

  // Punktemuster je Augenzahl (3×3-Raster, Feld 5 ist die Mitte).
  const MUSTER = [[], [4], [0, 8], [0, 4, 8], [0, 2, 6, 8], [0, 2, 4, 6, 8], [0, 2, 3, 5, 6, 8]];

  // Farben für bis zu vier Mitspieler – gedeckt, aber gut unterscheidbar.
  const FARBEN = ["#e8cd8e", "#7fb98a", "#d98878", "#8aa9d6"];
  const FARBNAMEN = ["Gold", "Grün", "Rot", "Blau"];

  function wuerfelEl(wert) {
    const d = document.createElement("span");
    d.className = "wuerfel";
    d.dataset.wert = String(wert || 0);
    for (let i = 0; i < 9; i++) {
      const p = document.createElement("i");
      if (wert && MUSTER[wert].includes(i)) p.className = "an";
      d.appendChild(p);
    }
    return d;
  }

  function setzeWuerfel(d, wert) {
    d.dataset.wert = String(wert);
    [...d.children].forEach((p, i) => {
      p.className = wert && MUSTER[wert].includes(i) ? "an" : "";
    });
  }

  const warte = (ms) => new Promise((r) => setTimeout(r, ms));

  // Zeigt kurz wirbelnde Augenzahlen und dann das Ergebnis.
  async function wuerfeln(d, ziel, dauer) {
    const lauf = dauer || 620;
    const ende = Date.now() + lauf;
    d.classList.add("rollt");
    while (Date.now() < ende) {
      setzeWuerfel(d, 1 + Math.floor(Math.random() * 6));
      await warte(70);
    }
    setzeWuerfel(d, ziel);
    d.classList.remove("rollt");
    d.classList.add("fertig");
    setTimeout(() => d.classList.remove("fertig"), 320);
  }

  // Connector-Anbindung: dieselben Rufe wie in den anderen Spielen.
  function netz({ pfad, room, id, melde }) {
    let busy = false;
    const fehler = (text) => { if (melde) melde(text); return { ok: false, game: null, fehler: text }; };

    async function ruf(aktion, koerper) {
      const url = pfad + "?room=" + encodeURIComponent(room) +
        (aktion ? "&action=" + encodeURIComponent(aktion) : "&id=" + encodeURIComponent(id));
      const r = await fetch(url, aktion ? {
        method: "POST", headers: { "Content-Type": "application/json" },
        body: JSON.stringify(Object.assign({ id }, koerper || {}))
      } : { cache: "no-store" });
      if (!r.ok) throw new Error((await r.text()).trim() || "Aktion fehlgeschlagen.");
      return (await r.json()).game || null;
    }

    async function hole() {
      if (busy) return { ok: false, game: null };
      busy = true;
      try {
        return { ok: true, game: await ruf(null) };
      } catch (e) {
        return fehler(e.message);
      } finally {
        busy = false;
      }
    }

    async function schicke(aktion, koerper) {
      // Läuft gerade eine Abfrage, wird der Klick nicht verworfen, sondern
      // nachgeholt, sobald der Platz frei ist.
      for (let i = 0; i < 60 && busy; i++) await warte(50);
      if (busy) return fehler("Der Server antwortet nicht – bitte noch einmal tippen.");
      busy = true;
      try {
        return { ok: true, game: await ruf(aktion, koerper) };
      } catch (e) {
        return fehler(e.message);
      } finally {
        busy = false;
      }
    }

    return { hole, schicke };
  }

  // Ein kleines Feld mit Figur in Spielerfarbe.
  function figur(farbe, name) {
    const f = document.createElement("span");
    f.className = "figur";
    f.style.setProperty("--farbe", FARBEN[farbe % FARBEN.length]);
    f.title = name || "";
    return f;
  }

  return { wuerfelEl, setzeWuerfel, wuerfeln, warte, netz, figur, FARBEN, FARBNAMEN };
})();
