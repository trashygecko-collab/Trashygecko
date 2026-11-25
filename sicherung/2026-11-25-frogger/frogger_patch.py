# -*- coding: utf-8 -*-
"""Patches für spiele/frogger/frogger.html.

Aufruf: python frogger_patch.py [--ziel PFAD]

Jede Ersetzung muss genau einmal vorkommen, sonst schreibt das Skript nichts.
"""

import argparse
import io
import sys

STANDARD = r"C:\Users\DarkP\Documents\Homepage\spiele\frogger\frogger.html"

ERSETZUNGEN = [
    # ---------------------------------------------------------------- Anzeige
    (
        '  <span>Zeit <b id="zeit">35</b></span>\n'
        '  <span>Rekord <b id="rekord">0</b></span>',
        '  <span>Zeit <b id="zeit">35</b></span>\n'
        '  <span>Buchten <b id="buchtenStand">0/5</b></span>\n'
        '  <span>Rekord <b id="rekord">0</b></span>',
    ),
    # Menüleiste: Levelauswahl und Level neu
    (
        '  <button id="btnNeu">Neues Spiel</button>\n'
        '  <button id="btnPause">Pause</button>',
        '  <button id="btnNeu">Neues Spiel</button>\n'
        '  <select id="levelWahl" aria-label="Level wählen"></select>\n'
        '  <button id="btnNochmal">Level neu</button>\n'
        '  <button id="btnPause">Pause</button>',
    ),
    (
        '.menueleiste button:hover{background:var(--hover)}',
        '.menueleiste button:hover{background:var(--hover)}\n'
        '.menueleiste select{font-family:inherit;font-size:13px;padding:7px 8px;border-radius:6px;\n'
        ' border:1px solid var(--rand-1);background:var(--panel);color:var(--gold)}',
    ),
    # ------------------------------------------------------- Zielkoordinaten
    # Die Buchten liegen bei den Spalten 1, 3, 5, 7, 9 – gezeichnet wurde
    # bisher auf der Zellgrenze ((i*2+1)*Z), geprüft wird die Zellmitte.
    # Das war der Fehler: Wer in die gezeichnete Bucht sprang, stand eine halbe
    # Zelle daneben und starb „zwischen den Buchten“.
    (
        " // Buchten\n for(let i=0;i<5;i++)zeichneBucht(i,((i*2)+1)*Z);",
        " // Buchten: Mittelpunkt der Spalten 1, 3, 5, 7, 9\n"
        " for(let i=0;i<5;i++)zeichneBucht(i,BUCHT_SPALTEN[i]*Z+Z/2);",
    ),
    # ---------------------------------------------------------- Konstanten
    (
        "const ZEIT_PRO_LEBEN=35;",
        "const ZEIT_PRO_LEBEN=35;\n"
        "// Die fünf Buchten sitzen in diesen Spalten (Mitte = Spalte*Z+Z/2).\n"
        "const BUCHT_SPALTEN=[1,3,5,7,9];",
    ),
    # ------------------------------------------------ Bucht aus der Spalte
    (
        " if(reihe===HEIM_ZEILE){\n"
        "  const spalte=frosch.zielSpalte;\n"
        "  const bucht=spalte%2===1&&spalte<=9?(spalte-1)/2:-1;",
        " if(reihe===HEIM_ZEILE){\n"
        "  const spalte=Math.round((frosch.x-Z/2)/Z);\n"
        "  const bucht=BUCHT_SPALTEN.indexOf(spalte);",
    ),
    # ------------------------------------- Reihenfolge: Sprung vor der Uhr
    # Vorher lief die Zeitprüfung vor dem Sprungabschluss: Lief die Zeit im
    # letzten Sprungbild ab, starb der Frosch genau beim Landen in der Bucht –
    # Game Over trotz Treffer.
    (
        " // Zeit\n"
        " zeit-=dt;\n"
        ' $("zeit").textContent=Math.max(0,Math.ceil(zeit));\n'
        " if(zeit<=0){\n"
        '  sterben("Zeit abgelaufen");\n'
        "  return;\n"
        " }\n"
        "\n"
        " // Sprung\n"
        " if(frosch.huepfend){\n"
        "  frosch.t=Math.min(1,frosch.t+dt/0.13);\n"
        "  const e=frosch.t<0.5?2*frosch.t*frosch.t:1-Math.pow(-2*frosch.t+2,2)/2;\n"
        "  frosch.x=frosch.startX+(frosch.zielX-frosch.startX)*e;\n"
        "  frosch.y=frosch.startY+(frosch.zielY-frosch.startY)*e;\n"
        "  if(frosch.t>=1)sprungFertig();\n"
        "  return;\n"
        " }",
        " // Sprung zuerst: Wer im letzten Sprungbild die Bucht erreicht, hat sie\n"
        " // erreicht – die Uhr darf den Treffer nicht mehr wegnehmen.\n"
        " if(frosch.huepfend){\n"
        "  frosch.t=Math.min(1,frosch.t+dt/0.13);\n"
        "  const e=frosch.t<0.5?2*frosch.t*frosch.t:1-Math.pow(-2*frosch.t+2,2)/2;\n"
        "  frosch.x=frosch.startX+(frosch.zielX-frosch.startX)*e;\n"
        "  frosch.y=frosch.startY+(frosch.zielY-frosch.startY)*e;\n"
        "  if(frosch.t>=1)sprungFertig();\n"
        "  return;\n"
        " }\n"
        "\n"
        " // Zeit\n"
        " zeit-=dt;\n"
        ' $("zeit").textContent=Math.max(0,Math.ceil(zeit));\n'
        " if(zeit<=0){\n"
        '  sterben("Zeit abgelaufen");\n'
        "  return;\n"
        " }",
    ),
    # ------------------------------------------------------- Leveldefinitionen
    (
        "function neueBahnen(){\n"
        " bahnen=[];\n"
        " const faktor=1+(level-1)*0.07;",
        "/* Zwölf Level mit eigener Verkehrslage. Die Grundwerte sind die alten,\n"
        "   damit Level 1 genau wie bisher spielbar ist; später kommen andere\n"
        "   Richtungen, dichtere Straßen und größere Flöße dazu. Ab Level 1 steigt\n"
        "   das Straßentempo um 4 % je Level, das Flusstempo nur um 1,5 % – die\n"
        "   Steigerung ist anfangs kaum zu merken und über zehn Level deutlich. */\n"
        "const LEVEL_DEFS=[\n"
        " {fluss:[[30,1,96,70],[36,-1,66,52],[44,1,128,84],[32,-1,44,46],[40,1,80,64]],\n"
        "  strasse:[[48,-1,52,96],[34,1,96,110],[60,-1,46,120],[42,1,56,84],[56,-1,104,130]]},\n"
        " {fluss:[[32,1,80,68],[38,-1,72,54],[42,1,112,82],[34,-1,50,48],[38,1,96,60]],\n"
        "  strasse:[[50,-1,52,92],[36,1,88,104],[58,-1,50,112],[44,1,56,80],[54,-1,96,120]]},\n"
        " {fluss:[[28,1,104,72],[40,-1,60,50],[46,1,120,86],[30,-1,48,52],[42,1,72,66]],\n"
        "  strasse:[[46,-1,56,100],[38,1,104,108],[62,-1,44,116],[40,1,60,88],[58,-1,112,126]]},\n"
        " {fluss:[[34,1,88,66],[34,-1,78,56],[40,1,104,80],[36,-1,54,44],[36,1,88,62]],\n"
        "  strasse:[[52,-1,48,88],[32,1,80,112],[56,-1,54,108],[46,1,52,76],[52,-1,88,118]]},\n"
        " {fluss:[[30,-1,96,64],[42,1,64,48],[48,-1,132,88],[28,1,40,50],[44,-1,76,58]],\n"
        "  strasse:[[44,1,52,94],[40,-1,100,102],[64,1,42,110],[38,-1,60,86],[60,1,120,124]]},\n"
        " {fluss:[[36,1,72,60],[32,-1,84,58],[38,1,96,78],[38,-1,58,42],[34,1,104,68]],\n"
        "  strasse:[[54,-1,44,84],[30,1,92,116],[54,-1,58,104],[48,1,48,72],[50,-1,80,114]]},\n"
        " {fluss:[[26,1,112,74],[44,-1,56,46],[50,1,140,90],[26,-1,44,54],[46,1,68,62]],\n"
        "  strasse:[[42,1,60,98],[42,-1,108,100],[66,1,40,106],[36,-1,64,90],[62,1,128,128]]},\n"
        " {fluss:[[38,-1,80,62],[30,1,92,60],[36,-1,112,76],[40,1,62,40],[32,-1,96,70]],\n"
        "  strasse:[[56,-1,48,80],[28,1,84,120],[52,-1,62,100],[50,1,44,68],[48,-1,72,110]]},\n"
        " {fluss:[[28,1,120,76],[46,-1,52,44],[42,1,96,74],[24,-1,48,56],[48,1,60,56]],\n"
        "  strasse:[[40,1,64,102],[44,-1,116,96],[68,1,38,102],[34,-1,68,94],[64,1,136,132]]},\n"
        " {fluss:[[40,-1,72,58],[28,1,100,62],[34,-1,124,74],[42,1,56,38],[30,-1,88,72]],\n"
        "  strasse:[[58,-1,52,76],[26,1,76,124],[50,-1,66,96],[52,1,40,64],[46,-1,64,106]]},\n"
        " {fluss:[[32,1,108,70],[48,-1,48,42],[46,1,84,68],[40,-1,64,50],[52,-1,52,54]],\n"
        "  strasse:[[38,1,68,106],[46,-1,124,92],[70,1,36,98],[32,-1,72,98],[66,1,144,136]]},\n"
        " {fluss:[[42,-1,64,54],[26,1,108,64],[32,-1,136,76],[44,1,52,36],[28,-1,80,74]],\n"
        "  strasse:[[60,-1,56,72],[24,1,68,128],[48,-1,70,92],[54,1,36,60],[44,-1,56,102]]}\n"
        "];\n"
        "\n"
        "function neueBahnen(){\n"
        " bahnen=[];\n"
        " const def=LEVEL_DEFS[(Math.max(1,level)-1)%LEVEL_DEFS.length];\n"
        " const faktor=1+(level-1)*0.04;          // Straße\n"
        " const flussFaktor=1+(level-1)*0.015;    // Fluss, nur ganz sacht",
    ),
    # Fluss-Bahnen aus der Definition
    (
        " // Fluss: Baumstämme und Schildkröten, abwechselnd beide Richtungen\n"
        " const fluss=[\n"
        '  {art:"holz",  reihe:1, tempo:30, richtung:1,  breite:96,  abstand:70},\n'
        '  {art:"schild",reihe:2, tempo:36, richtung:-1, breite:66,  abstand:52},\n'
        '  {art:"holz",  reihe:3, tempo:44, richtung:1,  breite:128, abstand:84},\n'
        '  {art:"schild",reihe:4, tempo:32, richtung:-1, breite:44,  abstand:46},\n'
        '  {art:"holz",  reihe:5, tempo:40, richtung:1,  breite:80,  abstand:64}\n'
        " ];\n"
        " fluss.forEach(b=>{",
        " // Fluss: Baumstämme und Schildkröten, Lage je Level\n"
        " const arten=[\"holz\",\"schild\",\"holz\",\"schild\",\"holz\"];\n"
        " const fluss=def.fluss.map((p,i)=>({art:arten[i], reihe:i+1,\n"
        "  tempo:p[0], richtung:p[1], breite:p[2], abstand:p[3]}));\n"
        " fluss.forEach(b=>{",
    ),
    (
        "  bahnen.push({...b,tempo:b.tempo*faktor,teile});\n"
        " });\n"
        "\n"
        " // Straße: Autos und Lastwagen\n"
        " const strasse=[\n"
        '  {art:"auto",  reihe:7,  tempo:48, richtung:-1, breite:52, abstand:96},\n'
        '  {art:"lkw",   reihe:8,  tempo:34, richtung:1,  breite:96, abstand:110},\n'
        '  {art:"auto",  reihe:9,  tempo:60, richtung:-1, breite:46, abstand:120},\n'
        '  {art:"auto",  reihe:10, tempo:42, richtung:1,  breite:56, abstand:84},\n'
        '  {art:"lkw",   reihe:11, tempo:56, richtung:-1, breite:104, abstand:130}\n'
        " ];",
        "  bahnen.push({...b,tempo:b.tempo*flussFaktor,teile});\n"
        " });\n"
        "\n"
        " // Straße: Autos und Lastwagen, Lage je Level\n"
        " const fahrArten=[\"auto\",\"lkw\",\"auto\",\"auto\",\"lkw\"];\n"
        " const strasse=def.strasse.map((p,i)=>({art:fahrArten[i], reihe:i+7,\n"
        "  tempo:p[0], richtung:p[1], breite:p[2], abstand:p[3]}));",
    ),
    # -------------------------------------------------------- Buchtenanzeige
    (
        "function zahlenZeigen(){\n"
        ' $("punkte").textContent=punkte;\n'
        ' $("level").textContent=level;',
        "function zahlenZeigen(){\n"
        ' $("punkte").textContent=punkte;\n'
        ' $("level").textContent=level;\n'
        ' $("buchtenStand").textContent=buchten.filter(Boolean).length+"/5";',
    ),
    # ---------------------------------------------- Fortschritt und Auswahl
    (
        "function rekordLaden(){\n"
        ' try{ return Number(localStorage.getItem("frogger-rekord")||0); }catch(e){ return 0; }\n'
        "}",
        "function rekordLaden(){\n"
        ' try{ return Number(localStorage.getItem("frogger-rekord")||0); }catch(e){ return 0; }\n'
        "}\n\n"
        "// Höchstes freigeschaltetes Level – bleibt im Browser.\n"
        "function erreichtLaden(){\n"
        " try{ return Math.max(1,Number(localStorage.getItem(\"frogger-erreicht\")||1)); }\n"
        " catch(e){ return 1; }\n"
        "}\n"
        "function erreichtMerken(n){\n"
        " if(n<=erreichtLaden())return;\n"
        " try{ localStorage.setItem(\"frogger-erreicht\",String(n)); }catch(e){}\n"
        " auswahlFuellen();\n"
        "}\n"
        "function auswahlFuellen(){\n"
        ' const wahl=$("levelWahl");\n'
        " if(!wahl)return;\n"
        " const bis=erreichtLaden();\n"
        " const alt=wahl.value;\n"
        " wahl.innerHTML=\"\";\n"
        " for(let i=1;i<=bis;i++){\n"
        '  const o=document.createElement("option");\n'
        "  o.value=String(i);\n"
        '  o.textContent="Level "+String(i).padStart(2,"0");\n'
        "  wahl.appendChild(o);\n"
        " }\n"
        " wahl.value=String(level<=bis?level:bis);\n"
        " if(alt&&Number(alt)<=bis)wahl.value=alt;\n"
        "}",
    ),
    # Levelwechsel: freischalten, Anzeige, Auswahl
    (
        "function levelGeschafft(){\n"
        ' zustand="level";\n'
        " punkte+=200;\n"
        " zahlenZeigen();\n"
        " rekordPruefen();\n"
        " const naechstes=level+1;",
        "function levelGeschafft(){\n"
        ' zustand="level";\n'
        " punkte+=200;\n"
        " erreichtMerken(level+1);\n"
        " zahlenZeigen();\n"
        " rekordPruefen();\n"
        " const naechstes=level+1;\n"
        " $(\"hinweis\").textContent=\"Alle fünf Buchten besetzt – Level \"+naechstes+\" wartet.\";",
    ),
    # ------------------------------------------------------- Level starten
    (
        "function neuesSpiel(){\n"
        " punkte=0; level=1; leben=3;",
        "function levelStarten(n){\n"
        " level=Math.max(1,n);\n"
        " punkte=0; leben=3;\n"
        " buchten=[false,false,false,false,false];\n"
        " neueBahnen();\n"
        " froschZurueck();\n"
        " zahlenZeigen();\n"
        " auswahlFuellen();\n"
        " zustand=\"laeuft\";\n"
        ' $("meldung").hidden=true;\n'
        ' $("btnPause").textContent="Pause";\n'
        " letzteZeit=performance.now();\n"
        ' $("hinweis").textContent="Level "+level+" – wischen oder Pfeile: ein Satz pro Feld.";\n'
        "}\n\n"
        "function neuesSpiel(){\n"
        " punkte=0; level=1; leben=3;",
    ),
    (
        " neueBahnen();\n"
        " froschZurueck();\n"
        " zahlenZeigen();\n"
        ' $("rekord").textContent=rekordLaden();',
        " neueBahnen();\n"
        " froschZurueck();\n"
        " zahlenZeigen();\n"
        " auswahlFuellen();\n"
        ' $("rekord").textContent=rekordLaden();',
    ),
    # Levelwechsel-Knopf: Auswahl nachziehen
    (
        "  $(\"meldung\").hidden=true;\n"
        "  level=naechstes;\n"
        "  buchten=[false,false,false,false,false];\n"
        "  neueBahnen();\n"
        "  froschZurueck();\n"
        "  zahlenZeigen();\n"
        '  zustand="laeuft";\n'
        "  letzteZeit=performance.now();",
        "  $(\"meldung\").hidden=true;\n"
        "  level=naechstes;\n"
        "  buchten=[false,false,false,false,false];\n"
        "  neueBahnen();\n"
        "  froschZurueck();\n"
        "  zahlenZeigen();\n"
        "  auswahlFuellen();\n"
        '  zustand="laeuft";\n'
        "  letzteZeit=performance.now();\n"
        '  $("hinweis").textContent="Level "+level+" – wischen oder Pfeile: ein Satz pro Feld.";',
    ),
    # ---------------------------------------------------------- Bedienung
    (
        ' else if(k===" "){ e.preventDefault(); if(zustand==="start")starten(); }\n'
        "});",
        ' else if(k===" "){ e.preventDefault(); if(zustand==="start")starten(); }\n'
        ' else if(k==="Enter"&&!$("meldung").hidden){ e.preventDefault(); $("meldungKnopf").click(); }\n'
        "});",
    ),
    (
        "$(\"btnNeu\").onclick=neuesSpiel;",
        "$(\"btnNeu\").onclick=neuesSpiel;\n"
        "$(\"btnNochmal\").onclick=()=>levelStarten(level);\n"
        '$("levelWahl").onchange=()=>levelStarten(Number($("levelWahl").value));',
    ),
]


def main():
    p = argparse.ArgumentParser()
    p.add_argument("--ziel", default=STANDARD)
    args = p.parse_args()
    s = io.open(args.ziel, encoding="utf-8").read()
    for i, (alt, neu) in enumerate(ERSETZUNGEN):
        n = s.count(alt)
        if n != 1:
            print("FEHLER %d: %d Treffer für:\n%s" % (i + 1, n, alt[:200]))
            return 1
        s = s.replace(alt, neu)
    io.open(args.ziel, "w", encoding="utf-8").write(s)
    print("%d Ersetzungen geschrieben: %s" % (len(ERSETZUNGEN), args.ziel))
    return 0


if __name__ == "__main__":
    sys.exit(main())
