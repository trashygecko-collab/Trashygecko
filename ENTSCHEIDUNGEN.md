# Entscheidungen – TrashyGecko Android

Jede Entscheidung gilt, bis sie hier ausdrücklich geändert wird. Wer eine verworfene
Lösung erneut vorschlägt, muss zuerst die Begründung unten widerlegen.

Format: **D-Nr · Entscheidung · Begründung · Folgen**.

---

## D-01 · Eine App: `spiele/android`, Paket `de.trashygecko.tmahjong`
Die Spiele-App bleibt das einzige Produkt. Der Paketname ist historisch (früher nur
Mahjong), wird aber **nicht** geändert: Android erkennt Updates an Paket-ID **und**
Signatur. Ein neuer Name würde jede bestehende Installation zu einer Neuinstallation
machen und alle lokalen Spielstände kosten.
*Folge:* Anzeigename „TrashyGecko“, interne Klasse `MainActivity` darf umbenannt werden,
die Application-ID nicht.

## D-02 · Spiele bleiben Web-Technik, die App wird nativ
91 Spiele als HTML/Canvas/JS neu zu schreiben wäre Monate Arbeit ohne Spielgewinn.
Nativ wird alles, was der Nutzer als **App** erlebt: Start, Spieleliste, Navigation,
Status, Einstellungen. Die Spiele laufen weiter in einem WebView mit Asset-Loader.
*Folge:* Es gibt genau **einen** WebView-Wirt für Spiele (`GameActivity`, aus
`MainActivity` herausgelöst), nicht mehrere.

## D-03 · Kein Go-Connector im APK
Der Connector ist ein Go-Programm für Windows/Linux. Ihn in die App zu holen hieße
zweite Engine, zweite Datenhaltung, doppelte Regeln. Lokales Mehrspieler läuft über den
vorhandenen NSD+TCP-Relay; Spiele werden einzeln angebunden, wenn ihre Regeln klein
genug sind (Vorbild: `GalgenLokal`).
*Folge:* Auf einem Handy ohne Internet sind Chat und Galgen mehrspielerfähig, andere
Spiele nicht. Das ist eine bewusste Grenze, keine offene Baustelle.

## D-04 · Die beiden Go-Bäume werden nicht zusammengelegt
`spiele/connector` und `spiele/connector-online` bleiben getrennt (Windows-Betrieb vs.
Render-Dienst) und werden über `sync.py` synchron gehalten. Eine Zusammenführung wurde
nie geprüft und würde den Betrieb auf beiden Seiten riskieren.

## D-05 · Der Spielekatalog hat genau eine Quelle
`spiele/spiele.html` (JS-Liste `SPIELE`) bleibt die Wahrheit. Beim Bauen erzeugt ein
Gradle-Task daraus `assets/katalog.json` für die native Ansicht. Kein zweiter
handgepflegter Katalog, kein HTML-Parsen zur Laufzeit.
*Folge:* Ein neues Spiel erscheint in der App automatisch, sobald es im Katalog steht
und die Datei existiert; der Build bricht bei einem toten Eintrag ab.

## D-06 · minSdk 21, targetSdk 34, Android 9 ist Pflicht
Android 9 (API 28) muss laufen – das tut es mit `minSdk 21`. Anheben auf 28 würde
Altgeräte ohne Nutzen aussperren. `targetSdk 34` bleibt, bis die lokale
Netzwerkberechtigung (ab API 37) umgesetzt ist.

## D-07 · Version wird fortgeschrieben, nie zurückgesetzt
Aktuell **1.15 / versionCode 16**. Die nächste Fassung ist 1.16 / 17. Ein Rücksetzen
bricht die Update-Kette im Play Store und bei `adb install -r`.

## D-08 · `apps/t-spiele` ist Altbestand
Die 144-KB-App (zwei Knöpfe in den Browser, `bauen.sh` ohne Gradle) wird nicht
weiterentwickelt und **nicht** auf Gradle umgestellt: sie ist nicht das Produkt, sondern
eine Verteilkopie für die Kachel „Meine Android App“. Sie bleibt unverändert liegen.
*Folge:* Die Regel „APK nur über Gradle“ gilt für **TrashyGecko.apk**, nicht für diese
Altlast.

## D-09 · Fernbedienung ausschließlich unter Extras
Bereits seit 1.14 so umgesetzt (Startbildschirm-Kachel und Katalogzeile entfernt,
`MainActivity` fängt `/WLAN-Fernbedienung/…` ab und leitet auf Extras). Bleibt so; die
Website behält ihre Katalogzeile.

## D-10 · Einstellungen werden nativ, Spielstände bleiben im WebView
Der Name und der Grafikstil gelten für alle Spiele und gehören deshalb in
`SharedPreferences`; die Spiele lesen sie über eine Brücke (`TGApp`) und schreiben sie
dorthin zurück. Spielstände (`localStorage`) bleiben, wo sie sind – sie sind
Spielzustand, nicht App-Einstellung.
*Folge:* Beim ersten Start werden vorhandene `localStorage`-Werte einmalig übernommen,
damit niemand seinen Stil verliert.

