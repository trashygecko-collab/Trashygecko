package de.trashygecko.tspiele;

import android.app.Activity;
import android.content.Intent;
import android.content.pm.PackageManager;
import android.graphics.Color;
import android.graphics.Typeface;
import android.graphics.drawable.GradientDrawable;
import android.net.Uri;
import android.os.Build;
import android.os.Bundle;
import android.util.TypedValue;
import android.view.Gravity;
import android.view.View;
import android.widget.LinearLayout;
import android.widget.TextView;
import android.widget.Toast;

/**
 * T.Spiele zeigt zwei Wege: oben den Connector mit Chat und Spielen, darunter
 * die eigene Homepage. Beides öffnet sich im Browser; die App bleibt im
 * Hintergrund stehen, damit man zurückwechseln kann.
 *
 * Der Startbildschirm ist bewusst festlich gehalten – tiefes Grün, Gold und
 * Serifenschrift, wie eine Einladung.
 */
public class MainActivity extends Activity {

    /** Chat und Spiele laufen über den Connector im Internet. */
    private static final String CONNECTOR = "https://trashygecko.onrender.com/connector";

    /** Die Homepage liegt bei GitHub Pages – bewusst nicht die Render-Adresse. */
    private static final String HOMEPAGE = "https://trashygecko-collab.github.io/Trashygecko/";

    // Farben: tiefes Grün, Gold, Elfenbein.
    private static final int GOLD = 0xFFD9BC7C;
    private static final int GOLD_HELL = 0xFFF1DCA8;
    private static final int ELFENBEIN = 0xFFEFE6D2;
    private static final int GRUEN_TIEF = 0xFF0A211B;
    private static final int GRUEN_MITTIG = 0xFF1C4A3D;

    /** Goldener Schnitt – so teilen sich die Abstände. */
    private static final float PHI = 1.618f;

    private float dichte;

    @Override
    protected void onCreate(Bundle zustand) {
        super.onCreate(zustand);
        dichte = getResources().getDisplayMetrics().density;

        // Morgen-Lob: Kanal anlegen, Erlaubnis erfragen, Wecker stellen.
        LobWecker.kanalAnlegen(this);
        erlaubnisFuerHinweise();
        LobWecker.plane(this);

        if (Build.VERSION.SDK_INT >= 21) {
            getWindow().setStatusBarColor(GRUEN_TIEF);
            getWindow().setNavigationBarColor(GRUEN_TIEF);
        }

        LinearLayout seite = new LinearLayout(this);
        seite.setOrientation(LinearLayout.VERTICAL);
        seite.setGravity(Gravity.CENTER_HORIZONTAL);
        seite.setBackgroundDrawable(hintergrund());
        seite.setPadding(px(32), px(62), px(32), px(44));

        // --- Kopf: Schmuckzeile, Titel, Untertitel, feine Goldlinie
        seite.addView(schrift("✦   ✦   ✦", GOLD, 12, 0.35f, Typeface.NORMAL));


        seite.addView(schrift("T.Spiele", GOLD_HELL, 38, 0.10f, Typeface.NORMAL),
                abstand(0, px(18), 0, 0));

        seite.addView(schrift("CHAT UND SPIELE", ELFENBEIN, 11, 0.32f, Typeface.NORMAL),
                abstand(0, px(12), 0, 0));

        View strich = new View(this);
        strich.setBackgroundColor(0x66D9BC7C);
        LinearLayout.LayoutParams strichMasse = new LinearLayout.LayoutParams(px(104), px(1));
        strichMasse.topMargin = px(30);
        strichMasse.gravity = Gravity.CENTER_HORIZONTAL;
        seite.addView(strich, strichMasse);

        // --- Die zwei Wege sitzen im goldenen Schnitt der Höhe.
        seite.addView(new View(this), new LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT, 0, 1f));

        seite.addView(knopf("Connector öffnen", "Chat und alle Spiele", CONNECTOR));

