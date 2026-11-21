package de.trashygecko.tmahjong;

import android.annotation.SuppressLint;
import android.app.Activity;
import android.app.AlertDialog;
import android.content.ActivityNotFoundException;
import android.content.Intent;
import android.content.pm.ApplicationInfo;
import android.graphics.Color;
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
import android.widget.TextView;
import android.widget.Toast;

import androidx.core.view.WindowCompat;
import androidx.core.view.WindowInsetsCompat;
import androidx.core.view.WindowInsetsControllerCompat;
import androidx.webkit.WebViewAssetLoader;

import java.io.InputStream;
import java.util.Locale;
import java.util.regex.Pattern;

/** Local-first Homepage and games host. All static content is loaded from the APK assets. */
public final class MainActivity extends Activity {
    public static final String EXTRA_START_PATH = "de.trashygecko.tmahjong.START_PATH";
    private static final String TAG = "TrashyGeckoHome";
    private static final String HOST = "appassets.androidplatform.net";
    private static final String LOCAL_ROOT = "https://" + HOST + "/";
    private static final String HOME_URL = LOCAL_ROOT + "app.html";
    private static final String ONLINE_HOST = "trashygecko.onrender.com";
    private static final String ONLINE_ROOT = "https://" + ONLINE_HOST;
    private static final Pattern SAFE_GAME = Pattern.compile("[A-Za-z0-9._/-]+\\.html");

    private WebView webView;
    private volatile boolean trustedLocalPage;
    private volatile String currentPagePath = "";
    private volatile boolean onlineGamePage;
    private AlertDialog modeDialog;

    @SuppressLint({"SetJavaScriptEnabled", "AddJavascriptInterface"})
    @Override protected void onCreate(Bundle state) {
        super.onCreate(state);
        vollbild();
        WebViewAssetLoader loader = HomepageAssets.baue(this);
        try {
            webView = new WebView(this);
        } catch (RuntimeException error) {
            showWebViewError();
            return;
        }
        if ((getApplicationInfo().flags & ApplicationInfo.FLAG_DEBUGGABLE) != 0) WebView.setWebContentsDebuggingEnabled(true);
        WebSettings settings = webView.getSettings();
        settings.setJavaScriptEnabled(true);
        settings.setDomStorageEnabled(true);
        settings.setAllowFileAccess(false);
        settings.setAllowContentAccess(false);
        settings.setSupportZoom(false);
        settings.setBuiltInZoomControls(false);
        settings.setTextZoom(100);
        // Die Seiten liegen im APK-Paket und ändern sich nur mit einem App-Update.
        // Ein WebView-Cache würde nach dem Update noch die alten Dateien zeigen.
        settings.setCacheMode(WebSettings.LOAD_NO_CACHE);
        settings.setJavaScriptCanOpenWindowsAutomatically(false);
        settings.setSupportMultipleWindows(false);
        webView.setBackgroundColor(0xFF0B241D);
        webView.setOverScrollMode(WebView.OVER_SCROLL_NEVER);
        webView.addJavascriptInterface(new HomepageBridge(), "TGApp");
        webView.setWebChromeClient(new WebChromeClient() {
            @Override public boolean onConsoleMessage(ConsoleMessage message) {
                Log.i(TAG, message.message() + " (" + message.sourceId() + ":" + message.lineNumber() + ")");
                return true;
            }
        });
        webView.setWebViewClient(new WebViewClient() {
            @Override public WebResourceResponse shouldInterceptRequest(WebView view, WebResourceRequest request) {
                return loader.shouldInterceptRequest(request.getUrl());
            }
            @Override public void onPageStarted(WebView view, String url, android.graphics.Bitmap icon) {
                trustedLocalPage = isLocalAsset(url);
                try { currentPagePath = Uri.parse(url).getPath() == null ? "" : Uri.parse(url).getPath(); } catch (RuntimeException ignored) { currentPagePath = ""; }
                if (trustedLocalPage) onlineGamePage = false;
                if (isMahjongPage(url)) webView.addJavascriptInterface(new MahjongBridge(), "TMahjong");
                else webView.removeJavascriptInterface("TMahjong");
            }
            @Override public void onPageFinished(WebView view, String url) {
                trustedLocalPage = isLocalAsset(url);
                try { currentPagePath = Uri.parse(url).getPath() == null ? "" : Uri.parse(url).getPath(); } catch (RuntimeException ignored) { currentPagePath = ""; }
                Log.i(TAG, "Seite geladen: " + url);
                if (trustedLocalPage && url.contains("/spiele/spiele.html")) injectGameModeChooser();
            }
            @Override public void onReceivedError(WebView view, WebResourceRequest request, WebResourceError error) {
                if (request.isForMainFrame()) Log.e(TAG, "Seitenfehler " + error.getErrorCode() + " " + error.getDescription() + ": " + request.getUrl());
            }
            // Ab API 26: ein abgestürzter WebView-Prozess darf die App nicht mit einem
            // schwarzen Bild zurücklassen. Ansicht verwerfen und den Startbildschirm neu aufbauen.
            @Override public boolean onRenderProcessGone(WebView view, android.webkit.RenderProcessGoneDetail detail) {
                Log.e(TAG, "WebView-Prozess beendet (Absturz=" + (detail != null && detail.didCrash()) + ")");
                if (webView != null && webView == view) {
                    webView.removeJavascriptInterface("TGApp");
                    webView.removeJavascriptInterface("TMahjong");
                    webView.destroy();
                    webView = null;
                }
                Toast.makeText(MainActivity.this, "Die Anzeige wurde neu gestartet.", Toast.LENGTH_SHORT).show();
                recreate();
                return true;
            }
            @Override public boolean shouldOverrideUrlLoading(WebView view, WebResourceRequest request) {
                return routeNavigation(request.getUrl());
            }
            @Override public boolean shouldOverrideUrlLoading(WebView view, String url) {
                return routeNavigation(Uri.parse(url));
            }
        });
        setContentView(webView);
        String path = getIntent() == null ? null : getIntent().getStringExtra(EXTRA_START_PATH);
        String target = localGameUrl(path);
        webView.loadUrl(target == null ? HOME_URL : target);
    }

