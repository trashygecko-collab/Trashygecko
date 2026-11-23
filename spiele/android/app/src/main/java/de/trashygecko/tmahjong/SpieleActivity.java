package de.trashygecko.tmahjong;

import android.app.Activity;
import android.app.AlertDialog;
import android.content.Intent;
import android.content.res.Configuration;
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
import android.widget.LinearLayout;
import android.widget.ScrollView;
import android.widget.TextView;

import org.json.JSONArray;
import org.json.JSONObject;

import java.io.ByteArrayOutputStream;
import java.io.InputStream;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.HashSet;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Locale;
import java.util.Map;
import java.util.Set;

/**
 * Die Spieleübersicht der App – nativ, aus `assets/katalog.json`.
 *
 * Die Datei entsteht beim Bauen aus `spiele/spiele.html` (Aufgabe
 * `katalogBauen`, D-15); hier wird sie nur gelesen. Ein Tipp auf ein Spiel
 * öffnet {@link GameActivity}, bei Online-Spielen vorher die Frage
 * „Lokal / Online“ – wie bisher im Katalog.
 *
 * Suche und Merkmalsfilter verhalten sich wie im HTML-Katalog: Die Suche
 * prüft Name und Suchwörter, jedes gewählte Merkmal muss zusätzlich zutreffen,
 * und angeboten werden nur Merkmale, die im Katalog wirklich vorkommen.
 */
public final class SpieleActivity extends Activity {
    private static final String TAG = "TrashyGeckoSpiele";

    /** Schlüssel für den gesicherten Zustand (Prozess-Tod, „Aktivitäten nicht behalten“). */
    private static final String ZUSTAND_SUCHE = "suche";
    private static final String ZUSTAND_MERKMALE = "merkmale";
    private static final String ZUSTAND_SCROLL = "scroll";

    private static final int BG = 0xFF0B241D;
    private static final int KARTE = 0xFF123028;
    private static final int CHIP = 0xFF1A3A30;
    private static final int CHIP_RAND = 0xFF2B4C40;
    private static final int GOLD = 0xFFE8CD8E;
    private static final int TEXT = 0xFFE9F1EB;
    private static final int MUTED = 0xFF9AA7A0;

    /** Kurzform für die Zeile unter dem Spielnamen. */
    private static final Map<String, String> MERKMAL_TEXT = new LinkedHashMap<>();
    /** Beschriftung der Filterknöpfe – Wortlaut wie im HTML-Katalog. */
    private static final Map<String, String> MERKMAL_FILTER = new LinkedHashMap<>();

    static {
        MERKMAL_TEXT.put("solo", "Solo");
        MERKMAL_TEXT.put("lokal", "Lokal");
        MERKMAL_TEXT.put("multi", "Mehrspieler");
        MERKMAL_TEXT.put("online", "Online");
        MERKMAL_TEXT.put("computer", "Computer");
        MERKMAL_TEXT.put("highscore", "Rekord");
        MERKMAL_TEXT.put("controller", "Pad");

        MERKMAL_FILTER.put("solo", "Einzelspieler");
        MERKMAL_FILTER.put("multi", "Mehrspieler");
        MERKMAL_FILTER.put("lokal", "Lokal am Gerät");
        MERKMAL_FILTER.put("online", "Online");
        MERKMAL_FILTER.put("computer", "Gegen Computer");
        MERKMAL_FILTER.put("highscore", "Highscore");
        MERKMAL_FILTER.put("controller", "Controller");
    }

    /** Ein Eintrag aus katalog.json. */
    private static final class Spiel {
        String name;
        String pfad;
        String kategorie;
        List<String> merkmale = new ArrayList<>();
        String suche = "";
        boolean online;
    }

    private final List<Spiel> spiele = new ArrayList<>();
    private final List<String> kategorien = new ArrayList<>();
    private final Map<String, String> kategorieNamen = new LinkedHashMap<>();
    private final List<String> aktiveMerkmale = new ArrayList<>();
    /** Einmal gebaute Zeilenansichten – ein Filterwechsel hängt sie nur um. */
    private final Map<Spiel, View> zeilen = new HashMap<>();
    private final Map<String, TextView> chips = new LinkedHashMap<>();

