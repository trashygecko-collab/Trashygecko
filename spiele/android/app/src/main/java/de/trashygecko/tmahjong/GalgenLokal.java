package de.trashygecko.tmahjong;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Locale;
import java.util.Map;
import java.util.Random;

/**
 * Galgenmännchen für einen Raum im lokalen Netz.
 *
 * Regeln und Texte sind ein Port von {@code spiele/connector/galgen.go}: Ein
 * gemeinsames Wort, die Spieler raten streng abwechselnd je einen Buchstaben.
 * Ein Treffer bringt Punkte, ein Fehlgriff kostet ein Leben (für alle). Wer das
 * letzte offene Feld füllt, gewinnt; sind vorher die Leben weg, hat keiner
 * gewonnen. Die Wortliste entspricht der der Seite, damit sich lokal und online
 * gleich anfühlen.
 */
final class GalgenLokal {
    static final int MAX_SPIELER = 6;

    /** Eine Runde wird nach dieser Zeit ohne Zug nicht abgeräumt – nur zur Info. */
    static final long STILLSTAND_MS = 30 * 60 * 1000L;

    private static final class Variante {
        final String name;
        final int fehler;
        final boolean hinweis;
        final int stufe;

        Variante(String name, int fehler, boolean hinweis, int stufe) {
            this.name = name;
            this.fehler = fehler;
            this.hinweis = hinweis;
            this.stufe = stufe;
        }
    }

    private static final class Sitz {
        final String id;
        String name;
        int punkte;
        int treffer;
        boolean bereit;

        Sitz(String id, String name) {
            this.id = id;
            this.name = name;
        }
    }

    private static final Map<String, Variante> VARIANTEN = new LinkedHashMap<>();

    static {
        VARIANTEN.put("klassisch", new Variante("Klassisch", 8, true, 3));
        VARIANTEN.put("streng", new Variante("Streng", 6, true, 3));
        VARIANTEN.put("ohnehilfe", new Variante("Ohne Hinweis", 8, false, 3));
        VARIANTEN.put("leicht", new Variante("Leicht", 10, true, 1));
    }

    /** Wort, Kategorie, Stufe – gleiche Liste wie in galgen.html. */
    private static final String[][] WORTE = {
            {"sonne", "Natur", "1"}, {"wolke", "Natur", "1"}, {"blume", "Natur", "1"},
            {"fluss", "Natur", "1"}, {"wiese", "Natur", "1"},
            {"apfel", "Essen", "1"}, {"brot", "Essen", "1"}, {"kaese", "Essen", "1"},
            {"milch", "Essen", "1"}, {"salat", "Essen", "1"},
            {"tisch", "Haus", "1"}, {"stuhl", "Haus", "1"}, {"fenster", "Haus", "2"},
            {"garten", "Haus", "1"}, {"teppich", "Haus", "2"},
            {"fahrrad", "Technik", "2"}, {"schraube", "Technik", "2"},
            {"batterie", "Technik", "2"}, {"lautsprecher", "Technik", "3"},
            {"gitarre", "Musik", "2"}, {"trommel", "Musik", "2"}, {"klavier", "Musik", "2"},
            {"konzert", "Musik", "2"},
            {"planet", "Weltraum", "2"}, {"rakete", "Weltraum", "2"},
            {"galaxie", "Weltraum", "3"}, {"teleskop", "Weltraum", "3"},
            {"schwimmen", "Sport", "2"}, {"springen", "Sport", "2"},
            {"torwart", "Sport", "2"}, {"stadion", "Sport", "2"},
            {"lehrer", "Beruf", "1"}, {"baecker", "Beruf", "2"}, {"gaertner", "Beruf", "2"},
            {"pilot", "Beruf", "1"},
            {"himmel", "Wetter", "2"}, {"gewitter", "Wetter", "3"},
            {"sonnenschein", "Wetter", "3"}, {"nebel", "Wetter", "1"},
            {"insel", "Reisen", "1"}, {"strand", "Reisen", "2"}, {"koffer", "Reisen", "2"},
            {"urlaub", "Reisen", "2"},
            {"bibliothek", "Orte", "3"}, {"brunnen", "Orte", "2"}, {"bahnhof", "Orte", "2"},
            {"museum", "Orte", "2"},
            {"elefant", "Tiere", "2"}, {"pinguin", "Tiere", "2"},
            {"schmetterling", "Tiere", "3"}, {"eichhoernchen", "Tiere", "3"},
            {"zitrone", "Obst", "2"}, {"banane", "Obst", "2"}, {"kirsche", "Obst", "2"},
            {"erdbeere", "Obst", "2"},
            {"schokolade", "Suess", "3"}, {"marmelade", "Suess", "3"}, {"kuchen", "Suess", "1"},
            {"keks", "Suess", "1"},
            {"fernseher", "Wohnen", "3"}, {"vorhang", "Wohnen", "2"}, {"kissen", "Wohnen", "2"},
            {"schrank", "Wohnen", "2"},
            {"zeitung", "Alltag", "2"}, {"schluessel", "Alltag", "3"}, {"brille", "Alltag", "2"},
            {"tasche", "Alltag", "2"},
    };

