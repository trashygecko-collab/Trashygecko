package de.trashygecko.tmahjong;

import android.annotation.SuppressLint;
import android.app.Activity;
import android.content.Intent;
import android.content.pm.ApplicationInfo;
import android.net.Uri;
import android.os.Bundle;
import android.util.Log;
import android.webkit.ConsoleMessage;
import android.webkit.JavascriptInterface;
import android.webkit.WebChromeClient;
import android.webkit.WebResourceError;
import android.webkit.WebResourceRequest;
import android.webkit.WebResourceResponse;
import android.webkit.WebSettings;
import android.webkit.WebView;
import android.webkit.WebViewClient;

import java.io.ByteArrayInputStream;
import java.io.IOException;
import java.io.InputStream;
import java.util.HashMap;
import java.util.Map;

import androidx.core.view.WindowCompat;
import androidx.core.view.WindowInsetsCompat;
import androidx.core.view.WindowInsetsControllerCompat;
import androidx.webkit.WebViewAssetLoader;

/**
 * TMahjong - das Mahjong-Spiel aus spiele/mahjong in einer WebView.
 *
 * Die Spieldateien liegen unter assets/www und werden ueber den
 * WebViewAssetLoader als https://appassets.androidplatform.net/... ausgeliefert.
 * Damit hat das Spiel eine echte Herkunft; localStorage und die
 * Sicherheitsregeln der Webfassung greifen unveraendert.
 *
 * Die einzige Bruecke ins Spiel ist TMahjong.verlassen() - der Knopf
 * "SPIEL VERLASSEN" ganz unten in den Einstellungen.
 */
public class MainActivity extends Activity {

    public static final String START_URL =
            "https://appassets.androidplatform.net/assets/www/brettspiele/mahjong/mahjong.html";

    private static final String HERKUNFT = "appassets.androidplatform.net";

    private static final String TAG = "TMahjong";

    private WebView webView;

