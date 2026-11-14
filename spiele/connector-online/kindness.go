package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

// Sprüche für den Namens-Klick im Chat. "a" ist der Absender, "b" der
// angeklickte Spieler; beide werden beim Senden ersetzt.
var toneLove = []string{
	`"a" findet "b" einfach zum Dahinschmelzen`,
	`"a" schenkt "b" das letzte Stück Schokolade`,
	`"a" deckt "b" liebevoll mit einer flauschigen Decke zu`,
	`"a" schickt "b" eine virtuelle Umarmung, die alle Sorgen vertreibt`,
	`"a" hat "b" einen heißen Kakao mit extra viel Sahne gemacht`,
	`"a" findet, dass "b" heute wieder besonders bezaubernd aussieht`,
	`"a" flüstert "b" zu, wie wertvoll "b" für die Welt ist`,
	`"a" streichelt "b" sanft über den Kopf und lächelt`,
	`"a" wünscht "b" einen Tag voller kleiner Wunder`,
	`"a" hält "b" ganz fest und lässt so schnell nicht mehr los`,
	`"a" hat für "b" extra die schönsten Blumen gepflückt`,
	`"a" bewundert "b" für den unglaublichen Mut und die Stärke`,
	`"a" schickt "b" einen Sonnenstrahl direkt ins Herz`,
	`"a" backt für "b" eine riesige Portion Lieblingskekse`,
	`"a" findet das Lachen von "b" ist die schönste Musik überhaupt`,
	`"a" beschützt "b" vor allen schlechten Träumen`,
	`"a" schaut "b" voller Bewunderung und Zuneigung an`,
	`"a" hinterlässt "b" eine kleine Nachricht, um den Tag zu versüßen`,
	`"a" reicht "b" eine helfende Hand, wann immer es nötig ist`,
	`"a" teilt mit "b" den allerletzten Regenschirm im Sturm`,
	`"a" ist unendlich dankbar, dass es "b" gibt`,
	`"a" schickt "b" ein Lächeln, das bis über beide Ohren geht`,
	`"a" hat "b" einen Platz an der Sonne freigehalten`,
	`"a" pustet alle Sorgen von "b" einfach weg`,
	`"a" hält "b" den Rücken frei, egal was passiert`,
	`"a" zaubert "b" mit einem Kompliment die Röte ins Gesicht`,
	`"a" schickt "b" eine ganze Ladung positive Energie`,
	`"a" hört "b" stundenlang geduldig und mit offenem Herzen zu`,
	`"a" findet, dass "b" ein echter Glücksgriff für diesen Chat ist`,
	`"a" spendet "b" Trost und Wärme an einem grauen Tag`,
	`"a" hat extra für "b" das Lieblingslied angemacht`,
	`"a" lächelt "b" an, als wäre "b" das achte Weltwunder`,
	`"a" schenkt "b" ungeteilte Aufmerksamkeit und ganz viel Zeit`,
	`"a" drückt "b" ganz fest die Daumen für alles, was ansteht`,
	`"a" bringt "b" das Frühstück ans Bett`,
	`"a" schickt "b" einen virtuellen Kuss auf die Stirn`,
	`"a" ist mächtig stolz auf alles, was "b" bisher erreicht hat`,
	`"a" hüllt "b" in eine Wolke aus purer Geborgenheit`,
	`"a" teilt mit "b" die schönsten Erinnerungen`,
	`"a" fängt "b" auf, falls "b" mal ins Stolpern gerät`,
	`"a" strahlt "b" an wie ein Honigkuchenpferd`,
	`"a" schenkt "b" ein offenes Ohr und ein großes Herz`,
	`"a" findet, dass "b" eine absolut inspirierende Persönlichkeit hat`,
	`"a" schickt "b" ein kleines Herzchen durch die Leitung`,
	`"a" wärmt "b" die kalten Hände`,
	`"a" ist immer für "b" da, egal zu welcher Uhrzeit`,
	`"a" feiert jeden noch so kleinen Erfolg von "b"`,
	`"a" schaut "b" tief in die Augen und nickt anerkennend`,
	`"a" hat für "b" extra die beste Aussicht reserviert`,
	`"a" schickt "b" einen Gruß voller Liebe und Harmonie`,
	`"a" baut "b" nach einem langen Tag wieder auf`,
	`"a" findet, dass die Welt durch "b" ein besserer Ort ist`,
	`"a" reicht "b" ein Taschentuch und drückt "b" ganz fest`,
	`"a" hat "b" einen kleinen Glücksbringer eingesteckt`,
	`"a" lächelt über die liebenswerte Art von "b"`,
	`"a" schickt "b" eine riesige Portion Zuversicht`,
	`"a" tanzt mit "b" vor lauter Freude durch den Raum`,
	`"a" gibt "b" das Gefühl, genau richtig zu sein`,
	`"a" schickt "b" einen warmen Sommerregen für die Seele`,
	`"a" hält "b" die Tür auf und verbeugt sich charmant`,
	`"a" hat für "b" die Sterne vom Himmel geholt`,
	`"a" schickt "b" ein virtuelles Dankeschön für alles`,
	`"a" bewundert die unendliche Geduld von "b"`,
	`"a" teilt mit "b" das Geheimnis der guten Laune`,
	`"a" schickt "b" einen frischen Kaffee direkt an den Platz`,
	`"a" findet, dass "b" ein echtes Goldstück ist`,
	`"a" drückt "b" so fest, dass kurz die Luft wegbleibt`,
	`"a" schickt "b" eine unsichtbare Krone, weil "b" großartig ist`,
	`"a" schenkt "b" ein Lächeln, das den ganzen Raum erhellt`,
	`"a" unterstützt "b" bei jedem noch so verrückten Plan`,
	`"a" schickt "b" eine Brise Rückenwind für alle Vorhaben`,
	`"a" findet, dass "b" einfach unbezahlbar ist`,
	`"a" reicht "b" die Hand für einen gemeinsamen Spaziergang`,
	`"a" schickt "b" einen virtuellen Blumenstrauß`,
	`"a" freut sich riesig, "b" hier zu sehen`,
	`"a" hat für "b" extra den roten Teppich ausgerollt`,
	`"a" schickt "b" eine Extra-Portion Glück`,
	`"a" findet die Kreativität von "b" absolut bewundernswert`,
	`"a" schenkt "b" ein ehrliches, tiefes Kompliment`,
	`"a" schickt "b" einen Schutzengel für den Rest des Tages`,
	`"a" teilt mit "b" die Freude am Leben`,
	`"a" schaut "b" an und denkt sich, wie schön das Leben ist`,
	`"a" schickt "b" eine virtuelle Kuscheldecke`,
	`"a" hat "b" ganz fest ins Herz geschlossen`,
	`"a" schickt "b" ein leises, liebevolles "Schön, dass es dich gibt"`,
	`"a" bewundert den großartigen Humor von "b"`,
	`"a" schenkt "b" eine Pause vom stressigen Alltag`,
	`"a" schickt "b" einen Regenbogen direkt in den Chat`,
	`"a" ist für "b" der Fels in der Brandung`,
	`"a" schickt "b" ganz viel Kraft und Ruhe`,
	`"a" findet, dass "b" das Herz am rechten Fleck hat`,
	`"a" reicht "b" eine Tasse Tee zur Entspannung`,
	`"a" schickt "b" ein kleines virtuelles Feuerwerk der Freude`,
	`"a" hält "b" immer einen Platz im Herzen frei`,
	`"a" schickt "b" einen lieben Gedanken aus der Ferne`,
	`"a" lächelt "b" an und weiß, dass alles gut wird`,
	`"a" schenkt "b" ein Stück unbeschwerte Leichtigkeit`,
	`"a" schickt "b" den allerbesten Wunsch für die Zukunft`,
	`"a" findet "b" einfach rundum wundervoll`,
	`"a" verabschiedet sich von "b" mit einem warmen Lächeln`,
}

