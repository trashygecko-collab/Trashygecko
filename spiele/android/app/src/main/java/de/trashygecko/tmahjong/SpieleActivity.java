package de.trashygecko.tmahjong;

import android.app.Activity;
import android.app.AlertDialog;
import android.content.Intent;
import android.content.res.Configuration;
import android.graphics.Bitmap;
import android.graphics.BitmapFactory;
import android.graphics.Color;
import android.graphics.Typeface;
import android.graphics.drawable.GradientDrawable;
import android.os.Bundle;
import android.text.Editable;
import android.text.TextWatcher;
import android.util.Log;
import android.util.TypedValue;
import android.view.Gravity;
import android.view.View;
import android.view.ViewTreeObserver;
import android.widget.EditText;
import android.widget.ImageView;
import android.widget.LinearLayout;
import android.widget.ScrollView;
import android.widget.TextView;

import org.json.JSONArray;
import org.json.JSONObject;

import java.io.ByteArrayOutputStream;
import java.io.InputStream;
import java.text.Normalizer;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.HashSet;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Locale;
import java.util.Map;
import java.util.Set;

/**
 * Die Spielhalle der App – nativ, aus `assets/katalog.json`.
 *
 * Die Datei entsteht beim Bauen aus `spiele/spiele.html` (Aufgabe
 * `katalogBauen`, D-15); hier wird sie nur gelesen. Die Startansicht zeigt
 * Highlights und Kategorie-Kacheln, eine Kategorie zeigt nur ihre Spiele, und
 * die Suche findet weiterhin alle Spiele unabhängig von der Kategorie.
 * Ein Tipp auf ein Spiel öffnet {@link GameActivity}, bei Online-Spielen
 * vorher die Frage „Lokal / Online“ – wie bisher im Katalog.
 *
 * Suche und Merkmalsfilter verhalten sich wie im HTML-Katalog: Die Suche
 * prüft Name, Suchwörter, Kurztext und Kategoriename, jedes gewählte Merkmal
 * muss zusätzlich zutreffen, und angeboten werden nur Merkmale, die im
 * Katalog wirklich vorkommen. Ein Spiel steht genau einmal im Katalog.
 */
public final class SpieleActivity extends Activity {
    private static final String TAG = "TrashyGeckoSpiele";

    /** Schlüssel für den gesicherten Zustand (Prozess-Tod, „Aktivitäten nicht behalten“). */
    private static final String ZUSTAND_SUCHE = "suche";
    private static final String ZUSTAND_MERKMALE = "merkmale";
    private static final String ZUSTAND_SCROLL = "scroll";
    private static final String ZUSTAND_ANSICHT = "ansicht";
    private static final String ZUSTAND_KATEGORIE = "kategorie";

    private static final String ANSICHT_START = "start";
    private static final String ANSICHT_KATEGORIE = "kategorie";
    /** Sammelansicht für Spiele, deren Kategorie im Katalog fehlt (alte Kataloge). */
    private static final String REST = "__rest";

    /** Nur der Chip-Füllton ist hier eigen; Fläche, Text und Gold kommen aus {@link Ansicht}. */
    private static final int CHIP = 0xFF1A3A30;

    /** Kurzform für die Zeile unter dem Spielnamen. */
    private static final Map<String, String> MERKMAL_TEXT = new LinkedHashMap<>();
    /** Beschriftung der Filterknöpfe – Wortlaut wie im HTML-Katalog. */
    private static final Map<String, String> MERKMAL_FILTER = new LinkedHashMap<>();

    static {
        MERKMAL_TEXT.put("solo", "Einzelspieler");
        MERKMAL_TEXT.put("lokal", "Lokal");
        MERKMAL_TEXT.put("zwei", "Zu zweit");
        MERKMAL_TEXT.put("multi", "Mehrspieler");
        MERKMAL_TEXT.put("online", "Online");
        MERKMAL_TEXT.put("computer", "Gegen Bots");
        MERKMAL_TEXT.put("highscore", "Rekord");
        MERKMAL_TEXT.put("controller", "Pad");

        MERKMAL_FILTER.put("solo", "Einzelspieler");
        MERKMAL_FILTER.put("multi", "Mehrspieler");
        MERKMAL_FILTER.put("lokal", "Lokal am Gerät");
        MERKMAL_FILTER.put("zwei", "Zu zweit");
        MERKMAL_FILTER.put("online", "Online");
        MERKMAL_FILTER.put("computer", "Gegen Computer");
        MERKMAL_FILTER.put("highscore", "Highscore");
        MERKMAL_FILTER.put("controller", "Controller");
    }

    /** Eine Kategorie aus katalog.json. */
    private static final class Kategorie {
        String id;
        String name;
        String text = "";
        /** Zeichen der Web-Halle (SVG) – leer, wenn der Katalog keins nennt. */
        String ikon = "";
        int akzent = Ansicht.GOLD;
    }

    /** Ein Eintrag aus katalog.json. */
    private static final class Spiel {
        String name;
        String pfad;
        String kategorie;
        String text = "";
        String suche = "";
        /** Name des Spielzeichens aus `symbole` (z. B. `krone`). */
        String symbol = "";
        /** Name, Suchwörter, Text und Kategorie in Kleinbuchstaben (mit und ohne Umlaute). */
        String suchtext = "";
        List<String> merkmale = new ArrayList<>();
        boolean online;
    }

    /** Ein hervorgehobenes Spiel samt Vorschaubild. */
    private static final class Highlight {
        Spiel spiel;
        String bild;
    }

    private final List<Spiel> spiele = new ArrayList<>();
    private final List<Kategorie> kategorien = new ArrayList<>();
    private final Map<String, Kategorie> kategorieNach = new LinkedHashMap<>();
    private final List<Highlight> highlights = new ArrayList<>();
    private final List<String> aktiveMerkmale = new ArrayList<>();
    /** Einmal gebaute Zeilenansichten – ein Filterwechsel hängt sie nur um. */
    private final Map<Spiel, View> zeilen = new HashMap<>();
    private final Map<String, TextView> chips = new LinkedHashMap<>();
    /** Einmal geladene Vorschaubilder; klein genug für die Kacheln. */
    private final Map<String, Bitmap> bilder = new HashMap<>();
    /** Spielzeichen aus `katalog.json` (Name → SVG-Inhalt), Quelle ist spiele.html. */
    private final Map<String, String> symbole = new HashMap<>();

