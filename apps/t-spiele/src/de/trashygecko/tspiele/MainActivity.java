package de.trashygecko.tspiele;

import android.app.Activity;
import android.content.Intent;
import android.content.pm.PackageManager;
import android.graphics.Color;
import android.graphics.Typeface;
import android.graphics.drawable.GradientDrawable;
import android.net.Uri;
import android.os.Bundle;
import android.util.TypedValue;
import android.view.Gravity;
import android.view.View;
import android.widget.LinearLayout;
import android.widget.TextView;
import android.widget.Toast;

/**
 * T.Spiele zeigt nur zwei Wege: oben den Connector mit Chat und Spielen,
 * unten die eigene Homepage. Beides wird im Browser geöffnet; die App selbst
 * bleibt im Hintergrund stehen, damit man zurückwechseln kann.
 */
public class MainActivity extends Activity {

    /** Chat und Spiele laufen über den Connector im Internet. */
    private static final String CONNECTOR = "https://trashygecko.onrender.com/connector";

    /** Die Homepage liegt bei GitHub Pages – bewusst nicht die Render-Adresse. */
    private static final String HOMEPAGE = "https://trashygecko-collab.github.io/Trashygecko/";

    private float dichte;

    @Override
    protected void onCreate(Bundle zustand) {
        super.onCreate(zustand);
        dichte = getResources().getDisplayMetrics().density;

        // Morgen-Lob: Kanal anlegen, Erlaubnis erfragen, Wecker stellen.
        LobWecker.kanalAnlegen(this);
        erlaubnisFuerHinweise();
        LobWecker.plane(this);

        LinearLayout seite = new LinearLayout(this);
        seite.setOrientation(LinearLayout.VERTICAL);
        seite.setBackgroundColor(Color.parseColor("#10121b"));
        seite.setPadding(px(20), px(28), px(20), px(16));

        TextView titel = new TextView(this);
        titel.setText("T.Spiele");
        titel.setTextColor(Color.parseColor("#4df5c2"));
        titel.setTextSize(TypedValue.COMPLEX_UNIT_SP, 30);
        titel.setTypeface(null, Typeface.BOLD);
        titel.setGravity(Gravity.CENTER);
        seite.addView(titel);

        TextView unter = new TextView(this);
        unter.setText("Chat & Spiele über den Connector");
        unter.setTextColor(Color.parseColor("#aab3c7"));
        unter.setTextSize(TypedValue.COMPLEX_UNIT_SP, 14);
        unter.setGravity(Gravity.CENTER);
        unter.setPadding(0, px(6), 0, px(26));
        seite.addView(unter);

        seite.addView(knopf("Connector öffnen", "Chat, Räume und alle Spiele", CONNECTOR, true));

        // Schiebt die Homepage an den unteren Rand des Bildschirms.
        seite.addView(new View(this), new LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT, 0, 1f));

        seite.addView(knopf("Meine Homepage", "trashygecko-collab.github.io/Trashygecko",
                HOMEPAGE, false));

        TextView fussnote = new TextView(this);
        fussnote.setText("Jeden Morgen um 7:30 ein Lob – abstellbar in den Android-Einstellungen.");
        fussnote.setTextSize(TypedValue.COMPLEX_UNIT_SP, 11);
        fussnote.setTextColor(Color.parseColor("#6f7891"));
        fussnote.setGravity(Gravity.CENTER);
        fussnote.setPadding(0, px(14), 0, 0);
        seite.addView(fussnote);

        setContentView(seite);
    }

    /** Ab Android 13 muss man Benachrichtigungen ausdrücklich erlauben. */
    private void erlaubnisFuerHinweise() {
        if (android.os.Build.VERSION.SDK_INT < 33) return;
        if (checkSelfPermission("android.permission.POST_NOTIFICATIONS")
                == PackageManager.PERMISSION_GRANTED) return;
        requestPermissions(new String[] {"android.permission.POST_NOTIFICATIONS"}, 1);
    }

    /** Baut einen antippbaren Kasten mit Beschriftung und kleiner Zeile darunter. */
    private View knopf(String beschriftung, String zusatz, final String adresse, boolean gross) {
        LinearLayout kasten = new LinearLayout(this);
        kasten.setOrientation(LinearLayout.VERTICAL);
        kasten.setGravity(Gravity.CENTER);
        kasten.setPadding(px(16), px(gross ? 22 : 16), px(16), px(gross ? 22 : 16));

        GradientDrawable form = new GradientDrawable();
        form.setColor(Color.parseColor(gross ? "#17252a" : "#1c2030"));
        form.setCornerRadius(px(14));
        form.setStroke(px(1), Color.parseColor(gross ? "#4df5c2" : "#394158"));
        kasten.setBackgroundDrawable(form);

        TextView name = new TextView(this);
        name.setText(beschriftung);
        name.setTextSize(TypedValue.COMPLEX_UNIT_SP, gross ? 20 : 16);
        name.setTextColor(Color.parseColor(gross ? "#4df5c2" : "#dfe6f5"));
        name.setTypeface(null, Typeface.BOLD);
        name.setGravity(Gravity.CENTER);
        kasten.addView(name);

        TextView klein = new TextView(this);
        klein.setText(zusatz);
        klein.setTextSize(TypedValue.COMPLEX_UNIT_SP, 13);
        klein.setTextColor(Color.parseColor("#8d97ad"));
        klein.setGravity(Gravity.CENTER);
        klein.setPadding(0, px(5), 0, 0);
        kasten.addView(klein);

        kasten.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View angetippt) {
                oeffne(adresse);
            }
        });
        kasten.setLayoutParams(new LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT,
                LinearLayout.LayoutParams.WRAP_CONTENT));
        return kasten;
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
