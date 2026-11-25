# -*- coding: utf-8 -*-
"""Patches für spiele/sokoban/sokoban.html – 100 Level, Fortschritt, Größe.

Aufruf: python sokoban_spiel_patch.py [--ziel PFAD]

Jede Ersetzung muss genau einmal vorkommen; sonst bricht das Skript ab und
schreibt nichts. So kann es gefahrlos auf beide Homepage-Ablagen laufen.
"""

import argparse
import io
import sys

STANDARD = r"C:\Users\DarkP\Documents\Homepage\spiele\sokoban\sokoban.html"

ERSETZUNGEN = [
    # 1. Statuszeile: aktuelles Level und Fortschritt
    (
        '  <span>Bestwert <b id="bestwert">–</b></span>\n </div>',
        '  <span>Level <b id="levelAnzeige">01 / 100</b></span>\n'
        '  <span>Geschafft <b id="fortschritt">0/100</b></span>\n'
        '  <span>Bestwert <b id="bestwert">–</b></span>\n </div>',
    ),
    # 2. Größere Level brauchen mehr Platz auf dem Schirm
    (
        "canvas{display:block;width:100%;max-width:480px;height:auto;",
        "canvas{display:block;max-width:100%;height:auto;",
    ),
    # 3. Leveldaten auslagern
    (
        '<script>\n(() => {\nconst canvas=document.getElementById("feld");',
        '<script src="level.js"></script>\n<script>\n(() => {\n'
        'const canvas=document.getElementById("feld");',
    ),
    # 4. altes Array umbenennen
    (
        'const LEVEL=[\n ["#######","#  .  #"',
        '// level.js liefert 100 Level; die alten zehn bleiben als Notnagel stehen.\n'
        'const LEVEL_ALT=[\n ["#######","#  .  #"',
    ),
    # 5. Level laden + Fortschritt
    (
        ' ["##########","#   #    #","# $ # .  #","#   #    #","#   @    #","#  .  $  #","##########"]\n];',
        ' ["##########","#   #    #","# $ # .  #","#   #    #","#   @    #","#  .  $  #","##########"]\n];\n'
        "const LEVEL=(window.SOKOBAN_LEVEL&&window.SOKOBAN_LEVEL.length)\n"
        " ?window.SOKOBAN_LEVEL.map(e=>Array.isArray(e)?e:e.karte)\n"
        " :LEVEL_ALT;\n\n"
        "// Fortschritt: welche Level schon geschafft sind (bleibt im Browser).\n"
        'const FERTIG="sokoban-fertig";\n'
        "function geschafftLaden(){\n"
        ' try{ return new Set(JSON.parse(localStorage.getItem(FERTIG)||"[]")); }catch(e){ return new Set(); }\n'
        "}\n"
        "function geschafftMerken(nr){\n"
        " const s=geschafftLaden();\n"
        " s.add(nr);\n"
        " try{ localStorage.setItem(FERTIG,JSON.stringify([...s].sort((a,b)=>a-b))); }catch(e){}\n"
        "}",
    ),
    # 6. Zellgröße an den Bildschirm anpassen
    (
        "const Z=40;",
        "let Z=40;\n"
        "// So viele Pixel wie möglich, aber nie mehr, als auf den Bildschirm passen –\n"
        "// hochkant wie quer, Handy wie Rechner.\n"
        "function feldGroesse(){\n"
        " const buehne=canvas.parentElement;\n"
        " const platzB=Math.max(240,(buehne.clientWidth||360)-18);\n"
        " const platzH=Math.max(200,Math.min(window.innerHeight*0.6,720));\n"
        " Z=Math.max(14,Math.min(64,Math.floor(Math.min(platzB/spalten,platzH/zeilen))));\n"
        "}",
    ),
    # 7. levelLaden: Größe berechnen, Anzeige aktualisieren
    (
        " zuege=0; schieben=0; verlauf=[]; geloest=false;\n"
        " canvas.width=spalten*Z;\n"
        " canvas.height=zeilen*Z;\n"
        ' $("meldung").hidden=true;\n'
        ' $("levelWahl").value=String(nummer);\n'
        " zahlenZeigen();\n"
        " bestwertZeigen();\n"
        ' $("hinweis").textContent="Wischen oder Pfeile: Der Arbeiter geht einen Schritt.";\n'
        " zeichne();\n"
        "}",
        " zuege=0; schieben=0; verlauf=[]; geloest=false;\n"
        " feldGroesse();\n"
        " canvas.width=spalten*Z;\n"
        " canvas.height=zeilen*Z;\n"
        ' $("meldung").hidden=true;\n'
        ' $("levelWahl").value=String(nummer);\n'
        " zahlenZeigen();\n"
        " bestwertZeigen();\n"
        " levelAnzeige();\n"
        ' $("hinweis").textContent="Wischen oder Pfeile: Der Arbeiter geht einen Schritt.";\n'
        " zeichne();\n"
        "}\n\n"
        "function levelAnzeige(){\n"
        " const fertig=geschafftLaden();\n"
        ' $("levelAnzeige").textContent=String(levelNr+1).padStart(2,"0")+" / "+LEVEL.length;\n'
        ' $("fortschritt").textContent=fertig.size+"/"+LEVEL.length;\n'
        ' const wahl=$("levelWahl");\n'
        " for(let i=0;i<wahl.options.length&&i<LEVEL.length;i++){\n"
        '  const nr=String(i+1).padStart(2,"0");\n'
        '  wahl.options[i].textContent="Level "+nr+(fertig.has(i)?" ✓":"");\n'
        " }\n"
        "}",
    ),
    # 8. Sieg: Level als geschafft merken
    (
        " const best=bestwertLaden(levelNr);\n"
        " if(!best||zuege<best){\n"
        '  try{ localStorage.setItem("sokoban-best-"+levelNr,String(zuege)); }catch(e){}\n'
        "  bestwertZeigen();\n"
        " }",
        " const best=bestwertLaden(levelNr);\n"
        " if(!best||zuege<best){\n"
        '  try{ localStorage.setItem("sokoban-best-"+levelNr,String(zuege)); }catch(e){}\n'
        "  bestwertZeigen();\n"
        " }\n"
        " geschafftMerken(levelNr);\n"
        " levelAnzeige();",
    ),
    # 9. Auswahl füllen + Neuzeichnen bei Fenstergröße
    (
        "LEVEL.forEach((_,i)=>{\n"
        ' const o=document.createElement("option");\n'
        " o.value=String(i);\n"
        ' o.textContent="Level "+(i+1);\n'
        ' $("levelWahl").appendChild(o);\n'
        "});",
        "LEVEL.forEach((_,i)=>{\n"
        ' const o=document.createElement("option");\n'
        " o.value=String(i);\n"
        ' o.textContent="Level "+String(i+1).padStart(2,"0");\n'
        ' $("levelWahl").appendChild(o);\n'
        "});\n"
        "levelAnzeige();\n"
        'window.addEventListener("resize",()=>{\n'
        " feldGroesse();\n"
        " canvas.width=spalten*Z;\n"
        " canvas.height=zeilen*Z;\n"
        " zeichne();\n"
        "});",
    ),
    # 10. Enter oder Leertaste nimmt das nächste Level an
    (
        ' else if(k==="z"||k==="Z"||k==="Backspace"){ e.preventDefault(); zurueck(); }\n'
        ' else if(k==="r"||k==="R"){ e.preventDefault(); levelLaden(levelNr); }',
        ' else if(k==="z"||k==="Z"||k==="Backspace"){ e.preventDefault(); zurueck(); }\n'
        ' else if(k==="r"||k==="R"){ e.preventDefault(); levelLaden(levelNr); }\n'
        ' else if((k==="Enter"||k===" ")&&!$("meldung").hidden){ e.preventDefault(); $("meldungKnopf").click(); }',
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
            print("FEHLER %d: %d Treffer für:\n%s" % (i + 1, n, alt[:120]))
            return 1
        s = s.replace(alt, neu)
    io.open(args.ziel, "w", encoding="utf-8").write(s)
    print("%d Ersetzungen geschrieben: %s" % (len(ERSETZUNGEN), args.ziel))
    return 0


if __name__ == "__main__":
    sys.exit(main())
