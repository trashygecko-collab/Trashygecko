package main

/* ===========================================================================
   BOT FÜR DIE WÜRFELSPIELE

   Die 18 Würfelspiele bringen ihre Regeln als JavaScript-Module mit; der
   Connector führt sie in einer kleinen JS-Maschine aus (wuerfel.go). Der Bot
   braucht deshalb keine eigene Regelkenntnis: Er liest aus sicht(z), welche
   Knöpfe, Auswahlmöglichkeiten und Würfel das Modul gerade anbietet – genau
   das, was ein Mensch auf der Seite sieht – und nimmt einen davon. Jeder Zug
   geht durch dieselbe Prüfung wie ein menschlicher Zug (p.aktion), schummeln
   kann er also nicht.

   Stufen:
     leicht  – nimmt eine zufällige der angebotenen Möglichkeiten.
     mittel  – schließt den Wurf ab (zweiter Knopf, z. B. „Halten“), sonst
               Hauptknopf, sonst die erste Auswahl, sonst ein Würfel.
     schwer  – dieselbe Zugwahl, denkt aber kürzer (botWartezeit).

   Warum mittel und schwer gleich ziehen: Die Regelmodule liefern nur die
   angebotenen Züge, keine Bewertung. Ein Würfelspiel ist ohnehin Glück; ein
   Bot, der die Regel nicht nachbaut, soll wenigstens immer regelgerecht und
   zügig spielen, statt eine Stärke vorzutäuschen, die er nicht hat.
   =========================================================================== */

import (
	"encoding/json"
	"time"
)

// wuerfelBotWahl ist ein Zug, den das Regelmodul gerade anbietet.
type wuerfelBotWahl struct {
	art   string
	daten json.RawMessage
}

// wuerfelBotTreiben lässt den Bot genau einen Zug machen, sobald seine
// Denkpause vorbei ist. Rückgabe: true, wenn ein Zug ausgeführt wurde.
func wuerfelBotTreiben(p *wuerfelPartie, jetzt time.Time) bool {
	if p == nil || p.Status != "active" || p.BotID == "" {
		return false
	}
	sitz := p.dran()
	if sitz < 0 || sitz >= len(p.Sitze) || p.Sitze[sitz].ID != p.BotID {
		p.BotFrist = time.Time{}
		return false
	}
	if p.BotFrist.IsZero() {
		p.BotFrist = jetzt.Add(botWartezeit(p.BotStufe))
		return false
	}
	if jetzt.Before(p.BotFrist) {
		return false
	}
	p.BotFrist = time.Time{}
	art, daten, ok := wuerfelBotZug(p.sichtWerte(), p.BotStufe)
	if !ok {
		return false
	}
	erfolg, err := p.aktion(art, daten)
	return err == nil && erfolg
}

// wuerfelBotWaehlbar liest die angebotenen Züge aus sicht(z). Die Reihenfolge
// ist die Vorliebe des Bots: erst der Zug, der den Wurf abschließt, dann der
// Hauptknopf, dann die Auswahl, zuletzt ein wählbarer Würfel. Ein abgeschalteter
// Knopf („an“: false) fällt weg – er würde nur eine Absage kassieren.
func wuerfelBotWaehlbar(sicht json.RawMessage) []wuerfelBotWahl {
	var s map[string]any
	if len(sicht) == 0 || json.Unmarshal(sicht, &s) != nil {
		return nil
	}
	var wahlen []wuerfelBotWahl
	knopf := func(v any) {
		k, ok := v.(map[string]any)
		if !ok {
			return
		}
		if an, gesetzt := k["an"].(bool); gesetzt && !an {
			return
		}
		art, _ := k["art"].(string)
		if art == "" {
			return
		}
		var daten json.RawMessage
		if d, da := k["daten"]; da && d != nil {
			if roh, err := json.Marshal(d); err == nil {
				daten = roh
			}
		}
		wahlen = append(wahlen, wuerfelBotWahl{art: art, daten: daten})
	}
	knopf(s["knopf2"])
	knopf(s["knopf"])
	if liste, ok := s["auswahl"].([]any); ok {
		for _, eintrag := range liste {
			knopf(eintrag)
		}
	}
	if wuerfel, ok := s["wuerfel"].([]any); ok {
		for i, w := range wuerfel {
			wm, ok := w.(map[string]any)
			if !ok {
				continue
			}
			if waehlbar, _ := wm["waehlbar"].(bool); !waehlbar {
				continue
			}
			daten, _ := json.Marshal(map[string]any{"index": i})
			wahlen = append(wahlen, wuerfelBotWahl{art: "wuerfel", daten: daten})
		}
	}
	return wahlen
}

// wuerfelBotZug wählt einen der angebotenen Züge. Ohne Angebot gibt es keinen
// Zug – dann wartet der Bot, statt etwas zu raten.
func wuerfelBotZug(sicht json.RawMessage, stufe string) (string, json.RawMessage, bool) {
	wahlen := wuerfelBotWaehlbar(sicht)
	if len(wahlen) == 0 {
		return "", nil, false
	}
	if botStufe(stufe) == botLeicht {
		wahl := wahlen[botZufall().Intn(len(wahlen))]
		return wahl.art, wahl.daten, true
	}
	return wahlen[0].art, wahlen[0].daten, true
}