    private final Random zufall = new Random();
    private final List<Sitz> spieler = new ArrayList<>();
    private final List<String> anzeige = new ArrayList<>();
    private final List<String> buchstaben = new ArrayList<>();
    private final List<String> letzte = new ArrayList<>();

    private String id = "";
    private String hostId = "";
    private String status = "";          // "", invited, spiel, ende
    private String variante = "klassisch";
    private String varianteName = "Klassisch";
    private String kategorie = "";
    private boolean hinweisZeigen = true;
    private String wort = "";
    private boolean geloest;
    private int fehler;
    private int maxFehler = 8;
    private int zug;
    private int runde = 1;
    private String sieger = "";
    private String note = "";

    boolean istLeer() {
        return status.isEmpty() || spieler.isEmpty();
    }

    /** Führt eine Aktion aus. Rückgabe ist der Fehlertext oder null bei Erfolg. */
    synchronized String aktion(String spielerId, String name, String aktion, JSONObject koerper) {
        String a = aktion == null ? "" : aktion.trim().toLowerCase(Locale.ROOT);
        if (a.isEmpty()) return "Unbekannte Aktion.";
        if ("create".equals(a)) {
            if (!status.isEmpty() && !"ende".equals(status)) {
                return "In diesem Spielraum läuft bereits eine Galgenmännchen-Runde.";
            }
            String v = koerper == null ? "" : koerper.optString("variante", "");
            neuesSpiel(spielerId, name, v);
            note = name + " eröffnet eine Runde Galgenmännchen (" + varianteName + ").";
            return null;
        }
        if ("join".equals(a)) {
            if (!"invited".equals(status)) return "Es gibt keine offene Runde.";
            if (sitzVon(spielerId) >= 0) return "Du sitzt schon am Tisch.";
            if (spieler.size() >= MAX_SPIELER) return "Der Tisch ist voll (höchstens " + MAX_SPIELER + " Spieler).";
            spieler.add(new Sitz(spielerId, name));
            return null;
        }
        if ("ready".equals(a)) {
            if (!"invited".equals(status)) return "Es gibt keine offene Runde.";
            int sitz = sitzVon(spielerId);
            if (sitz < 0) return "Du sitzt nicht an diesem Tisch.";
            spieler.get(sitz).bereit = !spieler.get(sitz).bereit;
            return null;
        }
        if ("start".equals(a)) {
            if (!"invited".equals(status)) return "Es gibt keine offene Runde.";
            if (sitzVon(spielerId) < 0) return "Du sitzt nicht an diesem Tisch.";
            for (Sitz s : spieler) if (!s.bereit) return "Noch nicht alle sind bereit.";
            status = "spiel";
            runde = 1;
            wortWaehlen();
            return null;
        }
        if ("raten".equals(a)) {
            if (!"spiel".equals(status)) return "Gerade läuft keine Runde.";
            int sitz = sitzVon(spielerId);
            if (sitz < 0) return "Du spielst nicht mit.";
            if (sitz != zug) return "Du bist nicht an der Reihe.";
            return raten(sitz, koerper == null ? "" : koerper.optString("buchstabe", ""));
        }
        if ("naechste".equals(a)) {
            if (!"ende".equals(status)) return "Es läuft keine beendete Runde.";
            if (sitzVon(spielerId) < 0) return "Du spielst nicht mit.";
            status = "spiel";
            runde++;
            wortWaehlen();
            return null;
        }
        if ("reset".equals(a)) {
            if (!status.isEmpty() && !hostId.equals(spielerId)) {
                return "Nur der Gastgeber kann die Runde schließen.";
            }
            leeren();
            return null;
        }
        return "Unbekannte Aktion.";
    }

    private String raten(int sitz, String eingabe) {
        String b = eingabe.trim().toLowerCase(Locale.ROOT);
        if (b.codePointCount(0, b.length()) != 1) return "Bitte genau einen Buchstaben raten.";
        if (buchstaben.contains(b)) return "Dieser Buchstabe wurde schon geraten.";
        buchstaben.add(b);
        int treffer = 0;
        for (int i = 0; i < wort.length(); i++) {
            if (String.valueOf(wort.charAt(i)).toLowerCase(Locale.ROOT).equals(b)) {
                anzeige.set(i, String.valueOf(wort.charAt(i)));
                treffer++;
            }
        }
        Sitz s = spieler.get(sitz);
        if (treffer > 0) {
            s.punkte += treffer;
            s.treffer++;
            note = s.name + " trifft „" + b.toUpperCase(Locale.ROOT) + "“ – " + treffer + " Buchstabe(n).";
            letzte.add("✔ " + s.name + ": " + b.toUpperCase(Locale.ROOT));
            if (fertig()) {
                geloest = true;
                beenden(true, s.id);
                note = "Wort gelöst: " + wort.toUpperCase(Locale.ROOT) + ". " + s.name
                        + " hat das letzte Feld gefüllt und gewinnt.";
            } else {
                weiter();
            }
        } else {
            fehler++;
            note = s.name + " rät „" + b.toUpperCase(Locale.ROOT) + "“ – daneben ("
                    + fehler + "/" + maxFehler + ").";
            letzte.add("✖ " + s.name + ": " + b.toUpperCase(Locale.ROOT));
            if (fehler >= maxFehler) {
                beenden(false, "");
                note = "Verloren! Das Wort war „" + wort.toUpperCase(Locale.ROOT)
                        + "“ – diesmal hat es keiner erraten.";
            } else {
                weiter();
            }
        }
        while (letzte.size() > 8) letzte.remove(0);
        return null;
    }

