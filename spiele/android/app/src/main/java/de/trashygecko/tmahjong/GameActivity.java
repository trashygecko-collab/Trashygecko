package de.trashygecko.tmahjong;

import android.annotation.SuppressLint;
import android.app.Activity;
import android.content.Intent;
import android.content.pm.ApplicationInfo;
import android.graphics.Color;
import android.net.Uri;
import android.os.Bundle;
import android.util.Log;
import android.webkit.JavascriptInterface;
import android.webkit.WebResourceError;
import android.webkit.WebResourceRequest;
import android.webkit.WebResourceResponse;
import android.webkit.WebView;
import android.webkit.WebViewClient;
import android.widget.TextView;
import android.widget.Toast;

import androidx.webkit.WebViewAssetLoader;

/**
 * Zeigt genau eine Spielseite.
 *
 * Vorher lagen Startseite, Katalog und Spiele im selben WebView von
 * {@link MainActivity}. Diese Activity übernimmt die Spielseiten: eigener
 * Verlauf, eigener Zurück-Weg, und der Katalog bleibt im Hintergrund erhalten
 * (Suche, Filter und Scrollposition gehen beim Spielen nicht verloren).
 *
 * Der Inhalt ist derselbe wie vorher: lokale Seiten kommen aus den Assets unter
 * {@code https://appassets.androidplatform.net/}, Online-Spiele von
 * {@code https://trashygecko.onrender.com/}. Am Spielverhalten ändert sich
 * nichts.
 */
public final class GameActivity extends Activity {
    /** Pfad relativ zum Web-Baum, z. B. {@code spiele/kartenspiele/uno/uno.html}. */
    public static final String EXTRA_PFAD = "de.trashygecko.tmahjong.SPIEL_PFAD";
    /** true: vom TrashyGecko-Server laden, false: aus dem App-Paket. */
    public static final String EXTRA_ONLINE = "de.trashygecko.tmahjong.SPIEL_ONLINE";

    private static final String TAG = "TrashyGeckoGame";
    private static final String MAHJONG_SEITE = "spiele/brettspiele/mahjong/mahjong.html";

    private WebView webView;
    private volatile String aktuelleSeite = "";

    @SuppressLint({"SetJavaScriptEnabled", "AddJavascriptInterface"})
    @Override protected void onCreate(Bundle zustand) {
        super.onCreate(zustand);
        WebBasis.vollbild(this);
        WebViewAssetLoader lader = HomepageAssets.baue(this);
        try {
            webView = new WebView(this);
        } catch (RuntimeException fehler) {
            zeigeWebViewFehler();
            return;
        }
        if ((getApplicationInfo().flags & ApplicationInfo.FLAG_DEBUGGABLE) != 0) {
            WebView.setWebContentsDebuggingEnabled(true);
        }
        WebBasis.einrichten(webView);
        webView.setWebChromeClient(WebBasis.protokoll(TAG));
        webView.setWebViewClient(new WebViewClient() {
            @Override public WebResourceResponse shouldInterceptRequest(WebView view, WebResourceRequest anfrage) {
                return lader.shouldInterceptRequest(anfrage.getUrl());
            }
            @Override public void onPageStarted(WebView view, String url, android.graphics.Bitmap symbol) {
                merkeSeite(url);
                if (istMahjong(url)) {
                    webView.addJavascriptInterface(new MahjongBruecke(), "TMahjong");
                } else {
                    webView.removeJavascriptInterface("TMahjong");
                }
            }
            @Override public void onPageFinished(WebView view, String url) {
                merkeSeite(url);
                Log.i(TAG, "Spielseite geladen: " + url);
            }
            @Override public void onReceivedError(WebView view, WebResourceRequest anfrage, WebResourceError fehler) {
                if (anfrage.isForMainFrame()) {
                    Log.e(TAG, "Seitenfehler " + fehler.getErrorCode() + " " + fehler.getDescription()
                            + ": " + anfrage.getUrl());
                }
            }
            @Override public boolean onRenderProcessGone(WebView view, android.webkit.RenderProcessGoneDetail detail) {
                Log.e(TAG, "WebView-Prozess beendet (Absturz=" + (detail != null && detail.didCrash()) + ")");
                if (webView != null && webView == view) {
                    webView.removeJavascriptInterface("TMahjong");
                    webView.destroy();
                    webView = null;
                }
                Toast.makeText(GameActivity.this, "Die Anzeige wurde neu gestartet.", Toast.LENGTH_SHORT).show();
                recreate();
                return true;
            }
            @Override public boolean shouldOverrideUrlLoading(WebView view, WebResourceRequest anfrage) {
                return leiteUm(anfrage.getUrl());
            }
            @Override public boolean shouldOverrideUrlLoading(WebView view, String url) {
                return leiteUm(Uri.parse(url));
            }
        });
        setContentView(webView);
        laden();
    }

