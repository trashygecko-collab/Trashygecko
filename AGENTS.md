# AGENTS.md – TrashyGecko Homepage & Android

**Lies diese Datei vor der ersten Änderung. Sie gilt für das ganze Projekt.**
Für den Android-Teil sind zusätzlich Pflicht:
[`ANDROID-ARCHITEKTUR.md`](ANDROID-ARCHITEKTUR.md) · [`ENTSCHEIDUNGEN.md`](ENTSCHEIDUNGEN.md) ·
[`ENTWICKLUNGSSTAND.md`](ENTWICKLUNGSSTAND.md) · [`PRUEFBERICHT-ANDROID.md`](PRUEFBERICHT-ANDROID.md)

---

## 1. Projekt auf einen Blick

| Pfad | Inhalt |
|---|---|
| `spiele/` | alle Spiele (HTML/JS/Canvas), `spiele/spiele.html` = Katalog (einzige Quelle der Spieleliste) |
| `spiele/connector/` | Go-Connector für Windows (Chat, Räume, Spiel-APIs), liefert die Homepage lokal aus |
| `spiele/connector-online/` | Go-Connector für den Render-Dienst, per `sync.py` mit `connector/` synchron |
| `spiele/halle.css`, `spiele/halle-online.js` | Optik und Online-Schicht der Spielhalle (D-27); Räume/Chat über den Connector, keine zweite Identität |
| `spiele/android/` | **die Android-App** (Gradle; Quellen, Assets entstehen beim Bauen) |
| `apps/t-spiele/` | Altlast: Zwei-Knopf-App in den Browser (kein Gradle, wird nicht weiterentwickelt) |
| `WLAN-Fernbedienung/` | Web-Modul der Fernbedienung (in der App nur unter Extras) |
| `spiele/quizshow/` | Scherzshow (Fragen, Moderatorleiste) |
| `index.html`, `main.css` | Startseite und gemeinsames Stylesheet der Website |

## 2. Zwei Ablagen, byte-gleich halten

Es gibt `Documents\Homepage\…` und `Documents\GitHub\Homepage\…`. **Jede** Änderung an
Seiten wird in beiden nachgezogen; danach Hashes vergleichen (`sha256sum`).
Der Go-Code hat eine Quelle (`spiele/connector`), `spiele/connector-online` bekommt die
Kopien über `python sync.py --nur-kopieren`; nie umgekehrt von Hand kopieren.
**Nichts committen und nichts veröffentlichen**, solange es nicht ausdrücklich verlangt ist.

## 3. Bauen

