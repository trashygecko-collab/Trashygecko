/* ===========================================================================
   Zentrale Spieleverwaltung für den Mehrspieler-Teil
   ---------------------------------------------------------------------------
   Diese Datei ist die einzige Liste der Spiele, die über den Connector zu
   mehreren gespielt werden. Sie wird geladen von:

     - spiele/connector/connector.html   (Spielauswahl im Multiplayer-Tab)

   Der Connector baut seine Auswahl daraus. Die Uebersicht (spiele.html) ist
   eine fertige Seite; ihre Merkmale „Multi“/„Online“ wurden aus dieser
   Liste abgeleitet und stehen fest in den Karten. Neue Eintraege hier also
   dort nachziehen (der Android-Bau liest beide Dateien). Die Go-Seite (spiele/connector/angebot.go: spielApi,
   spielAnzeige, tischspiel) wird von register_test.go gegen diese Datei
   geprüft: fehlt hier ein Spiel oder heißt anders, schlägt der Test fehl.

   Aufbau eines Eintrags:
     id      Kennung im Connector und in der Adresse (?game=…)
     name    Anzeigename (wie im Chat)
     kat     eine id aus „kategorien“
     pfad    Spielseite, absolut ab der Seitenwurzel
     api     Endpunkt ohne /api/ – der Connector hängt /api/ davor
     min/max kleinste und größte Spielerzahl am Tisch
     art     „tisch“     mehrere sitzen am Tisch, Beitritt über die Spielkarte
             „einladung“ einer lädt den anderen ein, dann geht es los
             „gemeinsam“ keine Züge, alle malen gleichzeitig
     text    ein Satz für die Auswahl

   Neues Spiel: einen Eintrag ergänzen, in Go spielApi/spielAnzeige/tischspiel
   nachziehen, fertig. Kein weiterer Ort muss angefasst werden.

   Ausnahme „malen“: kein Spiel der Spielhalle, sondern ein gemeinsames
   Kritzelbrett (eigene Kachel bei den Werkzeugen in spiele.html). Der Eintrag
   hier öffnet nur den Raum im Connector – deshalb steht malen/malen.html
   bewusst nicht in spieldaten.js.
   ========================================================================= */
