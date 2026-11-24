package de.trashygecko.tmahjong;

import android.content.Context;
import android.graphics.Canvas;
import android.graphics.Paint;
import android.graphics.Path;
import android.view.View;

import androidx.core.graphics.PathParser;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Locale;
import java.util.Map;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

/**
 * Ein Spielzeichen der Web-Halle, mit denselben Linien gezeichnet.
 *
 * Die Zeichen stehen in `spiele/spiele.html` (`SPIELSYMBOLE`, `IKONEN`) und kommen
 * über `katalog.json` herein. Gelesen werden nur `<path>`, `<circle>` und `<rect>` –
 * mehr benutzt die Vorlage nicht. Gezeichnet wird im 24 × 24-Raster mit runder
 * Linie; die Farbe gibt die Kategorie (Akzent) vor.
 */
final class Spielsymbol extends View {
    private static final float RASTER = 24f;
    private static final float STRICH = 1.7f;
    /** Einmal gelesene Zeichen – dieselbe Zeichenkette ergibt dieselben Pfade. */
    private static final Map<String, List<Path>> PFADE = new HashMap<>();
    private static final Pattern ELEMENT = Pattern.compile("<([a-zA-Z]+)([^>]*)>");
    private static final Pattern ATTRIBUT = Pattern.compile("([a-zA-Z-]+)=\"([^\"]*)\"");

    private List<Path> pfade;
    private final Paint stift = new Paint(Paint.ANTI_ALIAS_FLAG);

    Spielsymbol(Context context) {
        super(context);
        stift.setStyle(Paint.Style.STROKE);
        stift.setStrokeWidth(STRICH);
        stift.setStrokeCap(Paint.Cap.ROUND);
        stift.setStrokeJoin(Paint.Join.ROUND);
    }

    /** Setzt Zeichen und Farbe; eine leere Zeichenkette blendet die Ansicht leer. */
    void setze(String svg, int farbe) {
        pfade = svg == null || svg.isEmpty() ? null : pfadeFuer(svg);
        stift.setColor(farbe);
        invalidate();
    }

    @Override protected void onDraw(Canvas leinwand) {
        if (pfade == null || pfade.isEmpty()) return;
        float kante = Math.min(getWidth(), getHeight());
        if (kante <= 0) return;
        leinwand.save();
        leinwand.translate((getWidth() - kante) / 2f, (getHeight() - kante) / 2f);
        leinwand.scale(kante / RASTER, kante / RASTER);
        for (Path pfad : pfade) leinwand.drawPath(pfad, stift);
        leinwand.restore();
    }

    private static List<Path> pfadeFuer(String svg) {
        List<Path> gemerkt = PFADE.get(svg);
        if (gemerkt != null) return gemerkt;
        List<Path> pfade = new ArrayList<>();
        Matcher element = ELEMENT.matcher(svg);
        while (element.find()) {
            Map<String, String> werte = new HashMap<>();
            Matcher attribut = ATTRIBUT.matcher(element.group(2));
            while (attribut.find()) {
                werte.put(attribut.group(1).toLowerCase(Locale.ROOT), attribut.group(2));
            }
            String tag = element.group(1).toLowerCase(Locale.ROOT);
            try {
                if ("path".equals(tag) && werte.containsKey("d")) {
                    pfade.add(PathParser.createPathFromPathData(werte.get("d")));
                } else if ("circle".equals(tag)) {
                    Path kreis = new Path();
                    kreis.addCircle(zahl(werte, "cx"), zahl(werte, "cy"), zahl(werte, "r"),
                            Path.Direction.CW);
                    pfade.add(kreis);
                } else if ("rect".equals(tag)) {
                    float x = zahl(werte, "x");
                    float y = zahl(werte, "y");
                    float radius = zahl(werte, "rx");
                    Path kasten = new Path();
                    kasten.addRoundRect(x, y, x + zahl(werte, "width"), y + zahl(werte, "height"),
                            radius, radius, Path.Direction.CW);
                    pfade.add(kasten);
                }
            } catch (RuntimeException fehler) {
                // Ein unlesbares Zeichen darf die übrigen nicht aufhalten.
            }
        }
        PFADE.put(svg, pfade);
        return pfade;
    }

    private static float zahl(Map<String, String> werte, String name) {
        String wert = werte.get(name);
        if (wert == null || wert.isEmpty()) return 0f;
        try {
            return Float.parseFloat(wert);
        } catch (NumberFormatException fehler) {
            return 0f;
        }
    }
}