### Android-App (nur so entsteht die APK)
```bash
cd spiele/android
export JAVA_HOME=$(ls -d /c/Users/DarkP/Documents/arbeit/werkzeuge/t-spiele/jdk-17* | head -1)
export ANDROID_HOME=/c/Users/DarkP/Documents/arbeit/werkzeuge/t-spiele/sdk
/c/Users/DarkP/Documents/arbeit/werkzeuge/gradle-8.7/bin/gradle --no-daemon :app:packageTrashyGeckoApk
```
- Ergebnis: `app/build/outputs/apk/universal/TrashyGecko.apk`.
- `bash bauen.sh` baut nur die **Debug**-APK (`assembleDebug`).
- Release-Signatur braucht `spiele/android/keystore.properties`; der Keystore liegt
  außerhalb beider Repos unter `Documents\arbeit\werkzeuge\trashygecko-release\`
  (`trashygecko-release.jks`, Passwort in `passwort.txt` daneben).
  **Passwörter nie in den Chat, nie ins Repo, nie in eine Ausgabe schreiben.**
- Keine APK von Hand zusammenkopieren (kein aapt2/d8/apksigner-Bau für TrashyGecko).

### Connector
```bash
cd spiele/connector && go vet ./... && go test -run Quiz ./... && go build -o trashygecko-connector.exe .
```
Der Connector liest `connector.html` **einmal beim Start** → nach Seitenänderungen neu
starten. Er sucht sich den ersten freien Port ab 8765; soll er auf einem bestimmten
Port laufen, muss der darunter belegt sein.

### Seiten
Kein Build. Syntax prüfen: Skript aus der HTML in eine `.js` extrahieren und
`node --check` laufen lassen.

## 4. Regeln für Änderungen

- **Bestandsaufnahme vor dem Umbau.** Vor jeder Architekturänderung prüfen, was schon da
  ist (die App enthält mehr, als die Dateiliste vermuten lässt).
- **Keine funktionierende Funktion entfernen**, nur weil eine neue Architektur einfacher
  wäre. Was noch nicht funktioniert, wird dokumentiert und gezielt behoben.
- **Keine neuen Bibliotheken** ohne Eintrag in `ENTSCHEIDUNGEN.md`.
- **Online-Funktionen der Spielhalle** laufen über den Connector (D-27). Keine zweite
  Kennung, keinen zweiten Chat und keine Kopie der Spielseiten anlegen. Vorhanden sind nur
  `androidx.core` und `androidx.webkit`.
- **Spiele bleiben Web-Technik.** Native wird die App drumherum (siehe D-02).
- **Application-ID, Signatur und Version nicht zurücksetzen** (D-01, D-07).
- Vor größeren Eingriffen eine Sicherung unter `Homepage/sicherung/` anlegen
  (Namensschema `thema-jjjj-mm-tt/`); vorhandene Sicherungen nie löschen.
- Geheimnisse: `schluessel.jks`, `keystore.properties`, `passwort.txt`, `ANLEITUNG.txt`
  gehören **nicht** ins APK und nicht ins Repo (`ANLEITUNG.txt` ist schon ausgeschlossen).
- Nach jeder Arbeit: `ENTWICKLUNGSSTAND.md` fortschreiben, Testergebnisse in
  `PRUEFBERICHT-ANDROID.md` – nur wirklich ausgeführte Tests.

### Spielschnittstelle (`SPIELE-SCHNITTSTELLE.md`)

- **Jedes Spiel** bietet `window.<SpielName>` mit `version`, `bereit`, `start()`,
  `stop()`, `pause()`, `resume()`, `einstellungen()`, `zustand()`,
  `startbildschirm()`.
- **Kein Spiel spricht selbst mit dem Server.** Räume, Namen, Start und Ende
  koordiniert der Konnektor; Spiele kennen keine Räume, keine Spieler-IDs und
  keine `/api/`-Rufe.
- **Nur bekannte Felder, feste Fehlercodes:** falsche Typen und unbekannte
  Felder werden mit `{ok:false, fehler:'…'}` abgelehnt, der Spielzustand bleibt
  dabei unverändert.
- **Ereignisse statt Pollen:** `…:bereit` und `…:ende` (mit Ergebnis) sind
  Pflicht, `zustand()` liefert nur Kopien.
- **Leertaste und Enter starten nur ohne fokussiertes Bedienelement**; steht der
  Fokus auf Knopf, Auswahl oder Feld, löst der Browser selbst aus.
- **Vorbild:** `spiele/pimmel/spimmel.html` (Schnittstelle 1.1.0);
  `Super-Pimmelchen-61-Level-Einzeldatei.html` bleibt eingefrorene
  Sonderfassung 1.0.0 und wird nicht nachgezogen.

## 5. Fallen, die schon zugeschlagen haben

- **Debug-APK über release-signierte Installation** → `INSTALL_FAILED_UPDATE_INCOMPATIBLE`.
  Für Gerätetests release bauen oder Emulator nehmen; Deinstallieren löscht Spielstände.
- **WebView-Cache**: `LOAD_NO_CACHE` ist Absicht. Ohne das zeigt die App nach einem
  Update alte Seiten (`replaceChildren`-Fehler in 1.14).
- **`localStorage` hängt an der Herkunft** `https://appassets.androidplatform.net/` –
  deshalb liefert `HomepageAssets` über eine echte https-Herkunft aus und nicht über
  `file://`.
- **Versteckte Tabs/WebView drosseln Timer**; Messungen im Browser-Testfenster sind
  dadurch verfälscht.
- **Die verlinkten Homepage-Seiten liegen im APK** (D-24): `homepageKopieren` kopiert
  neben `spiele/` die 19 Ordner aus `homepageSeiten` in `app/build.gradle` (Tagebuch,
  Bücher, Rezepte, Anleitungen …). Eine neu verlinkte Rubrik gehört in diese Liste, sonst
  ist ihre Kachel im APK tot, obwohl sie im Netz geht. Draußen bleiben der alte Spiegel
  unter `css/` und `tagebuch/Textdokument (neu).txt`.
- **Der Spielekatalog wird beim Bauen verändert** (zwei Zeilen entfernt). Ändert jemand
  die Katalogzeilen, bricht `homepageKopieren` bewusst mit `GradleException`.
- **`assets/katalog.json` ist erzeugt**, nicht gepflegt (D-15, Aufgabe `katalogBauen`).
  Die Spieleliste der App nie von Hand ändern – sie kommt aus `spiele/spiele.html`.
- **Drehung baut die Activities nicht neu** (D-21, `configChanges`): `onCreate` läuft
  nicht noch einmal, Eingaben bleiben also erhalten. Dafür müssen breitenabhängige
  Layouts in `onConfigurationChanged` neu berechnet werden – sonst bleibt der Umbruch der
  alten Breite stehen und die Filter-Chips werden abgeschnitten (am 23.11.2026 so
  passiert und behoben).
- **Prozess-Tod deckt `onSaveInstanceState` ab** (D-22): Die Ansichten der Spieleliste
  haben keine IDs, das System sichert von sich aus nichts. Suchtext, Merkmale,
  Scrollposition, die Ansicht und die offene Kategorie werden in `SpieleActivity` selbst
  ins Bundle gelegt und in `onCreate`
  über `stelleWieder` zurückgeholt (Scroll erst nach dem ersten Layout). Neue
  Filterarten müssen dort mitgesichert werden.