    private LinearLayout liste;
    private LinearLayout chipLeiste;
    private TextView zaehler;
    private TextView zurueckChip;
    private EditText suchfeld;
    private ScrollView rollen;
    private String filter = "";
    /** Beim Wiederherstellen zeichnet noch nicht der Textwächter – das tut `stelleWieder`. */
    private boolean stelleWiederLaeuft;

    @Override protected void onCreate(Bundle zustand) {
        super.onCreate(zustand);
        WebBasis.vollbild(this);
        setContentView(baueAnsicht());
        ladeKatalog();
        stelleWieder(zustand);
    }

    /**
     * Wird die Activity vom System beendet (Prozess-Tod im Hintergrund, „Aktivitäten nicht
     * behalten“), entsteht sie neu. Suchtext, Merkmalswahl und Scrollposition werden deshalb
     * gesichert – die Drehung allein braucht das nicht (`configChanges`).
     */
    @Override protected void onSaveInstanceState(Bundle zustand) {
        super.onSaveInstanceState(zustand);
        zustand.putString(ZUSTAND_SUCHE, suchfeld.getText().toString());
        zustand.putStringArrayList(ZUSTAND_MERKMALE, new ArrayList<>(aktiveMerkmale));
        zustand.putInt(ZUSTAND_SCROLL, rollen.getScrollY());
    }

    /**
     * Holt den gesicherten Zustand zurück und zeichnet die Liste danach genau einmal – ohne
     * Sicherung ist das der erste Aufbau. Unbekannte Merkmale aus einem älteren Bundle
     * entfallen, weil `baueChips` nur vorhandene anbietet.
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
                    + aktiveMerkmale.size() + " Merkmale, Scroll " + scroll);
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
     * Prozess-Tod, den {@link #onSaveInstanceState} abdeckt. Die Chip-Leiste
     * wurde aber für die alte Breite umbrochen – im Querformat passen alle in eine Zeile,
     * im Hochformat wäre der Rest abgeschnitten. Deshalb hier mit der neuen Breite neu
     * umbrechen; die Auswahl steht in {@link #aktiveMerkmale} und bleibt bestehen.
     */
    @Override public void onConfigurationChanged(Configuration neu) {
        super.onConfigurationChanged(neu);
        WebBasis.vollbild(this);
        if (chipLeiste == null || spiele.isEmpty()) return;
        int scroll = rollen == null ? 0 : rollen.getScrollY();
        baueChips();
        if (rollen != null) rollen.post(() -> rollen.scrollTo(0, scroll));
    }

    // ---------------------------------------------------------------- Aufbau

    private View baueAnsicht() {
        LinearLayout wurzel = new LinearLayout(this);
        wurzel.setOrientation(LinearLayout.VERTICAL);
        wurzel.setBackgroundColor(BG);
        int rand = dp(16);
        wurzel.setPadding(rand, dp(12), rand, 0);

        LinearLayout kopf = new LinearLayout(this);
        kopf.setOrientation(LinearLayout.HORIZONTAL);
        kopf.setGravity(Gravity.CENTER_VERTICAL);

        TextView zurueck = new TextView(this);
        zurueck.setText("‹");
        zurueck.setTextColor(GOLD);
        zurueck.setTextSize(TypedValue.COMPLEX_UNIT_SP, 26);
        zurueck.setPadding(0, 0, dp(14), dp(2));
        zurueck.setOnClickListener(v -> finish());
        kopf.addView(zurueck);

        TextView titel = new TextView(this);
        titel.setText("Spiele");
        titel.setTextColor(GOLD);
        titel.setTypeface(Typeface.DEFAULT_BOLD);
        titel.setTextSize(TypedValue.COMPLEX_UNIT_SP, 24);
        LinearLayout.LayoutParams titelPlatz = new LinearLayout.LayoutParams(0,
                LinearLayout.LayoutParams.WRAP_CONTENT, 1f);
        kopf.addView(titel, titelPlatz);

        zaehler = new TextView(this);
        zaehler.setTextColor(MUTED);
        zaehler.setTextSize(TypedValue.COMPLEX_UNIT_SP, 13);
        kopf.addView(zaehler);
        wurzel.addView(kopf);

        suchfeld = new EditText(this);
        suchfeld.setHint("Spiel suchen – z. B. Mau Mau, Kniffel, Halma");
        suchfeld.setHintTextColor(MUTED);
        suchfeld.setTextColor(TEXT);
        suchfeld.setTextSize(TypedValue.COMPLEX_UNIT_SP, 15);
        suchfeld.setSingleLine(true);
        suchfeld.setBackground(flaeche(KARTE, dp(12)));
        suchfeld.setPadding(dp(14), dp(10), dp(14), dp(10));
        LinearLayout.LayoutParams suchPlatz = new LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT, LinearLayout.LayoutParams.WRAP_CONTENT);
        suchPlatz.topMargin = dp(12);
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
        chipPlatz.topMargin = dp(10);
        wurzel.addView(chipLeiste, chipPlatz);