    private LinearLayout liste;
    private LinearLayout chipLeiste;
    private TextView titel;
    private TextView zaehler;
    private TextView zurueckChip;
    private EditText suchfeld;
    private ScrollView rollen;
    private String filter = "";
    /** Beim Wiederherstellen zeichnet noch nicht der Textwächter – das tut `stelleWieder`. */
    private boolean stelleWiederLaeuft;
    /** Ansicht: Startseite mit Kacheln oder eine geöffnete Kategorie. */
    private String ansicht = ANSICHT_START;
    private String offeneKategorie;
    private boolean katalogFehler;

    @Override protected void onCreate(Bundle zustand) {
        super.onCreate(zustand);
        WebBasis.vollbild(this);
        setContentView(baueAnsicht());
        ladeKatalog();
        stelleWieder(zustand);
    }

    /**
     * Wird die Activity vom System beendet (Prozess-Tod im Hintergrund, „Aktivitäten nicht
     * behalten“), entsteht sie neu. Suchtext, Merkmalswahl, Scrollposition und die geöffnete
     * Ansicht werden deshalb gesichert – die Drehung allein braucht das nicht
     * (`configChanges`).
     */
    @Override protected void onSaveInstanceState(Bundle zustand) {
        super.onSaveInstanceState(zustand);
        zustand.putString(ZUSTAND_SUCHE, suchfeld.getText().toString());
        zustand.putStringArrayList(ZUSTAND_MERKMALE, new ArrayList<>(aktiveMerkmale));
        zustand.putInt(ZUSTAND_SCROLL, rollen.getScrollY());
        zustand.putString(ZUSTAND_ANSICHT, ansicht);
        zustand.putString(ZUSTAND_KATEGORIE, offeneKategorie);
    }

    /**
     * Holt den gesicherten Zustand zurück und zeichnet die Liste danach genau einmal – ohne
     * Sicherung ist das der erste Aufbau. Unbekannte Merkmale oder Kategorien aus einem
     * älteren Bundle entfallen, weil `baueChips` nur vorhandene anbietet.
     */
    private void stelleWieder(Bundle zustand) {
        int scroll = 0;
        if (zustand != null) {
            ArrayList<String> merkmale = zustand.getStringArrayList(ZUSTAND_MERKMALE);
            if (merkmale != null) {
                for (String merkmal : merkmale) {
                    if (MERKMAL_FILTER.containsKey(merkmal)) aktiveMerkmale.add(merkmal);
                }
            }

            String gespeicherteAnsicht = zustand.getString(ZUSTAND_ANSICHT, ANSICHT_START);
            String gespeicherteKategorie = zustand.getString(ZUSTAND_KATEGORIE, null);
            if (ANSICHT_KATEGORIE.equals(gespeicherteAnsicht)
                    && gespeicherteKategorie != null
                    && kategorieNach.containsKey(gespeicherteKategorie)) {
                ansicht = ANSICHT_KATEGORIE;
                offeneKategorie = gespeicherteKategorie;
            }

            String text = zustand.getString(ZUSTAND_SUCHE, "");
            if (text == null) text = "";
            if (!text.isEmpty()) {
                stelleWiederLaeuft = true;
                try {
                    suchfeld.setText(text);   // setzt `filter`; gezeichnet wird unten
                } finally {
                    stelleWiederLaeuft = false;
                }
                suchfeld.setSelection(text.length());
            }
            scroll = zustand.getInt(ZUSTAND_SCROLL, 0);
            Log.i(TAG, "Zustand wiederhergestellt: „" + text + "“, "
                    + aktiveMerkmale.size() + " Merkmale, Ansicht " + ansicht
                    + (offeneKategorie == null ? "" : " (" + offeneKategorie + ")")
                    + ", Scroll " + scroll);
        }

        chipAussehen();
        zeichne();

        // Erst nach dem Messen rollen – vor dem ersten Layout wäre die Inhaltshöhe 0.
        final int ziel = scroll;
        if (ziel > 0) {
            rollen.getViewTreeObserver().addOnGlobalLayoutListener(
                    new ViewTreeObserver.OnGlobalLayoutListener() {
                        @Override public void onGlobalLayout() {
                            ViewTreeObserver beobachter = rollen.getViewTreeObserver();
                            if (!beobachter.isAlive()) return;
                            beobachter.removeOnGlobalLayoutListener(this);
                            rollen.scrollTo(0, ziel);
                        }
                    });
        }
    }

    /**
     * Die Drehung baut die Activity nicht neu (`configChanges` im Manifest), also bleiben
     * Suchtext, gewählte Merkmale und Scrollposition von selbst erhalten – anders als beim
     * Prozess-Tod, den {@link #onSaveInstanceState} abdeckt. Die Chip-Leiste und das
     * Kategorie-Raster wurden aber für die alte Breite gebaut – im Querformat passen mehr
     * Spalten. Deshalb hier mit der neuen Breite neu aufbauen; die Auswahl steht in
     * {@link #aktiveMerkmale} und bleibt bestehen.
     */
    @Override public void onConfigurationChanged(Configuration neu) {
        super.onConfigurationChanged(neu);
        WebBasis.vollbild(this);
        if (chipLeiste == null || spiele.isEmpty()) return;
        int scroll = rollen == null ? 0 : rollen.getScrollY();
        baueChips();
        zeichne();
        if (rollen != null) rollen.post(() -> rollen.scrollTo(0, scroll));
    }

    /**
     * Zurück führt erst aus Suche und Filtern, dann aus einer Kategorie in die Übersicht
     * und erst zuletzt aus der Spieleliste heraus. Die Kopfzeile nutzt denselben Weg,
     * damit Bildschirmtaste und Pfeil sich gleich verhalten.
     */
    @Override public void onBackPressed() {
        if (!filter.isEmpty() || !aktiveMerkmale.isEmpty()) {
            zuruecksetzen();
            return;
        }
        if (ANSICHT_KATEGORIE.equals(ansicht)) {
            ansicht = ANSICHT_START;
            offeneKategorie = null;
            zeichne();
            return;
        }
        super.onBackPressed();
    }