        // Abstand der beiden Wege: 34 dp, die Luft nach unten ist gut halb so groß.
        seite.addView(new View(this), new LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT, px(34)));

        seite.addView(knopf("Meine Homepage", "", HOMEPAGE));

        seite.addView(new View(this), new LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT, 0, 1f / PHI));

        setContentView(seite);
    }

    /** Ab Android 13 muss man Benachrichtigungen ausdrücklich erlauben. */
    private void erlaubnisFuerHinweise() {
        if (Build.VERSION.SDK_INT < 33) return;
        if (checkSelfPermission("android.permission.POST_NOTIFICATIONS")
                == PackageManager.PERMISSION_GRANTED) return;
        requestPermissions(new String[] {"android.permission.POST_NOTIFICATIONS"}, 1);
    }

    /** Sanfter Verlauf von hellem Grün in der Mitte zu dunklem Grün am Rand. */
    private GradientDrawable hintergrund() {
        GradientDrawable form = new GradientDrawable(GradientDrawable.Orientation.TOP_BOTTOM,
                new int[] {GRUEN_MITTIG, GRUEN_TIEF});
        form.setGradientType(GradientDrawable.RADIAL_GRADIENT);
        form.setGradientRadius(getResources().getDisplayMetrics().heightPixels * 0.78f);
        form.setColors(new int[] {GRUEN_MITTIG, 0xFF123329, GRUEN_TIEF});
        return form;
    }

    /** Gold umrandeter Kasten mit Serifenschrift – eine gedruckte Einladung. */
    private View knopf(String beschriftung, String zusatz, final String adresse) {
        LinearLayout kasten = new LinearLayout(this);
        kasten.setOrientation(LinearLayout.VERTICAL);
        kasten.setGravity(Gravity.CENTER);
        kasten.setPadding(px(20), px(20), px(20), px(20));
        kasten.setClickable(true);
        kasten.setFocusable(true);

        GradientDrawable form = new GradientDrawable();
        form.setColor(0xF212332A);              // fast blickdichtes Tannengrün
        form.setCornerRadius(px(6));
        form.setStroke(px(1), 0xAAD9BC7C);      // feine Goldlinie
        kasten.setBackgroundDrawable(form);

        kasten.addView(schrift(beschriftung, GOLD_HELL, 19, 0.06f, Typeface.NORMAL));

        if (zusatz.length() > 0) {
            kasten.addView(schrift(zusatz, 0xFFC9BFA8, 12, 0.10f, Typeface.NORMAL),
                    abstand(0, px(7), 0, 0));
        }

        kasten.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View angetippt) {
                oeffne(adresse);
            }
        });

        LinearLayout.LayoutParams masse = new LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT,
                LinearLayout.LayoutParams.WRAP_CONTENT);
        kasten.setLayoutParams(masse);
        return kasten;
    }

    /** Text mit Serifenschrift, Sperrung und Wunschfarbe. */
    private TextView schrift(String text, int farbe, int groesse, float sperrung, int schnitt) {
        TextView feld = new TextView(this);
        feld.setText(text);
        feld.setTextColor(farbe);
        feld.setTextSize(TypedValue.COMPLEX_UNIT_SP, groesse);
        feld.setTypeface(Typeface.create(Typeface.SERIF, schnitt));
        feld.setGravity(Gravity.CENTER);
        if (Build.VERSION.SDK_INT >= 21) {
            feld.setLetterSpacing(sperrung);
        }
        return feld;
    }

    /** Maße mit Rand (alles schon in Pixeln) für einen Text, der sich selbst bemisst. */
    private LinearLayout.LayoutParams abstand(int randLinks, int randOben, int randRechts,
                                              int randUnten) {
        LinearLayout.LayoutParams masse = new LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.WRAP_CONTENT,
                LinearLayout.LayoutParams.WRAP_CONTENT);
        masse.leftMargin = randLinks;
        masse.topMargin = randOben;
        masse.rightMargin = randRechts;
        masse.bottomMargin = randUnten;
        masse.gravity = Gravity.CENTER_HORIZONTAL;
        return masse;
    }

    private void oeffne(String adresse) {
        try {
            startActivity(new Intent(Intent.ACTION_VIEW, Uri.parse(adresse)));
        } catch (Exception fehler) {
            // Kommt praktisch nur vor, wenn auf dem Gerät kein Browser ist.
            Toast.makeText(this, "Kein Browser gefunden – bitte einen installieren.",
                    Toast.LENGTH_LONG).show();
        }
    }

    private int px(int dp) {
        return Math.round(dp * dichte);
    }
}
