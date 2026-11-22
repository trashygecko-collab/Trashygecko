package main

// Aus der Karte in spiele/brettspiele/risiko/risiko.html erzeugt (22.11.2026):
// 47 Gebiete, 6 Kontinente, Nachbarschaften. Die Namen sind die einzigen
// Stellen, an denen sich Karte und Server berühren.

type risikoGebietDef struct {
	ID   string
	Name string
	Kont string
	Nach []string
}

type risikoKontDef struct {
	Name  string
	Bonus int
	Farbe string
}

var risikoGebiete = []risikoGebietDef{
	{ID: "alaska", Name: "Alaska", Kont: "nord", Nach: []string{"nwterr", "fernost", "sibirien"}},
	{ID: "nwterr", Name: "Nordwest-Territorien", Kont: "nord", Nach: []string{"alaska", "alberta", "ontario", "groenland"}},
	{ID: "groenland", Name: "Grönland", Kont: "nord", Nach: []string{"nwterr", "ontario", "britannien", "skandinavien"}},
	{ID: "alberta", Name: "Alberta", Kont: "nord", Nach: []string{"nwterr", "ontario", "westusa"}},
	{ID: "ontario", Name: "Ontario", Kont: "nord", Nach: []string{"nwterr", "alberta", "groenland", "westusa", "ostusa"}},
	{ID: "westusa", Name: "West-USA", Kont: "nord", Nach: []string{"alberta", "ontario", "ostusa", "mittelamerika"}},
	{ID: "ostusa", Name: "Ost-USA", Kont: "nord", Nach: []string{"ontario", "westusa", "mittelamerika", "westindien"}},
	{ID: "mittelamerika", Name: "Mittelamerika", Kont: "nord", Nach: []string{"westusa", "ostusa", "westindien", "kolumbien"}},
	{ID: "westindien", Name: "Westindien", Kont: "nord", Nach: []string{"ostusa", "mittelamerika", "kolumbien"}},
	{ID: "kolumbien", Name: "Kolumbien", Kont: "sued", Nach: []string{"mittelamerika", "westindien", "amazonien", "peru"}},
	{ID: "amazonien", Name: "Amazonien", Kont: "sued", Nach: []string{"kolumbien", "brasilien", "peru"}},
	{ID: "brasilien", Name: "Brasilien", Kont: "sued", Nach: []string{"amazonien", "peru", "argentinien"}},
	{ID: "peru", Name: "Peru", Kont: "sued", Nach: []string{"kolumbien", "amazonien", "brasilien", "chile", "argentinien"}},
	{ID: "chile", Name: "Chile", Kont: "sued", Nach: []string{"peru", "argentinien"}},
	{ID: "argentinien", Name: "Argentinien", Kont: "sued", Nach: []string{"brasilien", "peru", "chile"}},
	{ID: "skandinavien", Name: "Skandinavien", Kont: "europa", Nach: []string{"groenland", "britannien", "mitteleuropa", "osteuropa"}},
	{ID: "britannien", Name: "Britannien", Kont: "europa", Nach: []string{"groenland", "skandinavien", "westeuropa"}},
	{ID: "westeuropa", Name: "Westeuropa", Kont: "europa", Nach: []string{"britannien", "iberien", "mitteleuropa"}},
	{ID: "iberien", Name: "Iberien", Kont: "europa", Nach: []string{"westeuropa", "italien", "nordafrika"}},
	{ID: "italien", Name: "Italien", Kont: "europa", Nach: []string{"iberien", "mitteleuropa", "balkan", "nordafrika"}},
	{ID: "mitteleuropa", Name: "Mitteleuropa", Kont: "europa", Nach: []string{"skandinavien", "westeuropa", "italien", "balkan", "osteuropa"}},
	{ID: "balkan", Name: "Balkan", Kont: "europa", Nach: []string{"mitteleuropa", "italien", "osteuropa", "kaukasus"}},
	{ID: "osteuropa", Name: "Osteuropa", Kont: "europa", Nach: []string{"skandinavien", "mitteleuropa", "balkan", "kaukasus"}},
	{ID: "nordafrika", Name: "Nordafrika", Kont: "afrika", Nach: []string{"iberien", "italien", "westafrika", "zentralafrika", "ostafrika"}},
	{ID: "westafrika", Name: "Westafrika", Kont: "afrika", Nach: []string{"nordafrika", "zentralafrika", "suedafrika"}},
	{ID: "zentralafrika", Name: "Zentralafrika", Kont: "afrika", Nach: []string{"nordafrika", "westafrika", "ostafrika", "suedafrika"}},
	{ID: "ostafrika", Name: "Ostafrika", Kont: "afrika", Nach: []string{"nordafrika", "zentralafrika", "suedafrika", "horn", "arabien", "madagaskar"}},
	{ID: "horn", Name: "Horn von Afrika", Kont: "afrika", Nach: []string{"ostafrika", "arabien"}},
	{ID: "suedafrika", Name: "Südafrika", Kont: "afrika", Nach: []string{"westafrika", "zentralafrika", "ostafrika", "madagaskar"}},
	{ID: "madagaskar", Name: "Madagaskar", Kont: "afrika", Nach: []string{"suedafrika", "ostafrika"}},
	{ID: "kaukasus", Name: "Kaukasus", Kont: "asien", Nach: []string{"osteuropa", "balkan", "iran", "zentralasien"}},
	{ID: "iran", Name: "Iran", Kont: "asien", Nach: []string{"kaukasus", "arabien", "indien", "zentralasien"}},
	{ID: "arabien", Name: "Arabien", Kont: "asien", Nach: []string{"iran", "indien", "horn", "ostafrika"}},
	{ID: "indien", Name: "Indien", Kont: "asien", Nach: []string{"iran", "arabien", "china", "suedostasien", "zentralasien"}},
	{ID: "zentralasien", Name: "Zentralasien", Kont: "asien", Nach: []string{"kaukasus", "iran", "indien", "sibirien", "mongolei", "china"}},
	{ID: "sibirien", Name: "Sibirien", Kont: "asien", Nach: []string{"zentralasien", "mongolei", "fernost", "alaska"}},
	{ID: "mongolei", Name: "Mongolei", Kont: "asien", Nach: []string{"zentralasien", "sibirien", "china", "fernost"}},
	{ID: "china", Name: "China", Kont: "asien", Nach: []string{"zentralasien", "mongolei", "indien", "suedostasien", "japan"}},
	{ID: "suedostasien", Name: "Südostasien", Kont: "asien", Nach: []string{"indien", "china", "indonesien"}},
	{ID: "indonesien", Name: "Indonesien", Kont: "asien", Nach: []string{"suedostasien", "neuguinea", "ostaustralien", "westaustralien"}},
	{ID: "japan", Name: "Japan", Kont: "asien", Nach: []string{"china", "fernost"}},
	{ID: "fernost", Name: "Ferner Osten", Kont: "asien", Nach: []string{"sibirien", "mongolei", "japan", "alaska"}},
	{ID: "westaustralien", Name: "Westaustralien", Kont: "ozeanien", Nach: []string{"ostaustralien", "indonesien"}},
	{ID: "ostaustralien", Name: "Ostaustralien", Kont: "ozeanien", Nach: []string{"westaustralien", "indonesien", "neuguinea", "neuseeland"}},
	{ID: "neuguinea", Name: "Neuguinea", Kont: "ozeanien", Nach: []string{"indonesien", "ostaustralien", "pazifikinseln"}},
	{ID: "pazifikinseln", Name: "Pazifikinseln", Kont: "ozeanien", Nach: []string{"neuguinea", "neuseeland"}},
	{ID: "neuseeland", Name: "Neuseeland", Kont: "ozeanien", Nach: []string{"pazifikinseln", "ostaustralien"}},
}

var risikoKontinente = map[string]risikoKontDef{
	"nord":     {Name: "Nordland", Bonus: 5, Farbe: "#4d6f8a"},
	"sued":     {Name: "Südland", Bonus: 2, Farbe: "#7d8a4d"},
	"europa":   {Name: "Europa", Bonus: 5, Farbe: "#8a6a4d"},
	"afrika":   {Name: "Afrika", Bonus: 3, Farbe: "#8a4d5a"},
	"asien":    {Name: "Asien", Bonus: 7, Farbe: "#6a4d8a"},
	"ozeanien": {Name: "Ozeanien", Bonus: 2, Farbe: "#4d8a7a"},
}

// risikoGebietNach ist die Nachbarschaft als Karte, einmal beim Start gebaut.
var risikoGebietNach = func() map[string][]string {
	m := make(map[string][]string, len(risikoGebiete))
	for _, g := range risikoGebiete {
		m[g.ID] = g.Nach
	}
	return m
}()