    private void zurueck() {
        if (!filter.isEmpty() || !aktiveMerkmale.isEmpty()) {
            zuruecksetzen();
        } else if (ANSICHT_KATEGORIE.equals(ansicht)) {
            ansicht = ANSICHT_START;
            offeneKategorie = null;
            zeichne();
        } else {
            finish();
        }
    }

    // ---------------------------------------------------------------- Aufbau

    private View baueAnsicht() {
        LinearLayout wurzel = new LinearLayout(this);
        wurzel.setOrientation(LinearLayout.VERTICAL);
        wurzel.setBackgroundColor(Ansicht.BG);
        int rand = Ansicht.dp(this, 16);
        wurzel.setPadding(rand, Ansicht.dp(this, 12), rand, 0);

        LinearLayout kopf = new LinearLayout(this);
        kopf.setOrientation(LinearLayout.HORIZONTAL);
        kopf.setGravity(Gravity.CENTER_VERTICAL);

        TextView zurueck = new TextView(this);
        zurueck.setText("‹");
        zurueck.setTextColor(Ansicht.GOLD);
        zurueck.setTextSize(TypedValue.COMPLEX_UNIT_SP, 26);
        zurueck.setPadding(0, 0, Ansicht.dp(this, 14), Ansicht.dp(this, 14));
        zurueck.setContentDescription("Zurück");
        zurueck.setOnClickListener(v -> zurueck());
        kopf.addView(zurueck);

        titel = new TextView(this);
        titel.setText("Spielhalle");
        titel.setTextColor(Ansicht.GOLD);
        titel.setTypeface(Typeface.DEFAULT_BOLD);
        titel.setTextSize(TypedValue.COMPLEX_UNIT_SP, 24);
        LinearLayout.LayoutParams titelPlatz = new LinearLayout.LayoutParams(0,
                LinearLayout.LayoutParams.WRAP_CONTENT, 1f);
        kopf.addView(titel, titelPlatz);

        zaehler = new TextView(this);
        zaehler.setTextColor(Ansicht.MUTED);
        zaehler.setTextSize(TypedValue.COMPLEX_UNIT_SP, 13);
        kopf.addView(zaehler);
        wurzel.addView(kopf);

        suchfeld = new EditText(this);
        suchfeld.setHint("Spiel suchen – z. B. Kniffel");
        suchfeld.setHintTextColor(Ansicht.MUTED);
        suchfeld.setTextColor(Ansicht.TEXT);
        suchfeld.setTextSize(TypedValue.COMPLEX_UNIT_SP, 15);
        suchfeld.setSingleLine(true);
        suchfeld.setBackground(Ansicht.flaeche(Ansicht.KARTE, Ansicht.dp(this, 12)));
        suchfeld.setPadding(Ansicht.dp(this, 14), Ansicht.dp(this, 10), Ansicht.dp(this, 14), Ansicht.dp(this, 10));
        LinearLayout.LayoutParams suchPlatz = new LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT, LinearLayout.LayoutParams.WRAP_CONTENT);
        suchPlatz.topMargin = Ansicht.dp(this, 12);
        wurzel.addView(suchfeld, suchPlatz);
        suchfeld.addTextChangedListener(new TextWatcher() {
            @Override public void beforeTextChanged(CharSequence s, int a, int b, int c) { }
            @Override public void onTextChanged(CharSequence s, int a, int b, int c) { }
            @Override public void afterTextChanged(Editable s) {
                filter = s.toString().trim().toLowerCase(Locale.ROOT);
                if (!stelleWiederLaeuft) zeichne();
            }
        });

