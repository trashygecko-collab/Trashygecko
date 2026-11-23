package de.trashygecko.tmahjong;

import android.annotation.SuppressLint;
import android.app.Activity;
import android.util.Log;
import android.webkit.ConsoleMessage;
import android.webkit.WebChromeClient;
import android.webkit.WebSettings;
import android.webkit.WebView;

import androidx.core.view.WindowCompat;
import androidx.core.view.WindowInsetsCompat;
import androidx.core.view.WindowInsetsControllerCompat;

/**
 * Gemeinsame Einrichtung für alle WebView-Wirte (Start/Katalog und Spielseiten).
 *
 * Die Einstellungen sind bewusst an einer Stelle: Startseite, Katalog und Spiele
 * müssen dieselben Rechte und denselben Cache-Zustand haben, sonst zeigt die eine
 * Ansicht nach einem App-Update etwas anderes als die andere.
 */
final class WebBasis {
    /** Ziel der Online-Spiele (D-13: fest, nicht einstellbar). */
    static final String ONLINE_HOST = "trashygecko.onrender.com";
    static final String ONLINE_ROOT = "https://" + ONLINE_HOST + "/";

    private WebBasis() {
    }

    /** Konsolenmeldungen der Seiten landen im Log – für Start und Spiel gleich. */
    static WebChromeClient protokoll(final String tag) {
        return new WebChromeClient() {
            @Override public boolean onConsoleMessage(ConsoleMessage meldung) {
                Log.i(tag, meldung.message() + " (" + meldung.sourceId() + ":" + meldung.lineNumber() + ")");
                return true;
            }
        };
    }

    @SuppressLint({"SetJavaScriptEnabled"})
    static void einrichten(WebView view) {
        WebSettings einstellungen = view.getSettings();
        einstellungen.setJavaScriptEnabled(true);
        einstellungen.setDomStorageEnabled(true);
        einstellungen.setAllowFileAccess(false);
        einstellungen.setAllowContentAccess(false);
        einstellungen.setSupportZoom(false);
        einstellungen.setBuiltInZoomControls(false);
        einstellungen.setTextZoom(100);
        // Die Seiten liegen im APK und ändern sich nur mit einem App-Update.
        // Ein Cache würde nach dem Update noch die alten Dateien zeigen.
        einstellungen.setCacheMode(WebSettings.LOAD_NO_CACHE);
        einstellungen.setJavaScriptCanOpenWindowsAutomatically(false);
        einstellungen.setSupportMultipleWindows(false);
        view.setBackgroundColor(0xFF0B241D);
        view.setOverScrollMode(WebView.OVER_SCROLL_NEVER);
    }

    /** Randlos, Systemleisten nur bei Wischen sichtbar. */
    static void vollbild(Activity activity) {
        WindowCompat.setDecorFitsSystemWindows(activity.getWindow(), false);
        WindowInsetsControllerCompat steuerung =
                WindowCompat.getInsetsController(activity.getWindow(), activity.getWindow().getDecorView());
        steuerung.hide(WindowInsetsCompat.Type.systemBars());
        steuerung.setSystemBarsBehavior(WindowInsetsControllerCompat.BEHAVIOR_SHOW_TRANSIENT_BARS_BY_SWIPE);
    }

    /**
     * Nur einfache, relative HTML-Seiten sind erlaubt: kein "..", kein "%",
     * keine absoluten Pfade. Was durch diese Prüfung fällt, wird nicht geladen.
     */
    static boolean istSichereSeite(String pfad) {
        if (pfad == null) {
            return false;
        }
        String wert = pfad.trim().replace('\\', '/');
        if (wert.isEmpty() || wert.length() > 240) {
            return false;
        }
        if (wert.startsWith("/") || wert.contains("..") || wert.indexOf('%') >= 0) {
            return false;
        }
        return wert.matches("[A-Za-z0-9._/-]+\\.html");
    }
}