## D-11 · Keine neuen Bibliotheken
Vorhanden sind nur `androidx.core` und `androidx.webkit`. Die native Oberfläche wird mit
Board-Mitteln gebaut (Views, RecyclerView nur wenn unvermeidbar – dann als einzige
Ergänzung mit Begründung hier). Kein Compose, keine Navigation-Bibliothek, kein Kotlin.

## D-12 · Die echte Homepage kommt als „Extras → Homepage“ in die App
„Die richtige TrashyGecko-Startseite“ ist die Website (`index.html`). Sie als
App-Startbildschirm zu nehmen hieße wieder Webseite statt App. Sie wird deshalb
mitgeliefert und über Extras geöffnet; der Startbildschirm der App bleibt nativ.
*Offen:* Wenn du sie doch als Start willst, ist das eine Zeile – aber dann ist D-02
verletzt und muss neu entschieden werden.

*Stand 23.11.2026:* Der Start ist nativ umgesetzt. `MainActivity` baut die vier Kacheln
(Spiele, Lokales Netz, Online-Räume, Extras) als Views und zeigt den Netzstatus des
Geräts; `start/app.html` liegt nur noch als `index.html` im Paket. *Offen:* der Eintrag
„Extras → Homepage“ fehlt noch – dafür müsste die echte Website in die Assets.

## D-13 · Online-Ziel ist fest `https://trashygecko.onrender.com`
Kein konfigurierbarer Server in den Einstellungen: die Spiele-Seiten bringen ihre
`/api`-Pfade relativ mit, ein fremder Host würde Authentifizierung und Räume brechen.
*Folge:* Ein Umzug des Dienstes ist ein App-Update.

## D-14 · Tests nur auf vorhandener Hardware, nichts erfinden
Es stehen ein echtes Android-9-Gerät (Huawei FIG-LX1) und ein Emulator (Android 14) zur
Verfügung. Was dort nicht lief, steht in `PRUEFBERICHT-ANDROID.md` als „nicht getestet“.
Eine Debug-APK darf **nicht** über die release-signierte Installation gelegt werden
(Signaturkonflikt) – für Gerätetests entweder release-signiert bauen oder den Emulator
nehmen.

## D-15 · `katalog.json` entsteht beim Bauen aus `spiele/spiele.html`
Die App bekommt beim Bauen eine maschinenlesbare Spieleliste (`src/main/assets/katalog.json`,
Aufgabe `katalogBauen`): Name, Pfad, Kategorie, Kategoriename, Merkmale, Suchworte.
Quelle bleibt ausschließlich der Website-Katalog (D-05); die JSON-Datei wird nicht
eingecheckt und nicht von Hand gepflegt. Der Bau bricht ab, wenn der Katalog nicht
gelesen werden kann, ein Eintrag auf eine fehlende Datei zeigt oder 0 Spiele übrig sind.
Stand 23.11.2026: 94 Spiele, 9 Kategorien.

## D-16 · Spielseiten liegen in einer eigenen Activity
`GameActivity` zeigt eine einzelne Spielseite (lokal oder online), `MainActivity` behält
die Startseite. Damit hat ein Spiel seinen eigenen Verlauf und seine eigene Zurück-Taste,
und die Spieleliste behält Suche, Filter und Scrollposition.
*Folge:* Die WebView-Einstellungen liegen gemeinsam in `WebBasis`, damit beide Ansichten
nicht auseinanderlaufen. Die Mahjong-Brücke (`TMahjong`) wird nur noch dort injiziert.

## D-17 · Die Spieleliste ist nativ, `spiele.html` bleibt die Quelle
`SpieleActivity` zeigt den Katalog nativ (Suche, Kategorien, Merkmale) aus `katalog.json`;
`spiele.html` wird von der App nicht mehr geöffnet. Die HTML-Seite bleibt im APK
(Sicherheitsnetz) und bleibt die einzige Quelle der Daten (D-05, D-15).
*Folge:* Neue Spiele brauchen keine App-Änderung – eine Katalogzeile genügt, `katalogBauen`
nimmt sie mit. Änderungen an der HTML-Optik des Katalogs wirken nur noch auf der Website.

## D-18 · Eine Link-Regel für alle Activities
`WebNavigation` entscheidet, was mit einer Adresse passiert (WebView, Spieleliste,
Spielseite, Connector, Browser). Vorher stand die Regel doppelt in `MainActivity` und
`GameActivity` und lief bereits auseinander.
*Folge:* Neue Ziele werden nur dort ergänzt. Spielseiten leiten Spiel-Links nur um, wenn
der Aufrufer `spieleUmleiten = true` setzt (Start/Liste ja, Spielseite nein).

