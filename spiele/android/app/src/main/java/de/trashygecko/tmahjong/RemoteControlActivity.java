package de.trashygecko.tmahjong;

import android.annotation.SuppressLint;
import android.app.Activity;
import android.content.pm.ApplicationInfo;
import android.net.Uri;
import android.os.Bundle;
import android.os.Handler;
import android.os.Looper;
import android.util.Log;
import android.webkit.JavascriptInterface;
import android.webkit.WebChromeClient;
import android.webkit.ConsoleMessage;
import android.webkit.WebResourceRequest;
import android.webkit.WebView;
import android.webkit.WebViewClient;
import android.widget.TextView;

import androidx.webkit.WebViewAssetLoader;

import de.trashygecko.tmahjong.remote.RemoteBridgeService;

import org.json.JSONObject;

import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;
import java.util.concurrent.FutureTask;
import java.util.concurrent.ConcurrentHashMap;
import java.util.Map;

/** Isolated WebView host for the standalone Trashy Remote module. */
public final class RemoteControlActivity extends Activity {
    private static final String TAG = "TrashyRemote";
    private static final String HOST = "appassets.androidplatform.net";
    private static final String PATH = "/assets/www/fernbedienung/";
    private static final String START_URL = "https://" + HOST + PATH + "index.html";
    private final Handler mainHandler = new Handler(Looper.getMainLooper());
    private final ExecutorService requests = Executors.newFixedThreadPool(3);
    private final Map<String, Future<?>> inFlight = new ConcurrentHashMap<>();
    private WebView webView;
    private volatile boolean trustedPage;
    private volatile boolean destroyed;
    private RemoteBridgeService service;
    private NativeBridge nativeBridge;

    @SuppressLint({"SetJavaScriptEnabled", "AddJavascriptInterface"})
    @Override protected void onCreate(Bundle state) {
        super.onCreate(state);
        service = new RemoteBridgeService(this);
        WebViewAssetLoader loader = new WebViewAssetLoader.Builder()
                .addPathHandler("/assets/", new WebViewAssetLoader.AssetsPathHandler(this))
                .build();
        try { webView = new WebView(this); }
        catch (RuntimeException error) {
            TextView message = new TextView(this);
            message.setText("Die Fernbedienung benötigt eine verfügbare Android System WebView. Aktualisiere WebView/Chrome, sofern dein Gerät dies unterstützt.");
            message.setPadding(32, 32, 32, 32);
            setContentView(message);
            return;
        }
        if ((getApplicationInfo().flags & ApplicationInfo.FLAG_DEBUGGABLE) != 0) WebView.setWebContentsDebuggingEnabled(true);
        webView.setBackgroundColor(0xff10140f);
        webView.setOverScrollMode(WebView.OVER_SCROLL_NEVER);
        webView.getSettings().setJavaScriptEnabled(true);
        webView.getSettings().setDomStorageEnabled(true);
        webView.getSettings().setAllowFileAccess(false);
        webView.getSettings().setAllowContentAccess(false);
        webView.getSettings().setSupportZoom(false);
        // Oberfläche liegt im APK; nach einem App-Update darf kein alter Cache greifen.
        webView.getSettings().setCacheMode(android.webkit.WebSettings.LOAD_NO_CACHE);
        webView.getSettings().setJavaScriptCanOpenWindowsAutomatically(false);
        webView.getSettings().setSupportMultipleWindows(false);
        nativeBridge = new NativeBridge();
        webView.addJavascriptInterface(nativeBridge, "TGRemoteNative");
        webView.setWebChromeClient(new WebChromeClient() {
            @Override public boolean onConsoleMessage(ConsoleMessage message) {
                Log.i(TAG, "Web-Konsole: " + message.message() + " (" + message.sourceId() + ":" + message.lineNumber() + ")");
                return true;
            }
        });
        webView.setWebViewClient(new WebViewClient() {
            @Override public android.webkit.WebResourceResponse shouldInterceptRequest(WebView view, WebResourceRequest request) {
                return loader.shouldInterceptRequest(request.getUrl());
            }
            @Override public void onPageStarted(WebView view, String url, android.graphics.Bitmap favicon) {
                trustedPage = isRemoteAsset(url);
            }
            @Override public void onPageFinished(WebView view, String url) {
                trustedPage = isRemoteAsset(url);
                if (!trustedPage) Log.w(TAG, "Abweichende Remote-WebView-URL blockiert");
                else Log.i(TAG, "Fernbedienungsoberfläche geladen");
            }
            @Override public boolean onRenderProcessGone(WebView view, android.webkit.RenderProcessGoneDetail detail) {
                Log.e(TAG, "WebView-Prozess beendet (Absturz=" + (detail != null && detail.didCrash()) + ")");
                destroyed = true;
                trustedPage = false;
                if (webView != null && webView == view) { webView.removeJavascriptInterface("TGRemoteNative"); webView.destroy(); webView = null; }
                recreate();
                return true;
            }
            @Override public boolean shouldOverrideUrlLoading(WebView view, WebResourceRequest request) {
                Uri uri = request.getUrl();
                if (HOST.equals(uri.getHost()) && uri.getPath() != null && uri.getPath().startsWith(PATH)) return false;
                // The remote module has no external navigation requirement; never keep an untrusted origin in this WebView.
                return true;
            }
        });
        setContentView(webView);
        webView.loadUrl(START_URL);
    }

