# Docker-Abbild der Online-Fassung des Connectors.
#
# Gebaut wird aus dem Repo-Wurzelverzeichnis, weil dort die Webseite und der
# Ordner "spiele" liegen (die Online-Fassung liefert beides aus):
#
#   docker build -f spiele/connector-online/Dockerfile -t trashygecko-connector .
#   docker run -p 8080:8080 trashygecko-connector
#
# Auf Render übernimmt das render.yaml im Repo-Wurzelverzeichnis.

# 1) Programm bauen – nur der Go-Ordner, damit die Webseite den Build nicht
#    aufhält.
FROM golang:1.27 AS bauen
WORKDIR /bau
COPY spiele/connector-online/go.mod spiele/connector-online/go.sum ./
RUN go mod download
COPY spiele/connector-online/ ./
# Ohne cgo: das Ergebnis ist eine einzelne Datei, die auch in Alpine läuft.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /aus/connector .

# 2) Webseite einsammeln. Die beiden Programmordner bleiben draußen: dort
#    liegen Quelltext und die (großen) exen, die niemand herunterladen soll.
FROM alpine:3.20 AS webseite
WORKDIR /web
COPY . .
RUN rm -rf .git spiele/connector spiele/connector-online

# 3) Laufzeit
FROM alpine:3.20
WORKDIR /app
RUN adduser -D -u 10001 spieler
# Die Oberfläche sucht connector.html neben der Programmdatei …
COPY --from=bauen /aus/connector /app/connector
COPY --from=bauen /bau/connector.html /app/connector.html
# … und die Webseite im Ordner SITE_DIR.
COPY --from=webseite /web /app/webseite
ENV SITE_DIR=/app/webseite
ENV PORT=8080
EXPOSE 8080
USER spieler
CMD ["/app/connector"]