    private boolean routeNavigation(Uri uri) {
        String host = uri.getHost() == null ? "" : uri.getHost().toLowerCase(Locale.ROOT);
        String path = uri.getPath() == null ? "" : uri.getPath();
        if (HOST.equals(host)) {
            // Startbildschirm-Kacheln, die eine native Ansicht brauchen.
            if (path.equalsIgnoreCase("/lokal") || path.equalsIgnoreCase("/lokal/")) {
                openConnector("local");
                return true;
            }
            if (path.equalsIgnoreCase("/extras") || path.equalsIgnoreCase("/extras/")) {
                openConnector("extras");
                return true;
            }
            // Die Universalfernbedienung ist ausschließlich über „Extras“ erreichbar.
            if (path.equalsIgnoreCase("/WLAN-Fernbedienung/index.html") || path.startsWith("/WLAN-Fernbedienung/")) {
                openConnector("extras");
                return true;
            }
            return false;
        }
        if (ONLINE_HOST.equals(host)) {
            if (path.equals("/connector") || path.startsWith("/connector/")) {
                openConnector("online");
                return true;
            }
            if (onlineGamePage) return false;
            // The user's own TrashyGecko site stays in-app; other external sites use the browser.
            if (path.startsWith("/spiele/")) {
                onlineGamePage = true;
                return false;
            }
        }
        if ("http".equalsIgnoreCase(uri.getScheme()) || "https".equalsIgnoreCase(uri.getScheme())) {
            try { startActivity(new Intent(Intent.ACTION_VIEW, uri)); }
            catch (ActivityNotFoundException unavailable) { Toast.makeText(this, "Kein Browser für diesen Link verfügbar.", Toast.LENGTH_SHORT).show(); }
            return true;
        }
        return true;
    }

    private void openConnector(String screen) {
        Intent intent = new Intent(this, ConnectorActivity.class);
        intent.putExtra(ConnectorActivity.EXTRA_INITIAL_SCREEN, screen);
        startActivity(intent);
    }

    private void injectGameModeChooser() {
        if (webView == null) return;
        // One listener on the document: the catalog rebuilds its tiles on every search,
        // so listeners bound to single tiles would be lost with the old nodes.
        String script = "(function(){if(window.__tgAndroidModeChooser)return;window.__tgAndroidModeChooser=1;"
                + "document.addEventListener('click',function(ev){var t=ev.target;"
                + "if(!t||!t.closest)return;var a=t.closest('a.spiel');"
                + "if(!a||!a.querySelector('.m-online')||!window.TGApp)return;ev.preventDefault();"
                + "var n=a.querySelector('.spiel-name');"
                + "window.TGApp.chooseOnlineGame(a.getAttribute('href'),n?n.textContent:'');"
                + "},true);})();";
        webView.evaluateJavascript(script, null);
    }

    private String safeGamePath(String value) {
        if (!trustedLocalPage || value == null || value.length() > 240) return null;
        if (!"/spiele/spiele.html".equals(currentPagePath)) return null;
        String path = value.trim().replace('\\', '/');
        while (path.startsWith("./")) path = path.substring(2);
        if (path.startsWith("/") || path.contains("..") || path.indexOf('%') >= 0 || !SAFE_GAME.matcher(path).matches()) return null;
        return "/spiele/" + path;
    }

