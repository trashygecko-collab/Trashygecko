/* ==========================================================================
   TrashyGecko – Arcade Flipper · Darstellung, Steuerung, Sound
   Zeichnet den Tisch (Hauptfeld + Oberdeck), bedient Flipper und Stift.
   ========================================================================== */
"use strict";
(function () {
  var $ = function (id) { return document.getElementById(id); };
  var leinwand = $("leinwand");
  var ctx = leinwand.getContext("2d");
  var W = FL.W, H = FL.H, RK = FL.RK, KANAL = FL.KANAL, DECK = FL.DECK;

  var sp = FL.neuesSpiel(0);
  var laeuft = false;
  var pause = 0;
  var blitz = 0;
  var meldungZeit = 0;
  var ziehId = null;
  var flipperZeiger = {};
  var dpr = 1;
  var spur = {};              // Kugelindex → letzte Punkte

  /* ---------------------------------------------------------------- Farben */
  function tok(n) {
    var v = getComputedStyle(document.documentElement).getPropertyValue(n);
    return (v || "").trim() || "#888";
  }
  var F = {}, TF = {};
  function farbenLaden() {
    F = {
      glanz: tok("--fl-glanz"), s1: tok("--fl-schatten-1"), s2: tok("--fl-schatten-2"),
      s3: tok("--fl-schatten-3"), s4: tok("--fl-schatten-4"),
      stange1: tok("--fl-stange-1"), stange2: tok("--fl-stange-2"), stange3: tok("--fl-stange-3"),
      fahne1: tok("--fl-fahne-1"), fahne2: tok("--fl-fahne-2"), fahne3: tok("--fl-fahne-3"),
      fahneRand: tok("--fl-fahne-rand"), spinner: tok("--fl-spinner"), spinnerRand: tok("--fl-spinner-rand"),
      scoopAussen: tok("--fl-scoop-aussen"), scoopInnen: tok("--fl-scoop-innen"),
      rampe: tok("--fl-rampe"), rampeRand: tok("--fl-rampe-rand"), zinne: tok("--fl-zinne"),
      leitblech: tok("--fl-leitblech"), bumperRing: tok("--fl-bumper-ring"),
      bumperKante: tok("--fl-bumper-kante"), bumperKappe: tok("--fl-bumper-kappe"),
      arm1: tok("--fl-arm-1"), arm2: tok("--fl-arm-2"), arm3: tok("--fl-arm-3"),
      armRand: tok("--fl-arm-rand"), armLinie: tok("--fl-arm-linie"),
      lager: tok("--fl-lager"), lagerRand: tok("--fl-lager-rand"), lagerSchlitz: tok("--fl-lager-schlitz"),
      schraube: tok("--fl-schraube"), schraubSchlitz: tok("--fl-schraub-schlitz"),
      zierrand: tok("--fl-zierrand"), lampeAus: tok("--fl-lampe-aus"), spur: tok("--fl-spur"),
      kugelRand: tok("--fl-kugel-rand"), kugelEnde: tok("--fl-kugel-ende"), ring: tok("--fl-ring"),
      feder: tok("--fl-feder"), ladeWarn: tok("--fl-lade-warn"), lichtText: tok("--fl-licht-text"),
      punkte: tok("--punkte"), text: tok("--text")
    };
    var s = sp.tisch.satz;
    TF = {
      grund1: tok("--fl-t" + s + "-grund-1"), grund2: tok("--fl-t" + s + "-grund-2"),
      wand: tok("--fl-t" + s + "-wand"), licht: tok("--fl-t" + s + "-licht"),
      bumper1: tok("--fl-t" + s + "-bumper-1"), bumper2: tok("--fl-t" + s + "-bumper-2"),
      ziel: tok("--fl-t" + s + "-ziel"), kugel: tok("--fl-t" + s + "-kugel")
    };
  }
  function mitAlpha(hex, a) {
    var h = (hex || "").replace("#", "");
    if (h.length === 3) h = h[0] + h[0] + h[1] + h[1] + h[2] + h[2];
    if (h.length !== 6) return hex;
    var n = parseInt(h, 16);
    return "rgba(" + ((n >> 16) & 255) + "," + ((n >> 8) & 255) + "," + (n & 255) + "," + a + ")";
  }
  function rundeFlaeche(x, y, w, h, r) {
    ctx.beginPath();
    ctx.moveTo(x + r, y);
    ctx.lineTo(x + w - r, y); ctx.quadraticCurveTo(x + w, y, x + w, y + r);
    ctx.lineTo(x + w, y + h - r); ctx.quadraticCurveTo(x + w, y + h, x + w - r, y + h);
    ctx.lineTo(x + r, y + h); ctx.quadraticCurveTo(x, y + h, x, y + h - r);
    ctx.lineTo(x, y + r); ctx.quadraticCurveTo(x, y, x + r, y);
    ctx.closePath();
  }

  /* ---------------------------------------------------------------- Sound */
  var tonAn = true;
  try { tonAn = localStorage.getItem("flipper-ton") !== "aus"; } catch (e) { }
  var audio = null;
  function audioHolen() {
    if (!tonAn) return null;
    if (!audio) {
      var AC = window.AudioContext || window.webkitAudioContext;
      if (!AC) return null;
      try { audio = new AC(); } catch (e) { return null; }
    }
    if (audio.state === "suspended" && audio.resume) audio.resume();
    return audio;
  }
  function ton(freq, dauer, form, laut) {
    var a = audioHolen();
    if (!a) return;
    var o = a.createOscillator(), g = a.createGain();
    o.type = form || "square";
    o.frequency.setValueAtTime(freq, a.currentTime);
    o.frequency.exponentialRampToValueAtTime(Math.max(40, freq * 0.55), a.currentTime + dauer);
    g.gain.setValueAtTime(0.0001, a.currentTime);
    g.gain.exponentialRampToValueAtTime(laut || 0.06, a.currentTime + 0.008);
    g.gain.exponentialRampToValueAtTime(0.0001, a.currentTime + dauer);
    o.connect(g); g.connect(a.destination);
    o.start(); o.stop(a.currentTime + dauer + 0.02);
  }
  var klaenge = {
    flipper: function () { ton(520, 0.05, "square", 0.05); },
    stups: function () { ton(300, 0.04, "square", 0.03); },
    bumper: function () { ton(880, 0.09, "square", 0.075); },
    schleuder: function () { ton(660, 0.08, "sawtooth", 0.06); },
    ziel: function () { ton(1040, 0.1, "triangle", 0.07); },
    fallziel: function () { ton(760, 0.12, "square", 0.07); },
    gleiter: function () { ton(1180, 0.09, "triangle", 0.06); },
    spinner: function () { ton(1320, 0.06, "triangle", 0.05); },
    rampe: function () { ton(400, 0.18, "sawtooth", 0.07); },
    bahnende: function () { ton(760, 0.1, "triangle", 0.06); },
    scoop: function () { ton(300, 0.2, "square", 0.08); },
    loch: function () { ton(520, 0.14, "sine", 0.07); },
    licht: function () { ton(1200, 0.07, "sine", 0.06); },
    lichter: function () { ton(1600, 0.22, "sine", 0.08); },
    bonus: function () { ton(1200, 0.3, "triangle", 0.09); },
    bonusphase: function () { ton(520, 0.4, "square", 0.08); },
    mission: function () { ton(1400, 0.2, "triangle", 0.08); },
    jackpot: function () { ton(1800, 0.45, "sine", 0.1); },
    schloss: function () { ton(320, 0.22, "square", 0.08); },
    multiball: function () { ton(220, 0.5, "square", 0.09); },
    extra: function () { ton(900, 0.5, "triangle", 0.1); },
    kickback: function () { ton(180, 0.3, "sawtooth", 0.09); },
    klammer: function () { ton(640, 0.2, "sine", 0.07); },
    weiche: function () { ton(1000, 0.12, "square", 0.05); },
    aufzug: function () { ton(520, 0.25, "sawtooth", 0.06); },
    decksprung: function () { ton(420, 0.16, "triangle", 0.07); },
    deckausgang: function () { ton(760, 0.14, "triangle", 0.07); },
    abschuss: function () { ton(700, 0.09, "sawtooth", 0.05); },
    verlust: function () { ton(240, 0.32, "sine", 0.09); }
  };

  /* ---------------------------------------------------------------- Meldung */
  function meldung(titel, text, knopf, zahl) {
    $("meldungTitel").textContent = titel;
    $("meldungText").textContent = text;
    $("meldungKnopf").textContent = knopf;
    if (zahl) { $("meldungZahl").textContent = zahl; $("meldungZahl").hidden = false; }
    else { $("meldungZahl").hidden = true; $("meldungZahl").textContent = ""; }
    $("meldung").hidden = false;
  }
  function meldungZu() { $("meldung").hidden = true; }

  /* ---------------------------------------------------------------- Zahlen */
  function zeigeZahlen() {
    $("punkte").textContent = sp.punkte.toLocaleString("de-DE");
    $("multi").textContent = FL.multiWert(sp);
    $("kugelnText").textContent = Math.max(0, sp.kugelnGesamt - sp.kugelnUebrig);
    var b = "";
    for (var i = 0; i < Math.min(9, sp.kugelnGesamt); i++) {
      b += '<i class="' + (i < sp.kugelnGesamt - sp.kugelnUebrig ? "weg" : "da") + '"></i>';
    }
    $("kugelbild").innerHTML = b;
    var imSpiel = FL.imSpiel(sp);
    $("extras").textContent =
      "Kugeln im Spiel: " + imSpiel +
      " · Mission " + Math.min(sp.mission.stand + 1, sp.mission.folge.length) + "/" + sp.mission.folge.length +
      " · " + sp.mission.fertig + " erfüllt" +
      " · Schloss " + (sp.schlossOffen ? "offen" : sp.schloss.length + "/2") +
      (sp.jackpot ? " · JACKPOT" : "") +
      (sp.bonusZeit > 0 ? " · Bonus " + Math.ceil(sp.bonusZeit) + " s" : "") +
      (sp.kickback ? " · Kickback" : "");
    $("extras").className = sp.bonusZeit > 0 ? "extras bonus" : "extras";
  }
  function rekordHolen() {
    var w = 0;
    try { w = Number(localStorage.getItem("flipper-best-" + sp.tisch.id) || 0); } catch (e) { }
    return w;
  }
  function zeigeRekord() {
    var w = rekordHolen();
    $("rekord").textContent = w > 0 ? w.toLocaleString("de-DE") : "–";
  }

  /* --------------------------------------------------------------- Start */
  function tischInfo() {
    $("automatName").textContent = sp.tisch.name;
    $("hinweis").textContent = sp.tisch.tipp;
    zeigeRekord();
    zeigeZahlen();
  }
  function partieStarten(nr) {
    sp = FL.neuesSpiel(typeof nr === "number" ? nr : sp.tischNr);
    FL.partieStarten(sp);
    laeuft = true;
    pause = 0; spur = {}; blitz = 0; meldungZeit = 0;
    ziehId = null; flipperZeiger = {};
    $("links").classList.remove("drueckt");
    $("rechts").classList.remove("drueckt");
    farbenLaden();
    tischInfo();
    meldungZu();
    $("hinweis").textContent = "Kugel mit Leertaste oder dem Knopf schießen – je länger halten, desto stärker.";
    audioHolen();
  }
  function startbildschirm() {
    laeuft = false;
    var r = rekordHolen();
    meldung("FLIPPER", sp.tisch.name + "\n" + sp.tisch.tipp +
      "\n\nDrei Kugeln, zwei Flipper, ein Oberdeck.\n" +
      (r > 0 ? "Rekord: " + r.toLocaleString("de-DE") + " Punkte" : "Noch kein Rekord auf diesem Tisch."),
      "Start", "");
  }
  function spielEnde() {
    laeuft = false;
    var alt = rekordHolen();
    var neu = sp.punkte > alt;
    if (neu) { try { localStorage.setItem("flipper-best-" + sp.tisch.id, String(sp.punkte)); } catch (e) { } }
    zeigeRekord();
    var text = sp.tisch.name + "\n" + sp.zuege + " Abschüsse, " + sp.verloren + " Kugeln verloren, " +
      sp.mission.fertig + " Missionen, " + sp.multiball + " Multiball.\n" +
      (neu ? "\nNeuer Rekord!" : "\nRekord: " + Math.max(alt, sp.punkte).toLocaleString("de-DE"));
    meldung("PARTIE VORBEI", text, "Noch einmal", sp.punkte.toLocaleString("de-DE") + " Punkte");
  }

  /* ------------------------------------------------------------ Steuerung */
  function fluegelSetzen(links, an) {
    FL.fluegel(sp, links, an);
    $(links ? "links" : "rechts").classList.toggle("drueckt", !!an && sp.zustand !== "vorbei");
  }
  function alleFlipperLos() { fluegelSetzen(true, false); fluegelSetzen(false, false); }
  function stiftLaden(an, id) {
    if (sp.zustand !== "bereit") return;
    if (an) {
      ziehId = id === undefined ? "taste" : id;
      FL.abschussLaden(sp, true);
      $("stift").classList.add("drueckt");
    } else {
      if (ziehId === null) return;
      ziehId = null;
      FL.abschussLaden(sp, false);
      $("stift").classList.remove("drueckt");
      var kraft = sp.abschuss.kraft;
      if (FL.abschuss(sp, Math.max(0.28, kraft))) {
        spur = {};
        klaenge.abschuss();
        $("hinweis").textContent = kraft > 0.75 ? "Voller Schuss!" : "Kugel im Spiel.";
      }
    }
  }
  var tasten = {};
  window.addEventListener("keydown", function (e) {
    var t = e.key;
    if (t === "ArrowLeft" || t === "a" || t === "A") { fluegelSetzen(true, true); e.preventDefault(); }
    if (t === "ArrowRight" || t === "d" || t === "D") { fluegelSetzen(false, true); e.preventDefault(); }
    if (t === " " || t === "Spacebar") {
      if ($("meldung").hidden === false) { $("meldungKnopf").click(); }
      else if (!tasten[" "]) { tasten[" "] = true; stiftLaden(true); }
      e.preventDefault();
    }
    if (t === "Enter") {
      if (!$("meldung").hidden) $("meldungKnopf").click();
      e.preventDefault();
    }
  });
  window.addEventListener("keyup", function (e) {
    var t = e.key;
    if (t === "ArrowLeft" || t === "a" || t === "A") { fluegelSetzen(true, false); e.preventDefault(); }
    if (t === "ArrowRight" || t === "d" || t === "D") { fluegelSetzen(false, false); e.preventDefault(); }
    if (t === " " || t === "Spacebar") { tasten[" "] = false; stiftLaden(false); e.preventDefault(); }
  });
  window.addEventListener("blur", function () { alleFlipperLos(); tasten = {}; stiftLaden(false); });
  document.addEventListener("visibilitychange", function () {
    if (document.hidden) { alleFlipperLos(); tasten = {}; stiftLaden(false); }
  });

  function knopfFlipper(id, links) {
    var el = $(id);
    el.addEventListener("pointerdown", function (e) {
      e.preventDefault();
      if (el.setPointerCapture) { try { el.setPointerCapture(e.pointerId); } catch (err) { } }
      flipperZeiger[e.pointerId] = links;
      fluegelSetzen(links, true);
    });
    var los = function (e) {
      if (flipperZeiger[e.pointerId] === links) {
        delete flipperZeiger[e.pointerId];
        fluegelSetzen(links, false);
      }
    };
    el.addEventListener("pointerup", los);
    el.addEventListener("pointercancel", los);
    el.addEventListener("lostpointercapture", los);
  }
  knopfFlipper("links", true);
  knopfFlipper("rechts", false);

  var stiftEl = $("stift");
  stiftEl.addEventListener("pointerdown", function (e) {
    e.preventDefault();
    if (stiftEl.setPointerCapture) { try { stiftEl.setPointerCapture(e.pointerId); } catch (err) { } }
    stiftLaden(true, "knopf" + e.pointerId);
  });
  ["pointerup", "pointercancel", "lostpointercapture"].forEach(function (typ) {
    stiftEl.addEventListener(typ, function () { stiftLaden(false); });
  });

  function feldPunkt(e) {
    var r = leinwand.getBoundingClientRect();
    return { x: (e.clientX - r.left) / r.width * W, y: (e.clientY - r.top) / r.height * H };
  }
  leinwand.addEventListener("pointerdown", function (e) {
    e.preventDefault();
    if (leinwand.setPointerCapture) { try { leinwand.setPointerCapture(e.pointerId); } catch (err) { } }
    var p = feldPunkt(e);
    if (sp.zustand === "bereit" && (p.x > KANAL.x1 - 14 || p.y > 720)) {
      stiftLaden(true, "feld" + e.pointerId);
      return;
    }
    var links = p.x < W / 2;
    flipperZeiger[e.pointerId] = links;
    fluegelSetzen(links, true);
  });
  leinwand.addEventListener("pointermove", function (e) {
    if (ziehId !== "feld" + e.pointerId) return;
    var p = feldPunkt(e);
    var tiefe = Math.max(0, Math.min(1, (p.y - KANAL.stift) / 90));
    if (tiefe > sp.abschuss.kraft) sp.abschuss.kraft = tiefe;
  });
  function feldLos(e) {
    if (ziehId === "feld" + e.pointerId) { stiftLaden(false); return; }
    var links = flipperZeiger[e.pointerId];
    if (links !== undefined) { delete flipperZeiger[e.pointerId]; fluegelSetzen(links, false); }
  }
  leinwand.addEventListener("pointerup", feldLos);
  leinwand.addEventListener("pointercancel", feldLos);
  leinwand.addEventListener("lostpointercapture", feldLos);

  function baueTischLeiste() {
    var leiste = $("automatLeiste");
    leiste.innerHTML = "";
    FL.TISCHE.forEach(function (t, i) {
      var b = document.createElement("button");
      b.type = "button";
      b.textContent = t.name;
      b.dataset.auto = t.id;
      if (i === sp.tischNr) b.classList.add("aktiv");
      b.addEventListener("click", function () {
        if (laeuft) {
          $("hinweis").textContent = "Tischwechsel erst nach der Partie – jetzt läuft die Kugel.";
          return;
        }
        document.querySelectorAll("#automatLeiste button").forEach(function (x) { x.classList.remove("aktiv"); });
        b.classList.add("aktiv");
        sp = FL.neuesSpiel(i);
        laeuft = false;
        spur = {};
        farbenLaden();
        tischInfo();
        startbildschirm();
      });
      leiste.appendChild(b);
    });
  }

  /* ------------------------------------------------------------------ Ton */
  function tonUmschalten() {
    tonAn = !tonAn;
    try { localStorage.setItem("flipper-ton", tonAn ? "an" : "aus"); } catch (e) { }
    $("tonKnopf").textContent = tonAn ? "Ton an" : "Ton aus";
    $("tonKnopf").setAttribute("aria-pressed", tonAn ? "true" : "false");
    if (tonAn) { audioHolen(); ton(880, 0.08, "sine", 0.05); }
  }
  $("tonKnopf").addEventListener("click", tonUmschalten);
  $("tonKnopf").textContent = tonAn ? "Ton an" : "Ton aus";

  $("meldungKnopf").addEventListener("click", function () {
    if (sp.zustand === "vorbei" || !laeuft) partieStarten(sp.tischNr);
    else meldungZu();
  });

  /* ==================================================================== Zeichnen */

  function zeichneGrund() {
    var g = ctx.createLinearGradient(0, 0, 0, H);
    g.addColorStop(0, TF.grund1);
    g.addColorStop(1, TF.grund2);
    ctx.fillStyle = g;
    ctx.fillRect(0, 0, W, H);
    /* Lichtstrahlen, dezent */
    ctx.save();
    ctx.globalAlpha = 0.05;
    ctx.strokeStyle = TF.licht;
    ctx.lineWidth = 1;
    for (var i = -6; i < 18; i++) {
      ctx.beginPath(); ctx.moveTo(i * 44, 300); ctx.lineTo(i * 44 + 200, H); ctx.stroke();
    }
    ctx.restore();
    /* Spielfeldfläche mit Rand und Schrauben */
    ctx.strokeStyle = mitAlpha(TF.wand, 0.32);
    ctx.lineWidth = 2;
    rundeFlaeche(38, 306, 326, 466, 20); ctx.stroke();
    ctx.strokeStyle = F.zierrand; ctx.lineWidth = 1;
    rundeFlaeche(44, 312, 314, 454, 16); ctx.stroke();
    [[47, 318], [355, 318], [47, 560], [355, 560], [47, 760], [355, 760]].forEach(function (p) {
      ctx.beginPath(); ctx.arc(p[0], p[1], 3.2, 0, Math.PI * 2);
      ctx.fillStyle = F.schraube; ctx.fill();
      ctx.beginPath(); ctx.moveTo(p[0] - 1.5, p[1]); ctx.lineTo(p[0] + 1.5, p[1]);
      ctx.strokeStyle = F.schraubSchlitz; ctx.lineWidth = 1; ctx.stroke();
    });
    /* Lichteinsätze: Bereiche, die im Betrieb leuchten */
    ctx.save();
    ctx.globalAlpha = sp.bonusZeit > 0 ? 0.16 : 0.06;
    var rg = ctx.createRadialGradient(202, 430, 20, 202, 430, 190);
    rg.addColorStop(0, TF.licht); rg.addColorStop(1, "rgba(0,0,0,0)");
    ctx.fillStyle = rg;
    ctx.fillRect(30, 330, 344, 220);
    ctx.restore();
    /* Bahnmarkierungen links und rechts */
    ctx.save();
    ctx.strokeStyle = mitAlpha(TF.licht, 0.16);
    ctx.lineWidth = 8; ctx.lineCap = "round";
    ctx.beginPath(); ctx.moveTo(66, 300); ctx.lineTo(56, 460); ctx.stroke();
    ctx.beginPath(); ctx.moveTo(352, 300); ctx.lineTo(362, 460); ctx.stroke();
    ctx.restore();
    /* Zonenmarkierungen */
    ctx.textAlign = "center"; ctx.textBaseline = "middle";
    ctx.font = "bold 7px Arial";
    ctx.fillStyle = mitAlpha(F.text, 0.3);
    ctx.fillText("MITTELFELD", 202, 470);
    ctx.fillText("UNTERES FELD", 202, 645);
  }

  function zeichneGehaeuse() {
    ctx.save();
    var g = ctx.createLinearGradient(0, 0, W, H);
    g.addColorStop(0, mitAlpha(TF.wand, 0.75));
    g.addColorStop(0.5, mitAlpha(F.stange3 || TF.wand, 0.5));
    g.addColorStop(1, mitAlpha(TF.wand, 0.7));
    ctx.strokeStyle = g;
    ctx.lineWidth = 9;
    rundeFlaeche(5, 5, W - 10, H - 10, 16); ctx.stroke();
    ctx.strokeStyle = mitAlpha(F.zinne, 0.35);
    ctx.lineWidth = 1.4;
    rundeFlaeche(9.5, 9.5, W - 19, H - 19, 13); ctx.stroke();
    [[22, 22], [W - 22, 22], [22, H - 22], [W - 22, H - 22]].forEach(function (p) {
      ctx.beginPath(); ctx.arc(p[0], p[1], 4.5, 0, Math.PI * 2);
      ctx.fillStyle = F.schraube; ctx.fill();
      ctx.beginPath(); ctx.moveTo(p[0] - 2.2, p[1]); ctx.lineTo(p[0] + 2.2, p[1]);
      ctx.strokeStyle = F.schraubSchlitz; ctx.lineWidth = 1.2; ctx.stroke();
    });
    ctx.restore();
  }

  function zeichneDeck() {
    var d = sp.tisch.deck;
    /* Schatten unter dem erhöhten Deck */
    ctx.save();
    ctx.shadowColor = "rgba(0,0,0,0.55)";
    ctx.shadowBlur = 18; ctx.shadowOffsetY = 8;
    ctx.fillStyle = TF.grund2;
    rundeFlaeche(DECK.x1 - 6, DECK.y1 - 8, DECK.x2 - DECK.x1 + 12, DECK.y2 - DECK.y1 + 10, 16);
    ctx.fill();
    ctx.restore();
    /* Deckplatte */
    var g = ctx.createLinearGradient(0, DECK.y1, 0, DECK.y2);
    g.addColorStop(0, mitAlpha(TF.licht, 0.16));
    g.addColorStop(1, mitAlpha(TF.grund2, 0.96));
    ctx.fillStyle = g;
    rundeFlaeche(DECK.x1, DECK.y1, DECK.x2 - DECK.x1, DECK.y2 - DECK.y1, 14);
    ctx.fill();
    /* Rahmen */
    ctx.strokeStyle = TF.wand; ctx.lineWidth = 5;
    ctx.beginPath();
    ctx.moveTo(DECK.x1, DECK.y2); ctx.lineTo(DECK.x1, DECK.y1);
    ctx.lineTo(DECK.x1 + 44, DECK.y1 - 16); ctx.lineTo(DECK.x2 - 44, DECK.y1 - 16);
    ctx.lineTo(DECK.x2, DECK.y1); ctx.lineTo(DECK.x2, DECK.y2);
    ctx.stroke();
    ctx.strokeStyle = mitAlpha(F.zinne, 0.45); ctx.lineWidth = 1.6;
    ctx.beginPath();
    ctx.moveTo(DECK.x1 + 1, DECK.y2); ctx.lineTo(DECK.x1 + 1, DECK.y1 + 1);
    ctx.lineTo(DECK.x1 + 45, DECK.y1 - 15); ctx.lineTo(DECK.x2 - 45, DECK.y1 - 15);
    ctx.lineTo(DECK.x2 - 1, DECK.y1 + 1); ctx.lineTo(DECK.x2 - 1, DECK.y2);
    ctx.stroke();
    /* Deckboden mit Auslass in der Mitte */
    ctx.strokeStyle = TF.wand; ctx.lineWidth = 4;
    ctx.beginPath();
    ctx.moveTo(DECK.x1, DECK.y2); ctx.lineTo(160, DECK.y2);
    ctx.moveTo(240, DECK.y2); ctx.lineTo(DECK.x2, DECK.y2);
    ctx.stroke();
    /* Frontkante (Höhe des Decks) */
    var fk = ctx.createLinearGradient(0, DECK.y2, 0, DECK.y2 + 16);
    fk.addColorStop(0, mitAlpha(TF.wand, 0.55));
    fk.addColorStop(1, "rgba(0,0,0,0)");
    ctx.fillStyle = fk;
    ctx.fillRect(DECK.x1, DECK.y2, DECK.x2 - DECK.x1, 16);
    /* Beschriftung */
    ctx.textAlign = "center"; ctx.textBaseline = "middle";
    ctx.font = "bold 8px Arial";
    ctx.fillStyle = mitAlpha(TF.licht, 0.75);
    ctx.fillText("OBERDECK", 218, 126);
    ctx.font = "bold 7px Arial";
    ctx.fillStyle = mitAlpha(TF.licht, 0.45);
    ctx.fillText("AUFZUG →", 300, 150);

    /* Elemente auf dem Deck */
    d.ziele.forEach(function (z) {
      ctx.save();
      ctx.globalAlpha = z.runter ? 0.28 : 1;
      ctx.fillStyle = z.runter ? TF.grund2 : TF.ziel;
      ctx.strokeStyle = TF.licht; ctx.lineWidth = 1.4;
      rundeFlaeche(z.x, z.y, z.w, z.h, 4); ctx.fill(); ctx.stroke();
      if (z.blitz > 0) { ctx.globalAlpha = z.blitz; ctx.fillStyle = F.glanz; rundeFlaeche(z.x, z.y, z.w, z.h, 4); ctx.fill(); }
      ctx.restore();
    });
    /* Gleitziel */
    var gz = d.gleiter;
    ctx.save();
    ctx.shadowColor = TF.licht; ctx.shadowBlur = gz.blitz > 0 ? 16 : 4;
    ctx.fillStyle = gz.blitz > 0 ? F.glanz : TF.ziel;
    ctx.strokeStyle = TF.licht; ctx.lineWidth = 1.4;
    rundeFlaeche(gz.x, gz.y, gz.w, gz.h, 5); ctx.fill(); ctx.stroke();
    ctx.restore();
    /* Laufschiene des Gleiters */
    ctx.strokeStyle = mitAlpha(TF.wand, 0.35); ctx.lineWidth = 2;
    ctx.beginPath(); ctx.moveTo(gz.x0 - 4, gz.y + gz.h / 2); ctx.lineTo(gz.x1 + gz.w + 4, gz.y + gz.h / 2); ctx.stroke();
    /* Bumper */
    zeichneBumper(d.bumper);
    /* Jackpot-Loch / Ausgang */
    var lo = d.loch;
    ctx.save();
    ctx.shadowColor = sp.jackpot ? F.glanz : TF.licht;
    ctx.shadowBlur = sp.jackpot ? 20 : 5;
    ctx.beginPath(); ctx.arc(lo.x, lo.y, lo.r + 4, 0, Math.PI * 2);
    ctx.fillStyle = F.scoopAussen; ctx.fill();
    ctx.beginPath(); ctx.arc(lo.x, lo.y, lo.r, 0, Math.PI * 2);
    ctx.fillStyle = F.scoopInnen; ctx.fill();
    ctx.strokeStyle = sp.jackpot ? F.glanz : TF.licht;
    ctx.lineWidth = 2; ctx.stroke();
    ctx.fillStyle = sp.jackpot ? F.glanz : mitAlpha(TF.licht, 0.8);
    ctx.font = "bold 6.5px Arial"; ctx.textAlign = "center";
    ctx.fillText(sp.jackpot ? "JACKPOT" : "AUS", lo.x, lo.y + 2);
    ctx.restore();
    /* Halteklammer */
    var kl = d.klammer;
    ctx.save();
    ctx.shadowColor = TF.licht; ctx.shadowBlur = kl.haelt > 0 ? 18 : 4;
    ctx.beginPath(); ctx.arc(kl.x, kl.y, kl.r, 0, Math.PI * 2);
    ctx.fillStyle = mitAlpha(TF.licht, 0.12); ctx.fill();
    ctx.strokeStyle = kl.haelt > 0 ? F.glanz : TF.licht; ctx.lineWidth = 2.4; ctx.stroke();
    /* zwei Klemmen */
    ctx.strokeStyle = F.stange2; ctx.lineWidth = 3;
    ctx.beginPath();
    ctx.moveTo(kl.x - 9, kl.y - 72); ctx.lineTo(kl.x - 9, kl.y - 9);
    ctx.moveTo(kl.x + 9, kl.y - 72); ctx.lineTo(kl.x + 9, kl.y - 9);
    ctx.stroke();
    ctx.fillStyle = mitAlpha(TF.licht, 0.7);
    ctx.font = "bold 6px Arial"; ctx.textAlign = "center";
    ctx.fillText("KLAMMER", kl.x, kl.y + kl.r + 8);
    ctx.restore();
    /* Deck-Auslässen Pfeile */
    ctx.fillStyle = mitAlpha(TF.licht, 0.5);
    ctx.font = "bold 7px Arial"; ctx.textAlign = "center";
    ctx.fillText("▼", 200, 292);
    ctx.textAlign = "right";
    ctx.fillText("AUS ▶", 336, 250);
  }

  function zeichneRampen() {
    sp.tisch.rampen.forEach(function (r) {
      var b = r.bahn;
      ctx.save();
      ctx.lineCap = "round"; ctx.lineJoin = "round";
      /* Schiene als Band */
      ctx.strokeStyle = mitAlpha(F.rampeRand, 0.9);
      ctx.lineWidth = 11;
      ctx.beginPath(); ctx.moveTo(b[0].x, b[0].y);
      for (var i = 1; i < b.length; i++) ctx.lineTo(b[i].x, b[i].y);
      ctx.stroke();
      ctx.strokeStyle = F.rampe; ctx.lineWidth = 4;
      ctx.beginPath(); ctx.moveTo(b[0].x, b[0].y);
      for (var j = 1; j < b.length; j++) ctx.lineTo(b[j].x, b[j].y);
      ctx.stroke();
      ctx.strokeStyle = mitAlpha(TF.licht, r.blitz > 0 ? 0.95 : 0.5);
      ctx.lineWidth = 2;
      ctx.beginPath(); ctx.moveTo(b[0].x, b[0].y);
      for (var q = 1; q < b.length; q++) ctx.lineTo(b[q].x, b[q].y);
      ctx.stroke();
      /* Einfahrt */
      ctx.beginPath(); ctx.arc(r.ein.x, r.ein.y, r.ein.r, 0, Math.PI * 2);
      ctx.fillStyle = mitAlpha(TF.licht, 0.2); ctx.fill();
      ctx.strokeStyle = r.blitz > 0 ? F.glanz : TF.licht; ctx.lineWidth = 2.2; ctx.stroke();
      ctx.beginPath(); ctx.arc(r.ein.x, r.ein.y, r.ein.r * 0.42, 0, Math.PI * 2);
      ctx.fillStyle = TF.licht; ctx.fill();
      /* Auslauf-Pfeil */
      var a = r.aus;
      ctx.beginPath(); ctx.arc(a.x, a.y, 4, 0, Math.PI * 2);
      ctx.fillStyle = mitAlpha(TF.licht, 0.85); ctx.fill();
      /* Beschriftung */
      ctx.font = "bold 7px Arial"; ctx.textAlign = "center"; ctx.textBaseline = "middle";
      var lx = r.ein.x, ly = r.ein.y - r.ein.r - 10;
      var name = r.name.toUpperCase();
      var bw = ctx.measureText(name).width + 8;
      ctx.fillStyle = mitAlpha(TF.grund2, 0.8);
      rundeFlaeche(lx - bw / 2, ly - 6, bw, 12, 4); ctx.fill();
      ctx.fillStyle = mitAlpha(TF.licht, 0.95);
      ctx.fillText(name, lx, ly + 0.5);
      ctx.textBaseline = "alphabetic";
      ctx.restore();
    });
  }

  function zeichneWaende() {
    ctx.lineCap = "round";
    sp.tisch.waende.forEach(function (w) {
      ctx.strokeStyle = mitAlpha(TF.wand, 0.28); ctx.lineWidth = 10;
      ctx.beginPath(); ctx.moveTo(w.x1, w.y1); ctx.lineTo(w.x2, w.y2); ctx.stroke();
      ctx.strokeStyle = TF.wand; ctx.lineWidth = 4.5;
      ctx.beginPath(); ctx.moveTo(w.x1, w.y1); ctx.lineTo(w.x2, w.y2); ctx.stroke();
    });
    ctx.strokeStyle = mitAlpha(F.zinne, 0.4); ctx.lineWidth = 1.4;
    sp.tisch.waende.forEach(function (w) {
      ctx.beginPath(); ctx.moveTo(w.x1 - 1, w.y1); ctx.lineTo(w.x2 - 1, w.y2); ctx.stroke();
    });
    /* Weiche: zeigt, wohin die Schleife ausläuft */
    var lw = sp.tisch.rampen[1];
    if (lw) {
      ctx.save();
      ctx.fillStyle = mitAlpha(TF.licht, 0.8);
      ctx.font = "bold 7px Arial"; ctx.textAlign = "left";
      ctx.fillText(sp.weiche === 0 ? "WEICHE: OBEN" : "WEICHE: INLANE", 106, 500);
      ctx.beginPath();
      ctx.arc(98, 496, 4, 0, Math.PI * 2);
      ctx.fillStyle = sp.weiche === 0 ? TF.licht : mitAlpha(TF.licht, 0.3);
      ctx.fill();
      ctx.restore();
      /* Klappe am Auslauf */
      ctx.save();
      ctx.translate(80, 340);
      ctx.rotate(sp.weiche === 0 ? -0.5 : 0.55);
      ctx.fillStyle = mitAlpha(F.stange2, 0.9);
      rundeFlaeche(0, -2, 22, 4.5, 2); ctx.fill();
      ctx.restore();
    }
  }

  /* Inlane/Outlane wie am Automaten: eine Metallschiene trennt beide, oben ein
     Pfosten mit Gummi, an dem sich der Weg teilt. Die Schräge am Ende legt die
     Kugel auf den Flipper. */
  function zeichneLanes() {
    var ln = FL.LANES;
    [ln.links, ln.rechts].forEach(function (l) {
      ctx.save();
      ctx.lineCap = "round"; ctx.lineJoin = "round";
      ctx.beginPath();
      ctx.moveTo(l.wand[0], l.wand[1]);
      ctx.lineTo(l.knie[0], l.knie[1]);
      ctx.lineTo(l.ende[0], l.ende[1]);
      ctx.strokeStyle = mitAlpha(TF.grund2, 0.9); ctx.lineWidth = 8; ctx.stroke();
      ctx.strokeStyle = F.stange1; ctx.lineWidth = 4.6; ctx.stroke();
      ctx.strokeStyle = F.stange2; ctx.lineWidth = 2.6; ctx.stroke();
      ctx.strokeStyle = mitAlpha(F.glanz, 0.65); ctx.lineWidth = 1; ctx.stroke();
      /* Pfosten mit Gummiring */
      ctx.beginPath(); ctx.arc(l.wand[0], l.wand[1], 6, 0, Math.PI * 2);
      ctx.fillStyle = F.lager; ctx.fill();
      ctx.strokeStyle = F.lagerRand; ctx.lineWidth = 1.4; ctx.stroke();
      ctx.beginPath(); ctx.arc(l.wand[0], l.wand[1], 3.4, 0, Math.PI * 2);
      ctx.fillStyle = mitAlpha(TF.licht, 0.9); ctx.fill();
      ctx.restore();
    });
    /* Beschriftung längs der Gassen */
    ctx.save();
    ctx.font = "bold 6px Arial"; ctx.textAlign = "center"; ctx.textBaseline = "middle";
    ctx.fillStyle = mitAlpha(F.text, 0.5);
    [
      [74, 604, "INLANE", -Math.PI / 2], [325, 604, "INLANE", Math.PI / 2],
      [41, 640, "OUTLANE", -Math.PI / 2], [356, 640, "OUTLANE", Math.PI / 2]
    ].forEach(function (t) {
      ctx.save();
      ctx.translate(t[0], t[1]); ctx.rotate(t[3]);
      ctx.fillText(t[2], 0, 0);
      ctx.restore();
    });
    ctx.restore();
  }

  /* Apron unter den Flippern: deckt die Fläche ab, die Kugel verschwindet dahinter. */
  function zeichneApron() {
    ctx.save();
    var g = ctx.createLinearGradient(0, 756, 0, H);
    g.addColorStop(0, mitAlpha(F.scoopAussen, 0.82));
    g.addColorStop(1, mitAlpha(F.s4, 0.98));
    ctx.fillStyle = g;
    ctx.beginPath();
    ctx.moveTo(30, 766); ctx.lineTo(176, 758); ctx.lineTo(168, H); ctx.lineTo(30, H); ctx.closePath(); ctx.fill();
    ctx.beginPath();
    ctx.moveTo(372, 766); ctx.lineTo(226, 758); ctx.lineTo(234, H); ctx.lineTo(372, H); ctx.closePath(); ctx.fill();
    /* V-förmige Kugelrinne in der Mitte */
    ctx.beginPath();
    ctx.moveTo(168, H); ctx.lineTo(202, 780); ctx.lineTo(234, H); ctx.closePath(); ctx.fill();
    ctx.strokeStyle = mitAlpha(TF.wand, 0.3); ctx.lineWidth = 1.2;
    ctx.beginPath();
    ctx.moveTo(30, 766); ctx.lineTo(176, 758);
    ctx.moveTo(226, 758); ctx.lineTo(372, 766);
    ctx.stroke();
    ctx.fillStyle = mitAlpha(F.text, 0.4);
    ctx.font = "bold 7px Arial"; ctx.textAlign = "center"; ctx.textBaseline = "middle";
    ctx.fillText("DRAIN", 202, 802);
    ctx.restore();
  }

  function zeichneBumper(liste) {
    liste.forEach(function (b) {
      var puls = 0.5 + 0.5 * Math.sin(performance.now() / 260 + b.x);
      var g = ctx.createRadialGradient(b.x - b.r * 0.34, b.y - b.r * 0.42, 1, b.x, b.y, b.r + 5);
      g.addColorStop(0, F.glanz);
      g.addColorStop(0.22, TF.licht);
      g.addColorStop(0.48, TF.bumper1);
      g.addColorStop(1, TF.bumper2);
      ctx.save();
      ctx.shadowColor = b.blitz > 0 ? F.glanz : TF.licht;
      ctx.shadowBlur = b.blitz > 0 ? 24 : 4 + puls * 3;
      ctx.beginPath(); ctx.arc(b.x, b.y, b.r + 4, 0, Math.PI * 2);
      ctx.fillStyle = F.bumperRing; ctx.fill();
      ctx.lineWidth = 2.5; ctx.strokeStyle = F.bumperKante; ctx.stroke();
      ctx.beginPath(); ctx.arc(b.x, b.y, b.r, 0, Math.PI * 2);
      ctx.fillStyle = g; ctx.fill();
      ctx.lineWidth = 2; ctx.strokeStyle = b.blitz > 0 ? F.glanz : TF.licht; ctx.stroke();
      ctx.beginPath(); ctx.arc(b.x, b.y, b.r * 0.34, 0, Math.PI * 2);
      ctx.fillStyle = F.bumperKappe; ctx.fill();
      ctx.restore();
      if (b.blitz > 0) {
        ctx.beginPath(); ctx.arc(b.x, b.y, b.r + 10 * b.blitz, 0, Math.PI * 2);
        ctx.strokeStyle = mitAlpha(F.glanz, b.blitz * 0.5);
        ctx.lineWidth = 3; ctx.stroke();
      }
    });
  }

  function zeichneSchleudern() {
    sp.tisch.schleudern.forEach(function (s) {
      ctx.beginPath();
      ctx.moveTo(s.x1, s.y1); ctx.lineTo(s.x2, s.y2);
      ctx.lineTo(s.x1, s.y2);
      ctx.closePath();
      var sg = ctx.createLinearGradient(s.x1, s.y1, s.x2, s.y2);
      sg.addColorStop(0, mitAlpha(TF.licht, 0.35));
      sg.addColorStop(1, F.s3);
      ctx.fillStyle = sg; ctx.fill();
      ctx.strokeStyle = mitAlpha(TF.wand, 0.6); ctx.lineWidth = 2; ctx.stroke();
      if (s.blitz > 0) {
        ctx.strokeStyle = mitAlpha(F.glanz, s.blitz * 0.45);
        ctx.lineWidth = 12;
        ctx.beginPath(); ctx.moveTo(s.x1, s.y1); ctx.lineTo(s.x2, s.y2); ctx.stroke();
      }
      ctx.strokeStyle = s.blitz > 0 ? F.glanz : TF.licht;
      ctx.lineWidth = s.blitz > 0 ? 6 : 4;
      ctx.beginPath(); ctx.moveTo(s.x1, s.y1); ctx.lineTo(s.x2, s.y2); ctx.stroke();

    });
  }

  function zeichneZiele() {
    /* Aufrechtziele links (Mission) und rechts */
    sp.tisch.zieleOben.forEach(function (z, i) {
      var mission = sp.mission.folge.indexOf(i) === sp.mission.stand;
      ctx.save();
      ctx.fillStyle = z.blitz > 0 ? F.glanz : (mission ? mitAlpha(TF.licht, 0.9) : TF.ziel);
      ctx.strokeStyle = TF.licht; ctx.lineWidth = 1.6;
      rundeFlaeche(z.x, z.y, z.w, z.h, 5);
      ctx.fill(); ctx.stroke();
      if (mission) {
        ctx.strokeStyle = mitAlpha(F.glanz, 0.55 + 0.35 * Math.sin(performance.now() / 200));
        ctx.lineWidth = 2.4;
        rundeFlaeche(z.x - 2.5, z.y - 2.5, z.w + 5, z.h + 5, 7); ctx.stroke();
      }
      ctx.fillStyle = F.lichtText;
      ctx.font = "bold 8px Arial"; ctx.textAlign = "center"; ctx.textBaseline = "middle";
      ctx.fillText(mission ? "!" : String(i + 1), z.x + z.w / 2, z.y + z.h / 2);
      ctx.restore();
    });
    ctx.textBaseline = "alphabetic";
    /* Fallziele */
    sp.tisch.fallen.forEach(function (z) {
      ctx.save();
      if (z.runter) {
        ctx.globalAlpha = 0.25;
        ctx.fillStyle = TF.grund2;
        rundeFlaeche(z.x, z.y + 7, z.w, 4, 2); ctx.fill();
      } else {
        ctx.fillStyle = z.blitz > 0 ? F.glanz : TF.ziel;
        ctx.strokeStyle = TF.licht; ctx.lineWidth = 1.4;
        rundeFlaeche(z.x, z.y, z.w, z.h, 3);
        ctx.fill(); ctx.stroke();
        ctx.fillStyle = F.lichtText; ctx.font = "bold 6px Arial"; ctx.textAlign = "center"; ctx.textBaseline = "middle";
        ctx.fillText("FALL", z.x + z.w / 2, z.y + z.h / 2);
      }
      ctx.restore();
    });
    ctx.textBaseline = "alphabetic";
  }

  function zeichneRollover() {
    var namen = ["A", "B", "C"];
    ctx.save();
    ctx.textAlign = "center"; ctx.textBaseline = "middle";
    ctx.font = "bold 13px Georgia, serif";
    sp.tisch.rollover.forEach(function (r, i) {
      rundeFlaeche(r.x, r.y, r.w, r.h, 7);
      ctx.fillStyle = r.an ? TF.licht : F.s2;
      ctx.fill();
      ctx.strokeStyle = r.blitz > 0 ? F.glanz : mitAlpha(TF.licht, 0.75);
      ctx.lineWidth = 1.6; ctx.stroke();
      ctx.fillStyle = r.an ? F.lichtText : mitAlpha(TF.licht, 0.9);
      ctx.fillText(namen[i] || "·", r.x + r.w / 2, r.y + r.h / 2 + 1);
    });
    ctx.restore();
  }

  function zeichneSpinner() {
    var s = sp.tisch.spinner;
    ctx.save();
    ctx.translate(s.x, s.y); ctx.rotate(sp.spinnerWinkel);
    ctx.shadowColor = TF.licht; ctx.shadowBlur = 7;
    var g = ctx.createLinearGradient(0, -5, 0, 5);
    g.addColorStop(0, F.fahne1); g.addColorStop(0.5, F.fahne2); g.addColorStop(1, F.fahne3);
    ctx.fillStyle = g;
    rundeFlaeche(-s.len, -5, s.len * 2, 10, 4);
    ctx.fill();
    ctx.strokeStyle = F.fahneRand; ctx.lineWidth = 1.4; ctx.stroke();
    ctx.restore();
    ctx.beginPath(); ctx.arc(s.x, s.y, 7, 0, Math.PI * 2);
    ctx.fillStyle = F.spinner; ctx.fill();
    ctx.strokeStyle = F.spinnerRand; ctx.lineWidth = 2; ctx.stroke();
    ctx.fillStyle = mitAlpha(TF.licht, 0.6); ctx.font = "bold 6.5px Arial"; ctx.textAlign = "center";
    ctx.fillText("SPINNER", s.x, s.y + 18);
  }

  function zeichneScoop() {
    var s = sp.tisch.scoop;
    ctx.save();
    ctx.shadowColor = TF.licht; ctx.shadowBlur = s.blitz > 0 ? 20 : 5;
    ctx.beginPath(); ctx.arc(s.x, s.y, s.r + 5, 0, Math.PI * 2);
    ctx.fillStyle = F.scoopAussen; ctx.fill();
    ctx.beginPath(); ctx.arc(s.x, s.y, s.r, 0, Math.PI * 2);
    ctx.fillStyle = F.scoopInnen; ctx.fill();
    ctx.strokeStyle = TF.licht; ctx.lineWidth = 2; ctx.stroke();
    ctx.fillStyle = TF.licht; ctx.font = "bold 7px Arial"; ctx.textAlign = "center";
    ctx.fillText("KICK", s.x, s.y + 2.5);
    ctx.restore();
  }

  function zeichneSchloss() {
    var s = sp.tisch.schloss;
    ctx.save();
    ctx.shadowColor = sp.schlossOffen ? F.glanz : TF.licht;
    ctx.shadowBlur = sp.schlossOffen ? 18 : 4;
    ctx.beginPath(); ctx.arc(s.x, s.y, s.r + 4, 0, Math.PI * 2);
    ctx.fillStyle = F.scoopAussen; ctx.fill();
    ctx.beginPath(); ctx.arc(s.x, s.y, s.r, 0, Math.PI * 2);
    ctx.fillStyle = F.scoopInnen; ctx.fill();
    ctx.strokeStyle = sp.schlossOffen ? F.glanz : mitAlpha(TF.licht, 0.7);
    ctx.lineWidth = 2; ctx.stroke();
    ctx.fillStyle = sp.schlossOffen ? F.glanz : mitAlpha(TF.licht, 0.6);
    ctx.font = "bold 6.5px Arial"; ctx.textAlign = "center";
    ctx.fillText(sp.schlossOffen ? "FREI" : "ZU", s.x, s.y + 2);
    ctx.fillStyle = mitAlpha(TF.licht, 0.8);
    ctx.font = "bold 7px Arial";
    ctx.fillText(sp.schloss.length + "/2", s.x, s.y + s.r + 10);
    ctx.restore();
    /* Eingeschlossene Kugeln */
    for (var i = 0; i < sp.schloss.length; i++) {
      kugelBild(s.x - 6 + i * 12, s.y - s.r - 14, 0.85);
    }
  }

  function zeichneMagnet() {
    var m = sp.tisch.magnet;
    ctx.save();
    var aktiv = sp.bonusZeit > 0 || m.an > 0.1;
    if (m.an > 0.02) {
      ctx.strokeStyle = mitAlpha(TF.licht, 0.18 + 0.5 * m.an);
      ctx.lineWidth = 1.8;
      for (var r = 24; r <= 56; r += 16) {
        ctx.beginPath(); ctx.arc(m.x, m.y, r, 0, Math.PI * 2); ctx.stroke();
      }
    }
    ctx.beginPath(); ctx.arc(m.x, m.y, m.r, 0, Math.PI * 2);
    ctx.fillStyle = aktiv ? mitAlpha(TF.licht, 0.6) : F.s3;
    ctx.fill();
    ctx.strokeStyle = TF.licht; ctx.lineWidth = 1.8; ctx.stroke();
    ctx.fillStyle = F.lichtText; ctx.font = "bold 6px Arial"; ctx.textAlign = "center";
    ctx.fillText("MAG", m.x, m.y + 2);
    ctx.restore();
  }

  function zeichneFlipper() {
    sp.flipper.forEach(function (f) {
      ctx.save();
      ctx.translate(f.x, f.y); ctx.rotate(f.winkel);
      var L = f.len;
      ctx.shadowColor = TF.licht; ctx.shadowBlur = f.an ? 14 : 3;
      ctx.beginPath();
      ctx.moveTo(3, -9); ctx.lineTo(L - 10, -6);
      ctx.quadraticCurveTo(L + 2, 0, L - 10, 6);
      ctx.lineTo(3, 9); ctx.quadraticCurveTo(-3, 0, 3, -9);
      ctx.closePath();
      var g = ctx.createLinearGradient(0, -9, 0, 9);
      g.addColorStop(0, F.arm1); g.addColorStop(0.25, TF.licht);
      g.addColorStop(0.62, F.arm2); g.addColorStop(1, F.arm3);
      ctx.fillStyle = g; ctx.fill();
      ctx.shadowBlur = 0;
      ctx.lineWidth = 1.5; ctx.strokeStyle = F.armRand; ctx.stroke();
      ctx.beginPath(); ctx.moveTo(12, 0); ctx.lineTo(L - 12, 0);
      ctx.strokeStyle = F.armLinie; ctx.lineWidth = 1.3; ctx.stroke();
      ctx.beginPath(); ctx.arc(L - 9, 0, 6, 0, Math.PI * 2);
      ctx.fillStyle = mitAlpha(TF.licht, 0.35); ctx.fill();
      ctx.restore();
      ctx.beginPath(); ctx.arc(f.x, f.y, 10.5, 0, Math.PI * 2);
      ctx.fillStyle = F.lager; ctx.fill();
      ctx.lineWidth = 2; ctx.strokeStyle = F.lagerRand; ctx.stroke();
      ctx.beginPath(); ctx.arc(f.x, f.y, 4, 0, Math.PI * 2);
      ctx.fillStyle = TF.licht; ctx.fill();
      ctx.beginPath(); ctx.moveTo(f.x - 2, f.y); ctx.lineTo(f.x + 2, f.y);
      ctx.strokeStyle = F.lagerSchlitz; ctx.lineWidth = 1.3; ctx.stroke();
    });
  }

  function zeichneKanal() {
    /* Schacht */
    var kg = ctx.createLinearGradient(KANAL.x1, 0, KANAL.x2, 0);
    kg.addColorStop(0, F.s1);
    kg.addColorStop(0.45, mitAlpha(F.stange1, 0.5));
    kg.addColorStop(1, F.s2);
    ctx.fillStyle = kg;
    ctx.fillRect(KANAL.x1 + 2, 320, KANAL.x2 - KANAL.x1 - 4, KANAL.boden - 322);
    ctx.strokeStyle = mitAlpha(F.zinne, 0.3); ctx.lineWidth = 1.2;
    ctx.beginPath(); ctx.moveTo(KANAL.x1 + 6, 326); ctx.lineTo(KANAL.x1 + 6, KANAL.boden - 6); ctx.stroke();
    var hub = sp.abschuss.kraft * 18;
    var x = KANAL.mitte;
    ctx.fillStyle = TF.licht;
    rundeFlaeche(x - 5, KANAL.boden - 10 + hub, 10, 18, 4);
    ctx.fill();
    ctx.strokeStyle = F.s4; ctx.lineWidth = 1; ctx.stroke();
    ctx.strokeStyle = F.feder; ctx.lineWidth = 2;
    ctx.beginPath();
    var von = KANAL.boden + 12 + hub, bis = H - 14;
    for (var i = 0; i <= 6; i++) {
      var yy = von + (bis - von) * i / 6;
      var xx = x + (i === 0 ? 0 : (i % 2 ? 7 : -7));
      if (i === 0) ctx.moveTo(x, yy); else ctx.lineTo(xx, yy);
    }
    ctx.stroke();
    /* Ladeanzeige */
    var gx = KANAL.x1 + 4, gy = 660, gh = 70;
    ctx.fillStyle = F.s3; rundeFlaeche(gx, gy, 6, gh, 3); ctx.fill();
    var hh = Math.round(gh * sp.abschuss.kraft);
    if (hh > 1) {
      var g2 = ctx.createLinearGradient(0, gy + gh - hh, 0, gy + gh);
      g2.addColorStop(0, F.ladeWarn); g2.addColorStop(1, TF.licht);
      ctx.fillStyle = g2; rundeFlaeche(gx, gy + gh - hh, 6, hh, 3); ctx.fill();
    }
    ctx.strokeStyle = mitAlpha(TF.licht, 0.4); ctx.lineWidth = 1;
    rundeFlaeche(gx, gy, 6, gh, 3); ctx.stroke();
    if (sp.zustand === "bereit") {
      ctx.save();
      ctx.translate(x, 600); ctx.rotate(-Math.PI / 2);
      ctx.fillStyle = mitAlpha(TF.licht, 0.55); ctx.font = "bold 9px Arial"; ctx.textAlign = "center";
      ctx.fillText("HALTEN = STÄRKER", 0, 3);
      ctx.restore();
      /* wartende Kugel am Stift */
      kugelBild(x, KANAL.stift, 1);
    }
    ctx.save();
    ctx.translate(x, 500); ctx.rotate(-Math.PI / 2);
    ctx.fillStyle = mitAlpha(TF.licht, 0.35);
    ctx.font = "bold 8px Arial"; ctx.textAlign = "center";
    ctx.fillText("ABSCHUSS", 0, 0);
    ctx.restore();
  }

  function kugelBild(x, y, groesse) {
    var r = RK * (groesse || 1);
    var g = ctx.createRadialGradient(x - 3, y - 4, 1, x, y, r);
    g.addColorStop(0, F.glanz);
    g.addColorStop(0.5, TF.kugel);
    g.addColorStop(1, F.kugelEnde);
    ctx.save();
    ctx.shadowColor = "rgba(0,0,0,0.55)"; ctx.shadowBlur = 5; ctx.shadowOffsetY = 2;
    ctx.beginPath(); ctx.arc(x, y, r, 0, Math.PI * 2);
    ctx.fillStyle = g; ctx.fill();
    ctx.restore();
    ctx.strokeStyle = F.kugelRand; ctx.lineWidth = 1; ctx.stroke();
  }

  function zeichneKugeln() {
    var jetzt = performance.now();
    sp.kugeln.forEach(function (k, i) {
      if (k.zustand === "weg" || k.zustand === "schloss") return;
      var s = spur[i];
      if (!s || jetzt - s.t > 120 || Math.hypot(s.x - k.x, s.y - k.y) > 60) { spur[i] = { x: k.x, y: k.y, t: jetzt }; }
      if (k.ebene === "deck") {
        /* Kugel auf dem Oberdeck: Schattenwurf nach unten */
        ctx.save();
        ctx.globalAlpha = 0.35;
        ctx.beginPath(); ctx.ellipse(k.x + 4, k.y + 9, RK * 0.95, RK * 0.7, 0, 0, Math.PI * 2);
        ctx.fillStyle = "#000"; ctx.fill();
        ctx.restore();
      }
      ctx.save();
      if (k.zustand === "bahn") { ctx.shadowColor = TF.licht; ctx.shadowBlur = 14; }
      kugelBild(k.x, k.y, k.ebene === "deck" ? 1.06 : 1);
      ctx.restore();
      if (sp.kugeln.length > 1) {
        ctx.fillStyle = mitAlpha(F.text, 0.45);
        ctx.font = "bold 6px Arial"; ctx.textAlign = "center";
      }
    });
  }

  function zeichnePopups() {
    ctx.textAlign = "center";
    ctx.font = "bold 14px Georgia, serif";
    for (var i = sp.popups.length - 1; i >= 0; i--) {
      var p = sp.popups[i];
      ctx.globalAlpha = Math.max(0, Math.min(1, p.leben));
      ctx.fillStyle = F.punkte;
      ctx.fillText(p.text, p.x, p.y);
      ctx.globalAlpha = 1;
      p.leben -= 0.018;
      if (p.leben <= 0) sp.popups.splice(i, 1);
    }
  }

  function zeichne() {
    zeichneGrund();
    zeichneWaende();
    zeichneLanes();
    zeichneRampen();
    zeichneRollover();
    zeichneSpinner();
    zeichneScoop();
    zeichneSchloss();
    zeichneMagnet();
    zeichneBumper(sp.tisch.bumper);
    zeichneSchleudern();
    zeichneZiele();
    zeichneFlipper();
    zeichneDeck();
    zeichneKanal();
    zeichneGehaeuse();
    zeichneKugeln();
    zeichneApron();
    zeichnePopups();
    if (sp.bonusZeit > 0) {
      ctx.save();
      ctx.strokeStyle = mitAlpha(F.glanz, 0.25 + 0.15 * Math.sin(performance.now() / 180));
      ctx.lineWidth = 3;
      rundeFlaeche(3, 3, W - 6, H - 6, 14); ctx.stroke();
      ctx.restore();
    }
    if (meldungZeit > 0) {
      ctx.save();
      ctx.globalAlpha = Math.min(1, meldungZeit * 2);
      ctx.fillStyle = mitAlpha(TF.grund2, 0.85);
      rundeFlaeche(58, 700, 284, 26, 8); ctx.fill();
      ctx.fillStyle = TF.licht; ctx.font = "bold 11px Arial"; ctx.textAlign = "center";
      ctx.fillText(sp.meldung, 200, 717);
      ctx.restore();
    }
    if (blitz > 0) {
      ctx.fillStyle = mitAlpha(F.glanz, blitz * 0.14);
      ctx.fillRect(0, 0, W, H);
      blitz -= 0.05;
    }
  }

  /* Ereignisse aus dem Kern: Sound, Blitz, Kugelverlust, Spielende. */
  function verarbeite() {
    FL.ereignisseHolen(sp).forEach(function (e) {
      if (klaenge[e.art]) klaenge[e.art]();
      if (e.art === "rampe" || e.art === "bonus" || e.art === "scoop" || e.art === "lichter" ||
        e.art === "jackpot" || e.art === "multiball" || e.art === "mission") blitz = 1;
      if (e.art === "verlust") {
        pause = 0.9;
        spur = {};
        if (sp.zustand === "vorbei") spielEnde();
        else $("hinweis").textContent = sp.meldung;
      }
      if (e.art === "extra") $("hinweis").textContent = "Extrakugel!";
    });
  }

  /* ------------------------------------------------------------- Spielschleife */
  var sammler = 0, letzte = 0;
  function schleife(zeit) {
    if (!letzte) letzte = zeit;
    var dt = Math.min(0.06, (zeit - letzte) / 1000 || 0);
    letzte = zeit;
    if (!document.hidden) {
      if (pause > 0) pause -= dt;
      else if (laeuft) {
        sammler += dt;
        var n = 0;
        while (sammler >= FL.TAKT && n < 14) {
          FL.schritt(sp, FL.TAKT);
          sammler -= FL.TAKT;
          n++;
          if (sp.zustand === "bereit" || sp.zustand === "vorbei") { sammler = 0; break; }
        }
      }
      verarbeite();
      if (sp.meldung && sp.meldung !== "") { meldungZeit = 2.2; sp.meldung = ""; }
      meldungZeit = Math.max(0, meldungZeit - dt);
      ziehStand();
      zeichne();
      zeigeZahlen();
    }
    requestAnimationFrame(schleife);
  }

  function ziehStand() {
    var el = $("stift");
    if (sp.abschuss.kraft > 0.01) el.style.setProperty("--lade", Math.round(sp.abschuss.kraft * 100) + "%");
    else if (el.style.getPropertyValue("--lade")) el.style.removeProperty("--lade");
  }

  /* ------------------------------------------------------------- Tischgröße */
  function passeTisch() {
    var box = document.querySelector(".tisch");
    var lw = document.querySelector(".leinwand");
    if (!box || !lw) return;
    var r = box.getBoundingClientRect();
    var cs = getComputedStyle(box);
    var freiW = Math.max(80, r.width - parseFloat(cs.paddingLeft) - parseFloat(cs.paddingRight));
    var freiH = Math.max(80, r.height - parseFloat(cs.paddingTop) - parseFloat(cs.paddingBottom));
    var verh = W / H;                       // 440 : 820
    var h = Math.max(200, Math.min(freiH, freiW / verh));
    var w = h * verh;
    lw.style.width = Math.round(w) + "px";
    lw.style.height = Math.round(h) + "px";
    dpr = Math.min(2.5, window.devicePixelRatio || 1);
    var bw = Math.round(W * dpr), bh = Math.round(H * dpr);
    if (leinwand.width !== bw || leinwand.height !== bh) { leinwand.width = bw; leinwand.height = bh; }
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
  }
  window.addEventListener("resize", passeTisch);
  window.addEventListener("orientationchange", passeTisch);
  if (window.ResizeObserver) new ResizeObserver(function () { passeTisch(); }).observe(document.querySelector(".tisch"));
  document.getElementById("regeln").addEventListener("toggle", passeTisch);

  function stilNeu() { farbenLaden(); zeichne(); }
  if (window.TGStil && TGStil.beiWechsel) TGStil.beiWechsel(stilNeu);

  /* ------------------------------------------------------------------ Start */
  farbenLaden();
  baueTischLeiste();
  tischInfo();
  passeTisch();
  startbildschirm();
  zeichne();
  requestAnimationFrame(schleife);
  document.addEventListener("visibilitychange", function () { if (!document.hidden) { letzte = 0; sammler = 0; } });
})();
