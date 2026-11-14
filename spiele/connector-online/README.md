# Connector online

Dieselbe Spiele-Sammlung wie der lokale Connector, aber für einen Server im
Internet: kein Port-Suchen, kein automatisch geöffneter Browser, Adressen und
QR-Code zeigen auf die öffentliche Seite.

## Aktuell halten

Nach jeder Änderung am normalen Connector:

```
cd spiele\connector-online
python sync.py --linux
```

Das kopiert alle gemeinsamen Dateien (`*.go`, `connector.html`, `go.mod`,
`go.sum`) aus `../connector`, prüft sie mit `go vet` und baut

- `trashygecko-connector-online.exe` – zum Ausprobieren unter Windows,
- `trashygecko-connector-online-linux` – die Datei für den Server.

Eigene Dateien (`main.go`, `netz.go`, `status.go`, `qr.go`) bleiben dabei
unangetastet. Neue Spiele brauchen keine Extra-Arbeit: die Endpunkte stehen in
`server.go` (im Connector-Ordner) und werden mitkopiert.

## Auf den Server legen

Neben die Programmdatei gehört `connector.html`. Die Webseite selbst (der
Ordner mit `index.html` und `spiele/`) liegt irgendwo daneben, zum Beispiel:

```
/var/www/spiele/            <- Webseite (index.html, spiele/, ...)
/var/www/spiele/server/     <- trashygecko-connector-online-linux + connector.html
```

Starten:

```
SITE_DIR=/var/www/spiele PORT=8080 ./trashygecko-connector-online-linux
```

## Einstellungen

| Variable | Bedeutung | Standard |
| --- | --- | --- |
| `PORT` | Port, auf dem der Connector hört | `8080` |
| `ADDR` | Adresse, an die gebunden wird | alle (`:PORT`) |
| `SITE_DIR` | Ordner der Webseite | Ordner der Programmdatei bzw. dessen Elternordner, wenn dort `index.html` liegt |
| `PUBLIC_URL` | öffentliche Adresse für Link und QR-Code, z. B. `https://spiele.example.de` | aus der Anfrage (`X-Forwarded-Proto`/`Host`) |

Hinter nginx/Apache reicht ein Reverse Proxy auf `http://127.0.0.1:8080`;
`PUBLIC_URL` setzen, damit QR-Code und Adressfeld sicher `https` zeigen.

## Was online anders ist

- **Boss** ist, wer den Raum als Erster betritt (am Server-PC gibt es keinen
  Bildschirm, an dem jemand „der Host" wäre). Alle danach sind Spieler1,
  Spieler2, … – umbenennen wie gehabt.
- Adresse und QR-Code kommen aus `PUBLIC_URL` bzw. der Anfrage statt aus der
  LAN-IP.
- Kein Browser-Start, kein Durchprobieren von Ports.

## Docker und Render (Internet ohne eigenen PC-Betrieb)

`Dockerfile` baut die Online-Fassung samt Webseite in ein Abbild, `render.yaml`
im Repo-Wurzelverzeichnis beschreibt den Dienst für Render.

    docker build -f spiele/connector-online/Dockerfile -t trashygecko-connector .
    docker run -p 8080:8080 trashygecko-connector

Wichtig beim Bauen: Das **Repo-Wurzelverzeichnis** ist der Kontext (`-f` zeigt nur
auf die Dockerfile), weil dort die Webseite und `spiele/` liegen. Im Abbild steht
das Programm unter `/app/connector`, die Oberfläche daneben unter
`/app/connector.html`, die Webseite unter `/app/webseite` (`SITE_DIR`).
Die Ordner `spiele/connector` und `spiele/connector-online` kommen nicht mit ins
Abbild – dort liegen Quelltext und die großen exen.

Auf Render: „New" → „Blueprint" → Repository wählen. Der Free-Plan schläft nach
15 Minuten ohne Anfrage ein und wacht beim nächsten Aufruf in etwa einer Minute
wieder auf; dabei gehen Räume und laufende Partien verloren, die Seite tritt
danach automatisch wieder bei.

## Bei Render veröffentlichen (auch ohne eigene Domain)

1. Repo zu GitHub pushen (die Dateien in diesem Ordner müssen dort liegen):

       git add -A
       git commit -m "Connector: offenes Rundangebot, Docker/Render"
       git push

2. render.com öffnen, mit dem GitHub-Konto anmelden, "New +" -> "Blueprint"
   wählen und das Repository verbinden. Render findet `render.yaml` im
   Wurzelordner des Repos und legt den Dienst `trashygecko-connector` an.
3. "Apply" drücken. Render baut das Dockerfile (ein paar Minuten, Fortschritt
   unter "Logs") und startet den Dienst.
4. Danach ist die Seite erreichbar unter
   `https://trashygecko-connector.onrender.com` – der Connector liegt unter
   `/connector`. Adresse und QR-Code in der Seite zeigen automatisch diese
   Adresse, weil sie aus den Kopfzeilen des Proxys kommen.
5. Free-Plan: Ohne Besuch schläft der Dienst nach 15 Minuten ein; der nächste
   Aufruf weckt ihn in etwa einer Minute (Render zeigt eine Ladeseite).
   Räume und laufende Partien sind dann weg, die Seite tritt automatisch neu bei.

### Eigene Domain später

Eine eigene Domain ist **nicht nötig**. Wer eine will:

1. Domain kaufen (üblich: 5–15 € im Jahr), zum Beispiel bei IONOS, Namecheap
   oder Cloudflare.
2. In Render: Dienst -> "Settings" -> "Custom Domains" -> "Add Domain",
   z. B. `spiele.deine-domain.de`. Render zeigt den Zielnamen
   (CNAME auf `…onrender.com`).
3. Beim Domain-Anbieter diesen CNAME für `spiele` eintragen.
4. Warten, bis der Eintrag greift; Render holt das TLS-Zertifikat selbst.
5. Optional den Block `domains:` in `render.yaml` eintragen (siehe Kommentar
   dort), damit ein neuer Blueprint-Lauf die Domain gleich mitnimmt.

Liegt die Domain bei Cloudflare: den Eintrag auf "DNS only" (graue Wolke)
stellen, sonst scheitert die Zertifikatsprüfung von Render.
