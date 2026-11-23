package de.trashygecko.tmahjong;

import android.app.Activity;
import android.content.Intent;
import android.graphics.drawable.GradientDrawable;
import android.net.ConnectivityManager;
import android.net.NetworkInfo;
import android.os.Bundle;
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
 * Farben und Bausteine kommen aus {@link Ansicht} – dieselben wie in der Spieleliste.
 */
public final class MainActivity extends Activity {
    private TextView status;

    @Override protected void onCreate(Bundle zustand) {
        super.onCreate(zustand);
        WebBasis.vollbild(this);
        setContentView(baueAnsicht());
    }

    @Override protected void onResume() {
        super.onResume();
        status.setText(netzStatus());
    }

    // ---------------------------------------------------------------- Aufbau

    private View baueAnsicht() {
        ScrollView rollen = new ScrollView(this);
        rollen.setBackgroundColor(Ansicht.BG);
        rollen.setFillViewport(true);

        LinearLayout seite = new LinearLayout(this);
        seite.setOrientation(LinearLayout.VERTICAL);
        int rand = Ansicht.dp(this, 18);
        seite.setPadding(rand, Ansicht.dp(this, 22), rand, Ansicht.dp(this, 26));
        rollen.addView(seite, new ScrollView.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT, LinearLayout.LayoutParams.WRAP_CONTENT));

        LinearLayout kopf = new LinearLayout(this);
        kopf.setOrientation(LinearLayout.HORIZONTAL);
        kopf.setGravity(Gravity.CENTER_VERTICAL);
        TextView marke = Ansicht.text(this, "TRASHYGECKO", 12, Ansicht.GOLD, true);
        marke.setLetterSpacing(0.22f);
        kopf.addView(marke, new LinearLayout.LayoutParams(
                0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f));
        status = Ansicht.text(this, "", 12, Ansicht.MUTED, false);
        kopf.addView(status);
        seite.addView(kopf);

        TextView titel = Ansicht.text(this, "Was möchtest du spielen?", 28, Ansicht.GOLD, true);
        titel.setPadding(0, Ansicht.dp(this, 12), 0, Ansicht.dp(this, 6));
        seite.addView(titel);

        TextView unter = Ansicht.text(this, "Alles liegt auf dem Gerät – Spiele, Räume und Werkzeuge.",
                15, Ansicht.MUTED, false);
        unter.setPadding(0, 0, 0, Ansicht.dp(this, 20));
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

        TextView fuss = Ansicht.text(this, "Keine Konten, keine Werbung. Spiele und Seiten liegen "
                + "im App-Paket; nur Online-Räume brauchen ein Netz. Die Universalfernbedienung "
                + "findest du unter Extras.", 12, Ansicht.MUTED, false);
        fuss.setLineSpacing(0, 1.4f);
        fuss.setPadding(0, Ansicht.dp(this, 24), 0, 0);
        seite.addView(fuss);
        return rollen;
    }

    /** Eine Kachel: Punkt, Titel, Pfeil, darunter die Kurzbeschreibung. */
    private void kachel(LinearLayout seite, String titel, String beschreibung,
                        View.OnClickListener ziel) {
        LinearLayout karte = new LinearLayout(this);
        karte.setOrientation(LinearLayout.VERTICAL);
        karte.setPadding(Ansicht.dp(this, 16), Ansicht.dp(this, 16),
                Ansicht.dp(this, 16), Ansicht.dp(this, 14));
        karte.setBackground(Ansicht.flaeche(Ansicht.KARTE, Ansicht.dp(this, 12),
                Ansicht.RAND, Ansicht.dp(this, 1)));
        karte.setClickable(true);
        karte.setFocusable(true);
        LinearLayout.LayoutParams platz = new LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT, LinearLayout.LayoutParams.WRAP_CONTENT);
        platz.bottomMargin = Ansicht.dp(this, 12);
        seite.addView(karte, platz);

        LinearLayout zeile = new LinearLayout(this);
        zeile.setOrientation(LinearLayout.HORIZONTAL);
        zeile.setGravity(Gravity.CENTER_VERTICAL);
        View punkt = new View(this);
        punkt.setBackground(punktfarbe());
        LinearLayout.LayoutParams punktPlatz = new LinearLayout.LayoutParams(
                Ansicht.dp(this, 9), Ansicht.dp(this, 9));
        punktPlatz.rightMargin = Ansicht.dp(this, 10);
        zeile.addView(punkt, punktPlatz);
        zeile.addView(Ansicht.text(this, titel, 19, Ansicht.TEXT, true),
                new LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f));
        zeile.addView(Ansicht.text(this, "›", 20, Ansicht.GOLD, false));
        karte.addView(zeile);

        TextView unten = Ansicht.text(this, beschreibung, 13, Ansicht.MUTED, false);
        unten.setLineSpacing(0, 1.3f);
        unten.setPadding(0, Ansicht.dp(this, 6), 0, 0);
        karte.addView(unten);

        karte.setOnClickListener(ziel);
    }

    private GradientDrawable punktfarbe() {
        GradientDrawable form = new GradientDrawable();
        form.setShape(GradientDrawable.OVAL);
        form.setColor(Ansicht.GOLD);
        return form;
    }

    // -------------------------------------------------------------- Vernetzung

    private void starteConnector(String bildschirm) {
        startActivity(new Intent(this, ConnectorActivity.class)
                .putExtra(ConnectorActivity.EXTRA_INITIAL_SCREEN, bildschirm));
    }

    /**
     * Ohne Netz bleibt alles außer den Online-Räumen nutzbar – das sagt die Zeile.
     *
     * `getActiveNetworkInfo()` ist seit API 23 veraltet, auf Android 9 (API 28) aber
     * unverändert nutzbar und braucht nur `ACCESS_NETWORK_STATE`; erst ab Android 10
     * liefert es für Hintergrund-Apps seltener etwas, und die Zeile wird ohnehin bei
     * jedem `onResume` neu gelesen.
     */
    private String netzStatus() {
        ConnectivityManager netz = (ConnectivityManager) getSystemService(CONNECTIVITY_SERVICE);
        NetworkInfo aktiv = netz == null ? null : netz.getActiveNetworkInfo();
        if (aktiv == null || !aktiv.isConnected()) return "offline spielbar";
        String art = aktiv.getType() == ConnectivityManager.TYPE_WIFI ? "WLAN"
                : aktiv.getType() == ConnectivityManager.TYPE_MOBILE ? "Mobilfunk" : "Netz";
        return art + " · online bereit";
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
