"use strict";
var FL = require("../../spiele/flipper/flipper-kern.js");
var DT = 1 / 240;
FL.rettungAn(false);

var sp = FL.neuesSpiel(0);
FL.partieStarten(sp);
var k = FL.neuerBall(160, 640, 0, 0, "haupt");
sp.kugeln.push(k); sp.zustand = "spiel";

function info() {
  var sohle = sp.tisch.waende[8], sl = sp.tisch.schleudern[0];
  var c1 = FL.naechsterPunkt(k.x, k.y, sohle.x1, sohle.y1, sohle.x2, sohle.y2);
  var c2 = FL.naechsterPunkt(k.x, k.y, sl.x1, sl.y1, sl.x2, sl.y2);
  return "dSohle=" + Math.hypot(k.x - c1.x, k.y - c1.y).toFixed(2) +
    " c1=" + c1.x.toFixed(1) + "/" + c1.y.toFixed(1) +
    " dSling=" + Math.hypot(k.x - c2.x, k.y - c2.y).toFixed(2) +
    " c2=" + c2.x.toFixed(1) + "/" + c2.y.toFixed(1) +
    " |" + k.x.toFixed(2) + "/" + k.y.toFixed(2) + " v=" + k.vx.toFixed(1) + "/" + k.vy.toFixed(1);
}

console.log("start " + info());
for (var i = 1; i <= 40; i++) {
  FL.schritt(sp, DT);
  console.log("s" + i + " " + info());
}
var h = sp.tisch;
console.log("waende: " + h.waende.map(function (w, j) { return j + ":" + w.x1 + "," + w.y1 + "-" + w.x2 + "," + w.y2; }).join(" | "));
console.log("rettungspunkt zwerg: " + JSON.stringify(FL.sichererPunkt(h)));
