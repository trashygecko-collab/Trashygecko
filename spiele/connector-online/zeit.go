package main

// Zeitanzeige an einer Stelle.
//
// Gespeichert und gerechnet wird ausschließlich mit Zeitpunkten (time.Time).
// Die sind absolut und damit unabhängig von jeder Zone: Fristen, Rundenzeiten,
// „zuletzt aktiv" und alle Vergleiche laufen unverändert über time.Since,
// Sub und Add. Nur die Ausgabe – was ein Mensch als Uhrzeit liest – bekommt
// eine Zone, und zwar eine einzige: Europe/Berlin (MEZ/MESZ, Sommerzeit
// automatisch über die Zeitzonendatenbank).
//
// Warum das nötig war: Vorher stand an rund 200 Stellen
// time.Now().Format("15:04:05"). Das nimmt die Zone des Prozesses. Auf dem
// eigenen PC ist das Europe/Berlin, im Container (Render/Alpine) aber UTC –
// dort stand im Chat und in der Raumliste eine Stunde zu früh (Winterzeit),
// im Sommer zwei. Eine Umrechnung passiert jetzt genau einmal, hier.

import (
	"time"
	_ "time/tzdata" // Zeitzonendaten einbetten: Alpine hat keine /usr/share/zoneinfo
)

// anzeigeZone ist die Zone für alle Anzeigen. time/tzdata liefert die Daten,
// damit LoadLocation auch im schlanken Container funktioniert.
var anzeigeZone = func() *time.Location {
	zone, fehler := time.LoadLocation("Europe/Berlin")
	if fehler != nil {
		// Notnagel, falls die Datenbank fehlt: MEZ ohne Sommerzeit.
		return time.FixedZone("MEZ", 60*60)
	}
	return zone
}()

// uhrzeit liefert die Uhrzeit mit Sekunden in deutscher Ortszeit ("19:04:05").
func uhrzeit(t time.Time) string {
	return t.In(anzeigeZone).Format("15:04:05")
}

// uhrzeitKurz liefert die Uhrzeit ohne Sekunden ("19:04") – für die Raumliste.
func uhrzeitKurz(t time.Time) string {
	return t.In(anzeigeZone).Format("15:04")
}

// datum liefert den Kalendertag in deutscher Ortszeit ("2026-11-25").
// Wichtig für den Tageswechsel: um 00:30 Ortszeit ist in UTC noch der Vortag.
func datum(t time.Time) string {
	return t.In(anzeigeZone).Format("2006-01-02")
}