var toneJoke = []string{
	`"a" bewirft "b" aus dem Hinterhalt mit Wattebällchen`,
	`"a" erklärt "b" die Quantenmechanik anhand einer angebissenen Gurke`,
	`"a" stiehlt "b" heimlich den linken Schuh`,
	`"a" tut so, als wäre "b" eine seltene Zimmerpflanze, und gießt "b"`,
	`"a" schickt "b" eine Brieftaube mit einer leeren Nachricht`,
	`"a" fordert "b" zu einem epischen Duell im Daumenwrestling heraus`,
	`"a" versteckt alle Löffel von "b" und ersetzt sie durch Gabeln`,
	`"a" starrt "b" an, bis "b" freiwillig die Wahrheit über die Kekse sagt`,
	`"a" schickt "b" ein unsichtbares Geschenk und wartet auf Dank`,
	`"a" behauptet steif und fest, "b" wäre ein getarnter Geheimagent`,
	`"a" baut vor "b" eine Mauer aus leeren Kaffeetassen auf`,
	`"a" schickt "b" eine Rechnung für verbrauchte Atemluft`,
	`"a" versucht, "b" mit einem Toastbrot zu hypnotisieren`,
	`"a" verkleidet sich als Geist und erschreckt "b" im Chat`,
	`"a" schickt "b" eine Anleitung zum korrekten Atmen`,
	`"a" klaut "b" im Vorbeigehen den Wlan-Empfang`,
	`"a" fängt an, jedes Wort von "b" in Zeitlupe nachzusprechen`,
	`"a" schickt "b" ein Foto von einem Stein und nennt ihn "b"s Zwilling`,
	`"a" fordert "b" auf, sofort einen Tanz für den Regengott aufzuführen`,
	`"a" klebt "b" heimlich einen Zettel mit "Bitte schubsen" auf den Rücken`,
	`"a" schickt "b" eine Warnung vor aggressiven Eichhörnchen`,
	`"a" tut so, als hätte "b" gerade das Internet kaputt gemacht`,
	`"a" überreicht "b" feierlich die goldene Himbeere für unnützes Wissen`,
	`"a" schickt "b" ein Konzert auf einer unsichtbaren Luftgitarre`,
	`"a" schleicht sich an "b" heran und ruft ganz laut "Buh!"`,
	`"a" ersetzt das Wasser von "b" durch flüssige Verwirrung`,
	`"a" schickt "b" einen Gutschein für einmal dumm aus der Wäsche schauen`,
	`"a" beginnt eine ernsthafte Diskussion mit "b"s Schatten`,
	`"a" schickt "b" eine Abmahnung wegen zu guter Laune`,
	`"a" stülpt "b" einen Eimer über den Kopf und trommelt darauf`,
	`"a" behauptet, "b" schuldet "a" noch drei Kamele und ein Einhorn`,
	`"a" schickt "b" eine Einladung zur Weltmeisterschaft im Nichtstun`,
	`"a" kitzelt "b" mit einer imaginären Pfauenfeder`,
	`"a" tut so, als wäre "b" ein Roboter, und sucht den Ausschalter`,
	`"a" schickt "b" ein Paket voller Luft aus dem letzten Urlaub`,
	`"a" fordert "b" zu einem Wettrennen auf Bürostühlen heraus`,
	`"a" malt "b" im Schlaf einen Schnurrbart ins Gesicht`,
	`"a" schickt "b" den Beweis, dass Tauben eigentlich Drohnen sind`,
	`"a" versteckt sich hinter einem Grashalm vor "b"`,
	`"a" schickt "b" ein Ticket für den Hypetrain ins Nirgendwo`,
	`"a" versucht, "b" an einen fliegenden Händler zu verkaufen`,
	`"a" schickt "b" eine Sprachnachricht, in der nur geatmet wird`,
	`"a" bewirft "b" mit Konfetti aus dem Aktenvernichter`,
	`"a" erklärt "b" zum offiziellen König der Ausreden`,
	`"a" schickt "b" eine Karte für eine Schatzsuche im eigenen Kühlschrank`,
	`"a" tut so, als ob "b" unsichtbar wäre, und läuft durch "b" hindurch`,
	`"a" schickt "b" ein Rezept für angebranntes Wasser`,
	`"a" bindet heimlich die Schnürsenkel von "b" zusammen`,
	`"a" schickt "b" eine Nachricht aus der Zukunft: Es gibt keine Pizza mehr`,
	`"a" schaut "b" durch ein leeres Klopapierrohr an`,
	`"a" schickt "b" eine Liste mit Dingen, die absolut keinen Sinn ergeben`,
	`"a" fordert "b" auf, die Schwerkraft wieder einzuschalten`,
	`"a" schickt "b" einen Pokal für die beste Ausrede des Tages`,
	`"a" bewirft "b" mit Gummibärchen, bis "b" aufgibt`,
	`"a" schickt "b" eine Bedienungsanleitung für einen Holzlöffel`,
	`"a" tut so, als wäre "b"s Meinung ein seltenes Kunstwerk`,
	`"a" schickt "b" eine virtuelle Banane zum Ausrutschen`,
	`"a" hält "b" ein Schild mit der Aufschrift "Hier könnte Ihre Werbung stehen" vor`,
	`"a" schickt "b" den ultimativen Plan zur Übernahme der Bäckerei`,
	`"a" fängt an, vor "b" wie ein Pinguin zu watscheln`,
	`"a" schickt "b" eine Tüte Mitleid mit extra viel Salz`,
	`"a" behauptet, "b" hätte ein geheimes Doppelleben als Pantomime`,
	`"a" schickt "b" ein Abo für tägliche Sinnlos-Fakten`,
	`"a" versucht, "b" mit einem Laserpointer abzulenken`,
	`"a" schickt "b" eine Warnung vor spontaner Selbstentzündung`,
	`"a" baut "b" eine Falle aus Legosteinen vor dem Bett`,
	`"a" schickt "b" ein Zeugnis mit einer Sechs in Realitätsnähe`,
	`"a" tut so, als ob "b" ein Geist wäre, und ruft die Geisterjäger`,
	`"a" schickt "b" eine Einladung zu einem geheimen Treffen der Socken-Diebe`,
	`"a" setzt "b" einen Aluhut auf und nickt wissend`,
	`"a" schickt "b" ein Zertifikat für professionelles Prokrastinieren`,
	`"a" schlägt "b" mit einem riesigen Schaumstoff-Hammer`,
	`"a" schickt "b" eine Nachricht, die sich in fünf Sekunden selbst zerstört`,
	`"a" tut so, als würde "b" in einer anderen Sprache reden`,
	`"a" schickt "b" ein Foto von einem leeren Teller als Essenseinladung`,
	`"a" fordert "b" auf, sofort den Ernst der Lage zu verlassen`,
	`"a" schickt "b" einen Strafzettel wegen zu schnellen Tippens`,
	`"a" versteckt "b"s Autoschlüssel im Gefrierfach`,
	`"a" schickt "b" eine Wolke aus Glitzer, die nie wieder weggeht`,
	`"a" beginnt, jede Bewegung von "b" wie ein Sportreporter zu kommentieren`,
	`"a" schickt "b" ein Ticket für die erste Reihe beim Weltuntergang`,
	`"a" bewirft "b" mit imaginären Schneebällen`,
	`"a" schickt "b" eine Statistik über die Intelligenz von Toastbrot`,
	`"a" tut so, als ob "b" gerade ein Weltgeheimnis verraten hätte`,
	`"a" schickt "b" eine Dose Mitleid zum Selbstaufmachen`,
	`"a" klebt "b" ein Post-it auf die Stirn`,
	`"a" schickt "b" einen Link zu einer Seite, die nur "Nein" sagt`,
	`"a" fordert "b" zu einem Blickduell ohne Blinzeln heraus`,
	`"a" schickt "b" eine Anleitung zum Fangen von Einhörnern`,
	`"a" tut so, als ob "b" ein berühmter Rockstar wäre, und kreischt`,
	`"a" schickt "b" eine Rechnung für emotionalen Beistand`,
	`"a" schiebt "b" unauffällig die Schuld an allem in die Schuhe`,
	`"a" schickt "b" ein unsichtbares Fahrrad`,
	`"a" fragt "b", ob die Stimmen im Kopf heute auch so laut sind`,
	`"a" schickt "b" eine Urkunde für den ersten Platz im Chaos-Stiften`,
	`"a" wirft "b" einen Rettungsring im Trockenen zu`,
	`"a" schickt "b" eine Nachricht, die absolut nichts aussagt`,
	`"a" tut so, als ob "b"s Kopf ein Kristallball wäre`,
	`"a" schickt "b" ein virtuelles High-Five ins Gesicht`,
	`"a" verabschiedet sich von "b" mit einem eleganten Bauchklatscher`,
}