        // Merkmalsfilter: bricht wie im HTML-Katalog um, gefüllt nach dem Laden
        // des Katalogs (nur Merkmale, die wirklich vorkommen).
        chipLeiste = new LinearLayout(this);
        chipLeiste.setOrientation(LinearLayout.VERTICAL);
        LinearLayout.LayoutParams chipPlatz = new LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT, LinearLayout.LayoutParams.WRAP_CONTENT);
        chipPlatz.topMargin = Ansicht.dp(this, 10);
        wurzel.addView(chipLeiste, chipPlatz);

        rollen = new ScrollView(this);
        rollen.setClipToPadding(false);
        LinearLayout.LayoutParams rollPlatz = new LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT, 0, 1f);
        rollPlatz.topMargin = Ansicht.dp(this, 4);
        wurzel.addView(rollen, rollPlatz);

        liste = new LinearLayout(this);
        liste.setOrientation(LinearLayout.VERTICAL);
        liste.setPadding(0, 0, 0, Ansicht.dp(this, 24));
        rollen.addView(liste, new LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT, LinearLayout.LayoutParams.WRAP_CONTENT));
        return wurzel;
    }

    // ------------------------------------------------------------- Katalog

    private void ladeKatalog() {
        try (InputStream strom = getAssets().open("katalog.json")) {
            ByteArrayOutputStream puffer = new ByteArrayOutputStream();
            byte[] block = new byte[8192];
            int gelesen;
            while ((gelesen = strom.read(block)) > 0) {
                puffer.write(block, 0, gelesen);
            }
            JSONObject daten = new JSONObject(puffer.toString("UTF-8"));

            JSONArray kat = daten.optJSONArray("kategorien");
            if (kat != null) {
                for (int i = 0; i < kat.length(); i++) {
                    JSONObject eintrag = kat.optJSONObject(i);
                    if (eintrag == null) continue;
                    Kategorie kategorie = new Kategorie();
                    kategorie.id = eintrag.optString("id");
                    if (kategorie.id.isEmpty()) continue;
                    kategorie.name = eintrag.optString("name", kategorie.id);
                    kategorie.text = eintrag.optString("text", "");
                    kategorie.ikon = eintrag.optString("ikon", "");
                    kategorie.akzent = farbe(eintrag.optString("akzent", ""), Ansicht.GOLD);
                    kategorien.add(kategorie);
                    kategorieNach.put(kategorie.id, kategorie);
                }
            }

            JSONObject zeichen = daten.optJSONObject("symbole");
            if (zeichen != null) {
                for (java.util.Iterator<String> namen = zeichen.keys(); namen.hasNext(); ) {
                    String name = namen.next();
                    symbole.put(name, zeichen.optString(name, ""));
                }
            }

            JSONArray alle = daten.optJSONArray("spiele");
            if (alle != null) {
                for (int i = 0; i < alle.length(); i++) {
                    JSONObject eintrag = alle.optJSONObject(i);
                    if (eintrag == null) continue;
                    Spiel spiel = new Spiel();
                    spiel.name = eintrag.optString("name");
                    spiel.pfad = eintrag.optString("pfad");
                    spiel.kategorie = eintrag.optString("kategorie");
                    spiel.text = eintrag.optString("text", "");
                    spiel.suche = eintrag.optString("suche", "");
                    spiel.symbol = eintrag.optString("symbol", "");
                    JSONArray merk = eintrag.optJSONArray("merkmale");
                    if (merk != null) {
                        for (int m = 0; m < merk.length(); m++) {
                            spiel.merkmale.add(merk.optString(m));
                        }
                    }
                    spiel.online = spiel.merkmale.contains("online");
                    Kategorie kategorie = kategorieNach.get(spiel.kategorie);
                    String roh = spiel.name + " " + spiel.suche + " " + spiel.text + " "
                            + (kategorie == null ? "" : kategorie.name);
                    spiel.suchtext = (roh + " " + ohneUmlaute(roh)).toLowerCase(Locale.ROOT);
                    if (!spiel.name.isEmpty() && WebBasis.istSichereSeite(spiel.pfad)) {
                        spiele.add(spiel);
                    }
                }
            }

            JSONArray hervorgehoben = daten.optJSONArray("highlights");
            if (hervorgehoben != null) {
                for (int i = 0; i < hervorgehoben.length(); i++) {
                    JSONObject eintrag = hervorgehoben.optJSONObject(i);
                    if (eintrag == null) continue;
                    String pfad = eintrag.optString("pfad");
                    Spiel spiel = spielNach(pfad);
                    if (spiel == null) continue;
                    Highlight highlight = new Highlight();
                    highlight.spiel = spiel;
                    highlight.bild = eintrag.optString("bild", "");
                    highlights.add(highlight);
                }
            }

            Log.i(TAG, "Katalog: " + spiele.size() + " Spiele, " + kategorien.size()
                    + " Kategorien, " + highlights.size() + " Highlights");
            baueChips();
        } catch (Exception fehler) {
            Log.e(TAG, "Katalog nicht lesbar", fehler);
            katalogFehler = true;
        }
    }

    /**
     * Highlights nennen den Pfad relativ zu `spiele/` (`brettspiele/…`), der Katalog
     * dagegen mit Präfix (`spiele/brettspiele/…`) – beide Formen treffen dasselbe Spiel.
     */
    private Spiel spielNach(String pfad) {
        String relativ = pfad.startsWith("spiele/") ? pfad.substring("spiele/".length()) : pfad;
        for (Spiel spiel : spiele) {
            if (spiel.pfad.equals(pfad) || spiel.pfad.equals("spiele/" + relativ)) return spiel;
        }
        Log.w(TAG, "Highlight ohne Spiel: " + pfad);
        return null;
    }

    private static int farbe(String wert, int ersatz) {
        if (wert == null || wert.isEmpty()) return ersatz;
        try {
            return Color.parseColor(wert);
        } catch (IllegalArgumentException fehler) {
            return ersatz;
        }
    }

    private static String ohneUmlaute(String text) {
        String zerlegt = Normalizer.normalize(text, Normalizer.Form.NFD).replaceAll("\\p{M}+", "");
        return zerlegt.replace("ß", "ss");
    }

    private void zeigeHinweis(String text) {
        liste.removeAllViews();
        TextView hinweis = new TextView(this);
        hinweis.setText(text);
        hinweis.setTextColor(Ansicht.TEXT);
        hinweis.setTextSize(TypedValue.COMPLEX_UNIT_SP, 15);
        hinweis.setPadding(0, Ansicht.dp(this, 12), 0, 0);
        liste.addView(hinweis);
    }

    // -------------------------------------------------------------- Filter

    /** Baut die Filterleiste; leere Filter werden wie im HTML-Katalog nicht angeboten. */
    private void baueChips() {
        chipLeiste.removeAllViews();
        chips.clear();

        // Immer aus der aktuellen Anzeigebreite: beim Umbau nach einer Drehung ist die
        // alte Breite der Ansicht noch nicht neu gemessen.
        int platz = getResources().getDisplayMetrics().widthPixels - Ansicht.dp(this, 32);
        LinearLayout zeile = chipZeile();
        chipLeiste.addView(zeile);
        int belegt = 0;

        zurueckChip = chip("Zurücksetzen");
        zurueckChip.setOnClickListener(v -> zuruecksetzen());
        belegt = haenge(zeile, zurueckChip, belegt, platz);

        Set<String> genutzt = new HashSet<>();
        for (Spiel spiel : spiele) genutzt.addAll(spiel.merkmale);

        for (String merkmal : MERKMAL_FILTER.keySet()) {
            if (!genutzt.contains(merkmal)) continue;
            final String wahl = merkmal;
            TextView chip = chip(MERKMAL_FILTER.get(merkmal));
            chip.setContentDescription("Filter " + MERKMAL_FILTER.get(merkmal));
            chip.setOnClickListener(v -> {
                if (aktiveMerkmale.contains(wahl)) {
                    aktiveMerkmale.remove(wahl);
                } else {
                    aktiveMerkmale.add(wahl);
                }
                chipAussehen();
                zeichne();
            });
            chips.put(merkmal, chip);
            chip.measure(View.MeasureSpec.makeMeasureSpec(0, View.MeasureSpec.UNSPECIFIED),
                    View.MeasureSpec.makeMeasureSpec(0, View.MeasureSpec.UNSPECIFIED));
            int breite = chip.getMeasuredWidth() + Ansicht.dp(this, 8);
            if (belegt > 0 && belegt + breite > platz) {
                zeile = chipZeile();
                chipLeiste.addView(zeile);
                belegt = 0;
            }
            belegt = haenge(zeile, chip, belegt, platz);
        }
        chipAussehen();
    }

    private LinearLayout chipZeile() {
        LinearLayout zeile = new LinearLayout(this);
        zeile.setOrientation(LinearLayout.HORIZONTAL);
        zeile.setGravity(Gravity.CENTER_VERTICAL);
        LinearLayout.LayoutParams platz = new LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT, LinearLayout.LayoutParams.WRAP_CONTENT);
        platz.bottomMargin = Ansicht.dp(this, 8);
        zeile.setLayoutParams(platz);
        return zeile;
    }

    /** Misst „Zurücksetzen“ mit und hängt den Chip in die laufende Zeile. */
    private int haenge(LinearLayout zeile, TextView chip, int belegt, int platz) {
        if (chip.getMeasuredWidth() == 0) {
            chip.measure(View.MeasureSpec.makeMeasureSpec(0, View.MeasureSpec.UNSPECIFIED),
                    View.MeasureSpec.makeMeasureSpec(0, View.MeasureSpec.UNSPECIFIED));
        }
        zeile.addView(chip);
        return belegt + chip.getMeasuredWidth() + Ansicht.dp(this, 8);
    }

    private TextView chip(String beschriftung) {
        TextView chip = new TextView(this);
        chip.setText(beschriftung);
        chip.setTextSize(TypedValue.COMPLEX_UNIT_SP, 13);
        chip.setGravity(Gravity.CENTER);
        chip.setMinHeight(Ansicht.dp(this, 38));
        chip.setPadding(Ansicht.dp(this, 14), Ansicht.dp(this, 8), Ansicht.dp(this, 14), Ansicht.dp(this, 8));
        chip.setClickable(true);
        chip.setFocusable(true);
        LinearLayout.LayoutParams platz = new LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.WRAP_CONTENT, LinearLayout.LayoutParams.WRAP_CONTENT);
        platz.rightMargin = Ansicht.dp(this, 8);
        chip.setLayoutParams(platz);
        return chip;
    }

    private void chipAussehen() {
        for (Map.Entry<String, TextView> eintrag : chips.entrySet()) {
            aussehen(eintrag.getValue(), aktiveMerkmale.contains(eintrag.getKey()));
        }
        if (zurueckChip != null) {
            aussehen(zurueckChip, false);
            zurueckChip.setEnabled(istGefiltert());
            zurueckChip.setAlpha(istGefiltert() ? 1f : 0.55f);
        }
    }

    private void aussehen(TextView chip, boolean an) {
        GradientDrawable form = new GradientDrawable();
        form.setCornerRadius(Ansicht.dp(this, 19));
        form.setColor(an ? Ansicht.GOLD : CHIP);
        form.setStroke(Math.max(1, Ansicht.dp(this, 1)), an ? Ansicht.GOLD : Ansicht.RAND);
        chip.setBackground(form);
        chip.setTextColor(an ? Ansicht.BG : Ansicht.TEXT);
        chip.setSelected(an);
    }

    private void zuruecksetzen() {
        aktiveMerkmale.clear();
        boolean hatteText = suchfeld.getText().length() > 0;
        if (hatteText) {
            suchfeld.setText("");     // afterTextChanged zeichnet neu
        }
        chipAussehen();
        if (!hatteText) zeichne();
    }

    private boolean istGefiltert() {
        return !filter.isEmpty() || !aktiveMerkmale.isEmpty();
    }

    // ------------------------------------------------------------ Zeichnen

    private void zeichne() {
        if (liste == null) return;
        long beginn = System.nanoTime();
        liste.removeAllViews();

        if (katalogFehler) {
            titel.setText("Spielhalle");
            zaehler.setText("");
            zeigeHinweis("Der Spielekatalog fehlt im App-Paket. Bitte die App neu installieren.");
            return;
        }

        Kategorie offen = offeneKategorie == null ? null : kategorieNach.get(offeneKategorie);
        boolean trefferAnsicht = istGefiltert();

        if (trefferAnsicht) {
            titel.setText(!filter.isEmpty() ? "Suche" : (offen == null ? "Gefiltert" : offen.name));
            // Die Suche findet alle Spiele, ein Merkmalsfilter bleibt in der offenen Kategorie.
            int sichtbar = zeichneTreffer(filter.isEmpty() && ANSICHT_KATEGORIE.equals(ansicht) && offen != null
                    ? offen : null);
            zaehler.setText(sichtbar + (filter.isEmpty() ? " Spiele" : " Treffer"));
        } else if (offen != null) {
            titel.setText(offen.name);
            int sichtbar = zeichneKategorie(offen);
            zaehler.setText(sichtbar + " Spiele");
        } else if (REST.equals(offeneKategorie)) {
            titel.setText("Weitere Spiele");
            int sichtbar = zeichneRest();
            zaehler.setText(sichtbar + " Spiele");
        } else {
            titel.setText("Spielhalle");
            zeichneStart();
            zaehler.setText(spiele.size() + " Spiele");
        }

        rollen.scrollTo(0, 0);
        Log.i(TAG, "Liste: " + liste.getChildCount() + " Ansichten in "
                + ((System.nanoTime() - beginn) / 1000000L) + " ms");
    }

    /** Startansicht: Highlights und Kategorie-Kacheln, keine lange Spieleliste. */
    private void zeichneStart() {
        if (!highlights.isEmpty()) {
            liste.addView(ueberschrift("Highlights"));
            List<View> karten = new ArrayList<>();
            for (Highlight highlight : highlights) karten.add(highlightKarte(highlight));
            haengeRaster(karten, spaltenHighlights());
        }

        liste.addView(ueberschrift("Kategorien"));
        List<View> kacheln = new ArrayList<>();
        for (Kategorie kategorie : kategorien) {
            kacheln.add(kategorieKachel(kategorie, anzahlIn(kategorie.id)));
        }
        // Spiele ohne bekannte Kategorie sollen nicht verschwinden.
        int rest = restAnzahl();
        if (rest > 0) kacheln.add(restKachel(rest));
        haengeRaster(kacheln, spaltenKategorien());
    }

    /** Eine geöffnete Kategorie: Beschreibung und ausschließlich ihre Spiele. */
    private int zeichneKategorie(Kategorie kategorie) {
        if (!kategorie.text.isEmpty()) {
            TextView beschreibung = Ansicht.text(this, kategorie.text, 13, Ansicht.MUTED, false);
            beschreibung.setLineSpacing(0, 1.3f);
            beschreibung.setPadding(Ansicht.dp(this, 4), 0, Ansicht.dp(this, 4), Ansicht.dp(this, 10));
            liste.addView(beschreibung);
        }

        int sichtbar = 0;
        for (Spiel spiel : spiele) {
            if (!kategorie.id.equals(spiel.kategorie) || !passt(spiel)) continue;
            liste.addView(zeile(spiel));
            sichtbar++;
        }
        if (sichtbar == 0) {
            zeigeHinweis("Kein Spiel dieser Kategorie passt zu Suche und Filtern.");
        }

        // Direkt weiter zur nächsten Kategorie – wie in der Web-Halle.
        List<View> andere = new ArrayList<>();
        for (Kategorie weitere : kategorien) {
            if (weitere.id.equals(kategorie.id)) continue;
            int anzahl = anzahlIn(weitere.id);
            if (anzahl == 0) continue;
            andere.add(kategorieKachel(weitere, anzahl));
        }
        if (!andere.isEmpty()) {
            liste.addView(ueberschrift("Andere Kategorien"));
            haengeRaster(andere, spaltenKategorien());
        }
        return sichtbar;
    }

    /** Spiele ohne bekannte Kategorie – nur für ältere Kataloge. */
    private int zeichneRest() {
        TextView beschreibung = Ansicht.text(this, "Spiele ohne bekannte Kategorie.", 13, Ansicht.MUTED, false);
        beschreibung.setPadding(Ansicht.dp(this, 4), 0, Ansicht.dp(this, 4), Ansicht.dp(this, 10));
        liste.addView(beschreibung);
        int sichtbar = 0;
        for (Spiel spiel : spiele) {
            if (kategorieNach.containsKey(spiel.kategorie) || !passt(spiel)) continue;
            liste.addView(zeile(spiel));
            sichtbar++;
        }
        if (sichtbar == 0) zeigeHinweis("Kein Spiel passt zu Suche und Filtern.");
        return sichtbar;
    }

    /**
     * Trefferliste (Suche oder Merkmalsfilter), nach Kategorien gruppiert. Ist eine
     * Kategorie geöffnet, bleibt die Trefferliste auf sie beschränkt; die Suche findet
     * dagegen weiterhin alle Spiele.
     */
    private int zeichneTreffer(Kategorie nur) {
        int sichtbar = 0;
        for (Kategorie kategorie : kategorien) {
            if (nur != null && !nur.id.equals(kategorie.id)) continue;
            List<Spiel> treffer = new ArrayList<>();
            for (Spiel spiel : spiele) {
                if (!kategorie.id.equals(spiel.kategorie) || !passt(spiel)) continue;
                treffer.add(spiel);
            }
            if (treffer.isEmpty()) continue;
            liste.addView(ueberschrift(kategorie.name + "  ·  " + treffer.size()));
            for (Spiel spiel : treffer) {
                liste.addView(zeile(spiel));
                sichtbar++;
            }
        }

        if (nur == null) {
            List<Spiel> rest = new ArrayList<>();
            for (Spiel spiel : spiele) {
                if (kategorieNach.containsKey(spiel.kategorie) || !passt(spiel)) continue;
                rest.add(spiel);
            }
            if (!rest.isEmpty()) {
                liste.addView(ueberschrift("Weitere  ·  " + rest.size()));
                for (Spiel spiel : rest) {
                    liste.addView(zeile(spiel));
                    sichtbar++;
                }
            }
        }

        if (sichtbar == 0) {
            zeigeHinweis("Kein Spiel passt zu Suche und Filtern. „Zurücksetzen“ zeigt wieder alle.");
        }
        return sichtbar;
    }

    private int anzahlIn(String kategorieId) {
        int anzahl = 0;
        for (Spiel spiel : spiele) {
            if (kategorieId.equals(spiel.kategorie)) anzahl++;
        }
        return anzahl;
    }

    private int restAnzahl() {
        int anzahl = 0;
        for (Spiel spiel : spiele) {
            if (!kategorieNach.containsKey(spiel.kategorie)) anzahl++;
        }
        return anzahl;
    }

    private View ueberschrift(String text) {
        TextView kopf = new TextView(this);
        kopf.setText(text);
        kopf.setTextColor(Ansicht.GOLD);
        kopf.setTypeface(Typeface.DEFAULT_BOLD);
        kopf.setTextSize(TypedValue.COMPLEX_UNIT_SP, 15);
        kopf.setPadding(Ansicht.dp(this, 4), Ansicht.dp(this, 16), Ansicht.dp(this, 4), Ansicht.dp(this, 8));
        return kopf;
    }

    private int spaltenKategorien() {
        int breite = getResources().getConfiguration().screenWidthDp;
        if (breite >= 900) return 6;
        if (breite >= 600) return 3;
        return 2;
    }

    private int spaltenHighlights() {
        return getResources().getConfiguration().screenWidthDp >= 900 ? 4 : 2;
    }

    /**
     * Legt Karten in Zeilen zu je `spalten` Stück. Die Karten einer Zeile werden vorher
     * auf die Spaltenbreite gemessen und bekommen die Höhe der höchsten Karte – sonst
     * stehen Kacheln mit zweizeiliger Beschreibung tiefer als ihre Nachbarn. Die letzte
     * Zeile wird mit leeren Flächen aufgefüllt, damit die Karten ihre Breite behalten.
     */
    private void haengeRaster(List<View> teile, int spalten) {
        if (teile.isEmpty()) return;
        int breite = getResources().getDisplayMetrics().widthPixels - Ansicht.dp(this, 32);
        int abstand = Ansicht.dp(this, 8);
        int spaltenBreite = Math.max(1, (breite - abstand * (spalten - 1)) / spalten);

        for (int start = 0; start < teile.size(); start += spalten) {
            int ende = Math.min(start + spalten, teile.size());
            List<View> reihe = teile.subList(start, ende);
            int hoehe = 0;
            for (View teil : reihe) {
                teil.measure(View.MeasureSpec.makeMeasureSpec(spaltenBreite, View.MeasureSpec.EXACTLY),
                        View.MeasureSpec.makeMeasureSpec(0, View.MeasureSpec.UNSPECIFIED));
                hoehe = Math.max(hoehe, teil.getMeasuredHeight());
            }

            LinearLayout zeile = new LinearLayout(this);
            zeile.setOrientation(LinearLayout.HORIZONTAL);
            for (int i = 0; i < spalten; i++) {
                boolean gefuellt = i < reihe.size();
                View teil = gefuellt ? reihe.get(i) : new View(this);
                LinearLayout.LayoutParams platz = new LinearLayout.LayoutParams(
                        0, gefuellt ? hoehe : 1, 1f);
                if (i < spalten - 1) platz.rightMargin = abstand;
                teil.setLayoutParams(platz);
                zeile.addView(teil);
            }
            LinearLayout.LayoutParams platz = new LinearLayout.LayoutParams(
                    LinearLayout.LayoutParams.MATCH_PARENT, LinearLayout.LayoutParams.WRAP_CONTENT);
            platz.bottomMargin = abstand;
            liste.addView(zeile, platz);
        }
    }

    /** Große Karte für ein Highlight – Bild, Name, Kurztext und Startmöglichkeit. */
    private View highlightKarte(Highlight highlight) {
        LinearLayout karte = new LinearLayout(this);
        karte.setOrientation(LinearLayout.VERTICAL);
        karte.setBackground(Ansicht.flaeche(Ansicht.KARTE, Ansicht.dp(this, 12), Ansicht.RAND,
                Math.max(1, Ansicht.dp(this, 1))));
        karte.setClipToOutline(true);
        karte.setClickable(true);
        karte.setFocusable(true);
        karte.setMinimumHeight(Ansicht.dp(this, 186));
        karte.setContentDescription(highlight.spiel.name + " spielen");

        ImageView bild = new ImageView(this);
        bild.setScaleType(ImageView.ScaleType.CENTER_CROP);
        karte.addView(bild, new LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT, Ansicht.dp(this, 92)));
        ladeBild(bild, highlight.bild);

        LinearLayout innen = new LinearLayout(this);
        innen.setOrientation(LinearLayout.VERTICAL);
        innen.setPadding(Ansicht.dp(this, 10), Ansicht.dp(this, 8),
                Ansicht.dp(this, 10), Ansicht.dp(this, 10));
        karte.addView(innen, new LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT, LinearLayout.LayoutParams.WRAP_CONTENT));

        TextView name = Ansicht.text(this, highlight.spiel.name, 14, Ansicht.TEXT, true);
        name.setMaxLines(2);
        innen.addView(name);

        if (!highlight.spiel.text.isEmpty()) {
            TextView text = Ansicht.text(this, highlight.spiel.text, 12, Ansicht.MUTED, false);
            text.setLineSpacing(0, 1.25f);
            text.setPadding(0, Ansicht.dp(this, 2), 0, 0);
            innen.addView(text);
        }

        TextView start = Ansicht.text(this, "Spielen ›", 12, Ansicht.GOLD, true);
        start.setPadding(0, Ansicht.dp(this, 5), 0, 0);
        innen.addView(start);

        karte.setOnClickListener(v -> starte(highlight.spiel));
        return karte;
    }

    /** Kachel einer Kategorie: Symbol, Name, Kurzbeschreibung und Anzahl. */
    private View kategorieKachel(final Kategorie kategorie, int anzahl) {
        LinearLayout karte = new LinearLayout(this);
        karte.setOrientation(LinearLayout.HORIZONTAL);
        karte.setBackground(Ansicht.flaeche(Ansicht.KARTE, Ansicht.dp(this, 12), Ansicht.RAND,
                Math.max(1, Ansicht.dp(this, 1))));
        karte.setClipToOutline(true);
        karte.setClickable(true);
        karte.setFocusable(true);
        karte.setMinimumHeight(Ansicht.dp(this, 116));
        karte.setContentDescription(kategorie.name + " – " + anzahl + " Spiele");

        View streifen = new View(this);
        streifen.setBackgroundColor(kategorie.akzent);
        karte.addView(streifen, new LinearLayout.LayoutParams(
                Ansicht.dp(this, 4), LinearLayout.LayoutParams.MATCH_PARENT));

        LinearLayout innen = new LinearLayout(this);
        innen.setOrientation(LinearLayout.VERTICAL);
        innen.setPadding(Ansicht.dp(this, 10), Ansicht.dp(this, 9),
                Ansicht.dp(this, 10), Ansicht.dp(this, 9));
        karte.addView(innen, new LinearLayout.LayoutParams(0,
                LinearLayout.LayoutParams.WRAP_CONTENT, 1f));

        Spielsymbol zeichen = new Spielsymbol(this);
        zeichen.setze(kategorie.ikon, kategorie.akzent);
        LinearLayout.LayoutParams zeichenPlatz = new LinearLayout.LayoutParams(
                Ansicht.dp(this, 24), Ansicht.dp(this, 24));
        innen.addView(zeichen, zeichenPlatz);

        TextView name = Ansicht.text(this, kategorie.name, 15, Ansicht.TEXT, true);
        name.setPadding(0, Ansicht.dp(this, 2), 0, 0);
        innen.addView(name);

        if (!kategorie.text.isEmpty()) {
            TextView text = Ansicht.text(this, kategorie.text, 11, Ansicht.MUTED, false);
            text.setLineSpacing(0, 1.25f);
            text.setPadding(0, Ansicht.dp(this, 2), 0, 0);
            innen.addView(text);
        }

        TextView zahl = Ansicht.text(this, anzahl + (anzahl == 1 ? " Spiel" : " Spiele"),
                11, Ansicht.MUTED, false);
        zahl.setPadding(0, Ansicht.dp(this, 6), 0, 0);
        innen.addView(zahl);

        karte.setOnClickListener(v -> {
            ansicht = ANSICHT_KATEGORIE;
            offeneKategorie = kategorie.id;
            zeichne();
        });
        return karte;
    }

    /** Auffangkachel für Spiele mit unbekannter Kategorie (nur bei alten Katalogen). */
    private View restKachel(int anzahl) {
        Kategorie rest = new Kategorie();
        rest.id = REST;
        rest.name = "Weitere Spiele";
        rest.text = "Spiele ohne bekannte Kategorie.";
        rest.akzent = Ansicht.MUTED;
        return kategorieKachel(rest, anzahl);
    }

    /** Vorschaubild aus den App-Assets; einmal geladen und verkleinert. */
    private void ladeBild(ImageView ziel, String relativ) {
        if (relativ == null || relativ.isEmpty()) {
            ziel.setVisibility(View.GONE);
            return;
        }
        Bitmap bitmap = bilder.get(relativ);
        if (bitmap == null) {
            String pfad = "www/spiele/" + relativ;
            try (InputStream strom = getAssets().open(pfad)) {
                BitmapFactory.Options optionen = new BitmapFactory.Options();
                // 1392×752 gezeigt auf höchstens ~450 px – ein Viertel reicht und spart Speicher.
                optionen.inSampleSize = 3;
                bitmap = BitmapFactory.decodeStream(strom, null, optionen);
                if (bitmap != null) bilder.put(relativ, bitmap);
            } catch (Exception fehler) {
                Log.w(TAG, "Vorschaubild fehlt: " + pfad, fehler);
            }
        }
        if (bitmap == null) {
            ziel.setVisibility(View.GONE);
        } else {
            ziel.setImageBitmap(bitmap);
        }
    }

    private boolean passt(Spiel spiel) {
        if (!filter.isEmpty()) {
            if (!spiel.suchtext.contains(filter)
                    && !spiel.suchtext.contains(ohneUmlaute(filter))) return false;
        }
        for (String merkmal : aktiveMerkmale) {
            if (!spiel.merkmale.contains(merkmal)) return false;
        }
        return true;
    }

    /** Karte eines Spiels – einmal gebaut, danach nur noch eingehängt. */
    private View zeile(Spiel spiel) {
        View vorhanden = zeilen.get(spiel);
        if (vorhanden != null) return vorhanden;

        LinearLayout zeile = new LinearLayout(this);
        zeile.setOrientation(LinearLayout.HORIZONTAL);
        zeile.setBackground(Ansicht.flaeche(Ansicht.KARTE, Ansicht.dp(this, 12), Ansicht.RAND,
                Math.max(1, Ansicht.dp(this, 1))));
        zeile.setPadding(Ansicht.dp(this, 14), Ansicht.dp(this, 11),
                Ansicht.dp(this, 14), Ansicht.dp(this, 11));
        zeile.setClickable(true);
        zeile.setFocusable(true);
        zeile.setContentDescription(spiel.name + " spielen");
        LinearLayout.LayoutParams platz = new LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT, LinearLayout.LayoutParams.WRAP_CONTENT);
        platz.bottomMargin = Ansicht.dp(this, 8);
        zeile.setLayoutParams(platz);

        Kategorie kategorie = kategorieNach.get(spiel.kategorie);
        Spielsymbol zeichen = new Spielsymbol(this);
        zeichen.setze(symbole.get(spiel.symbol),
                kategorie == null ? Ansicht.GOLD : kategorie.akzent);
        LinearLayout.LayoutParams zeichenPlatz = new LinearLayout.LayoutParams(
                Ansicht.dp(this, 26), Ansicht.dp(this, 26));
        zeichenPlatz.rightMargin = Ansicht.dp(this, 11);
        zeichenPlatz.topMargin = Ansicht.dp(this, 1);
        zeile.addView(zeichen, zeichenPlatz);

        LinearLayout innen = new LinearLayout(this);
        innen.setOrientation(LinearLayout.VERTICAL);
        zeile.addView(innen, new LinearLayout.LayoutParams(0,
                LinearLayout.LayoutParams.WRAP_CONTENT, 1f));

        TextView name = new TextView(this);
        name.setText(spiel.name);
        name.setTextColor(Ansicht.TEXT);
        name.setTypeface(Typeface.DEFAULT_BOLD);
        name.setTextSize(TypedValue.COMPLEX_UNIT_SP, 16);
        innen.addView(name);

        if (!spiel.text.isEmpty()) {
            TextView beschreibung = new TextView(this);
            beschreibung.setText(spiel.text);
            beschreibung.setTextColor(Ansicht.MUTED);
            beschreibung.setTextSize(TypedValue.COMPLEX_UNIT_SP, 12);
            beschreibung.setLineSpacing(0, 1.25f);
            beschreibung.setPadding(0, Ansicht.dp(this, 2), 0, 0);
            innen.addView(beschreibung);
        }

        String merkmale = merkmaleText(spiel);
        if (!merkmale.isEmpty()) {
            TextView unten = new TextView(this);
            unten.setText(merkmale);
            unten.setTextColor(Ansicht.MUTED);
            unten.setTextSize(TypedValue.COMPLEX_UNIT_SP, 11);
            unten.setPadding(0, Ansicht.dp(this, 4), 0, 0);
            innen.addView(unten);
        }

        zeile.setOnClickListener(v -> starte(spiel));
        zeilen.put(spiel, zeile);
        return zeile;
    }

    private String merkmaleText(Spiel spiel) {
        StringBuilder text = new StringBuilder();
        for (String merkmal : spiel.merkmale) {
            String wort = MERKMAL_TEXT.containsKey(merkmal) ? MERKMAL_TEXT.get(merkmal) : merkmal;
            if (text.length() > 0) text.append("  ·  ");
            text.append(wort);
        }
        return text.toString();
    }

    // -------------------------------------------------------------- Starten

    /** Online-Spiele fragen zuerst nach dem Modus, alle anderen starten sofort lokal. */
    private void starte(final Spiel spiel) {
        if (!spiel.online) {
            oeffne(spiel, false);
            return;
        }
        new AlertDialog.Builder(this)
                .setTitle(spiel.name)
                .setMessage("Lokal/offline öffnet die eingebettete Spielseite. Online nutzt die "
                        + "vorhandenen Spielräume und den TrashyGecko-Server; dafür ist Internet erforderlich.")
                .setNegativeButton("Lokal / offline", (dialog, which) -> oeffne(spiel, false))
                .setPositiveButton("Online-Spiel", (dialog, which) -> oeffne(spiel, true))
                .setNeutralButton("Abbrechen", null)
                .show();
    }

    private void oeffne(Spiel spiel, boolean online) {
        Intent absicht = new Intent(this, GameActivity.class);
        absicht.putExtra(GameActivity.EXTRA_PFAD, spiel.pfad);
        absicht.putExtra(GameActivity.EXTRA_ONLINE, online);
        startActivity(absicht);
    }
}
