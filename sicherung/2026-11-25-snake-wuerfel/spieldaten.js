/* Spieldaten – die eine Quelle für Spielhalle (spiele.html) und Connector.

   Kategorien, Spiele, Kurztexte und Highlights stehen nur noch hier. Der
   Connector laedt diese Datei; die Spielhalle (spiele.html) ist eine fertige
   Seite, deren Karten aus dieser Liste entstanden sind. Nach einer Aenderung
   hier muss spiele.html nachgezogen werden (der Android-Bau erzeugt seine
   Fassung dagegen selbst aus dieser Datei).

   Merkmale (an den Spielseiten geprüft, Stand 24.11.2026):
     solo     allein spielbar (Solo-Modus oder Gegner-KI)
     lokal    Mehrspieler am selben Gerät
     zwei     lokal mit zwei Personen spielbar (Arschloch: lokal erst ab drei)
     computer, highscore, controller   wie bisher

   „online“ steht bewusst nicht in den Merkmalen: online ist nur, was im
   multiplayer-register.js steht (echte Connector-Anbindung). Connector und
   Spielhalle leiten das daraus ab; in spiele.html stehen die Merkmale fest
   in den Karten und muessen dort nachgezogen werden.

   Nicht in dieser Liste: malen/malen.html. Das ist ein gemeinsames
   Kritzelbrett ohne Ziel, Punkte und Sieger; es steht als Werkzeug-Kachel in
   spiele.html und nur für den gemeinsamen Raum im Register. */