var toneEvil = []string{
	`"a" schubst "b" ohne mit der Wimper zu zucken in eine tiefe Pfütze`,
	`"a" wünscht "b", dass beim Händewaschen immer die Ärmel runterrutschen`,
	`"a" isst vor den Augen von "b" das letzte Stück Kuchen auf`,
	`"a" blockiert "b" im wichtigsten Moment absichtlich das Internet`,
	`"a" tritt "b" mit voller Absicht auf die frisch geputzten Schuhe`,
	`"a" verpetzt "b" wegen einer absoluten Kleinigkeit`,
	`"a" wünscht "b" einen Legostein direkt unter den nackten Fuß`,
	`"a" ignoriert "b" eiskalt und tut so, als wäre niemand da`,
	`"a" verrät "b" das Ende der Lieblingsserie ohne Vorwarnung`,
	`"a" zieht "b" den Stuhl weg, kurz bevor "b" sich hinsetzt`,
	`"a" lacht "b" schadenfroh aus, als "b" über die eigenen Füße stolpert`,
	`"a" streut "b" heimlich Salz in den morgendlichen Kaffee`,
	`"a" lässt "b" im strömenden Regen vor der verschlossenen Tür stehen`,
	`"a" löscht absichtlich den Spielstand von "b"`,
	`"a" gibt "b" die Schuld für einen Fehler, den "a" selbst gemacht hat`,
	`"a" schickt "b" in die völlig falsche Richtung`,
	`"a" wünscht "b", dass das Kopfkissen auf beiden Seiten warm ist`,
	`"a" stiehlt "b" die beste Idee und verkauft sie als die eigene`,
	`"a" schlägt "b" die Tür direkt vor der Nase zu`,
	`"a" lässt "b" absichtlich ins offene Messer laufen`,
	`"a" verdirbt "b" mit einem einzigen Satz die komplette Vorfreude`,
	`"a" dreht "b" im Winter heimlich die Heizung ab`,
	`"a" schaut "b" verächtlich von oben bis unten an`,
	`"a" lässt "b" stundenlang auf eine wichtige Antwort warten`,
	`"a" wünscht "b" einen dauerhaften Ohrwurm von einem schrecklichen Lied`,
	`"a" sabotiert heimlich die Arbeit von "b"`,
	`"a" stellt "b" vor allen anderen bloß`,
	`"a" klaut "b" den letzten Parkplatz direkt vor der Nase weg`,
	`"a" wünscht "b", dass der Akku immer bei einem Prozent ausgeht`,
	`"a" zerreißt die Pläne von "b" in der Luft`,
	`"a" schließt "b" versehentlich-absichtlich auf dem Balkon aus`,
	`"a" gibt "b" einen Rat, der garantiert in einer Katastrophe endet`,
	`"a" grinst hämisch, als "b" in die vorbereitete Falle tappt`,
	`"a" lässt "b" die ganze Drecksarbeit alleine machen`,
	`"a" wünscht "b" ewigen Juckreiz an einer Stelle, die man nicht erreicht`,
	`"a" schaltet "b" mitten im spannenden Film den Fernseher aus`,
	`"a" verbreitet ein fieses Gerücht über "b"`,
	`"a" verweigert "b" jede Hilfe in der größten Not`,
	`"a" wünscht "b", dass der Löffel immer in der Suppe versinkt`,
	`"a" nimmt "b" das letzte Wort und verdreht den gesamten Sinn`,
	`"a" lässt "b" mit den Konsequenzen komplett alleine hängen`,
	`"a" wünscht "b" einen Einkaufswagen mit einem klemmenden Rad`,
	`"a" zerstört absichtlich die Sandburg von "b"`,
	`"a" zeigt mit dem Finger auf "b" und lacht laut los`,
	`"a" wünscht "b", dass beim Kekse-Tunken immer die Hälfte abbricht`,
	`"a" zieht "b" gnadenlos über den Tisch`,
	`"a" fällt "b" bei der erstbesten Gelegenheit in den Rücken`,
	`"a" wünscht "b" eine langsame Internetverbindung beim Streaming`,
	`"a" lässt "b" absichtlich auflaufen`,
	`"a" nimmt "b" die Luft aus den Fahrradreifen`,
	`"a" wünscht "b", dass die Socken im Schuh immer nach vorne rutschen`,
	`"a" kommentiert das Scheitern von "b" mit einem kalten Applaus`,
	`"a" schneidet "b" mitten im Satz eiskalt das Wort ab`,
	`"a" wünscht "b" eine Mücke im Schlafzimmer, die man nie findet`,
	`"a" setzt "b" auf die schwarze Liste`,
	`"a" trinkt den letzten Schluck Milch aus und stellt die leere Packung zurück`,
	`"a" wünscht "b", dass der Toast immer auf die beschmierte Seite fällt`,
	`"a" lässt "b" bei einer Gruppenarbeit komplett im Stich`,
	`"a" schaut weg, als "b" dringend Unterstützung braucht`,
	`"a" wünscht "b" einen ewigen Stau auf dem Weg in den Urlaub`,
	`"a" nutzt die Gutmütigkeit von "b" schamlos aus`,
	`"a" zerkratzt absichtlich die Lieblings-CD von "b"`,
	`"a" wünscht "b", dass beim Niesen immer jemand dazwischenredet`,
	`"a" schickt "b" eine falsche Uhrzeit für einen wichtigen Termin`,
	`"a" lässt "b" vor verschlossener Kasse stehen, während "a" geht`,
	`"a" wünscht "b" einen Klebestreifen, bei dem man den Anfang nicht findet`,
	`"a" verrät das größte Geheimnis von "b" an alle weiter`,
	`"a" schaltet "b" den Wecker aus, damit "b" verschläft`,
	`"a" wünscht "b", dass die Kopfhörer immer an der Türklinke hängen bleiben`,
	`"a" lässt "b" absichtlich über ein Hindernis stolpern`,
	`"a" zieht "b" mit fiesen Kommentaren komplett runter`,
	`"a" wünscht "b" einen Papierschnitt genau zwischen den Fingern`,
	`"a" verweigert "b" den Zutritt zur Gruppe`,
	`"a" isst "b"s beschriftetes Mittagessen aus dem Kühlschrank`,
	`"a" wünscht "b", dass die Jeans direkt nach dem Waschen zu eng ist`,
	`"a" lächelt falsch und plant schon die nächste Gemeinheit gegen "b"`,
	`"a" lässt "b" den ganzen Weg umsonst laufen`,
	`"a" wünscht "b" eine Ampel, die immer genau vor "b" auf Rot springt`,
	`"a" wirft die Mühe von "b" achtlos in den Papierkorb`,
	`"a" gibt "b" die unmöglichste Aufgabe und lehnt sich zurück`,
	`"a" streicht "b" kommentarlos von der Einladungsliste`,
	`"a" macht sich über die größte Schwäche von "b" lustig`,
	`"a" wünscht "b" einen Stift, der genau dann leer ist, wenn man ihn braucht`,
	`"a" lässt "b" ins Leere greifen`,
	`"a" wünscht "b", dass der Sicherheitsgurt im Auto ständig klemmt`,
	`"a" verdreht die Augen, sobald "b" den Raum betritt`,
	`"a" gibt "b" das Gefühl, komplett überflüssig zu sein`,
	`"a" wünscht "b" eine Tastatur, bei der die Leertaste klemmt`,
	`"a" reißt "b" das letzte Blatt Klopapier vor der Nase weg`,
	`"a" schaut "b" an und schüttelt nur abwertend den Kopf`,
	`"a" wünscht "b", dass die Milch im Kaffee sofort ausflockt`,
	`"a" lässt "b" für eine Sache bezahlen, die "a" kaputt gemacht hat`,
	`"a" wünscht "b" einen Reißverschluss, der genau in der Mitte klemmt`,
	`"a" dreht "b" demonstrativ den Rücken zu`,
	`"a" wünscht "b" einen Sonnenbrand an einer Stelle, die man nicht eincremen kann`,
	`"a" lässt "b" absichtlich im Dunkeln tappen`,
	`"a" wünscht "b", dass der Zug immer genau eine Sekunde vorher abfährt`,
	`"a" verpasst "b" einen verbalen Tiefschlag`,
	`"a" schließt die Tür hinter sich ab und lässt "b" einfach stehen`,
}