window.TGMultiplayer = {
  "stand": "2026-11-22",
  "kategorien": [
    {
      "id": "wuerfel",
      "name": "Würfelspiele",
      "text": "Kniffel, 10.000, Mäxchen und andere Runden, in denen die Würfel entscheiden. Gewürfelt wird auf dem Server, alle sehen dasselbe Ergebnis."
    },
    {
      "id": "karten",
      "name": "Kartenspiele",
      "text": "UNO, Mau Mau, Poker, SKYJO und mehr. Die Karten liegen auf dem Server, jeder sieht nur seine eigene Hand."
    },
    {
      "id": "brett",
      "name": "Brettspiele",
      "text": "Schach, Mühle, Halma, Mensch ärgere dich nicht und andere Brett- und Denkspiele für zwei bis sechs Spieler."
    },
    {
      "id": "weitere",
      "name": "Weitere Multiplayer-Spiele",
      "text": "Geschicklichkeit und Tempo: Pong, Tron, Bomberman, Snake, Tetris, Malen und die Rate- und Rechenspiele."
    }
  ],
  "spiele": [
    { "id": "kniffel", "name": "Kniffel", "kat": "wuerfel", "pfad": "/spiele/brettspiele/kniffel/kniffel.html", "api": "kniffel", "min": 2, "max": 6, "art": "tisch",
      "text": "Fünf Würfel, drei Würfe, 13 Kategorien – Yahtzee mit bis zu sechs Spielern." },
    { "id": "zehntausend", "name": "10.000", "kat": "wuerfel", "pfad": "/spiele/brettspiele/zehntausend/zehntausend.html", "api": "zehntausend", "min": 2, "max": 6, "art": "tisch",
      "text": "Sechs Würfel, Punkte sammeln und sichern – wer zuerst 10.000 hat, gewinnt." },
    { "id": "maexchen", "name": "Lügen Mäxchen", "kat": "wuerfel", "pfad": "/spiele/brettspiele/maexchen/maexchen.html", "api": "maexchen", "min": 2, "max": 6, "art": "tisch",
      "text": "Zwei verdeckte Würfel, Ansage und Zweifel – wer zu hoch lügt, verliert." },
    { "id": "chicago", "name": "Chicago", "kat": "wuerfel", "pfad": "/spiele/brettspiele/chicago/chicago.html", "api": "chicago", "min": 2, "max": 6, "art": "tisch",
      "text": "Chicago: Filze, Würfe und der beste Wurf der Runde." },
    { "id": "shipcrew", "name": "Ship Captain Crew", "kat": "wuerfel", "pfad": "/spiele/brettspiele/shipcrew/shipcrew.html", "api": "shipcrew", "min": 2, "max": 6, "art": "tisch",
      "text": "Schiff, Kapitän, Crew – drei Würfel in der richtigen Reihenfolge, dann 500 Punkte." },
    { "id": "pig", "name": "Pig", "kat": "wuerfel", "pfad": "/spiele/brettspiele/pig/pig.html", "api": "pig", "min": 2, "max": 4, "art": "tisch",
      "text": "Würfeln und sammeln, aber vor der Eins aufhören – sonst ist die Runde weg." },

    { "id": "der-turm", "name": "Der Turm", "kat": "wuerfel", "pfad": "/spiele/wuerfelspiele/der-turm.html", "api": "wuerfel", "min": 2, "max": 6, "art": "tisch",
      "text": "Jeder hat eine eigene Stichzahl; Treffer holen Steine aus dem Turm – oder geben sie weiter." },
    { "id": "eins-bis-sechzehn", "name": "Eins bis Sechzehn", "kat": "wuerfel", "pfad": "/spiele/wuerfelspiele/eins-bis-sechzehn.html", "api": "wuerfel", "min": 2, "max": 6, "art": "tisch",
      "text": "Bilde die Zahlen 1 bis 16 aus drei Würfeln – erst vorwärts, dann rückwärts." },
    { "id": "fenster", "name": "Fenster würfeln", "kat": "wuerfel", "pfad": "/spiele/wuerfelspiele/fenster.html", "api": "wuerfel", "min": 2, "max": 6, "art": "tisch",
      "text": "Drei Würfel, nur Zweien zählen: eine = 1, zwei = 4, drei = 9. Ziel sind 30." },
    { "id": "rentmeister", "name": "Rentmeister", "kat": "wuerfel", "pfad": "/spiele/wuerfelspiele/rentmeister.html", "api": "wuerfel", "min": 2, "max": 6, "art": "tisch",
      "text": "Sechs feste Zahlen von 3 bis 8, jede mal einem gewählten Würfel – Summe sammeln." },
    { "id": "streichhoelzer", "name": "Streichhölzer auswürfeln", "kat": "wuerfel", "pfad": "/spiele/wuerfelspiele/streichhoelzer.html", "api": "wuerfel", "min": 2, "max": 6, "art": "tisch",
      "text": "Zehn Hölzer: Einsen wandern nach links, Sechsen in die Mitte. Wer zuletzt Hölzer hat, gewinnt." },
    { "id": "teure-sieben", "name": "Teure Sieben", "kat": "wuerfel", "pfad": "/spiele/wuerfelspiele/teure-sieben.html", "api": "wuerfel", "min": 2, "max": 6, "art": "tisch",
      "text": "Streiche 2–6 und 8–12; die Sieben erlaubt zwei Striche, doppelte Zahlen wandern weiter." },
    { "id": "unter-oder-ueber", "name": "Unter oder über sieben", "kat": "wuerfel", "pfad": "/spiele/wuerfelspiele/unter-oder-ueber.html", "api": "wuerfel", "min": 2, "max": 6, "art": "tisch",
      "text": "Sage vor dem Wurf unter, über oder genau sieben an – richtig bringt Punkte, falsch kostet." },
    { "id": "wuerfelpoker", "name": "Würfelpoker", "kat": "wuerfel", "pfad": "/spiele/wuerfelspiele/wuerfelpoker.html", "api": "wuerfel", "min": 2, "max": 6, "art": "tisch",
      "text": "Fünf Würfel, zwei Nachwürfe, drei Blätter – volle Pokerwertung, höchste Summe gewinnt." },
    { "id": "skyjo", "name": "SKYJO", "kat": "karten", "pfad": "/spiele/kartenspiele/skyjo/skyjo.html", "api": "skyjo", "min": 2, "max": 8, "art": "einladung",
      "text": "Karten aufdecken und tauschen, am Ende zählt die niedrigste Summe." },
    { "id": "maumau", "name": "Mau Mau", "kat": "karten", "pfad": "/spiele/kartenspiele/maumau/maumau.html", "api": "maumau", "min": 2, "max": 7, "art": "tisch",
      "text": "Mau Mau: Farbe oder Zahl bedienen, sieben Karten ziehen, Mau sagen." },
    { "id": "uno", "name": "UNO", "kat": "karten", "pfad": "/spiele/kartenspiele/uno/uno.html", "api": "uno", "min": 2, "max": 4, "art": "tisch",
      "text": "UNO mit Aussetzen, Richtungswechsel, +2 und +4." },
    { "id": "schwimmen", "name": "31 Schwimmen", "kat": "karten", "pfad": "/spiele/kartenspiele/schwimmen/schwimmen.html", "api": "schwimmen", "min": 2, "max": 9, "art": "tisch",
      "text": "31 Schwimmen: drei Karten tauschen und klopfen – wer am nächsten an 31 ist." },
    { "id": "arschloch", "name": "Arschloch", "kat": "karten", "pfad": "/spiele/kartenspiele/arschloch/arschloch.html", "api": "arschloch", "min": 3, "max": 9, "art": "tisch",
      "text": "Arschloch: Karten abwerfen und am Ende die Plätze tauschen." },
    { "id": "poker", "name": "Poker", "kat": "karten", "pfad": "/spiele/kartenspiele/poker/poker.html", "api": "poker", "min": 2, "max": 6, "art": "tisch",
      "text": "Texas Hold'em mit Blinds, Einsatzrunden und Showdown." },
    { "id": "cabo", "name": "Cabo", "kat": "karten", "pfad": "/spiele/kartenspiele/cabo/cabo.html", "api": "cabo", "min": 2, "max": 4, "art": "tisch",
      "text": "Cabo: vier verdeckte Karten, tauschen und möglichst niedrig bleiben." },
    { "id": "siebzehn", "name": "17 und 4", "kat": "karten", "pfad": "/spiele/kartenspiele/siebzehn/siebzehn.html", "api": "siebzehn", "min": 2, "max": 6, "art": "tisch",
      "text": "17 und 4: gegen die Bank auf 21 kommen, ohne zu überkaufen." },
    { "id": "baccara", "name": "Baccara", "kat": "karten", "pfad": "/spiele/kartenspiele/baccara/baccara.html", "api": "baccara", "min": 2, "max": 6, "art": "tisch",
      "text": "Baccara: auf Spieler oder Bank setzen, näher an neun gewinnt." },
    { "id": "sechsundsechzig", "name": "Sechsundsechzig", "kat": "karten", "pfad": "/spiele/kartenspiele/sechsundsechzig/sechsundsechzig.html", "api": "sechsundsechzig", "min": 2, "max": 2, "art": "tisch",
      "text": "Sechsundsechzig: 24 Karten, Stiche und 66 Punkte." },

    { "id": "tictactoe", "name": "Tic Tac Toe", "kat": "brett", "pfad": "/spiele/brettspiele/tictac/tictactoe.html", "api": "tictactoe", "min": 2, "max": 2, "art": "einladung",
      "text": "Drei gewinnt auf dem Neon-Brett." },
    { "id": "viergewinnt", "name": "Vier gewinnt", "kat": "brett", "pfad": "/spiele/brettspiele/viergewinnt/viergewinnt.html", "api": "viergewinnt", "min": 2, "max": 2, "art": "einladung",
      "text": "Vier gewinnt: Steine fallen lassen und vier in eine Reihe bringen." },
    { "id": "dame", "name": "Dame", "kat": "brett", "pfad": "/spiele/brettspiele/dame/dame.html", "api": "dame", "min": 2, "max": 2, "art": "einladung",
      "text": "Dame mit Schlagzwang – wer keine Figur mehr hat, verliert." },
    { "id": "paare", "name": "Paare finden", "kat": "brett", "pfad": "/spiele/brettspiele/paare/paare.html", "api": "paare", "min": 2, "max": 2, "art": "einladung",
      "text": "Paare finden: beide lösen dasselbe Brett – wer schneller ist, gewinnt." },
    { "id": "menschaerger", "name": "Mensch ärgere dich nicht", "kat": "brett", "pfad": "/spiele/brettspiele/menschaerger/menschaerger.html", "api": "maensync", "min": 2, "max": 4, "art": "tisch",
      "text": "Mensch ärgere dich nicht mit zwei bis vier Spielern am Tisch." },
    { "id": "kaesekaestche", "name": "Käsekästchen", "kat": "brett", "pfad": "/spiele/brettspiele/kaesekaestche/kaesekaestche.html", "api": "kaesekaestche", "min": 2, "max": 2, "art": "einladung",
      "text": "Käsekästchen: Linien ziehen und Kästchen schließen." },
    { "id": "sim", "name": "Sim", "kat": "brett", "pfad": "/spiele/brettspiele/sim/sim.html", "api": "sim", "min": 2, "max": 2, "art": "einladung",
      "text": "Sim: Linien setzen – wer das letzte Dreieck schließt, verliert." },
    { "id": "kalaha", "name": "Kalaha", "kat": "brett", "pfad": "/spiele/brettspiele/kalaha/kalaha.html", "api": "kalaha", "min": 2, "max": 2, "art": "einladung",
      "text": "Kalaha: Steine verteilen und im eigenen Kasten sammeln." },
    { "id": "streichholzspiel", "name": "21 Streichhölzer", "kat": "brett", "pfad": "/spiele/brettspiele/streichholzspiel/streichholzspiel.html", "api": "streichholz", "min": 2, "max": 2, "art": "einladung",
      "text": "21 Streichhölzer: reihum ein bis drei nehmen – wer das letzte nimmt, verliert." },
    { "id": "muehle", "name": "Mühle", "kat": "brett", "pfad": "/spiele/brettspiele/muehle/muehle.html", "api": "muehle", "min": 2, "max": 2, "art": "einladung",
      "text": "Mühle: setzen, ziehen, Mühlen schließen und Steine abräumen." },
    { "id": "reversi", "name": "Reversi", "kat": "brett", "pfad": "/spiele/brettspiele/reversi/reversi.html", "api": "reversi", "min": 2, "max": 2, "art": "einladung",
      "text": "Reversi: Steine einschließen und umdrehen – am Ende zählt die Mehrheit." },
    { "id": "hold", "name": "Hold the Line", "kat": "brett", "pfad": "/spiele/brettspiele/hold/holdtheline.html", "api": "hold", "min": 2, "max": 2, "art": "einladung",
      "text": "Hold the Line: die eigene Linie gegen den Gegner halten." },
    { "id": "mastermind", "name": "Mastermind", "kat": "brett", "pfad": "/spiele/brettspiele/mastermind/mastermind.html", "api": "mastermind", "min": 2, "max": 2, "art": "einladung",
      "text": "Mastermind: denselben Farbcode knacken, wer zuerst vier schwarze Treffer hat gewinnt." },
    { "id": "memory", "name": "Memory", "kat": "brett", "pfad": "/spiele/brettspiele/memory/memory.html", "api": "memory", "min": 2, "max": 2, "art": "einladung",
      "text": "Memory: Kartenpaare finden, abwechselnd, wer ein Paar hat bleibt dran." },
    { "id": "halma", "name": "Halma", "kat": "brett", "pfad": "/spiele/brettspiele/halma/halma.html", "api": "halma", "min": 2, "max": 6, "art": "tisch",
      "text": "Halma: die eigenen Figuren ins gegenüberliegende Haus bringen." },
    { "id": "risiko", "name": "Risiko", "kat": "brett", "pfad": "/spiele/brettspiele/risiko/risiko.html", "api": "risiko", "min": 2, "max": 6, "art": "tisch",
      "text": "Risiko: Welteroberung mit Armeen, Würfeln und geheimen Missionen für zwei bis sechs Spieler." },
    { "id": "labyrinth", "name": "Das verrückte Labyrinth", "kat": "brett", "pfad": "/spiele/brettspiele/verruecktes-labyrinth/verruecktes-labyrinth.html", "api": "labyrinth", "min": 2, "max": 4, "art": "tisch",
      "text": "Das verrückte Labyrinth: Gänge verschieben und Gegenstände einsammeln." },
    { "id": "schach", "name": "Schach", "kat": "brett", "pfad": "/spiele/brettspiele/schach/schach.html", "api": "schach", "min": 2, "max": 2, "art": "einladung",
      "text": "Schach mit allen Regeln – matt oder Aufgabe." },
    { "id": "go", "name": "Go", "kat": "brett", "pfad": "/spiele/brettspiele/go/go.html", "api": "go", "min": 2, "max": 2, "art": "einladung",
      "text": "Go auf dem 9×9-Brett: Gebiet gewinnen und Steine schlagen." },
    { "id": "backgammon", "name": "Backgammon", "kat": "brett", "pfad": "/spiele/brettspiele/backgammon/backgammon.html", "api": "backgammon", "min": 2, "max": 2, "art": "einladung",
      "text": "Backgammon: 24 Punkte, Würfel und der eigene Heimbereich." },
    { "id": "schiffe", "name": "Schiffe versenken", "kat": "brett", "pfad": "/spiele/brettspiele/schiffe/schiffe.html", "api": "schiffe", "min": 2, "max": 2, "art": "tisch",
      "text": "Schiffe versenken: Flotten setzen und die des Gegners treffen." },
    { "id": "bingo", "name": "Bingo", "kat": "brett", "pfad": "/spiele/brettspiele/bingo/bingo.html", "api": "bingo", "min": 2, "max": 6, "art": "tisch",
      "text": "Bingo: Zahlen ziehen und Reihen auf der eigenen Karte abhaken." },
    { "id": "domino", "name": "Domino", "kat": "brett", "pfad": "/spiele/brettspiele/domino/domino.html", "api": "domino", "min": 2, "max": 2, "art": "tisch",
      "text": "Domino: passende Steine anlegen, wer keinen hat, zieht nach." },
    { "id": "schlangen", "name": "Schlangen und Leitern", "kat": "brett", "pfad": "/spiele/brettspiele/schlangen/schlangen.html", "api": "schlangen", "min": 2, "max": 4, "art": "tisch",
      "text": "Schlangen und Leitern: würfeln, klettern und wieder zurückfallen." },
    { "id": "gaensespiel", "name": "Gänsespiel", "kat": "brett", "pfad": "/spiele/brettspiele/gaensespiel/gaensespiel.html", "api": "gaensespiel", "min": 2, "max": 4, "art": "tisch",
      "text": "Gänsespiel: das Rennen ums Brett mit seinen Glücksfeldern." },

    { "id": "pong", "name": "Pong", "kat": "weitere", "pfad": "/spiele/pong/pong.html", "api": "pong", "min": 2, "max": 2, "art": "tisch",
      "text": "Pong: Schläger unten und oben, der Ball wird mit jedem Schlag schneller." },
    { "id": "tron", "name": "Tron", "kat": "weitere", "pfad": "/spiele/tron/tron.html", "api": "tron", "min": 2, "max": 2, "art": "tisch",
      "text": "Tron: Lichtspuren ziehen – wer anstößt, verliert die Runde." },
    { "id": "bomberman", "name": "Bomberman", "kat": "weitere", "pfad": "/spiele/bomberman/bomberman.html", "api": "bomberman", "min": 2, "max": 2, "art": "tisch",
      "text": "Bomberman: Bomben legen und den Gegner in die Flammen locken." },
    { "id": "bubble", "name": "Bubble Shooter", "kat": "weitere", "pfad": "/spiele/bubbleshooter/bubbleshooter.html", "api": "bubble", "min": 2, "max": 2, "art": "tisch",
      "text": "Bubble Shooter: Blasen schießen und gleiche Farben sprengen." },
    { "id": "snake", "name": "Snake", "kat": "weitere", "pfad": "/spiele/snake/snake.html", "api": "snake", "min": 2, "max": 2, "art": "tisch",
      "text": "Snake: fressen, wachsen und nicht in die eigene Spur fahren." },
    { "id": "tetris", "name": "Tetris", "kat": "weitere", "pfad": "/spiele/tetris/tetris.html", "api": "tetris", "min": 2, "max": 2, "art": "einladung",
      "text": "Tetris: Reihen füllen – beide sehen das Feld des anderen." },
    { "id": "einsxeins", "name": "1×1-Rennen", "kat": "weitere", "pfad": "/spiele/1x1/1x1.html", "api": "einsxeins", "min": 2, "max": 7, "art": "tisch",
      "text": "1×1-Rennen: Kopfrechnen um die Wette – richtig rechnen bringt Punkte." },
    { "id": "galgen", "name": "Galgenmännchen", "kat": "weitere", "pfad": "/spiele/brettspiele/galgen/galgen.html", "api": "galgen", "min": 2, "max": 6, "art": "tisch",
      "text": "Galgenmännchen: reihum raten – wer das letzte Feld füllt, gewinnt." },
    { "id": "quizshow", "name": "Scherzshow", "kat": "weitere", "pfad": "/spiele/quizshow/quizshow.html", "api": "quiz", "min": 2, "max": 6, "art": "tisch",
      "text": "Scherzfragen im Studio: vier Antworten, schnelle Finger oder ruhige Punkte – Wissen ist ausdrücklich nicht gefragt." },
    { "id": "wortsuche", "name": "Wortsuche", "kat": "weitere", "pfad": "/spiele/brettspiele/wortsuche/wortsuche.html", "api": "wortsuche", "min": 2, "max": 6, "art": "tisch",
      "text": "Wortsuche: beide suchen im selben Gitter – wer mehr Wörter findet." },
    { "id": "malen", "name": "Malen", "kat": "weitere", "pfad": "/spiele/malen/malen.html", "api": "malen", "min": 2, "max": 8, "art": "gemeinsam",
      "text": "Malen: eine gemeinsame Fläche, alle malen gleichzeitig darauf." }
  ]
};