    /** Lädt die Seite aus dem Intent; ohne gültigen Pfad zurück zur Startseite. */
    private void laden() {
        Intent absicht = getIntent();
        String pfad = absicht == null ? null : absicht.getStringExtra(EXTRA_PFAD);
        boolean online = absicht != null && absicht.getBooleanExtra(EXTRA_ONLINE, false);
        if (!WebBasis.istSichereSeite(pfad)) {
            Log.w(TAG, "Ungültiger Spielpfad: " + pfad);
            finish();
            return;
        }
        String ziel = (online ? WebBasis.ONLINE_ROOT : HomepageAssets.ROOT) + pfad.trim();
        webView.loadUrl(ziel);
    }

    /**
     * Link-Regel wie überall (siehe {@link WebNavigation}); Spielseiten bleiben
     * hier im WebView, damit ein Spiel sein eigenes Fenster behält.
     */
    private boolean leiteUm(Uri uri) {
        return WebNavigation.starte(this, WebNavigation.pruefe(uri, false), uri);
    }

    private void merkeSeite(String url) {
        try {
            String pfad = Uri.parse(url).getPath();
            aktuelleSeite = pfad == null ? "" : pfad.replaceFirst("^/", "");
        } catch (RuntimeException unbrauchbar) {
            aktuelleSeite = "";
        }
    }

    private boolean istMahjong(String url) {
        try {
            Uri uri = Uri.parse(url);
            return "https".equals(uri.getScheme()) && HomepageAssets.HOST.equals(uri.getHost())
                    && ("/" + MAHJONG_SEITE).equals(uri.getPath());
        } catch (RuntimeException unbrauchbar) {
            return false;
        }
    }

    private void zeigeWebViewFehler() {
        TextView meldung = new TextView(this);
        meldung.setText("TrashyGecko benötigt eine verfügbare Android System WebView. "
                + "Aktiviere oder aktualisiere WebView/Chrome und starte die App erneut.");
        meldung.setTextColor(Color.WHITE);
        meldung.setPadding(32, 32, 32, 32);
        setContentView(meldung);
    }

    @Override public void onWindowFocusChanged(boolean hatFokus) {
        super.onWindowFocusChanged(hatFokus);
        if (hatFokus) {
            WebBasis.vollbild(this);
        }
    }

    /** „SPIEL VERLASSEN“ im Mahjong ruft das hier auf. */
    private final class MahjongBruecke {
        @JavascriptInterface public void verlassen() {
            if (!MAHJONG_SEITE.equals(aktuelleSeite)) {
                return;
            }
            runOnUiThread(() -> {
                if (webView != null && webView.canGoBack()) {
                    webView.goBack();
                } else {
                    finish();
                }
            });
        }
    }

    @Override protected void onNewIntent(Intent absicht) {
        super.onNewIntent(absicht);
        setIntent(absicht);
        if (webView != null) {
            laden();
        }
    }

    @Override public void onBackPressed() {
        if (webView != null && webView.canGoBack()) {
            webView.goBack();
        } else {
            finish();
        }
    }

    @Override protected void onPause() {
        if (webView != null) {
            webView.onPause();
        }
        super.onPause();
    }

    @Override protected void onResume() {
        super.onResume();
        if (webView != null) {
            webView.onResume();
        }
    }

    @Override protected void onDestroy() {
        if (webView != null) {
            webView.removeJavascriptInterface("TMahjong");
            webView.stopLoading();
            webView.destroy();
            webView = null;
        }
        super.onDestroy();
    }
}