// toneSentences liefert die Sprüche einer Art, oder nil bei unbekannter Art.
func toneSentences(kind string) []string {
	switch kind {
	case "love":
		return toneLove
	case "joke":
		return toneJoke
	case "evil":
		return toneEvil
	}
	return nil
}

// toneHandler schickt einen Spruch von "a" über "b" in den Chat. Der Absender
// ist der, der den Namen angeklickt hat, der Empfänger der angeklickte Spieler.
func toneHandler(w http.ResponseWriter, r *http.Request) {
	code := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("room")))
	if code == "" {
		http.Error(w, "Spielraum-Code fehlt", http.StatusBadRequest)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "POST erforderlich", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		ID       string `json:"id"`
		TargetID string `json:"targetId"`
		Kind     string `json:"kind"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Ungültige Daten", http.StatusBadRequest)
		return
	}

	kind := strings.ToLower(strings.TrimSpace(input.Kind))
	sentences := toneSentences(kind)
	if sentences == nil {
		http.Error(w, "Unbekannte Art.", http.StatusBadRequest)
		return
	}

	mu.Lock()
	room := rooms[code]
	if room == nil {
		mu.Unlock()
		http.Error(w, "Spielraum nicht gefunden", http.StatusNotFound)
		return
	}
	room.cleanPlayers()
	sender, ok := room.Players[input.ID]
	if !ok {
		mu.Unlock()
		http.Error(w, "Bitte erneut dem Spielraum beitreten.", http.StatusForbidden)
		return
	}
	target, ok := room.Players[input.TargetID]
	if !ok || input.TargetID == input.ID {
		mu.Unlock()
		http.Error(w, "Dieser Spieler ist nicht verfügbar.", http.StatusBadRequest)
		return
	}

	// Nicht öfter als der normale Chat: höchstens fünf in zehn Sekunden.
	now := time.Now()
	recent := []time.Time{}
	for _, t := range room.MsgTimes[input.ID] {
		if now.Sub(t) < 10*time.Second {
			recent = append(recent, t)
		}
	}
	if len(recent) >= 5 {
		mu.Unlock()
		http.Error(w, "Zu viele Nachrichten in kurzer Zeit", http.StatusTooManyRequests)
		return
	}
	room.MsgTimes[input.ID] = append(recent, now)

	text := sentences[rand.Intn(len(sentences))]
	text = strings.ReplaceAll(text, `"a"`, `"`+sender.Name+`"`)
	text = strings.ReplaceAll(text, `"b"`, `"`+target.Name+`"`)

	room.addMessage(Message{
		Name:     sender.Name,
		Text:     text,
		Time:     now.Format("15:04:05"),
		SenderID: input.ID,
		Tone:     kind,
	})
	room.LastActive = now
	mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}