    private String localGameUrl(String value) {
        if (value == null || value.length() > 240) return null;
        String path = value.trim().replace('\\', '/');
        if (path.startsWith("/")) path = path.substring(1);
        if (path.startsWith("spiele/")) path = path.substring("spiele/".length());
        if (path.contains("..") || path.indexOf('%') >= 0 || !SAFE_GAME.matcher(path).matches()) return null;
        return LOCAL_ROOT + "spiele/" + path;
    }

    private void showGameModeDialog(String path, String name) {
        runOnUiThread(() -> {
            // A quick double tap can arrive twice before the dialog is on screen.
            if (modeDialog != null && modeDialog.isShowing()) return;
            modeDialog = new AlertDialog.Builder(this)
                .setTitle(name == null || name.trim().isEmpty() ? "Spielmodus" : name.trim().substring(0, Math.min(80, name.trim().length())))
                .setMessage("Lokal/offline öffnet die eingebettete Spielseite. Online nutzt die vorhandenen Spielräume und den TrashyGecko-Server; dafür ist Internet erforderlich.")
                .setNegativeButton("Lokal / offline", (dialog, which) -> {
                    String local = LOCAL_ROOT + path.substring(1);
                    onlineGamePage = false;
                    webView.loadUrl(local);
                })
                .setPositiveButton("Online-Spiel", (dialog, which) -> {
                    onlineGamePage = true;
                    webView.loadUrl(ONLINE_ROOT + path);
                })
                .setNeutralButton("Abbrechen", null)
                .show();
        });
    }

    private static boolean isLocalAsset(String url) {
        try { Uri uri = Uri.parse(url); return "https".equals(uri.getScheme()) && HOST.equals(uri.getHost()); }
        catch (RuntimeException ignored) { return false; }
    }

    private static boolean isMahjongPage(String url) {
        try { return isLocalAsset(url) && "/spiele/brettspiele/mahjong/mahjong.html".equals(Uri.parse(url).getPath()); }
        catch (RuntimeException ignored) { return false; }
    }

    private void showWebViewError() {
        TextView message = new TextView(this);
        message.setText("TrashyGecko benötigt eine verfügbare Android System WebView. Aktiviere oder aktualisiere WebView/Chrome und starte die App erneut.");
        message.setTextColor(Color.WHITE);
        message.setPadding(32, 32, 32, 32);
        setContentView(message);
    }

    private void vollbild() {
        WindowCompat.setDecorFitsSystemWindows(getWindow(), false);
        WindowInsetsControllerCompat controller = WindowCompat.getInsetsController(getWindow(), getWindow().getDecorView());
        controller.hide(WindowInsetsCompat.Type.systemBars());
        controller.setSystemBarsBehavior(WindowInsetsControllerCompat.BEHAVIOR_SHOW_TRANSIENT_BARS_BY_SWIPE);
    }

    @Override public void onWindowFocusChanged(boolean hasFocus) {
        super.onWindowFocusChanged(hasFocus);
        if (hasFocus) vollbild();
    }

    private final class HomepageBridge {
        @JavascriptInterface public void chooseOnlineGame(String relativePath, String name) {
            String path = safeGamePath(relativePath);
            if (path != null) showGameModeDialog(path, name);
        }
    }

    private final class MahjongBridge {
        @JavascriptInterface public void verlassen() {
            if (!"/spiele/brettspiele/mahjong/mahjong.html".equals(currentPagePath)) return;
            runOnUiThread(() -> {
                if (webView != null && webView.canGoBack()) webView.goBack();
                else if (webView != null) webView.loadUrl(HOME_URL);
            });
        }
    }

    @Override protected void onNewIntent(Intent intent) {
        super.onNewIntent(intent);
        setIntent(intent);
        String target = localGameUrl(intent == null ? null : intent.getStringExtra(EXTRA_START_PATH));
        if (target != null && webView != null) webView.loadUrl(target);
    }
    @Override public void onBackPressed() {
        if (webView != null && webView.canGoBack()) webView.goBack();
        else finishAffinity();
    }

    @Override protected void onPause() { if (webView != null) webView.onPause(); super.onPause(); }
    @Override protected void onResume() { super.onResume(); if (webView != null) webView.onResume(); }
    @Override protected void onDestroy() {
        trustedLocalPage = false;
        if (modeDialog != null) { modeDialog.dismiss(); modeDialog = null; }
        if (webView != null) { webView.removeJavascriptInterface("TGApp"); webView.removeJavascriptInterface("TMahjong"); webView.stopLoading(); webView.destroy(); webView = null; }
        super.onDestroy();
    }
}