## D-19 · Merkmalsfilter verhalten sich wie im HTML-Katalog
Die native Liste bietet genau die Merkmale an, die im Katalog vorkommen, verknüpft mehrere
gewählte Merkmale mit **UND** und kombiniert sie mit der Textsuche – dieselbe Regel wie
`passt(s, q)` in `spiele.html`. „Zurücksetzen“ leert beides und bleibt sichtbar.
*Folge:* Neue Merkmale brauchen nur eine Zeile in `MERKMAL_FILTER` (`SpieleActivity`);
ob ein Chip erscheint, entscheidet der Katalog. Die Zeilenansichten werden einmal gebaut
und beim Filtern nur umgehängt – kein `RecyclerView`, keine neue Bibliothek (D-11).
Ändert sich die Filterregel im HTML, muss `SpieleActivity.passt` mitgezogen werden.

## D-20 · Die Spieleliste hat genau einen Einstieg: die Startseite
Die Kachel „Spiele“ in `start/app.html` öffnet `SpieleActivity`; das ist der einzige Weg.
Die Karte „Alle Spiele (offline)“ in `ConnectorActivity` ist am 23.11.2026 entfernt worden:
sie lag auf der Connector-Startseite, die man von der App aus nur über den Kopfknopf
„Start“ erreicht, und bot dieselbe Liste ein zweites Mal an. Verschieben auf den
Lokal-Bildschirm wurde verworfen – dort geht es um Räume und Chat, spielbar ist dort nur
der Galgen.
*Folge:* Der Connector kümmert sich um Räume (lokal/online) und Extras, nicht um die
Spieleliste. Ein Einstieg dorthin wird dort nicht wieder ergänzt – sonst gibt es zwei
Türen zu einer Liste, und die Texte laufen wie bei „91 Spiele“ auseinander.

## D-21 · Drehung baut keine Activity neu – Breitenabhängiges wird nachgezogen
Alle Activities tragen `configChanges` für `orientation|screenSize|…`, werden bei einer
Drehung also nicht neu erstellt. Suchtext, gewählte Chips und Scrollposition der
Spieleliste bleiben dadurch erhalten; es geht nichts über `onSaveInstanceState` verloren.
*Folge:* Alles, was von der Breite abhängt, muss `onConfigurationChanged` neu berechnen –
zurzeit nur der Umbruch der Filter-Chips. Die Breite dafür kommt aus
`getResources().getDisplayMetrics()`, nicht aus `View.getWidth()` (im Callback noch alt).
Prozess-Tod ist damit **nicht** abgedeckt – dafür sorgt D-22.

---

## D-22 · Prozess-Tod: Zustand der Spieleliste geht über `onSaveInstanceState`
Die Drehung baut die Activity nicht neu (D-21), ein beendeter Prozess schon. Weil die
Ansichten der Spieleliste keine IDs haben, sichert das System von sich aus nichts –
`SpieleActivity` legt Suchtext, gewählte Merkmale und Scrollposition deshalb selbst ins
Bundle und stellt sie in `onCreate` wieder her.
*Folge:* Die drei Werte sind dort die einzige Wahrheit über einen Neustart; wer eine neue
Filterart ergänzt, muss sie ebenfalls sichern. Unbekannte Merkmalsnamen aus einem alten
Bundle werden verworfen, damit kein Chip ohne Beschriftung entsteht; die Scrollposition
wird erst nach dem ersten Layout gesetzt (`addOnGlobalLayoutListener`), sonst klemmt sie
auf 0.
*Nachtrag 23.11.2026:* Erst wiederherstellen, dann **einmal** zeichnen; solange
`suchfeld.setText` den Textwächter auslöst, ist `stelleWiederLaeuft` gesetzt. Vorher lief
`zeichne()` vor `stelleWieder`, dadurch blieb ein gesichertes Merkmal ohne Suchtext
ungefiltert stehen.

## Verworfene Wege

| Vorschlag | Warum verworfen |
|---|---|
| Alles in Kotlin/Compose neu schreiben | Neue Abhängigkeiten (D-11), Risiko für alle Spiele, kein Mehrwert |
| Spiele nativ nachbauen | Monate Arbeit, 91 Spiele, laufende Pflege doppelt |
| Go-Connector in die App einbetten (gomobile) | Zweite Engine, Konflikte mit lokalen Räumen (D-03) |
| Nur die Website im WebView (Status quo von `T.Spiele`) | Genau das, was abgelöst werden soll |
| Spielekatalog als zweite JSON-Datei pflegen | Zwei Wahrheiten, sofort auseinanderlaufend (D-05) |
| Eigener Server pro Raum auf dem Handy (HTTP-Server im Gerät) | Android verbietet/Hintergrunddienste, Sicherheitsrisiko; NSD-Relay reicht |