    @SuppressLint({"SetJavaScriptEnabled", "AddJavascriptInterface"})
    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);

        vollbild();

        final WebViewAssetLoader lader = new WebViewAssetLoader.Builder()
                .addPathHandler("/assets/", new WebViewAssetLoader.AssetsPathHandler(this))
                // Die Spielseiten laden gemeinsame Dateien mit absolutem Pfad:
                // /spiele/schrift.css und /spiele/schrift/*.woff2.  Sie liegen in
                // den Assets unter spiele/ und bekommen hier dieselbe Adresse wie
                // auf der Homepage - kein 404, kein Systemschrift-Rueckfall.
                .addPathHandler("/spiele/", new WebViewAssetLoader.PathHandler() {
                    @Override
                    public WebResourceResponse handle(String pfad) {
                        return spieleDatei(pfad);
                    }
                })
                .build();

        webView = new WebView(this);

        // In Debug-Builds mitschreiben, was im Spiel passiert (chrome://inspect
        // am Rechner und Logcat). Release-Builds bleiben unveraendert.
        if ((getApplicationInfo().flags & ApplicationInfo.FLAG_DEBUGGABLE) != 0) {
            WebView.setWebContentsDebuggingEnabled(true);
        }

        WebSettings einstellungen = webView.getSettings();
        einstellungen.setJavaScriptEnabled(true);
        einstellungen.setDomStorageEnabled(true);      // Spielstand (localStorage)
        einstellungen.setAllowFileAccess(false);
        einstellungen.setAllowContentAccess(false);
        einstellungen.setSupportZoom(false);
        einstellungen.setBuiltInZoomControls(false);
        einstellungen.setTextZoom(100);                // Systemschrift skaliert nicht mit

        webView.setBackgroundColor(0xFF0B241D);
        webView.setOverScrollMode(WebView.OVER_SCROLL_NEVER);

        webView.setWebViewClient(new WebViewClient() {
            @Override
            public WebResourceResponse shouldInterceptRequest(WebView view, WebResourceRequest anfrage) {
                return lader.shouldInterceptRequest(anfrage.getUrl());
            }

            @Override
            public void onPageFinished(WebView view, String url) {
                Log.i(TAG, "Seite geladen: " + url);
            }

            @Override
            public void onReceivedError(WebView view, WebResourceRequest anfrage,
                                        WebResourceError fehler) {
                Log.e(TAG, "Ladefehler " + fehler.getErrorCode() + " "
                        + fehler.getDescription() + ": " + anfrage.getUrl());
            }

            @Override
            public void onReceivedHttpError(WebView view, WebResourceRequest anfrage,
                                            WebResourceResponse antwort) {
                Log.e(TAG, "HTTP " + antwort.getStatusCode() + " " + antwort.getReasonPhrase()
                        + ": " + anfrage.getUrl());
            }

            @Override
            public boolean shouldOverrideUrlLoading(WebView view, WebResourceRequest anfrage) {
                Uri ziel = anfrage.getUrl();

                // Alles aus den App-Assets bleibt in der WebView, externe Links
                // (z.B. "ZURUECK ZU DEN SPIELEN") gehen an den Browser.
                if (HERKUNFT.equals(ziel.getHost())) {
                    return false;
                }

                startActivity(new Intent(Intent.ACTION_VIEW, ziel));

                return true;
            }
        });

        webView.setWebChromeClient(new WebChromeClient() {
            @Override
            public boolean onConsoleMessage(ConsoleMessage meldung) {
                Log.i(TAG, "Konsole: " + meldung.message() + " ("
                        + meldung.sourceId() + ":" + meldung.lineNumber() + ")");

                return true;
            }
        });

        webView.addJavascriptInterface(new Bruecke(), "TMahjong");

        setContentView(webView);
        webView.loadUrl(START_URL);
    }

    /**
     * Liefert eine Datei aus assets/spiele/ fuer die Adresse /spiele/... .
     *
     * Je nach Fassung des WebViewAssetLoader kommt der Pfad mit oder ohne
     * "/spiele/" an; beides fuehrt auf denselben Assetnamen.
     */
    private WebResourceResponse spieleDatei(String pfad) {
        if (pfad == null || pfad.contains("..")) {
            return null;
        }

        String name = pfad;
        int stelle = name.indexOf("/spiele/");

        if (stelle >= 0) {
            name = name.substring(stelle + "/spiele/".length());
        }

        while (name.startsWith("/")) {
            name = name.substring(1);
        }

        if (name.isEmpty()) {
            return null;
        }

        String typ = "application/octet-stream";
        String kodierung = null;

        if (name.endsWith(".css")) {
            typ = "text/css";
            kodierung = "utf-8";
        } else if (name.endsWith(".woff2")) {
            typ = "font/woff2";
        } else if (name.endsWith(".js")) {
            typ = "application/javascript";
            kodierung = "utf-8";
        }

        Map<String, String> kopf = new HashMap<>();
        kopf.put("Cache-Control", "max-age=86400");

        try {
            InputStream strom = getAssets().open("spiele/" + name);

            return new WebResourceResponse(typ, kodierung, 200, "OK", kopf, strom);
        } catch (IOException fehler) {
            Log.w(TAG, "Nicht in den Assets: spiele/" + name);

            return new WebResourceResponse("text/plain", "utf-8", 404, "Nicht gefunden",
                    kopf, new ByteArrayInputStream(new byte[0]));
        }
    }

    /** Randlos und ohne Systemleisten - die App laeuft im Vollbild. */
    private void vollbild() {
        WindowCompat.setDecorFitsSystemWindows(getWindow(), false);

        WindowInsetsControllerCompat steuerung =
                WindowCompat.getInsetsController(getWindow(), getWindow().getDecorView());

        steuerung.hide(WindowInsetsCompat.Type.systemBars());
        steuerung.setSystemBarsBehavior(
                WindowInsetsControllerCompat.BEHAVIOR_SHOW_TRANSIENT_BARS_BY_SWIPE);
    }

    @Override
    public void onWindowFocusChanged(boolean hatFokus) {
        super.onWindowFocusChanged(hatFokus);

        // Nach Wischen oder einem Dialog die Leisten wieder ausblenden.
        if (hatFokus) {
            vollbild();
        }
    }

    /** Wird aus dem Spiel aufgerufen: window.TMahjong.verlassen() */
    private class Bruecke {

        @JavascriptInterface
        public void verlassen() {
            runOnUiThread(new Runnable() {
                @Override
                public void run() {
                    finishAffinity();
                }
            });
        }
    }

    @Override
    public void onBackPressed() {
        // Zurueck beendet die App - im Vollbild gibt es keine Leiste zu verlassen.
        finishAffinity();
    }

    @Override
    protected void onPause() {
        super.onPause();

        if (webView != null) {
            webView.onPause();
        }
    }

    @Override
    protected void onResume() {
        super.onResume();

        if (webView != null) {
            webView.onResume();
        }
    }

    @Override
    protected void onDestroy() {
        if (webView != null) {
            webView.destroy();
            webView = null;
        }

        super.onDestroy();
    }
}
