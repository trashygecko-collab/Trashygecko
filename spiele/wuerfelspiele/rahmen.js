/* TrashyGecko Würfelspiele – gemeinsamer Rahmen.
   ---------------------------------------------------------------------------
   Hier steckt die Oberfläche: Spieler, Zugwechsel, Würfelbild, Tasten, Punkte,
   Verlauf, Neustart. Die Regeln jedes Spiels liegen in einer eigenen Datei
   unter spiele/ und melden sich mit TGW.anmelden({...}) an.

   Ein Regelmodul liefert:
     id, titel, unter, regeln:[...], spieler:{min,max,standard}, varianten:[{id,text}]
     start(z)                 – Anfangszustand füllen (Spieler sind schon da)
     sicht(z)                 – was angezeigt wird (siehe zeichne)
     aktion(z, art, daten)    – "wurf" | "weiter" | "waehlen" | eigener Artname

   Für wiederkehrende Abläufe gibt es zwei Spiel-Bausteine weiter unten im
   selben Skript: TGW.plusMinusRunden (vier Würfel, einige davon rot) und
   TGW.wurfDannWahl (würfeln, dann eine Möglichkeit wählen). Sie stehen hier,
   weil rahmen.js das einzige Skript ist, das jede Seite vor ihrem Regelmodul
   lädt.

   Der Rahmen stellt im Zustand bereit:
     z.spieler[]  {name, punkte, …eigene Felder}
     z.dran       Index des Spielers am Zug
     z.wuerfel    zuletzt gewürfelte Zahlen
     z.ende       null oder {sieger:[…], text}
     z.wurf(n)    würfelt n Zahlen, liefert sie und schreibt sie nach z.wuerfel
     z.log        Verlaufszeilen (max. 8)
   --------------------------------------------------------------------------- */
