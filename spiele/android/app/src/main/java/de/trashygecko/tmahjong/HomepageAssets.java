package de.trashygecko.tmahjong;

import android.content.Context;
import android.net.Uri;
import android.util.Log;
import android.webkit.WebResourceResponse;
import androidx.webkit.WebViewAssetLoader;

import java.io.ByteArrayInputStream;
import java.io.IOException;
import java.net.URLConnection;
import java.util.Collections;
import java.util.HashMap;
import java.util.Locale;
import java.util.Map;

/** Liefert die eingebettete Homepage unter einer echten https-Herkunft aus. */
final class HomepageAssets {
    static final String HOST = "appassets.androidplatform.net";
    static final String ROOT = "https://" + HOST + "/";

    private static final String TAG = "TrashyGeckoAssets";
    private static final Map<String, String> MIME = mimeTypes();

    private HomepageAssets() {
    }

    static WebViewAssetLoader baue(Context context) {
        return new WebViewAssetLoader.Builder()
                .addPathHandler("/assets/", new WebViewAssetLoader.AssetsPathHandler(context))
                .addPathHandler("/", pfad -> lade(context, pfad))
                .build();
    }

    static boolean istLokal(String url) {
        try {
            Uri uri = Uri.parse(url);
            return "https".equals(uri.getScheme()) && HOST.equals(uri.getHost());
        } catch (RuntimeException unbrauchbar) {
            return false;
        }
    }

    static WebResourceResponse lade(Context context, String angefragt) {
        String pfad = angefragt == null ? "" : angefragt;
        while (pfad.startsWith("/")) pfad = pfad.substring(1);
        if (pfad.isEmpty() || pfad.endsWith("/")) pfad += "index.html";
        for (String teil : pfad.split("/")) {
            if (teil.isEmpty() || ".".equals(teil) || "..".equals(teil) || teil.indexOf('\\') >= 0) {
                return nichtGefunden();
            }
        }
        String endung = pfad.lastIndexOf('.') < 0 ? "" : pfad.substring(pfad.lastIndexOf('.') + 1).toLowerCase(Locale.ROOT);
        String typ = MIME.get(endung);
        if (typ == null) typ = URLConnection.guessContentTypeFromName(pfad);
        if (typ == null) typ = "application/octet-stream";
        String kodierung = typ.startsWith("text/") || typ.contains("javascript") || typ.contains("json")
                || typ.contains("xml") || typ.contains("svg") ? "utf-8" : null;
        try {
            Map<String, String> kopf = Collections.singletonMap("Cache-Control", "no-cache");
            return new WebResourceResponse(typ, kodierung, 200, "OK", kopf,
                    context.getAssets().open("www/" + pfad));
        } catch (IOException fehlt) {
            Log.w(TAG, "Lokale Datei fehlt: www/" + pfad);
            return nichtGefunden();
        }
    }

    static WebResourceResponse nichtGefunden() {
        try {
            return new WebResourceResponse("text/plain", "utf-8", 404, "Not Found",
                    Collections.<String, String>emptyMap(), new ByteArrayInputStream(new byte[0]));
        } catch (RuntimeException unmoeglich) {
            return null;
        }
    }

    private static Map<String, String> mimeTypes() {
        Map<String, String> werte = new HashMap<>();
        werte.put("html", "text/html"); werte.put("htm", "text/html"); werte.put("css", "text/css");
        werte.put("js", "application/javascript"); werte.put("mjs", "application/javascript");
        werte.put("json", "application/json"); werte.put("svg", "image/svg+xml");
        werte.put("wasm", "application/wasm"); werte.put("woff", "font/woff");
        werte.put("woff2", "font/woff2"); werte.put("ttf", "font/ttf"); werte.put("otf", "font/otf");
        werte.put("mp3", "audio/mpeg"); werte.put("ogg", "audio/ogg"); werte.put("wav", "audio/wav");
        werte.put("mp4", "video/mp4"); werte.put("webm", "video/webm"); werte.put("png", "image/png");
        werte.put("jpg", "image/jpeg"); werte.put("jpeg", "image/jpeg"); werte.put("gif", "image/gif");
        werte.put("webp", "image/webp"); werte.put("ico", "image/x-icon"); werte.put("pdf", "application/pdf");
        werte.put("txt", "text/plain");
        return Collections.unmodifiableMap(werte);
    }
}