- **Reihenfolge bei der Wiederherstellung** (D-22): erst `stelleWieder`, dann **einmal**
  `zeichne`. Während `suchfeld.setText` den Textwächter auslöst, ist `stelleWiederLaeuft`
  gesetzt – sonst entstünde ein zweiter Listenaufbau, und ein gesichertes Merkmal ohne
  Suchtext bliebe ungefiltert stehen (am 23.11.2026 so gefunden und behoben).
- **Prozess-Tod auf dem Emulator testen**: Zustand setzen → Home →
  `adb shell am kill de.trashygecko.tmahjong` (`pidof` prüfen: leer) → über die
  Übersicht (Recents) zurück. Ein Start per `am start … MAIN/LAUNCHER` räumt den Task
  auf `MainActivity` ab und taugt dafür nicht.
- **Drei Ansichten in der App-Spielhalle** (D-26): Start (`ANSICHT_START`), Kategorie
  (`ANSICHT_KATEGORIE` + `offeneKategorie`) und Trefferliste. Zurück heißt Filter
  löschen → Kategorie verlassen → Activity beenden; `zurueck()` und `onBackPressed`
  nehmen denselben Weg. Kategorie-Kacheln und Highlights sind Karten aus `katalog.json`;
  die Kategorie-Symbole stehen in `symbolFuer`.
- **Einen zweiten Einstieg zur Spieleliste gibt es nicht** (D-20): die Kachel „Spiele“ auf
  der Startseite ist der einzige Weg. Der Connector bietet bewusst keine Spiele-Karte an –
  die frühere lag auf einem Bildschirm, den man von der App aus nicht direkt erreicht.
- **Merkmalsfilter und Suche gibt es zweimal**: `passt(s, q)` in `spiele.html` und
  `passt(spiel)` in `SpieleActivity`. Beide verknüpfen Chips mit UND und kombinieren sie
  mit der Suche (D-19); die App sucht wie das Web in Name, Suchwörtern, Kurztext und
  Kategoriename. Wer die Regel ändert, ändert sie an beiden Stellen. Neue
  Merkmalsnamen brauchen eine Zeile in `MERKMAL_FILTER`, sonst fehlt der Chip.
- **Die Startansicht ist nativ** (D-12): `MainActivity` baut die vier Kacheln (Spiele,
  Lokales Netz, Online-Räume, Extras) als Views; die Kopfzeile zeigt den Netzstatus des
  Geräts. `start/app.html` liegt nur noch als `index.html` im Paket, damit „zurück zum
  Start“ aus einer Spielseite landet – wer die Kacheln ändert, ändert `MainActivity`.
- **Farben und kleine Bausteine stehen in `Ansicht`** (`BG`, `KARTE`, `RAND`, `GOLD`,
  `TEXT`, `MUTED`, `flaeche`, `text`, `dp`). Neue native Ansichten holen sie dort, statt
eigene Farbkonstanten anzulegen – sonst laufen Start und Spieleliste farblich auseinander
(so waren beide Grün- und Grautöne vor dem 23.11.2026 leicht verschieden).
- **Ein WebView, der ins Internet zeigt, braucht drei Zustände** (D-23): vor dem Laden das
  Netz prüfen, `onPageFinished` → „erreichbar“, `onReceivedError` im Hauptframe → deutsche
  Meldung mit „Erneut versuchen“. Chromiums englische Fehlerseite ist kein Nutzertext, und
  die Statuszeile darf den Erfolg nicht vorwegnehmen.
- **Die echte Homepage liegt als `www/homepage.html` im Paket** (D-12) und wird nur über
  „Extras → Homepage“ geöffnet (`HomepageActivity`). Sie entsteht beim Bauen aus der
  `index.html` der Wurzel; `www/index.html` bleibt die alte App-Startseite als Rückweg.
  Von der Website ist nur diese Startseite dabei, nicht Tagebuch, Bücher und Rezepte.
- **Spielseiten laufen in `GameActivity`** (D-16). Wer WebView-Einstellungen ändert,
  ändert sie in `WebBasis` – sonst laufen die Ansichten auseinander.
- **Emulator braucht Zeit**: `adb wait-for-device`, `getprop sys.boot_completed` prüfen,
  sonst kommen `Can't find service: package`-Fehler.
- **`connector`-Bäume nicht zusammenlegen** (D-04) und **keinen Go-Server ins APK** (D-03).

## 6. Was ein Agent nicht ohne Rückfrage darf

- Spiele nativ nachbauen oder die App auf Kotlin/Compose umstellen.
- Die beiden Go-Connector-Bäume zusammenführen.
- Version, Paket-ID oder Signatur ändern.
- Veröffentlichen, committen, Play Store, Löschen von Sicherungen.
- Tests als bestanden melden, die nicht gelaufen sind („baut“ ≠ „läuft“).