    private static boolean isRemoteAsset(String value) {
        try {
            Uri uri = Uri.parse(value);
            return "https".equals(uri.getScheme()) && HOST.equals(uri.getHost())
                    && uri.getPath() != null && uri.getPath().startsWith(PATH);
        } catch (RuntimeException ignored) { return false; }
    }

    private final class NativeBridge {
        @JavascriptInterface public void request(String requestId, String method, String payloadText) {
            if (destroyed || !trustedPage || requestId == null || !requestId.matches("[A-Za-z0-9_-]{1,80}")) return;
            if (payloadText == null || payloadText.length() > 32768) {
                deliver(requestId, error("Anfrage ist zu groß."));
                return;
            }
            final JSONObject payload;
            try { payload = new JSONObject(payloadText); }
            catch (Exception invalid) { deliver(requestId, error("Ungültige Anfrage.")); return; }
            FutureTask<Void> task = new FutureTask<>(() -> {
                JSONObject result;
                try { result = service.execute(method, payload); }
                catch (Throwable failure) { result = error("Native Netzwerkfunktion fehlgeschlagen."); }
                deliver(requestId, result);
                return null;
            });
            inFlight.put(requestId, task);
            requests.execute(task);
        }
    }

    private void deliver(String requestId, JSONObject result) {
        inFlight.remove(requestId);
        if (destroyed || webView == null || result == null) return;
        String script = "window.__trashyRemoteDeliver&&window.__trashyRemoteDeliver(" + JSONObject.quote(requestId) + "," + result.toString() + ");";
        mainHandler.post(() -> {
            if (!destroyed && webView != null && trustedPage) webView.evaluateJavascript(script, null);
        });
    }

    private static JSONObject error(String message) {
        try { return new JSONObject().put("ok", false).put("connected", false).put("message", message); }
        catch (Exception ignored) { return new JSONObject(); }
    }

    @Override protected void onPause() { if (webView != null) webView.onPause(); super.onPause(); }
    @Override protected void onResume() { super.onResume(); if (webView != null) webView.onResume(); }
    @Override public void onBackPressed() { finish(); }

    @Override protected void onDestroy() {
        destroyed = true;
        trustedPage = false;
        for (Future<?> task : inFlight.values()) task.cancel(true);
        inFlight.clear();
        requests.shutdownNow();
        if (service != null) service.shutdown();
        if (webView != null) { webView.removeJavascriptInterface("TGRemoteNative"); webView.stopLoading(); webView.destroy(); webView = null; }
        super.onDestroy();
    }
}