(function () {
  "use strict";

  var TGW = (window.TGW = window.TGW || { spiele: {}, reihenfolge: [] });
  TGW.anmelden = function (def) {
    TGW.spiele[def.id] = def;
    TGW.reihenfolge.push(def.id);
  };

  /* --- Zufall --------------------------------------------------------- */
  function rngAus(saat) {
    var a = saat >>> 0;
    return function () {
      a = (a + 0x6d2b79f5) >>> 0;
      var t = a;
      t = Math.imul(t ^ (t >>> 15), t | 1);
      t ^= t + Math.imul(t ^ (t >>> 7), t | 61);
      return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
    };
  }
  TGW.rngAus = rngAus;

  TGW.wurf = function (z, n) {
    var werte = [];
    for (var i = 0; i < n; i++) werte.push(1 + Math.floor(z.rng() * 6));
    z.wuerfel = werte;
    return werte;
  };
  TGW.naechster = function (z) {
    z.dran = (z.dran + 1) % z.spieler.length;
  };
  TGW.sieger = function (z, indizes, text) {
    z.ende = { sieger: indizes.slice(), text: text || "" };
  };
  TGW.hoechste = function (werte) {
    var max = -Infinity;
    werte.forEach(function (w) { if (w > max) max = w; });
    var siege = [];
    werte.forEach(function (w, i) { if (w === max) siege.push(i); });
    return { max: max, siege: siege };
  };
  TGW.note = function (z, text) {
    z.log.unshift(text);
    if (z.log.length > 8) z.log.length = 8;
  };
  /* Höchster Wert gewinnt, sobald alle fertig sind. Liefert true, wenn das
     Spiel damit entschieden ist. */
  TGW.endeNachHoechstem = function (z, feld, was) {
    if (!z.spieler.every(function (p) { return p.fertig === true; })) return false;
    var r = TGW.hoechste(z.spieler.map(feld));
    TGW.sieger(z, r.siege, (was || "Höchste Summe") + ": " + r.max);
    return true;
  };

  /* --- Spiel-Bausteine --------------------------------------------------
     Die zwei Vorlagen unten bauen ein ganzes Regelmodul; sie gehören zur
     Regel-Seite, nicht zur Oberfläche. */

  /* Vier Würfel, `minus` davon rot, über fünf Runden – Pärchen und Beelzebub
     unterscheiden sich nur in der Zahl der Minuswürfel. */
  TGW.plusMinusRunden = function (def) {
    var RUNDEN = 5;
    var plus = 4 - def.minus;

    function wert(w) {
      var s = 0;
      for (var i = 0; i < w.length; i++) s += i < plus ? w[i] : -w[i];
      return s;
    }
    function rechnung(w) {
      return w.map(function (x, i) {
        return (i === 0 ? "" : i < plus ? " + " : " − ") + x;
      }).join("");
    }

    return {
      id: def.id,
      titel: def.titel,
      unter: def.unter,
      regeln: def.regeln,
      spieler: { min: 1, max: 6, standard: 2 },
      start: function (z) {
        z.spieler.forEach(function (p) { p.punkte = 0; p.runde = 0; p.fertig = false; p.letzte = null; });
      },
      sicht: function (z) {
        var p = z.spieler[z.dran];
        var roh = z.wuerfel || [];
        return {
          wuerfel: roh.map(function (w, i) { return { wert: w, negativ: i >= plus }; }),
          summe: roh.length === 4 ? wert(roh) : null,
          summeText: "Runde",
          hinweis: "Runde " + (p.runde + 1) + " von " + RUNDEN,
          klein: def.klein,
          knopf: { text: "Würfeln", art: "wurf" },
          spieler: z.spieler.map(function (q) {
            return {
              name: q.name,
              wert: q.punkte,
              neben: q.fertig ? "fertig" : "Runde " + (q.runde + 1) + "/" + RUNDEN
            };
          }),
          tabelle: {
            titel: "Runden",
            spalten: ["Spieler", "Letzte Runde", "Gesamt"],
            zeilen: z.spieler.map(function (q) {
              return [q.name, q.letzte === null ? "—" : (q.letzte > 0 ? "+" : "") + q.letzte, q.punkte];
            })
          }
        };
      },
      aktion: function (z, art) {
        if (art !== "wurf") return false;
        var p = z.spieler[z.dran];
        var w = z.wurf(4);
        var runde = wert(w);
        p.letzte = runde;
        p.punkte += runde;
        p.runde += 1;
        TGW.note(z, p.name + ": " + rechnung(w) + " = " + runde + " (" + p.punkte + ")");
        if (p.runde >= RUNDEN) p.fertig = true;
        if (TGW.endeNachHoechstem(z, function (q) { return q.punkte; }, "Höchste Gesamtsumme")) return;
        if (p.fertig) TGW.naechster(z);
      }
    };
  };

  /* Ein Zug in zwei Schritten: erst würfeln, dann genau eine Möglichkeit aus
     der Auswahl nehmen (Hohe Hausnummer, Rentmeister). Das Modul liefert
     start, sicht, wahl, wurf und setze; die Aktionen heißen „wurf“ und
     „waehlen“. */
  TGW.wurfDannWahl = function (def) {
    return {
      id: def.id,
      titel: def.titel,
      unter: def.unter,
      regeln: def.regeln,
      spieler: { min: 1, max: 6, standard: 2 },
      start: function (z) {
        def.start(z);
        z.phase = "wurf";
      },
      sicht: function (z) {
        var s = def.sicht(z);
        if (z.phase === "wurf") s.knopf = { text: "Würfeln", art: "wurf" };
        else s.auswahl = def.wahl(z);
        return s;
      },
      aktion: function (z, art, daten) {
        if (art === "wurf") {
          if (z.phase !== "wurf") return false;
          def.wurf(z);
          z.phase = "waehlen";
          return;
        }
        if (art === "waehlen") {
          if (z.phase !== "waehlen" || def.setze(z, daten) === false) return false;
          z.phase = "wurf";
          return;
        }
        return false;
      }
    };
  };

  /* --- Würfelbild ------------------------------------------------------ */
  var AUGEN = { 1: [4], 2: [0, 8], 3: [0, 4, 8], 4: [0, 2, 6, 8], 5: [0, 2, 4, 6, 8], 6: [0, 2, 3, 5, 6, 8] };

  function wuerfelEl(wert, negativ) {
    var d = document.createElement("div");
    d.className = "wuerfel" + (negativ ? " negativ" : "");
    d.setAttribute("role", "img");
    d.setAttribute("aria-label", (negativ ? "Minuswürfel " : "Würfel ") + wert);
    var an = AUGEN[wert] || [];
    for (var i = 0; i < 9; i++) {
      var s = document.createElement("span");
      s.className = "punkt" + (an.indexOf(i) >= 0 ? " an" : "");
      d.appendChild(s);
    }
    return d;
  }
  TGW.wuerfelEl = wuerfelEl;

  /* --- Zustand --------------------------------------------------------- */
  var spiel = null;      // Regelmodul
  var z = null;          // Partiezustand
  var ruhig = false;     // prefers-reduced-motion
  var el = {};           // DOM-Verweise

  function neuerZustand(namen, varianten, saat) {
    var zustand = {
      spieler: namen.map(function (n) { return { name: n, punkte: 0 }; }),
      dran: 0,
      wuerfel: [],
      ende: null,
      log: [],
      varianten: varianten || {},
      rng: rngAus(saat)
    };
    zustand.wurf = function (n) { return TGW.wurf(zustand, n); };
    return zustand;
  }

  /* --- Start der Partie ------------------------------------------------ */
  function namenAusFeld() {
    var namen = [];
    var felder = el.namen.querySelectorAll("input");
    for (var i = 0; i < felder.length; i++) {
      var wert = felder[i].value.trim() || "Spieler " + (i + 1);
      namen.push(wert.slice(0, 14));
    }
    return namen;
  }

  function variantenAusFeld() {
    var werte = {};
    var felder = el.varianten.querySelectorAll("input[type=checkbox]");
    for (var i = 0; i < felder.length; i++) werte[felder[i].value] = felder[i].checked;
    return werte;
  }

  function startePartie() {
    var namen = namenAusFeld();
    try { localStorage.setItem("tg-wuerfel-namen", JSON.stringify(namen)); } catch (e) { /* egal */ }
    z = neuerZustand(namen, variantenAusFeld(), (Date.now() ^ (Math.random() * 0xffffffff)) >>> 0);
    spiel.start(z);
    el.einrichten.hidden = true;
    el.tafel.hidden = false;
    zeichne();
  }

  /* --- Zeichnen -------------------------------------------------------- */
  function zeichne() {
    var s = spiel.sicht(z) || {};

    /* Spielerleiste */
    el.spieler.innerHTML = "";
    var anzeige = s.spieler || z.spieler.map(function (p) {
      return { name: p.name, wert: p.punkte, neben: "" };
    });
    anzeige.forEach(function (eintrag, i) {
      var d = document.createElement("div");
      d.className = "spieler" + (i === z.dran && !z.ende ? " aktiv" : "");
      var kopf = document.createElement("div");
      kopf.className = "name";
      var n = document.createElement("span");
      n.textContent = eintrag.name != null ? eintrag.name : (z.spieler[i] ? z.spieler[i].name : "");
      var jetzt = document.createElement("span");
      jetzt.className = "jetzt";
      jetzt.textContent = i === z.dran && !z.ende ? "am Zug" : "";
      kopf.append(n, jetzt);
      var wert = document.createElement("div");
      wert.className = "wert";
      wert.textContent = eintrag.wert != null ? eintrag.wert : "";
      var neben = document.createElement("div");
      neben.className = "neben";
      neben.textContent = eintrag.neben || "";
      d.append(kopf, wert, neben);
      el.spieler.appendChild(d);
    });

    /* Würfel */
    el.wuerfel.innerHTML = "";
    (s.wuerfel || []).forEach(function (w, i) {
      var ding = typeof w === "object" && w !== null;
      var d = wuerfelEl(ding ? w.wert : w, ding && w.negativ);
      if (ding && w.waehlbar) {
        d.classList.add("waehlbar");
        if (w.gewaehlt) d.classList.add("gewaehlt");
        d.setAttribute("role", "button");
        d.tabIndex = 0;
        d.setAttribute("aria-pressed", w.gewaehlt ? "true" : "false");
        d.addEventListener("click", function () { druecken("wuerfel", { index: i }); });
        d.addEventListener("keydown", function (ereignis) {
          if (ereignis.code === "Enter" || ereignis.code === "Space") {
            ereignis.preventDefault();
            druecken("wuerfel", { index: i });
          }
        });
      }
      el.wuerfel.appendChild(d);
    });
    if (s.summe != null) {
      var summe = document.createElement("div");
      summe.className = "wurf-summe";
      summe.innerHTML = "<b></b><span></span>";
      summe.querySelector("b").textContent = s.summe;
      summe.querySelector("span").textContent = s.summeText || "Summe";
      el.wuerfel.appendChild(summe);
    }

    /* Hinweis */
    el.hinweis.innerHTML = "";
    var gross = document.createElement("div");
    gross.className = "gross";
    gross.textContent = s.hinweis || "";
    el.hinweis.appendChild(gross);
    if (s.klein) {
      var klein = document.createElement("div");
      klein.className = "klein";
      klein.textContent = s.klein;
      el.hinweis.appendChild(klein);
    }

    /* Tasten */
    el.tasten.innerHTML = "";
    function taste(knopf, klasse) {
      if (!knopf) return;
      var b = document.createElement("button");
      b.type = "button";
      b.className = "knopf " + (klasse || "");
      b.textContent = knopf.text;
      b.disabled = knopf.an === false;
      b.addEventListener("click", function () { druecken(knopf.art, knopf.daten); });
      el.tasten.appendChild(b);
    }
    if (z.ende) {
      var neu = { text: "Neues Spiel", art: "neu" };
      taste(neu, "haupt");
    } else {
      taste(s.knopf, "haupt");
      taste(s.knopf2, "");
      (s.auswahl || []).forEach(function (wahl) { taste(wahl, "wahl"); });
    }

    /* Leisten (Zahlen zum Abstreichen) */
    el.leisten.innerHTML = "";
    (s.leisten || []).forEach(function (leiste) {
      var box = document.createElement("div");
      box.className = "leiste";
      var t = document.createElement("div");
      t.className = "titel";
      t.textContent = leiste.titel;
      var reihe = document.createElement("div");
      reihe.className = "zahlen";
      (leiste.zahlen || []).forEach(function (zahl) {
        var w = document.createElement("span");
        w.className = "zahl" + (zahl.aus ? " aus" : "") + (zahl.frisch ? " frisch" : "");
        w.textContent = zahl.wert;
        reihe.appendChild(w);
      });
      box.append(t, reihe);
      el.leisten.appendChild(box);
    });

    /* Tabelle */
    el.tabelle.innerHTML = "";
    if (s.tabelle) {
      var tabelle = document.createElement("table");
      tabelle.className = "tabelle";
      if (s.tabelle.titel) {
        var cap = document.createElement("caption");
        cap.textContent = s.tabelle.titel;
        tabelle.appendChild(cap);
      }
      var thead = document.createElement("thead");
      var trh = document.createElement("tr");
      (s.tabelle.spalten || []).forEach(function (spalte, i) {
        var th = document.createElement("th");
        th.textContent = spalte;
        if (i > 0) th.className = "zahl-spalte";
        trh.appendChild(th);
      });
      thead.appendChild(trh);
      tabelle.appendChild(thead);
      var tbody = document.createElement("tbody");
      (s.tabelle.zeilen || []).forEach(function (zeile, i) {
        var tr = document.createElement("tr");
        if (i === z.dran && !z.ende) tr.className = "dran";
        zeile.forEach(function (wert, j) {
          var td = document.createElement(j === 0 ? "th" : "td");
          td.textContent = wert;
          if (j > 0) td.className = "zahl-spalte";
          tr.appendChild(td);
        });
        tbody.appendChild(tr);
      });
      tabelle.appendChild(tbody);
      el.tabelle.appendChild(tabelle);
    }

    /* Sieg */
    el.sieg.innerHTML = "";
    if (z.ende) {
      var kasten = document.createElement("div");
      kasten.className = "sieg";
      var b = document.createElement("b");
      b.textContent = z.ende.sieger.length === 1
        ? z.spieler[z.ende.sieger[0]].name + " gewinnt"
        : "Gleichstand: " + z.ende.sieger.map(function (i) { return z.spieler[i].name; }).join(", ");
      var p = document.createElement("p");
      p.textContent = z.ende.text || "";
      kasten.append(b, p);
      el.sieg.appendChild(kasten);
    }

    /* Verlauf */
    el.verlauf.innerHTML = "";
    (s.verlauf || z.log).forEach(function (zeile) {
      var li = document.createElement("li");
      li.textContent = zeile;
      el.verlauf.appendChild(li);
    });
  }

  /* --- Aktionen -------------------------------------------------------- */
  function druecken(art, daten) {
    if (art === "neu") { startePartie(); return; }
    if (!z || z.ende) return;
    var ergebnis = spiel.aktion(z, art, daten);
    if (ergebnis === false) return;
    if (art === "wurf" && !ruhig) {
      animiere();          // kurz rollen lassen, dann den echten Stand zeigen
      return;
    }
    zeichne();
  }

  function animiere() {
    var dauer = 340;
    var start = Date.now();
    var tick = setInterval(function () {
      var wuerfel = el.wuerfel.querySelectorAll(".wuerfel");
      if (!wuerfel.length || Date.now() - start > dauer) {
        clearInterval(tick);
        zeichne();
        return;
      }
      for (var i = 0; i < wuerfel.length; i++) {
        var neu = wuerfelEl(1 + Math.floor(Math.random() * 6), wuerfel[i].classList.contains("negativ"));
        neu.classList.add("rollend");
        wuerfel[i].replaceWith(neu);
        wuerfel = el.wuerfel.querySelectorAll(".wuerfel");
      }
    }, 55);
  }

  /* --- Aufbau der Seite ------------------------------------------------ */
  function baueSeite() {
    var kopf = document.createElement("header");
    kopf.className = "kopf";
    kopf.innerHTML = '<div><marke>TrashyGecko Würfelspiele</marke><h1></h1><p></p></div>' +
      '<div class="zurueck"><a href="../spiele.html">← alle Spiele</a></div>';
    kopf.querySelector("h1").textContent = spiel.titel;
    kopf.querySelector("p").textContent = spiel.unter || "";

    var rahmen = document.createElement("div");
    rahmen.className = "rahmen";
    rahmen.appendChild(kopf);

    /* Einrichten */
    el.einrichten = document.createElement("section");
    el.einrichten.className = "tafel einrichten";
    var titel = document.createElement("div");
    titel.className = "mini-titel";
    titel.textContent = "Mitspieler";
    el.einrichten.appendChild(titel);

    el.namen = document.createElement("div");
    el.namen.className = "aufbau";
    var gespeichert = [];
    try { gespeichert = JSON.parse(localStorage.getItem("tg-wuerfel-namen") || "[]"); } catch (e) { gespeichert = []; }
    var anzahl = spiel.spieler && spiel.spieler.standard ? spiel.spieler.standard : 2;
    for (var i = 0; i < anzahl; i++) el.namen.appendChild(namensZeile(i, gespeichert[i] || ""));
    el.einrichten.appendChild(el.namen);

    var steuerung = document.createElement("div");
    steuerung.className = "tasten";
    steuerung.style.justifyContent = "flex-start";
    var weniger = document.createElement("button");
    weniger.type = "button";
    weniger.className = "knopf";
    weniger.textContent = "− Spieler";
    weniger.addEventListener("click", function () {
      var felder = el.namen.querySelectorAll(".namenszeile").length;
      if (felder > (spiel.spieler.min || 1)) el.namen.lastChild.remove();
    });
    var mehr = document.createElement("button");
    mehr.type = "button";
    mehr.className = "knopf";
    mehr.textContent = "+ Spieler";
    mehr.addEventListener("click", function () {
      var felder = el.namen.querySelectorAll(".namenszeile").length;
      if (felder < (spiel.spieler.max || 6)) el.namen.appendChild(namensZeile(felder, ""));
    });
    var los = document.createElement("button");
    los.type = "button";
    los.className = "knopf haupt";
    los.textContent = "Partie starten";
    los.addEventListener("click", startePartie);
    steuerung.append(weniger, mehr, los);
    el.einrichten.appendChild(steuerung);

    el.varianten = document.createElement("div");
    el.varianten.className = "rundenzahl";
    (spiel.varianten || []).forEach(function (variante) {
      var label = document.createElement("label");
      label.style.display = "flex";
      label.style.gap = "6px";
      label.style.alignItems = "center";
      var box = document.createElement("input");
      box.type = "checkbox";
      box.value = variante.id;
      var text = document.createElement("span");
      text.textContent = variante.text;
      label.append(box, text);
      el.varianten.appendChild(label);
    });
    if ((spiel.varianten || []).length) el.einrichten.appendChild(el.varianten);

    var regelnBox = document.createElement("details");
    regelnBox.className = "regeln";
    var summary = document.createElement("summary");
    summary.textContent = "Regeln";
    var liste = document.createElement("ol");
    (spiel.regeln || []).forEach(function (zeile) {
      var li = document.createElement("li");
      li.textContent = zeile;
      liste.appendChild(li);
    });
    regelnBox.append(summary, liste);
    el.einrichten.appendChild(regelnBox);
    rahmen.appendChild(el.einrichten);

    /* Spieltafel */
    el.tafel = document.createElement("section");
    el.tafel.className = "tafel aufbau";
    el.tafel.hidden = true;
    el.spieler = document.createElement("div");
    el.spieler.className = "spieler-liste";
    el.wuerfel = document.createElement("div");
    el.wuerfel.className = "wurf-flaeche";
    el.hinweis = document.createElement("div");
    el.hinweis.className = "hinweis";
    el.tasten = document.createElement("div");
    el.tasten.className = "tasten";
    el.leisten = document.createElement("div");
    el.leisten.className = "leisten";
    el.tabelle = document.createElement("div");
    el.sieg = document.createElement("div");
    el.verlauf = document.createElement("ul");
    el.verlauf.className = "verlauf";
    el.tafel.append(el.spieler, el.wuerfel, el.hinweis, el.tasten, el.leisten, el.tabelle, el.sieg);
    var verlaufKasten = document.createElement("details");
    verlaufKasten.className = "regeln";
    var verlaufTitel = document.createElement("summary");
    verlaufTitel.textContent = "Verlauf";
    verlaufKasten.append(verlaufTitel, el.verlauf);
    el.tafel.appendChild(verlaufKasten);
    rahmen.appendChild(el.tafel);

    /* Fußzeile: Neustart */
    var fuss = document.createElement("div");
    fuss.className = "tasten";
    fuss.style.marginTop = "12px";
    var neu = document.createElement("button");
    neu.type = "button";
    neu.className = "knopf";
    neu.textContent = "Neu aufstellen";
    neu.addEventListener("click", function () {
      el.tafel.hidden = true;
      el.einrichten.hidden = false;
    });
    fuss.appendChild(neu);
    rahmen.appendChild(fuss);

    document.body.appendChild(rahmen);

    /* Tastatur: Leertaste = Haupttaste, 1–9 = Auswahl */
    document.addEventListener("keydown", function (ereignis) {
      if (!z || z.ende || el.tafel.hidden) return;
      if (ereignis.target && /input|textarea/i.test(ereignis.target.tagName)) return;
      var s = spiel.sicht(z) || {};
      if (ereignis.code === "Space" || ereignis.code === "Enter") {
        if (s.knopf) { ereignis.preventDefault(); druecken(s.knopf.art, s.knopf.daten); }
      } else if (/^Digit[1-9]$/.test(ereignis.code) && s.auswahl) {
        var wahl = s.auswahl[Number(ereignis.code.slice(5)) - 1];
        if (wahl) { ereignis.preventDefault(); druecken(wahl.art, wahl.daten); }
      }
    });
  }

  function namensZeile(nummer, wert) {
    var zeile = document.createElement("div");
    zeile.className = "namenszeile";
    var marke = document.createElement("span");
    marke.style.color = "var(--matt)";
    marke.style.fontSize = "13px";
    marke.style.minWidth = "70px";
    marke.textContent = "Spieler " + (nummer + 1);
    var feld = document.createElement("input");
    feld.type = "text";
    feld.maxLength = 14;
    feld.placeholder = "Name";
    feld.value = wert || "";
    zeile.append(marke, feld);
    return zeile;
  }

  /* --- Los ------------------------------------------------------------- */
  function los() {
    ruhig = !!(window.matchMedia && window.matchMedia("(prefers-reduced-motion: reduce)").matches);
    var id = window.TGW_SPIEL || "";
    spiel = TGW.spiele[id];
    if (!spiel) {
      document.body.textContent = "Dieses Würfelspiel konnte nicht geladen werden.";
      return;
    }
    document.title = spiel.titel + " – TrashyGecko";
    baueSeite();
  }

  if (document.readyState === "loading") document.addEventListener("DOMContentLoaded", los);
  else los();
})();
