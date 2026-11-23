package de.trashygecko.tmahjong;

import android.annotation.SuppressLint;
import android.app.Activity;
import android.content.pm.ApplicationInfo;
import android.graphics.Color;
import android.net.Uri;
import android.os.Bundle;
import android.util.Log;
import android.webkit.RenderProcessGoneDetail;
import android.webkit.WebResourceError;
import android.webkit.WebResourceRequest;
import android.webkit.WebResourceResponse;
import android.webkit.WebView;
import android.webkit.WebViewClient;
import android.widget.TextView;
import android.widget.Toast;

import androidx.webkit.WebViewAssetLoader;

/**
 * Die echte TrashyGecko-Startseite aus dem App-Paket (D-12).
 *
 * Die Website bleibt die Startseite der Website – die App zeigt sie nur an, und
 * zwar ausschließlich über „Extras → Homepage“. Der Startbildschirm der App
 * bleibt nativ (D-02). Die Seite liegt als `homepage.html` in den Assets
 * (kopiert aus `index.html`), `main.css` liegt daneben; der Weg zurück führt
 * über „Zurück“ in die App.
 */
public final class HomepageActivity extends Activity {
    private static final String TAG = "TrashyGeckoHomepage";
    private static final String SEITE = "homepage.html";

    private WebView webView;

    @SuppressLint("SetJavaScriptEnabled")
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
            @Override public void onPageFinished(WebView view, String url) {
                Log.i(TAG, "Homepage geladen: " + url);
            }
            @Override public void onReceivedError(WebView view, WebResourceRequest anfrage, WebResourceError fehler) {
                if (anfrage.isForMainFrame()) {
                    Log.e(TAG, "Seitenfehler " + fehler.getErrorCode() + " " + fehler.getDescription()
                            + ": " + anfrage.getUrl());
                }
            }
            // Ab API 26: ein abgestürzter WebView-Prozess darf die App nicht mit einem
            // schwarzen Bild zurücklassen. Ansicht verwerfen und neu aufbauen.
            @Override public boolean onRenderProcessGone(WebView view, RenderProcessGoneDetail detail) {
                Log.e(TAG, "WebView-Prozess beendet (Absturz=" + (detail != null && detail.didCrash()) + ")");
                if (webView != null && webView == view) {
                    webView.destroy();
                    webView = null;
                }
                Toast.makeText(HomepageActivity.this, "Die Anzeige wurde neu gestartet.", Toast.LENGTH_SHORT).show();
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
        webView.loadUrl(HomepageAssets.ROOT + SEITE);
    }

    /**
     * Dieselbe Link-Regel wie überall (siehe {@link WebNavigation}), aber mit
     * Spielumleitung: „Spiele“ auf der Homepage öffnet die native Liste.
     */
    private boolean leiteUm(Uri uri) {
        return WebNavigation.starte(this, WebNavigation.pruefe(uri, true), uri);
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
            webView.stopLoading();
            webView.destroy();
            webView = null;
        }
        super.onDestroy();
    }
}