        rollen = new ScrollView(this);
        rollen.setClipToPadding(false);
        LinearLayout.LayoutParams rollPlatz = new LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT, 0, 1f);
        rollPlatz.topMargin = dp(4);
        wurzel.addView(rollen, rollPlatz);

        liste = new LinearLayout(this);
        liste.setOrientation(LinearLayout.VERTICAL);
        liste.setPadding(0, 0, 0, dp(24));
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
                    String id = eintrag.optString("id");
                    kategorien.add(id);
                    kategorieNamen.put(id, eintrag.optString("name", id));
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
                    spiel.suche = eintrag.optString("suche", "");
                    JSONArray merk = eintrag.optJSONArray("merkmale");
                    if (merk != null) {
                        for (int m = 0; m < merk.length(); m++) {
                            spiel.merkmale.add(merk.optString(m));
                        }
                    }
                    spiel.online = spiel.merkmale.contains("online");
                    if (!spiel.name.isEmpty() && WebBasis.istSichereSeite(spiel.pfad)) {
                        spiele.add(spiel);
                    }
                }
            }
            Log.i(TAG, "Katalog: " + spiele.size() + " Spiele, " + kategorien.size() + " Kategorien");
            baueChips();
        } catch (Exception fehler) {
            Log.e(TAG, "Katalog nicht lesbar", fehler);
            zeigeHinweis("Der Spielekatalog fehlt im App-Paket. Bitte die App neu installieren.");
        }
    }

    private void zeigeHinweis(String text) {
        liste.removeAllViews();
        TextView hinweis = new TextView(this);
        hinweis.setText(text);
        hinweis.setTextColor(TEXT);
        hinweis.setTextSize(TypedValue.COMPLEX_UNIT_SP, 15);
        hinweis.setPadding(0, dp(12), 0, 0);
        liste.addView(hinweis);
    }

    // -------------------------------------------------------------- Filter

    /** Baut die Filterleiste; leere Filter werden wie im HTML-Katalog nicht angeboten. */
    private void baueChips() {
        chipLeiste.removeAllViews();
        chips.clear();

        // Immer aus der aktuellen Anzeigebreite: beim Umbau nach einer Drehung ist die
        // alte Breite der Ansicht noch nicht neu gemessen.
        int platz = getResources().getDisplayMetrics().widthPixels - dp(32);
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
            int breite = chip.getMeasuredWidth() + dp(8);
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
        platz.bottomMargin = dp(8);
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
        return belegt + chip.getMeasuredWidth() + dp(8);
    }

    private TextView chip(String beschriftung) {
        TextView chip = new TextView(this);
        chip.setText(beschriftung);
        chip.setTextSize(TypedValue.COMPLEX_UNIT_SP, 13);
        chip.setGravity(Gravity.CENTER);
        chip.setMinHeight(dp(38));
        chip.setPadding(dp(14), dp(8), dp(14), dp(8));
        chip.setClickable(true);
        chip.setFocusable(true);
        LinearLayout.LayoutParams platz = new LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.WRAP_CONTENT, LinearLayout.LayoutParams.WRAP_CONTENT);
        platz.rightMargin = dp(8);
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
        form.setCornerRadius(dp(19));
        form.setColor(an ? GOLD : CHIP);
        form.setStroke(Math.max(1, dp(1)), an ? GOLD : CHIP_RAND);
        chip.setBackground(form);
        chip.setTextColor(an ? BG : TEXT);
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
        int sichtbar = 0;

        for (String id : kategorien) {
            List<Spiel> treffer = new ArrayList<>();
            for (Spiel spiel : spiele) {
                if (!id.equals(spiel.kategorie) || !passt(spiel)) continue;
                treffer.add(spiel);
            }
            if (treffer.isEmpty()) continue;

            liste.addView(ueberschrift(anzahlText(id, treffer.size())));

            for (Spiel spiel : treffer) {
                liste.addView(zeile(spiel));
                sichtbar++;
            }
        }

        // Spiele ohne bekannte Kategorie sollen nicht verschwinden.
        List<Spiel> rest = new ArrayList<>();
        for (Spiel spiel : spiele) {
            if (!kategorien.contains(spiel.kategorie) && passt(spiel)) rest.add(spiel);
        }
        if (!rest.isEmpty()) {
            liste.addView(ueberschrift("Weitere  ·  " + rest.size()));
            for (Spiel spiel : rest) {
                liste.addView(zeile(spiel));
                sichtbar++;
            }
        }

        zaehler.setText(istGefiltert()
                ? sichtbar + " von " + spiele.size()
                : spiele.size() + " Spiele");
        if (sichtbar == 0 && istGefiltert()) {
            zeigeHinweis("Kein Spiel passt zu Suche und Filtern. „Zurücksetzen“ zeigt wieder alle.");
        }
        rollen.scrollTo(0, 0);
        Log.i(TAG, "Liste: " + sichtbar + " Zeilen in "
                + ((System.nanoTime() - beginn) / 1000000L) + " ms");
    }

    /** Kategorie-Kopf; bei aktivem Filter wie im HTML-Katalog „x / gesamt“. */
    private String anzahlText(String id, int treffer) {
        String name = kategorieNamen.containsKey(id) ? kategorieNamen.get(id) : id;
        if (!istGefiltert()) return name + "  ·  " + treffer;
        int gesamt = 0;
        for (Spiel spiel : spiele) {
            if (id.equals(spiel.kategorie)) gesamt++;
        }
        return name + "  ·  " + treffer + " / " + gesamt;
    }

    private View ueberschrift(String text) {
        TextView kopf = new TextView(this);
        kopf.setText(text);
        kopf.setTextColor(GOLD);
        kopf.setTypeface(Typeface.DEFAULT_BOLD);
        kopf.setTextSize(TypedValue.COMPLEX_UNIT_SP, 15);
        kopf.setPadding(dp(4), dp(16), dp(4), dp(8));
        return kopf;
    }

    private boolean passt(Spiel spiel) {
        if (!filter.isEmpty()) {
            String name = spiel.name.toLowerCase(Locale.ROOT);
            if (!name.contains(filter) && !spiel.suche.contains(filter)) return false;
        }
        for (String merkmal : aktiveMerkmale) {
            if (!spiel.merkmale.contains(merkmal)) return false;
        }
        return true;
    }

    /** Zeile eines Spiels – einmal gebaut, danach nur noch eingehängt. */
    private View zeile(Spiel spiel) {
        View vorhanden = zeilen.get(spiel);
        if (vorhanden != null) return vorhanden;

        LinearLayout zeile = new LinearLayout(this);
        zeile.setOrientation(LinearLayout.VERTICAL);
        zeile.setBackground(flaeche(KARTE, dp(12)));
        zeile.setPadding(dp(14), dp(11), dp(14), dp(11));
        zeile.setClickable(true);
        zeile.setFocusable(true);
        LinearLayout.LayoutParams platz = new LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT, LinearLayout.LayoutParams.WRAP_CONTENT);
        platz.bottomMargin = dp(8);
        zeile.setLayoutParams(platz);

        TextView name = new TextView(this);
        name.setText(spiel.name);
        name.setTextColor(TEXT);
        name.setTextSize(TypedValue.COMPLEX_UNIT_SP, 16);
        zeile.addView(name);

        String merkmale = merkmaleText(spiel);
        if (!merkmale.isEmpty()) {
            TextView unten = new TextView(this);
            unten.setText(merkmale);
            unten.setTextColor(MUTED);
            unten.setTextSize(TypedValue.COMPLEX_UNIT_SP, 12);
            unten.setPadding(0, dp(3), 0, 0);
            zeile.addView(unten);
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

    // -------------------------------------------------------------- Kleinkram

    private GradientDrawable flaeche(int farbe, int radius) {
        GradientDrawable form = new GradientDrawable();
        form.setColor(farbe);
        form.setCornerRadius(radius);
        return form;
    }

    private int dp(int wert) {
        return Math.round(wert * getResources().getDisplayMetrics().density);
    }
}