    private void neuesSpiel(String hostName, String name, String varianteWahl) {
        Variante v = VARIANTEN.get(varianteWahl);
        if (v == null) {
            varianteWahl = "klassisch";
            v = VARIANTEN.get(varianteWahl);
        }
        id = Long.toHexString(zufall.nextLong()).toUpperCase(Locale.ROOT);
        hostId = hostName;
        status = "invited";
        variante = varianteWahl;
        varianteName = v.name;
        maxFehler = v.fehler;
        hinweisZeigen = v.hinweis;
        spieler.clear();
        spieler.add(new Sitz(hostName, name));
        anzeige.clear();
        buchstaben.clear();
        letzte.clear();
        wort = "";
        kategorie = "";
        geloest = false;
        fehler = 0;
        zug = 0;
        runde = 1;
        sieger = "";
    }

    private void wortWaehlen() {
        int stufe = VARIANTEN.get(variante).stufe;
        List<String[]> passend = new ArrayList<>();
        for (String[] w : WORTE) if (Integer.parseInt(w[2]) <= stufe) passend.add(w);
        String[] w = passend.get(zufall.nextInt(passend.size()));
        wort = w[0];
        kategorie = w[1];
        anzeige.clear();
        buchstaben.clear();
        letzte.clear();
        fehler = 0;
        geloest = false;
        for (int i = 0; i < wort.length(); i++) anzeige.add("");
        for (Sitz s : spieler) {
            s.punkte = 0;
            s.treffer = 0;
        }
        zug = spieler.isEmpty() ? 0 : zufall.nextInt(spieler.size());
        note = "Das Wort hat " + wort.length() + " Buchstaben.";
    }

    private int sitzVon(String spielerId) {
        for (int i = 0; i < spieler.size(); i++) {
            if (spieler.get(i).id.equals(spielerId)) return i;
        }
        return -1;
    }

    private boolean fertig() {
        for (String a : anzeige) if (a.isEmpty()) return false;
        return true;
    }

    private void weiter() {
        if (!spieler.isEmpty()) zug = (zug + 1) % spieler.size();
    }

    private void beenden(boolean gewonnen, String siegerId) {
        status = "ende";
        sieger = gewonnen ? siegerId : "";
    }

    private void leeren() {
        status = "";
        spieler.clear();
        anzeige.clear();
        buchstaben.clear();
        letzte.clear();
        wort = "";
        kategorie = "";
        note = "";
        sieger = "";
        hostId = "";
        runde = 1;
        fehler = 0;
    }

    /** Zustand als JSON – Feldnamen wie in galgen.go, damit die Seite gleich bleibt. */
    synchronized JSONObject json() throws JSONException {
        JSONObject g = new JSONObject();
        g.put("id", id);
        g.put("hostId", hostId);
        g.put("status", status);
        g.put("variante", variante);
        g.put("varianteName", varianteName);
        g.put("kategorie", kategorie);
        g.put("hinweisZeigen", hinweisZeigen);
        JSONArray anzeigeJson = new JSONArray();
        for (String a : anzeige) anzeigeJson.put(a);
        g.put("anzeige", anzeigeJson);
        g.put("geloest", geloest);
        JSONArray buchstabenJson = new JSONArray();
        for (String b : buchstaben) buchstabenJson.put(b);
        g.put("buchstaben", buchstabenJson);
        g.put("fehler", fehler);
        g.put("maxFehler", maxFehler);
        JSONArray spielerJson = new JSONArray();
        for (Sitz s : spieler) {
            JSONObject eintrag = new JSONObject();
            eintrag.put("id", s.id);
            eintrag.put("name", s.name);
            eintrag.put("punkte", s.punkte);
            eintrag.put("treffer", s.treffer);
            eintrag.put("ready", s.bereit);
            spielerJson.put(eintrag);
        }
        g.put("spieler", spielerJson);
        g.put("zug", zug);
        g.put("runde", runde);
        if (!sieger.isEmpty()) g.put("sieger", sieger);
        if (!note.isEmpty()) g.put("note", note);
        JSONArray letzteJson = new JSONArray();
        for (String z : letzte) letzteJson.put(z);
        g.put("letzte", letzteJson);
        return g;
    }
}
