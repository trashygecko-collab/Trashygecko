package de.trashygecko.tmahjong;

import android.content.Context;
import android.graphics.Typeface;
import android.graphics.drawable.GradientDrawable;
import android.util.TypedValue;
import android.widget.TextView;

/**
 * Farbbild und kleine Bausteine der nativen Ansichten.
 *
 * Die Werte standen vorher in {@link MainActivity} und {@link SpieleActivity} doppelt und
 * wichen leicht voneinander ab; Start und Spieleliste sollen gleich aussehen.
 */
final class Ansicht {
    static final int BG = 0xFF0B241D;
    static final int KARTE = 0xFF123028;
    static final int RAND = 0xFF2B4C40;
    static final int GOLD = 0xFFE8CD8E;
    static final int TEXT = 0xFFE9F1EB;
    static final int MUTED = 0xFF9AA7A0;

    private Ansicht() {
    }

    /** Abgerundete Fläche ohne Rand. */
    static GradientDrawable flaeche(int farbe, int radius) {
        GradientDrawable form = new GradientDrawable();
        form.setColor(farbe);
        form.setCornerRadius(radius);
        return form;
    }

    /** Abgerundete Fläche mit Kante (Karten); die Breite kommt in Pixeln herein. */
    static GradientDrawable flaeche(int farbe, int radius, int randfarbe, int randbreite) {
        GradientDrawable form = flaeche(farbe, radius);
        form.setStroke(randbreite, randfarbe);
        return form;
    }

    static TextView text(Context context, String wert, int groesse, int farbe, boolean fett) {
        TextView feld = new TextView(context);
        feld.setText(wert);
        feld.setTextSize(TypedValue.COMPLEX_UNIT_SP, groesse);
        feld.setTextColor(farbe);
        if (fett) feld.setTypeface(Typeface.DEFAULT_BOLD);
        return feld;
    }

    static int dp(Context context, int wert) {
        return Math.round(wert * context.getResources().getDisplayMetrics().density);
    }
}
