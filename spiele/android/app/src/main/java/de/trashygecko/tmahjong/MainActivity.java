package de.trashygecko.tmahjong;

import android.app.Activity;
import android.content.Intent;
import android.graphics.Typeface;
import android.graphics.drawable.GradientDrawable;
import android.net.ConnectivityManager;
import android.net.NetworkInfo;
import android.os.Bundle;
import android.util.TypedValue;
import android.view.Gravity;
import android.view.View;
import android.widget.LinearLayout;
import android.widget.ScrollView;
import android.widget.TextView;

/**
 * Startansicht der App – nativ, vier Kacheln wie auf der alten HTML-Startseite (D-12).
 *
 * Spiele, Katalog und Connector öffnen eigene Activities; die HTML-Startseite liegt nur
 * noch als `index.html` im Paket, damit „zurück zum Start“ aus einer Spielseite landet.
 */
public final class MainActivity extends Activity {
    private static final int BG = 0xFF07140F;
    private static final int KARTE = 0xFF122B22;
    private static final int LINIE = 0xFF24503F;
    private static final int GOLD = 0xFFE8CD8E;
    private static final int TEXT = 0xFFEAF3EC;
    private static final int MUTED = 0xFF9FB8AA;

    private TextView status;

    @Override protected void onCreate(Bundle zustand) {
        super.onCreate(zustand);
        WebBasis.vollbild(this);
        setContentView(baueAnsicht());
    }

    @Override protected void onResume() {
        super.onResume();
        if (status != null) status.setText(netzStatus());
    }

    // ---------------------------------------------------------------- Aufbau

    private View baueAnsicht() {
        ScrollView rollen = new ScrollView(this);
        rollen.setBackgroundColor(BG);
        rollen.setFillViewport(true);

        LinearLayout seite = new LinearLayout(this);
        seite.setOrientation(LinearLayout.VERTICAL);
        int rand = dp(18);
        seite.setPadding(rand, dp(22), rand, dp(26));
        rollen.addView(seite, new ScrollView.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT, LinearLayout.LayoutParams.WRAP_CONTENT));

        LinearLayout kopf = new LinearLayout(this);
        kopf.setOrientation(LinearLayout.HORIZONTAL);
        kopf.setGravity(Gravity.CENTER_VERTICAL);
        TextView marke = text("TRASHYGECKO", 12, GOLD, true);
        marke.setLetterSpacing(0.22f);
        kopf.addView(marke, new LinearLayout.LayoutParams(
                0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f));
        status = text("", 12, MUTED, false);
        kopf.addView(status);
        seite.addView(kopf);

        TextView titel = text("Was möchtest du spielen?", 28, GOLD, true);
        titel.setPadding(0, dp(12), 0, dp(6));
        seite.addView(titel);

        TextView unter = text("Alles liegt auf dem Gerät – Spiele, Räume und Werkzeuge.", 15, MUTED, false);
        unter.setPadding(0, 0, 0, dp(20));
        seite.addView(unter);

        kachel(seite, "Spiele",
                "Spiele, Werkzeuge und Editoren. Einzelspieler läuft ohne Netz.",
                v -> startActivity(new Intent(this, SpieleActivity.class)));
        kachel(seite, "Lokales Netz",
                "Raum aufmachen oder einem beitreten – ohne Internet, im selben WLAN.",
                v -> starteConnector("local"));
        kachel(seite, "Online-Räume",
                "Chat und Spiele mit mehreren – auch die Würfelspiele. Braucht Internet.",
                v -> starteConnector("online"));
        kachel(seite, "Extras",
                "Zusatzfunktionen – dort liegt auch die Universalfernbedienung für Fernseher im WLAN.",
                v -> starteConnector("extras"));

