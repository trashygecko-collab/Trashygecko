// Simulation der drei Pick-Bot-Stufen – nutzt exakt dieselben Regeln und
// Strategien wie pig.html (DP-Wertiteration für "schwer").
const ZIEL = 100;
let dpW = null, dpIterationen = 0;

function dpBauen() {
  if (dpW) return dpW;
  const N = ZIEL, idx = (a, b) => a * N + b;
  let W = new Float64Array(N * N).fill(0.5);
  const V = new Float64Array(N + 8);
  let it = 0;
  for (; it < 60; it++) {
    const neu = new Float64Array(N * N);
    let maxDiff = 0;
    for (let p = 0; p < N; p++) {
      const rest = N - p;
      for (let q = 0; q < N; q++) {
        const gegner = 1 - W[idx(q, p)];
        for (let t = rest - 1; t >= 1; t--) {
          const sicher = (p + t >= N) ? 1 : 1 - W[idx(q, p + t)];
          let weiter = gegner / 6;
          for (let a = 2; a <= 6; a++) { const tn = t + a; weiter += (tn >= rest ? 1 : V[tn]) / 6; }
          V[t] = sicher > weiter ? sicher : weiter;
        }
        let w0 = gegner / 6;
        for (let a = 2; a <= 6; a++) w0 += ((a >= rest ? 1 : V[a])) / 6;
        neu[idx(p, q)] = w0;
        const d = Math.abs(w0 - W[idx(p, q)]);
        if (d > maxDiff) maxDiff = d;
      }
    }
    W = neu;
    if (maxDiff < 1e-6) break;
  }
  dpIterationen = it + 1;
  dpW = W;
  return W;
}

function dpEntscheidung(p, q, t) {
  const W = dpBauen(), N = ZIEL, idx = (a, b) => a * N + b;
  p = Math.max(0, Math.min(N - 1, Math.round(p)));
  q = Math.max(0, Math.min(N - 1, Math.round(q)));
  t = Math.max(0, Math.min(N - 1, Math.round(t)));
  const rest = N - p;
  const sicherWert = (p + t >= N) ? 1 : 1 - W[idx(q, p + t)];
  let weiterWert = (1 - W[idx(q, p)]) / 6;
  if (t > 0) {
    const V = new Float64Array(N + 8);
    for (let tt = rest - 1; tt >= 1; tt--) {
      const s = (p + tt >= N) ? 1 : 1 - W[idx(q, p + tt)];
      let w = (1 - W[idx(q, p)]) / 6;
      for (let a = 2; a <= 6; a++) { const tn = tt + a; w += (tn >= rest ? 1 : V[tn]) / 6; }
      V[tt] = s > w ? s : w;
    }
    weiterWert = (1 - W[idx(q, p)]) / 6;
    for (let a = 2; a <= 6; a++) { const tn = t + a; weiterWert += (tn >= rest ? 1 : V[tn]) / 6; }
  }
  return { sicherWert, weiterWert, sicher: sicherWert > weiterWert };
}

function entscheidung(stufe, p, t, gegner) {
  if (t === 0) return "wuerfeln";
  if (stufe === "schwer") return dpEntscheidung(p, gegner, t).sicher ? "sichern" : "wuerfeln";
  if (stufe === "normal") {
    let grenze = 21;
    if (p - gegner >= 25) grenze -= 4;
    if (gegner - p >= 25) grenze += 6;
    if (p + t >= ZIEL) return "sichern";
    if (t >= grenze + 4) return "sichern";
    if (t >= grenze - 3) return Math.random() < 0.18 ? "wuerfeln" : "sichern";
    if (Math.random() < 0.07) return t > 30 ? "wuerfeln" : "sichern";
    return "wuerfeln";
  }
  const grenze = 55 + (p > gegner ? 6 : 0) - (gegner - p > 35 ? 6 : 0);
  if (p + t >= ZIEL) return "sichern";
  if (t >= 58) return Math.random() < 0.6 ? "sichern" : "wuerfeln";
  if (t >= grenze) return Math.random() < 0.35 ? "sichern" : "wuerfeln";
  if (t >= 10 && t <= 24 && Math.random() < 0.3) return "sichern";
  if (Math.random() < 0.5) return t > 26 ? "sichern" : "wuerfeln";
  return "wuerfeln";
}

function w6() { return 1 + Math.floor(Math.random() * 6); }

function partie(stufen) {
  const n = stufen.length;
  const p = new Array(n).fill(0);
  let zug = 0, zuege = 0;
  while (zuege < 4000) {
    zuege++;
    const gegner = Math.max(...p.filter((_, i) => i !== zug));
    let t = 0;
    for (;;) {
      const augen = w6();
      if (augen === 1) { t = 0; break; }
      t += augen;
      if (entscheidung(stufen[zug], p[zug], t, gegner) === "sichern") {
        p[zug] += t; t = 0; break;
      }
      if (p[zug] + t >= ZIEL) { p[zug] += t; t = 0; break; }
    }
    if (p[zug] >= ZIEL) return { sieger: zug, punkte: p.slice(), zuege };
    zug = (zug + 1) % n;
  }
  return { sieger: -1, punkte: p.slice(), zuege };
}

function lauf(stufen, runden) {
  const sieg = new Array(stufen.length).fill(0);
  const punkte = new Array(stufen.length).fill(0);
  let zuege = 0;
  for (let i = 0; i < runden; i++) {
    const r = partie(stufen);
    zuege += r.zuege;
    if (r.sieger >= 0) sieg[r.sieger]++;
    r.punkte.forEach((x, k) => (punkte[k] += x));
  }
  return {
    runden,
    sieg: sieg.map((x) => ((100 * x) / runden).toFixed(1) + " %"),
    punkte: punkte.map((x) => (x / runden).toFixed(1)),
    zuege: (zuege / runden).toFixed(1)
  };
}

const t0 = Date.now();
dpBauen();
const dpMs = Date.now() - t0;
console.log(`DP: ${dpIterationen} Iterationen, ${dpMs} ms, W(0,0)=${dpW[0].toFixed(4)}`);

const R = 20000;
const faelle = [
  [["schwer", "normal"], "schwer vs normal"],
  [["schwer", "leicht"], "schwer vs leicht"],
  [["normal", "leicht"], "normal vs leicht"],
  [["schwer", "schwer"], "schwer vs schwer"],
  [["normal", "normal"], "normal vs normal"],
  [["schwer", "normal", "leicht", "leicht"], "4er: schwer+normal+2xleicht"],
  [["schwer", "schwer", "normal", "normal"], "4er: 2xschwer+2xnormal"]
];
for (const [stufen, name] of faelle) {
  console.log(name.padEnd(32), JSON.stringify(lauf(stufen, R)));
}

// Entscheidungsprofil: wann sichert der schwere Bot?
const profil = [];
for (let t = 2; t <= 40; t += 2) {
  profil.push(t + ":" + (dpEntscheidung(0, 0, t).sicher ? "S" : "w"));
}
console.log("schwer bei 0/0, Zugsumme 2..40:", profil.join(" "));
for (const p of [0, 20, 60, 85]) {
  const zeile = [];
  for (let t = 2; t <= 40; t += 2) zeile.push(t + ":" + (dpEntscheidung(p, 40, t).sicher ? "S" : "w"));
  console.log(`schwer eigen ${p} / Gegner 40:`, zeile.join(" "));
}
