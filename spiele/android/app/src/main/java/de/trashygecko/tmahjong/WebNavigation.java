package de.trashygecko.tmahjong;

import android.app.Activity;
import android.content.ActivityNotFoundException;
import android.content.Intent;
import android.net.Uri;
import android.widget.Toast;

import java.util.Locale;

/**
 * Die eine Stelle, an der entschieden wird, was mit einem Link passiert.
 *
 * Vorher stand dieselbe Regel in {@link MainActivity} und in
 * {@link GameActivity} – zweimal, mit feinen Unterschieden. Jetzt fragen beide
 * nur noch hier. Die Regeln:
 *
 * - eigene Seiten (`appassets…`) bleiben im WebView, außer sie brauchen eine
 *   native Ansicht: `/lokal`, `/extras`, `/WLAN-Fernbedienung/` und der Katalog;
 * - Spielseiten bekommen eine eigene Activity (nur wenn der Aufrufer das
 *   erlaubt – in einer Spielseite sollen Links im selben WebView bleiben);
 * - der eigene Server (`trashygecko.onrender.com`) bleibt in der App, nur
 *   fremde Adressen gehen an den Browser.
 */
final class WebNavigation {
    static final int BLEIBT = 0;
    static final int KATALOG = 1;
    static final int SPIEL = 2;
    static final int KONNECTOR_LOKAL = 3;
    static final int KONNECTOR_EXTRAS = 4;
    static final int KONNECTOR_ONLINE = 5;
    static final int BROWSER = 6;

    private static final String KATALOG_SEITE = "/spiele/spiele.html";

    /** Ergebnis der Prüfung: was zu tun ist und – bei einem Spiel – welches. */
    static final class Ziel {
        final int art;
        final String pfad;
        final boolean online;

        private Ziel(int art, String pfad, boolean online) {
            this.art = art;
            this.pfad = pfad;
            this.online = online;
        }
    }

    private static final Ziel BLEIBT_IM_WEBVIEW = new Ziel(BLEIBT, null, false);

    private WebNavigation() {
    }

    /**
     * @param spieleUmleiten true: Spielseiten in {@link GameActivity} öffnen
     *                       (Start/Katalog), false: im WebView lassen (Spielseite)
     */
    static Ziel pruefe(Uri uri, boolean spieleUmleiten) {
        if (uri == null) {
            return new Ziel(BROWSER, null, false);
        }
        String host = uri.getHost() == null ? "" : uri.getHost().toLowerCase(Locale.ROOT);
        String pfad = uri.getPath() == null ? "" : uri.getPath();

        if (HomepageAssets.HOST.equals(host)) {
            if (pfad.equalsIgnoreCase("/lokal") || pfad.equalsIgnoreCase("/lokal/")) {
                return new Ziel(KONNECTOR_LOKAL, null, false);
            }
            if (pfad.equalsIgnoreCase("/extras") || pfad.equalsIgnoreCase("/extras/")) {
                return new Ziel(KONNECTOR_EXTRAS, null, false);
            }
            // Die Universalfernbedienung ist ausschließlich über „Extras“ erreichbar.
            if (pfad.startsWith("/WLAN-Fernbedienung/")) {
                return new Ziel(KONNECTOR_EXTRAS, null, false);
            }
            if (KATALOG_SEITE.equalsIgnoreCase(pfad)) {
                return new Ziel(KATALOG, null, false);
            }
            if (spieleUmleiten && istSpielseite(pfad)) {
                return new Ziel(SPIEL, pfad.substring(1), false);
            }
            return BLEIBT_IM_WEBVIEW;
        }

        if (WebBasis.ONLINE_HOST.equals(host)) {
            if (pfad.equals("/connector") || pfad.startsWith("/connector/")) {
                return new Ziel(KONNECTOR_ONLINE, null, false);
            }
            if (spieleUmleiten && istSpielseite(pfad)) {
                return new Ziel(SPIEL, pfad.substring(1), true);
            }
            return BLEIBT_IM_WEBVIEW;
        }

        return new Ziel(BROWSER, null, false);
    }

    /** Führt aus, was {@link #pruefe} ergeben hat. false = im WebView laden. */
    static boolean starte(Activity activity, Ziel ziel, Uri uri) {
        switch (ziel.art) {
            case BLEIBT:
                return false;
            case KATALOG:
                activity.startActivity(new Intent(activity, SpieleActivity.class));
                return true;
            case SPIEL:
                Intent spiel = new Intent(activity, GameActivity.class);
                spiel.putExtra(GameActivity.EXTRA_PFAD, ziel.pfad);
                spiel.putExtra(GameActivity.EXTRA_ONLINE, ziel.online);
                activity.startActivity(spiel);
                return true;
            case KONNECTOR_LOKAL:
                return starteKonnector(activity, "local");
            case KONNECTOR_EXTRAS:
                return starteKonnector(activity, "extras");
            case KONNECTOR_ONLINE:
                return starteKonnector(activity, "online");
            default:
                if ("http".equalsIgnoreCase(uri.getScheme()) || "https".equalsIgnoreCase(uri.getScheme())) {
                    try {
                        activity.startActivity(new Intent(Intent.ACTION_VIEW, uri));
                    } catch (ActivityNotFoundException fehlt) {
                        Toast.makeText(activity, "Kein Browser für diesen Link verfügbar.", Toast.LENGTH_SHORT).show();
                    }
                }
                return true;
        }
    }

    private static boolean starteKonnector(Activity activity, String bildschirm) {
        activity.startActivity(new Intent(activity, ConnectorActivity.class)
                .putExtra(ConnectorActivity.EXTRA_INITIAL_SCREEN, bildschirm));
        return true;
    }

    /** Spielseite = HTML-Seite unter /spiele/, aber nicht der Katalog selbst. */
    private static boolean istSpielseite(String pfad) {
        return pfad.startsWith("/spiele/") && pfad.endsWith(".html")
                && !KATALOG_SEITE.equalsIgnoreCase(pfad)
                && WebBasis.istSichereSeite(pfad.substring(1));
    }
}