        TextView fuss = text("Keine Konten, keine Werbung. Spiele und Seiten liegen im App-Paket; "
                + "nur Online-Räume brauchen ein Netz. Die Universalfernbedienung findest du unter Extras.",
                12, MUTED, false);
        fuss.setLineSpacing(0, 1.4f);
        fuss.setPadding(0, dp(24), 0, 0);
        seite.addView(fuss);
        return rollen;
    }

    /** Eine Kachel: Punkt, Titel, Pfeil, darunter die Kurzbeschreibung. */
    private void kachel(LinearLayout seite, String titel, String beschreibung,
                        View.OnClickListener ziel) {
        LinearLayout karte = new LinearLayout(this);
        karte.setOrientation(LinearLayout.VERTICAL);
        karte.setPadding(dp(16), dp(16), dp(16), dp(14));
        karte.setBackground(flaeche(KARTE, dp(12), LINIE));
        karte.setClickable(true);
        karte.setFocusable(true);
        LinearLayout.LayoutParams platz = new LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT, LinearLayout.LayoutParams.WRAP_CONTENT);
        platz.bottomMargin = dp(12);
        seite.addView(karte, platz);

        LinearLayout zeile = new LinearLayout(this);
        zeile.setOrientation(LinearLayout.HORIZONTAL);
        zeile.setGravity(Gravity.CENTER_VERTICAL);
        View punkt = new View(this);
        punkt.setBackground(punkt());
        LinearLayout.LayoutParams punktPlatz = new LinearLayout.LayoutParams(dp(9), dp(9));
        punktPlatz.rightMargin = dp(10);
        zeile.addView(punkt, punktPlatz);
        zeile.addView(text(titel, 19, TEXT, true), new LinearLayout.LayoutParams(
                0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f));
        zeile.addView(text("›", 20, GOLD, false));
        karte.addView(zeile);

        TextView unten = text(beschreibung, 13, MUTED, false);
        unten.setLineSpacing(0, 1.3f);
        unten.setPadding(0, dp(6), 0, 0);
        karte.addView(unten);

        karte.setOnClickListener(ziel);
    }

    // -------------------------------------------------------------- Vernetzung

    private void starteConnector(String bildschirm) {
        startActivity(new Intent(this, ConnectorActivity.class)
                .putExtra(ConnectorActivity.EXTRA_INITIAL_SCREEN, bildschirm));
    }

    /** Ohne Netz bleibt alles außer den Online-Räumen nutzbar – das sagt die Zeile. */
    private String netzStatus() {
        ConnectivityManager netz = (ConnectivityManager) getSystemService(CONNECTIVITY_SERVICE);
        NetworkInfo aktiv = netz == null ? null : netz.getActiveNetworkInfo();
        if (aktiv == null || !aktiv.isConnected()) return "offline spielbar";
        String art = aktiv.getType() == ConnectivityManager.TYPE_WIFI ? "WLAN"
                : aktiv.getType() == ConnectivityManager.TYPE_MOBILE ? "Mobilfunk" : "Netz";
        return art + " · online bereit";
    }

    // -------------------------------------------------------------- Kleinkram

    private GradientDrawable flaeche(int farbe, int radius, int randfarbe) {
        GradientDrawable form = new GradientDrawable();
        form.setColor(farbe);
        form.setCornerRadius(radius);
        form.setStroke(Math.max(1, dp(1)), randfarbe);
        return form;
    }

    private GradientDrawable punkt() {
        GradientDrawable form = new GradientDrawable();
        form.setShape(GradientDrawable.OVAL);
        form.setColor(GOLD);
        return form;
    }

    private TextView text(String wert, int groesse, int farbe, boolean fett) {
        TextView feld = new TextView(this);
        feld.setText(wert);
        feld.setTextSize(TypedValue.COMPLEX_UNIT_SP, groesse);
        feld.setTextColor(farbe);
        if (fett) feld.setTypeface(Typeface.DEFAULT_BOLD);
        return feld;
    }

    private int dp(int wert) {
        return Math.round(wert * getResources().getDisplayMetrics().density);
    }

    // ------------------------------------------------------------ Lebenszyklus

    @Override public void onWindowFocusChanged(boolean hatFokus) {
        super.onWindowFocusChanged(hatFokus);
        if (hatFokus) {
            WebBasis.vollbild(this);
        }
    }

    @Override public void onBackPressed() {
        finishAffinity();
    }
}