window.TGSpieldaten = (function () {
 var KATEGORIEN = [
    { id: "wuerfel", name: "Würfelspiele", akzent: "#8fd0a8",
      text: "Kniffel, 10.000, Mäxchen – hier entscheiden die Würfel." },
    { id: "karten", name: "Kartenspiele", akzent: "#e8cd8e",
      text: "UNO, Mau Mau, Poker und Solitär – alles mit einem Blatt Karten." },
    { id: "brett", name: "Brettspiele", akzent: "#7fb7e8",
      text: "Schach, Mühle, Halma, Risiko – Brett- und Gesellschaftsspiele." },
    { id: "denk", name: "Denkspiele", akzent: "#b9a0e8",
      text: "Sudoku, Memory, Mastermind – knobeln, merken und kombinieren." },
    { id: "arcade", name: "Arcade und Geschicklichkeit", akzent: "#f0a35e",
      text: "Pac-Man, Tetris, Snake, Flipper – Klassiker und ruhige Finger." },
    { id: "quiz", name: "Quiz und Partyspiele", akzent: "#ef9e9e",
      text: "Rateshows und Partyspiele – schnell antworten, gemeinsam lachen." }
  ];
 var SPIELE = [
    /* --- Kartenspiele -------------------------------------------------- */
    ["Arschloch", "kartenspiele/arschloch/arschloch.html", "karten", "lokal"],
    ["Baccara", "kartenspiele/baccara/baccara.html", "karten", "", "baccarat"],
    ["Cabo", "kartenspiele/cabo/cabo.html", "karten", ""],
    ["Mau Mau", "kartenspiele/maumau/maumau.html", "karten", "lokal zwei", "maumau"],
    ["Poker", "kartenspiele/poker/poker.html", "karten", "", "texas holdem"],
    ["Sechsundsechzig", "kartenspiele/sechsundsechzig/sechsundsechzig.html", "karten", "", "66"],
    ["17 und 4", "kartenspiele/siebzehn/siebzehn.html", "karten", "", "blackjack siebzehn"],
    ["SKYJO", "kartenspiele/skyjo/skyjo.html", "karten", "lokal zwei"],
    ["Solitär", "kartenspiele/solitaer/solitaer.html", "karten", "solo", "klondike spider freecell pyramid tripeaks patience solitaire"],
    ["31 Schwimmen", "kartenspiele/schwimmen/schwimmen.html", "karten", "lokal zwei", "einunddreissig knack"],
    ["UNO", "kartenspiele/uno/uno.html", "karten", "lokal zwei"],

    /* --- Würfelspiele -------------------------------------------------- */
    ["10.000", "brettspiele/zehntausend/zehntausend.html", "wuerfel", "lokal zwei", "10000 zehntausend"],
    ["Auf nach Boston", "brettspiele/boston/boston.html", "wuerfel", "lokal zwei", "boston"],
    ["Chicago", "brettspiele/chicago/chicago.html", "wuerfel", "lokal zwei"],
    ["Kniffel", "brettspiele/kniffel/kniffel.html", "wuerfel", "solo lokal zwei", "yahtzee"],
    ["Lügen Mäxchen", "brettspiele/maexchen/maexchen.html", "wuerfel", "lokal zwei", "maexchen mäxchen"],
    ["Pig", "brettspiele/pig/pig.html", "wuerfel", "lokal zwei", "schweinchen"],
    ["Ship Captain Crew", "brettspiele/shipcrew/shipcrew.html", "wuerfel", "lokal zwei", "schiffe versenken"],

    /* --- Würfelspiel-Sammlung (lokal am selben Gerät) ------------------- */
    ["Beelzebub", "wuerfelspiele/beelzebub.html", "wuerfel", "solo lokal zwei"],
    ["Der Turm", "wuerfelspiele/der-turm.html", "wuerfel", "lokal zwei"],
    ["Die böse Drei", "wuerfelspiele/boese-drei.html", "wuerfel", "solo lokal zwei"],
    ["Die böse Eins", "wuerfelspiele/boese-eins.html", "wuerfel", "solo lokal zwei"],
    ["Dufte Hundert", "wuerfelspiele/dufte-hundert.html", "wuerfel", "solo lokal zwei"],
    ["Eins bis Sechzehn", "wuerfelspiele/eins-bis-sechzehn.html", "wuerfel", "solo lokal zwei"],
    ["Fenster würfeln", "wuerfelspiele/fenster.html", "wuerfel", "solo lokal zwei"],
    ["Gestrichene Zwölf", "wuerfelspiele/gestrichene-zwoelf.html", "wuerfel", "solo lokal zwei"],
    ["Hin und her", "wuerfelspiele/hin-und-her.html", "wuerfel", "solo lokal zwei"],
    ["Hohe Hausnummer", "wuerfelspiele/hohe-hausnummer.html", "wuerfel", "solo lokal zwei"],
    ["Jule", "wuerfelspiele/jule.html", "wuerfel", "solo lokal zwei", "stumme jule"],
    ["Pärchen", "wuerfelspiele/paerchen.html", "wuerfel", "solo lokal zwei"],
    ["Plus-Minus", "wuerfelspiele/plus-minus.html", "wuerfel", "solo lokal zwei"],
    ["Rentmeister", "wuerfelspiele/rentmeister.html", "wuerfel", "solo lokal zwei"],
    ["Streichhölzer auswürfeln", "wuerfelspiele/streichhoelzer.html", "wuerfel", "lokal zwei", "streichhoelzer"],
    ["Teure Sieben", "wuerfelspiele/teure-sieben.html", "wuerfel", "lokal zwei"],
    ["Unter oder über sieben", "wuerfelspiele/unter-oder-ueber.html", "wuerfel", "solo lokal zwei"],
    ["Würfelpoker", "wuerfelspiele/wuerfelpoker.html", "wuerfel", "solo lokal zwei", "poker würfel"],

    /* --- Brettspiele --------------------------------------------------- */
    ["Backgammon", "brettspiele/backgammon/backgammon.html", "brett", "lokal zwei"],
    ["Bingo", "brettspiele/bingo/bingo.html", "brett", "solo"],
    ["Dame", "brettspiele/dame/dame.html", "brett", "lokal zwei", "checkers"],
    ["Das verrückte Labyrinth", "brettspiele/verruecktes-labyrinth/verruecktes-labyrinth.html", "brett", "lokal zwei computer", "gaenge schieben schaetze labyrinth"],
    ["Domino", "brettspiele/domino/domino.html", "brett", "lokal zwei computer"],
    ["Gänsespiel", "brettspiele/gaensespiel/gaensespiel.html", "brett", "lokal zwei", "ganzespiel"],
    ["Go", "brettspiele/go/go.html", "brett", "lokal zwei", "goban baduk"],
    ["Halma", "brettspiele/halma/halma.html", "brett", "solo lokal zwei computer", "sternhalma sprunghalma"],
    ["Hold the Line", "brettspiele/hold/holdtheline.html", "brett", "lokal zwei"],
    ["Käsekästchen", "brettspiele/kaesekaestche/kaesekaestche.html", "brett", "lokal zwei", "punkte verbinden dots boxes"],
    ["Kalaha", "brettspiele/kalaha/kalaha.html", "brett", "lokal zwei", "mancala"],
    ["Mahjong", "brettspiele/mahjong/mahjong.html", "brett", "solo", "mahjongg steine"],
    ["Mensch ärgere dich nicht", "brettspiele/menschaerger/menschaerger.html", "brett", "lokal zwei", "ludo"],
    ["Mühle", "brettspiele/muehle/muehle.html", "brett", "lokal zwei", "muehle"],
    ["Reversi", "brettspiele/reversi/reversi.html", "brett", "lokal zwei", "othello"],
    ["Risiko – Welt erobern", "brettspiele/risiko/risiko.html", "brett", "lokal zwei computer", "erobern weltkarte strategie armeen wuerfel"],
    ["Schach", "brettspiele/schach/schach.html", "brett", "lokal zwei", "chess"],
    ["Schiffe versenken", "brettspiele/schiffe/schiffe.html", "brett", "", "battleship flotte"],
    ["Schlangen und Leitern", "brettspiele/schlangen/schlangen.html", "brett", "lokal zwei", "leiterlispiel"],
    ["Sim", "brettspiele/sim/sim.html", "brett", "lokal zwei", "verbindungslinien"],
    ["Tic Tac Toe", "brettspiele/tictac/tictactoe.html", "brett", "lokal zwei", "tic-tac-toe drei gewinnt"],
    ["Vier gewinnt", "brettspiele/viergewinnt/viergewinnt.html", "brett", "lokal zwei", "connect four"],

    /* --- Denkspiele ---------------------------------------------------- */
    ["1×1-Training", "1x1/1x1.html", "denk", "solo", "einmaleins rechnen mathe"],
    ["Bejeweled", "bejeweled/bejeweled.html", "denk", "solo highscore", "juwelen match 3"],
    ["Galgenmännchen", "brettspiele/galgen/galgen.html", "denk", "solo", "hangman galgen"],
    ["Lights Out", "lightsout/lightsout.html", "denk", "solo highscore", "lichter"],
    ["Mastermind", "brettspiele/mastermind/mastermind.html", "denk", "solo highscore", "farbcode"],
    ["Memory", "brettspiele/memory/memory.html", "denk", "solo highscore", "kartenspiel gedaechtnis"],
    ["Minensucher", "minesweeper/minesweeper.html", "denk", "solo highscore", "minesweeper"],
    ["Paare finden", "brettspiele/paare/paare.html", "denk", "solo", "memory paare"],
    ["15-Puzzle", "puzzle15/puzzle15.html", "denk", "solo highscore", "schiebepuzzle puzzel"],
    ["SameGame", "samegame/samegame.html", "denk", "solo highscore", "steine wegklicken"],
    ["Sokoban", "sokoban/sokoban.html", "denk", "solo highscore", "kisten schieben"],
    ["Steine-Solitär", "brettspiele/solitaer/solitaer.html", "denk", "solo highscore", "solitär solitaire peg steine"],
    ["21 Streichhölzer", "brettspiele/streichholzspiel/streichholzspiel.html", "denk", "lokal zwei", "streichhoelzer nim"],
    ["Sudoku", "sudoku/sudoku.html", "denk", "solo highscore"],
    ["Wortsuche", "brettspiele/wortsuche/wortsuche.html", "denk", "solo", "wortsalat buchstaben"],

    /* --- Quiz und Partyspiele ------------------------------------------ */
    ["Scherzshow", "quizshow/quizshow.html", "quiz", "", "quiz scherzfragen witze rate show"],

    /* --- Arcade und Geschicklichkeit ----------------------------------- */
    ["Acia", "acia/acia.html", "arcade", "solo lokal zwei", "farben bloecke leveleditor"],
    ["Arkanoid", "arkanoid/arkanoid.html", "arcade", "solo highscore", "breakout schlaeger"],
    ["Bubble Shooter", "bubbleshooter/bubbleshooter.html", "arcade", "solo lokal zwei highscore", "blasen"],
    ["Flappy Bird", "flappybird/flappybird.html", "arcade", "solo highscore", "flappy"],
    ["Frogger", "frogger/frogger.html", "arcade", "solo highscore", "frosch"],
    ["Gold Miner", "goldminer/goldminer.html", "arcade", "solo", "goldgraeber haken"],
    ["Pac-Man", "pacman/pacman.html", "arcade", "solo highscore", "pacman geister"],
    ["Pong", "pong/pong.html", "arcade", "lokal zwei computer highscore"],
    ["Snake", "snake/snake.html", "arcade", "solo highscore", "nokia schlange"],
    ["Space Invaders", "invaders/invaders.html", "arcade", "solo highscore", "invaders alien"],
    ["Super Pimmelchen", "pimmel/spimmel.html", "arcade", "solo highscore", "spimmel retro"],
    ["Tetris", "tetris/tetris.html", "arcade", "solo", "kloetze steine"],
    ["Tron", "tron/tron.html", "arcade", "lokal zwei", "lichtrennen lightcycle"],
    ["Voigt Zone", "zone/voigtzone.html", "arcade", "solo", "qix flaechen"],
    ["Bomberman", "bomberman/bomberman.html", "arcade", "lokal zwei", "bombe"],
    ["Flipper", "flipper/flipper.html", "arcade", "solo highscore", "pinball"],
    ["Labyrinth Plus!", "brettspiele/labyrinth/labyrinth.html", "arcade", "solo highscore", "kugel kippen"],
    ["Tower Bloxx", "towerbloxx/towerbloxx.html", "arcade", "solo highscore", "turm stapeln"],
    ["Doodle Jump", "doodlejump/doodlejump.html", "arcade", "solo highscore", "springen plattformen"],
    ["Durst", "durst/durst.html", "arcade", "solo", "bier level abenteuer"],
    ["Rotlicht", "rotlicht/rotlicht.html", "arcade", "solo", "idle clicker wirtschaft imperium"]
  ];
 var TEXTE = {
    "kartenspiele/arschloch/arschloch.html": "Karten abwerfen und am Ende die Plätze tauschen.",
    "kartenspiele/baccara/baccara.html": "Auf Spieler oder Bank setzen – näher an neun gewinnt.",
    "kartenspiele/cabo/cabo.html": "Vier verdeckte Karten tauschen und möglichst niedrig bleiben.",
    "kartenspiele/maumau/maumau.html": "Farbe oder Zahl bedienen, sieben ziehen, Mau sagen.",
    "kartenspiele/poker/poker.html": "Texas Hold'em mit Blinds, Einsatzrunden und Showdown.",
    "kartenspiele/sechsundsechzig/sechsundsechzig.html": "24 Karten, Stiche sammeln und 66 Punkte erreichen.",
    "kartenspiele/siebzehn/siebzehn.html": "Gegen die Bank auf 21 kommen, ohne zu überkaufen.",
    "kartenspiele/skyjo/skyjo.html": "Karten aufdecken und tauschen – die niedrigste Summe gewinnt.",
    "kartenspiele/solitaer/solitaer.html": "Klondike, Spider und mehr – Kartenlegen in Ruhe.",
    "kartenspiele/schwimmen/schwimmen.html": "Drei Karten tauschen und klopfen – wer am nächsten an 31 ist.",
    "kartenspiele/uno/uno.html": "Farben wechseln, aussetzen und mit +2 und +4 ärgern.",

    "brettspiele/zehntausend/zehntausend.html": "Sechs Würfel, Punkte sichern – wer zuerst 10.000 hat.",
    "brettspiele/boston/boston.html": "Würfelrunde mit Ansagen, auch die Würfe der anderen zählen.",
    "brettspiele/chicago/chicago.html": "Filze, Würfe und der beste Wurf der Runde.",
    "brettspiele/kniffel/kniffel.html": "Fünf Würfel, drei Würfe, 13 Kategorien.",
    "brettspiele/maexchen/maexchen.html": "Zwei verdeckte Würfel, Ansage und Zweifel.",
    "brettspiele/pig/pig.html": "Würfeln und sammeln – aber vor der Eins aufhören.",
    "brettspiele/shipcrew/shipcrew.html": "Schiff, Kapitän, Crew – dann 500 Punkte sammeln.",
    "wuerfelspiele/beelzebub.html": "Drei Würfel zählen plus, ein roter zählt minus.",
    "wuerfelspiele/der-turm.html": "Eigene Stichzahl treffen und Steine aus dem Turm holen.",
    "wuerfelspiele/boese-drei.html": "So oft würfeln wie du willst – eine Drei löscht den Zug.",
    "wuerfelspiele/boese-eins.html": "Fünf Würfe werden addiert – eine Eins stoppt sofort.",
    "wuerfelspiele/dufte-hundert.html": "Pasch wird multipliziert, sonst addiert – Ziel 100.",
    "wuerfelspiele/eins-bis-sechzehn.html": "Die Zahlen 1 bis 16 aus drei Würfeln bilden.",
    "wuerfelspiele/fenster.html": "Nur Zweien zählen – eine 1, zwei 4, drei 9.",
    "wuerfelspiele/gestrichene-zwoelf.html": "Summen von 2 bis 12 würfeln und streichen.",
    "wuerfelspiele/hin-und-her.html": "Jede Zahl umschalten – alle sechs gleichzeitig streichen.",
    "wuerfelspiele/hohe-hausnummer.html": "Drei Würfe auf Hunderter, Zehner und Einer verteilen.",
    "wuerfelspiele/jule.html": "Die Treppe 1 bis 6 hoch und zurück würfeln.",
    "wuerfelspiele/paerchen.html": "Vier Würfel, zwei zählen plus, zwei minus.",
    "wuerfelspiele/plus-minus.html": "Neun Würfe im Muster plus, plus, minus.",
    "wuerfelspiele/rentmeister.html": "Sechs feste Zahlen, jede mal einem gewählten Würfel.",
    "wuerfelspiele/streichhoelzer.html": "Einsen nach links, Sechsen in die Mitte – Hölzer räumen.",
    "wuerfelspiele/teure-sieben.html": "2–6 und 8–12 streichen, die Sieben zählt doppelt.",
    "wuerfelspiele/unter-oder-ueber.html": "Vor dem Wurf ansagen – richtig bringt Punkte.",
    "wuerfelspiele/wuerfelpoker.html": "Fünf Würfel, zwei Nachwürfe, volle Pokerwertung.",

    "brettspiele/backgammon/backgammon.html": "24 Punkte, Würfel und der eigene Heimbereich.",
    "brettspiele/bingo/bingo.html": "Zahlen ziehen und Reihen auf der eigenen Karte abhaken.",
    "brettspiele/dame/dame.html": "Mit Schlagzwang – wer keine Figur mehr hat, verliert.",
    "brettspiele/verruecktes-labyrinth/verruecktes-labyrinth.html": "Gänge verschieben und Gegenstände einsammeln.",
    "brettspiele/domino/domino.html": "Passende Steine anlegen, sonst nachziehen.",
    "brettspiele/gaensespiel/gaensespiel.html": "Das Rennen ums Brett mit seinen Glücksfeldern.",
    "brettspiele/go/go.html": "Auf 9×9 Gebiet gewinnen und Steine schlagen.",
    "brettspiele/halma/halma.html": "Figuren ins gegenüberliegende Haus bringen.",
    "brettspiele/hold/holdtheline.html": "Die eigene Linie gegen den Gegner halten.",
    "brettspiele/kaesekaestche/kaesekaestche.html": "Linien ziehen und Kästchen schließen.",
    "brettspiele/kalaha/kalaha.html": "Steine verteilen und im eigenen Kasten sammeln.",
    "brettspiele/mahjong/mahjong.html": "Steinpaare finden und das Brett abräumen.",
    "brettspiele/menschaerger/menschaerger.html": "Würfeln, rauswerfen und als Erster im Ziel.",
    "brettspiele/muehle/muehle.html": "Setzen, ziehen, Mühlen schließen und Steine abräumen.",
    "brettspiele/reversi/reversi.html": "Steine einschließen und umdrehen – Mehrheit gewinnt.",
    "brettspiele/risiko/risiko.html": "Armeen, Würfel und Missionen – die Welt erobern.",
    "brettspiele/schach/schach.html": "Alle Regeln, matt oder Aufgabe.",
    "brettspiele/schiffe/schiffe.html": "Flotten setzen und die des Gegners treffen.",
    "brettspiele/schlangen/schlangen.html": "Würfeln, klettern und wieder zurückfallen.",
    "brettspiele/sim/sim.html": "Linien setzen – wer das letzte Dreieck schließt, verliert.",
    "brettspiele/tictac/tictactoe.html": "Drei gewinnt auf dem Neon-Brett.",
    "brettspiele/viergewinnt/viergewinnt.html": "Steine fallen lassen und vier in eine Reihe bringen.",

    "1x1/1x1.html": "Einmaleins üben – gegen die Zeit oder andere.",
    "bejeweled/bejeweled.html": "Juwelen tauschen und Reihen sprengen.",
    "brettspiele/galgen/galgen.html": "Buchstaben raten, bevor das Männchen fertig ist.",
    "lightsout/lightsout.html": "Lichter ausschalten – jeder Klick kippt die Nachbarn.",
    "brettspiele/mastermind/mastermind.html": "Den Farbcode in möglichst wenigen Versuchen knacken.",
    "brettspiele/memory/memory.html": "Kartenpaare finden und im Kopf behalten.",
    "minesweeper/minesweeper.html": "Minen markieren, ohne auf eine zu treten.",
    "brettspiele/paare/paare.html": "Dasselbe Brett lösen – wer ist schneller?",
    "puzzle15/puzzle15.html": "Steine schieben, bis die Zahlen stimmen.",
    "samegame/samegame.html": "Gleichfarbige Blöcke wegklicken, bis nichts mehr bleibt.",
    "sokoban/sokoban.html": "Kisten in die richtigen Ecken schieben.",
    "brettspiele/solitaer/solitaer.html": "Steine überspringen, bis nur einer übrig bleibt.",
    "brettspiele/streichholzspiel/streichholzspiel.html": "Reihum ein bis drei nehmen – wer das letzte nimmt, verliert.",
    "sudoku/sudoku.html": "Zahlen 1 bis 9 in jede Reihe, Spalte und Box.",
    "brettspiele/wortsuche/wortsuche.html": "Wörter im Buchstabengitter finden.",

    "quizshow/quizshow.html": "Vier Antworten, schnelle Finger – Wissen ist nicht gefragt.",

    "acia/acia.html": "Farbblöcke verbinden und eigene Level bauen.",
    "arkanoid/arkanoid.html": "Mit dem Schläger den Ball halten und Steine abbrechen.",
    "bubbleshooter/bubbleshooter.html": "Blasen schießen und gleiche Farben sprengen.",
    "flappybird/flappybird.html": "Durch die Lücken flattern, so weit es geht.",
    "frogger/frogger.html": "Über Straße und Fluss nach Hause.",
    "goldminer/goldminer.html": "Mit dem Haken Gold und Schätze hochziehen.",
    "pacman/pacman.html": "Punkte fressen und den Geistern entkommen.",
    "pong/pong.html": "Schläger unten und oben – der Ball wird schneller.",
    "snake/snake.html": "Fressen, wachsen und nicht in die eigene Spur fahren.",
    "invaders/invaders.html": "Reihen von Aliens abschießen und Deckung nutzen.",
    "pimmel/spimmel.html": "Retro-Hüpfer mit eigenen Leveln.",
    "tetris/tetris.html": "Reihen füllen, bevor der Stapel oben ankommt.",
    "tron/tron.html": "Lichtspuren ziehen – wer anstößt, verliert.",
    "zone/voigtzone.html": "Flächen abgrenzen und Gegner aussperren.",
    "bomberman/bomberman.html": "Bomben legen und den Gegner in die Flammen locken.",
    "flipper/flipper.html": "Die Kugel mit den Flippern oben halten.",
    "brettspiele/labyrinth/labyrinth.html": "Kugel durchs Labyrinth kippen und Ziele treffen.",
    "towerbloxx/towerbloxx.html": "Stockwerke stapeln, bis der Turm wächst.",
    "doodlejump/doodlejump.html": "Plattform für Plattform nach oben springen.",
    "durst/durst.html": "Level-Abenteuer – Flaschen sammeln und weiterkommen.",
    "rotlicht/rotlicht.html": "Vom Straßenstrich zum Luxusimperium – Idle-Clicker mit Personal, Ausbau und Offline-Ertrag."
  };
 var HIGHLIGHTS = [
    { pfad: "brettspiele/mahjong/mahjong.html", bild: "bilder/highlights/mahjong.jpg",
      text: "Steine paaren, Schichten freilegen – ruhiges Denkspiel mit 144 Steinen." },
    { pfad: "brettspiele/zehntausend/zehntausend.html", bild: "bilder/highlights/zehntausend.jpg",
      text: "Sechs Würfel, sichere Punkte, ein Tisch für zwei bis sechs – auch online." },
    { pfad: "brettspiele/risiko/risiko.html", bild: "bilder/highlights/risiko.jpg",
      text: "Weltkarte, Armeen, Würfel – Eroberung für zwei bis sechs Spieler." },
    { pfad: "brettspiele/verruecktes-labyrinth/verruecktes-labyrinth.html", bild: "bilder/highlights/labyrinth.jpg",
      text: "Gänge schieben, Schätze einsammeln – das Brett verändert sich mit jedem Zug." },
    { pfad: "flipper/flipper.html", bild: "bilder/highlights/flipper.jpg",
      text: "Silberkugel, Flipperfinger, Mittelstift – eine Runde am Tisch." },
    { pfad: "quizshow/quizshow.html", bild: "bilder/highlights/quizshow.jpg",
      text: "Scherzshow mit Moderator: Fragen, Antworten, Punkte – auch mit Freunden." }
  ];
 return { kategorien: KATEGORIEN, spiele: SPIELE, texte: TEXTE, highlights: HIGHLIGHTS };
})();
