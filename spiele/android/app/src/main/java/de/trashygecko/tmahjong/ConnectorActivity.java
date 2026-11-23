package de.trashygecko.tmahjong;

import android.app.Activity;
import android.app.AlertDialog;
import android.content.Intent;
import android.content.pm.PackageInfo;
import android.content.pm.PackageManager;
import android.graphics.Color;
import android.graphics.Typeface;
import android.app.ActivityManager;
import android.net.ConnectivityManager;
import android.net.NetworkInfo;
import android.net.Uri;
import android.net.wifi.WifiManager;
import android.os.Bundle;
import android.os.Build;
import android.view.Gravity;
import android.view.View;
import android.webkit.JavascriptInterface;
import android.webkit.WebChromeClient;
import android.webkit.WebResourceRequest;
import android.webkit.WebResourceResponse;
import android.webkit.WebSettings;
import android.webkit.WebView;
import android.webkit.WebViewClient;

import androidx.webkit.WebViewAssetLoader;
import android.widget.Button;
import android.widget.EditText;
import android.widget.LinearLayout;
import android.widget.ScrollView;
import android.widget.TextView;

import org.json.JSONException;
import org.json.JSONObject;

import java.util.ArrayList;
import java.util.List;
import java.util.Locale;

/** Native entry screen for online play and lightweight app-to-app LAN rooms. */
public final class ConnectorActivity extends Activity implements LocalNetworkCoordinator.Listener {
    public static final String EXTRA_INITIAL_SCREEN = "de.trashygecko.tmahjong.INITIAL_SCREEN";
    private static final String ONLINE_URL = "https://trashygecko.onrender.com/connector";
    private static final int BG = Color.rgb(11, 36, 29);
    private static final int CARD = Color.rgb(20, 54, 44);
    private static final int TEXT = Color.rgb(239, 244, 234);
    private static final int MUTED = Color.rgb(174, 194, 181);
    private static final int ACCENT = Color.rgb(232, 205, 142);

    private LinearLayout root, body, roomList, playerList;
    private ScrollView bodyScroll;
    private TextView status, code, playersTitle, chatLog;
    private EditText chatInput;
    private WebView onlineView;
    private WebView spielView;
    private GalgenLokal galgen;
    private String gastStand = "{\"game\":null}";
    private LocalNetworkCoordinator network;
    private WifiManager.WifiLock wifiLock;
    private final List<LocalNetworkCoordinator.Room> rooms = new ArrayList<>();
    private boolean onOnlinePage;
    private boolean onHomeScreen = true;

    @Override protected void onCreate(Bundle state) {
        super.onCreate(state);
        getWindow().getDecorView().setSystemUiVisibility(
                View.SYSTEM_UI_FLAG_FULLSCREEN | View.SYSTEM_UI_FLAG_HIDE_NAVIGATION |
                        View.SYSTEM_UI_FLAG_IMMERSIVE_STICKY | View.SYSTEM_UI_FLAG_LAYOUT_STABLE);
        network = new LocalNetworkCoordinator(this, geraeteName(), this);
        buildShell();
        showInitialScreen(getIntent());
    }

    private void showInitialScreen(Intent intent) {
        String screen = intent == null ? "" : intent.getStringExtra(EXTRA_INITIAL_SCREEN);
        if ("local".equals(screen)) showLocal();
        else if ("online".equals(screen)) openOnline();
        else if ("extras".equals(screen)) showExtras();
        else showHome();
    }

    @Override protected void onNewIntent(Intent intent) {
        super.onNewIntent(intent);
        setIntent(intent);
        showInitialScreen(intent);
    }
    private void buildShell() {
        root = new LinearLayout(this); root.setOrientation(LinearLayout.VERTICAL); root.setPadding(dp(18), dp(14), dp(18), dp(14)); root.setBackgroundColor(BG);
        LinearLayout header = new LinearLayout(this); header.setGravity(Gravity.CENTER_VERTICAL); header.setOrientation(LinearLayout.HORIZONTAL);
        TextView brand = text("TRASHYGECKO", 15, ACCENT, true); header.addView(brand, new LinearLayout.LayoutParams(0, -2, 1));
        Button home = button("Start"); home.setOnClickListener(v -> showHome()); header.addView(home);
        Button extras = button("Extras"); extras.setOnClickListener(v -> showExtras()); header.addView(extras);
        Button online = button("Online"); online.setOnClickListener(v -> openOnline()); header.addView(online);
        root.addView(header);
        status = text("Bereit", 12, MUTED, false); status.setPadding(0, dp(8), 0, dp(8)); root.addView(status);
        bodyScroll = new ScrollView(this); bodyScroll.setFillViewport(true);
        body = new LinearLayout(this); body.setOrientation(LinearLayout.VERTICAL); body.setPadding(0, 0, 0, dp(12));
        bodyScroll.addView(body, new ScrollView.LayoutParams(-1, -2));
        root.addView(bodyScroll, new LinearLayout.LayoutParams(-1, 0, 1));
        setContentView(root);
    }

    private static String geraeteName() {
        String modell = Build.MODEL == null ? "" : Build.MODEL.trim();
        if (modell.isEmpty() || "sdk".equalsIgnoreCase(modell)) return "Spieler";
        return modell.length() > 24 ? modell.substring(0, 24) : modell;
    }

    private void showHome() {
        schliesseSpiel();
        onHomeScreen = true; onOnlinePage = false; showNativeBody(); body.removeAllViews();
        TextView title = text("Spielen mit TrashyGecko", 27, TEXT, true); title.setPadding(0, dp(24), 0, dp(4)); body.addView(title);
        body.addView(text("Wähle, wie du spielen möchtest.", 15, MUTED, false));
        space(body, 26);
        addCard(body, "Lokales Netzwerk", "Geräte im selben WLAN oder Handy-Hotspot finden und einen gemeinsamen Raum öffnen.", "Lokalen Raum öffnen", v -> showLocal());
        addCard(body, "Online verbinden", "Vorhandene TrashyGecko-Räume, Chat und Spiele über den bestehenden Online-Connector.", "Online-Connector öffnen", v -> openOnline());
        TextView note = text("Lokale Räume koordinieren Nachrichten und Spielereignisse. Für Mehrspieler muss das jeweilige Spiel diese Ereignisse unterstützen.", 12, MUTED, false);
        note.setPadding(0, dp(20), 0, dp(8)); body.addView(note);
        TextView diagnostics = text(compatibilitySummary(), 11, MUTED, false);
        diagnostics.setPadding(0, dp(8), 0, dp(20)); body.addView(diagnostics);
    }

    private void showExtras() {
        schliesseSpiel();
        onHomeScreen = false; onOnlinePage = false; showNativeBody(); body.removeAllViews();
        body.addView(text("Extras", 25, TEXT, true));
        body.addView(text("Optionale Zusatzfunktionen für TrashyGecko.", 14, MUTED, false));
        space(body, 18);
        if (isRemoteAssetAvailable()) {
            addCard(body, "WLAN-Fernbedienung", "Eigenständige Zusatzfunktion. Android-/Google-TV-Steuerung funktioniert nur bei kompatiblen Geräten mit Android TV Remote v2; Samsung-TVs lassen sich derzeit manuell erfassen, aber nicht steuern.", "Fernbedienung öffnen", v -> startActivity(new Intent(this, RemoteControlActivity.class)));
        } else {
            body.addView(text("Die Fernbedienung ist in dieser APK nicht enthalten. Der Einstieg wurde deshalb nicht angezeigt.", 14, MUTED, false));
        }
    }

    private boolean isRemoteAssetAvailable() {
        try (java.io.InputStream input = getAssets().open("www/fernbedienung/index.html")) {
            return input.read() >= 0;
        } catch (java.io.IOException unavailable) {
            return false;
        }
    }
    private void showLocal() {
        schliesseSpiel();
        onHomeScreen = false; onOnlinePage = false; showNativeBody(); body.removeAllViews();
        body.addView(text("Lokales Netzwerk", 25, TEXT, true));
        body.addView(text("Suche nach TrashyGecko-Räumen im WLAN oder Handy-Hotspot. Die Geräte müssen im selben lokalen Netz sein.", 13, MUTED, false));
        LinearLayout actions = new LinearLayout(this); actions.setOrientation(LinearLayout.HORIZONTAL); actions.setPadding(0, dp(12), 0, dp(10));
        Button scan = button("Räume suchen"); scan.setOnClickListener(v -> network.discover()); actions.addView(scan, new LinearLayout.LayoutParams(0, -2, 1));
        Button host = button("Raum erstellen"); host.setOnClickListener(v -> promptCreate()); actions.addView(host, new LinearLayout.LayoutParams(0, -2, 1));
        body.addView(actions);
        body.addView(text("Gefundene Räume", 16, ACCENT, true));
        roomList = new LinearLayout(this); roomList.setOrientation(LinearLayout.VERTICAL); body.addView(roomList);
        body.addView(text("Raumcode", 16, ACCENT, true));
        code = text("Kein eigener Raum", 14, TEXT, false); body.addView(code);
        playersTitle = text("Mitspieler", 16, ACCENT, true); playersTitle.setPadding(0, dp(8), 0, 0); body.addView(playersTitle);
        playerList = new LinearLayout(this); playerList.setOrientation(LinearLayout.VERTICAL); body.addView(playerList);
        chatLog = text("Nachrichten erscheinen hier.", 12, MUTED, false); chatLog.setPadding(0, dp(8), 0, dp(4)); body.addView(chatLog);
        LinearLayout chat = new LinearLayout(this); chat.setOrientation(LinearLayout.HORIZONTAL);
        chatInput = new EditText(this); chatInput.setSingleLine(true); chatInput.setHint("Nachricht im Raum"); chatInput.setTextColor(TEXT); chatInput.setHintTextColor(MUTED); chat.addView(chatInput, new LinearLayout.LayoutParams(0, -2, 1));
        Button send = button("Senden"); send.setOnClickListener(v -> sendChat()); chat.addView(send); body.addView(chat);
        Button leave = button("Raum verlassen / stoppen"); leave.setOnClickListener(v -> { network.stop(); galgen = null; gastStand = "{\"game\":null}"; schliesseSpiel(); code.setText("Kein eigener Raum"); status.setText("Raum beendet."); }); body.addView(leave);
        if (network.isHosting() && network.getRoomCode() != null) code.setText(network.getRoomCode() + " · Gastgeber");
        space(body, 18);
        addCard(body, "Galgenmännchen im Raum",
                "Ein Gerät führt die Runde, die Züge laufen über diesen Raum. Eröffne die Runde im Spiel, die anderen treten dort bei und drücken „Bereit“. Zwei bis sechs Geräte, alle im selben WLAN oder Hotspot.",
                "Galgen öffnen", v -> openGalgen());
    }

    private void openGalgen() {
        if (network == null || !(network.isHosting() || network.isVerbunden())) {
            status.setText("Erst einen Raum erstellen oder beitreten.");
            return;
        }
        onOnlinePage = false;
        if (galgen == null) galgen = new GalgenLokal();
        try {
            spielView = new WebView(this);
        } catch (RuntimeException fehler) {
            spielView = null;
            status.setText("WebView kann auf diesem Gerät nicht gestartet werden.");
            return;
        }
        WebSettings einstellungen = spielView.getSettings();
        einstellungen.setJavaScriptEnabled(true);
        einstellungen.setDomStorageEnabled(true);
        einstellungen.setAllowFileAccess(false);
        einstellungen.setAllowContentAccess(false);
        spielView.setBackgroundColor(BG);
        spielView.setWebChromeClient(new WebChromeClient());
        WebViewAssetLoader assets = HomepageAssets.baue(this);
        spielView.setWebViewClient(new WebViewClient() {
            @Override public WebResourceResponse shouldInterceptRequest(WebView view, WebResourceRequest request) {
                return assets.shouldInterceptRequest(request.getUrl());
            }
        });
        spielView.addJavascriptInterface(new GalgenBruecke(), "TGLokal");
        bodyScroll.removeAllViews();
        bodyScroll.addView(spielView, new ScrollView.LayoutParams(-1, -1));
        String adresse = HomepageAssets.ROOT + "spiele/brettspiele/galgen/galgen.html?lokal=1&id="
                + Uri.encode(network.getPlayerId()) + "&name=" + Uri.encode(network.getPlayerName());
        spielView.loadUrl(adresse);
        status.setText(network.isHosting()
                ? "Galgen im eigenen Raum · Code " + network.getRoomCode()
                : "Galgen im Raum · „Zurück“ führt zur Raumübersicht");
    }

    private void schliesseSpiel() {
        if (spielView == null) return;
        spielView.removeJavascriptInterface("TGLokal");
        spielView.stopLoading();
        spielView.destroy();
        spielView = null;
    }

    /** Aktueller Stand für die Seite: beim Gastgeber aus dem Spiel, beim Gast der letzte Empfang. */
    private String stand() {
        try {
            if (network != null && network.isHosting() && galgen != null && !galgen.istLeer()) {
                JSONObject antwort = new JSONObject();
                antwort.put("game", galgen.json());
                return antwort.toString();
            }
        } catch (JSONException ignoriert) {
            // unten kommt der letzte Empfang
        }
        return gastStand;
    }

    private void antworte(JSONObject antwort) {
        String json = antwort.toString();
        runOnUiThread(() -> {
            if (spielView == null) return;
            spielView.evaluateJavascript("window.TGLokalAntwort&&window.TGLokalAntwort(" + json + ");", null);
        });
    }

    private final class GalgenBruecke {
        @JavascriptInterface public String hole() {
            return stand();
        }

        @JavascriptInterface public void schicke(String roh) {
            try {
                JSONObject anfrageJson = new JSONObject(roh == null || roh.isEmpty() ? "{}" : roh);
                String anfrage = anfrageJson.optString("anfrage");
                String aktion = anfrageJson.optString("aktion");
                JSONObject koerper = anfrageJson.optJSONObject("koerper");
                if (network == null) return;
                if (network.isHosting()) {
                    String fehler = galgen == null ? "Kein Spiel offen."
                            : galgen.aktion(network.getPlayerId(), network.getPlayerName(), aktion, koerper);
                    JSONObject antwort = new JSONObject();
                    antwort.put("anfrage", anfrage);
                    if (fehler != null) {
                        antwort.put("fehler", fehler);
                    } else {
                        antwort.put("game", galgen == null || galgen.istLeer() ? JSONObject.NULL : galgen.json());
                    }
                    antworte(antwort);
                    if (fehler == null) {
                        JSONObject stand = new JSONObject();
                        stand.put("game", galgen == null || galgen.istLeer() ? JSONObject.NULL : galgen.json());
                        network.sendeAllen("state", stand);
                    }
                } else {
                    JSONObject payload = new JSONObject();
                    payload.put("anfrage", anfrage);
                    payload.put("aktion", aktion);
                    payload.put("koerper", koerper == null ? new JSONObject() : koerper);
                    network.send("move", payload);
                }
            } catch (JSONException fehler) {
                // Die Seite wartet kurz und meldet dann selbst einen Fehlschlag.
            }
        }
    }

    private void promptCreate() {
        EditText name = new EditText(this); name.setHint("Raumname"); name.setText("Spieleabend");
        new AlertDialog.Builder(this).setTitle("Lokalen Raum erstellen").setMessage("Du koordinierst diesen Raum auf deinem Gerät.").setView(name)
                .setPositiveButton("Erstellen", (d, w) -> { network.createRoom(name.getText().toString(), "TrashyGecko"); code.setText("Wird gestartet …"); })
                .setNegativeButton("Abbrechen", null).show();
    }

    private void promptJoin(LocalNetworkCoordinator.Room room) {
        EditText pin = new EditText(this); pin.setHint("8-stelliger Raumcode vom Gastgeber"); pin.setInputType(android.text.InputType.TYPE_CLASS_NUMBER); pin.setSingleLine(true);
        new AlertDialog.Builder(this).setTitle("Raum beitreten").setMessage(room.toString()).setView(pin)
                .setPositiveButton("Beitreten", (d, w) -> network.joinRoom(room, pin.getText().toString().trim(), "Spieler", ""))
                .setNegativeButton("Abbrechen", null).show();
    }

    private void openOnline() {
        schliesseSpiel();
        onHomeScreen = false; onOnlinePage = true; bodyScroll.removeAllViews();
        try { onlineView = new WebView(this); }
        catch (RuntimeException e) {
            onOnlinePage = false;
            status.setText("WebView kann auf diesem Gerät nicht gestartet werden.");
            showNativeBody(); body.removeAllViews();
            body.addView(text("Aktualisiere Android System WebView oder Google Chrome, sofern dein Gerät dies unterstützt. Offline-Mahjong und lokale Raumfunktionen bleiben separat verfügbar.", 14, MUTED, false));
            return;
        }
        onlineView.getSettings().setJavaScriptEnabled(true); onlineView.getSettings().setDomStorageEnabled(true); onlineView.setBackgroundColor(BG);
        onlineView.setWebChromeClient(new WebChromeClient());
        onlineView.setWebViewClient(new WebViewClient() {
            @Override public boolean shouldOverrideUrlLoading(WebView view, WebResourceRequest request) {
                Uri uri = request.getUrl();
                if ("trashygecko.onrender.com".equals(uri.getHost())) return false;
                if ("appassets.androidplatform.net".equals(uri.getHost())) return false;
                startActivity(new Intent(Intent.ACTION_VIEW, uri)); return true;
            }
        });
        bodyScroll.addView(onlineView, new ScrollView.LayoutParams(-1, -1));
        onlineView.loadUrl(ONLINE_URL);
        status.setText("Online-Connector · bestehender Server");
    }

    private void showNativeBody() {
        if (bodyScroll.getChildCount() == 0 || bodyScroll.getChildAt(0) != body) {
            bodyScroll.removeAllViews();
            bodyScroll.addView(body, new ScrollView.LayoutParams(-1, -2));
        }
    }

    private void sendChat() {
        String value = chatInput.getText().toString().trim(); if (value.isEmpty()) return;
        try { JSONObject p = new JSONObject(); p.put("text", value); network.send("chat", p); chatInput.setText(""); }
        catch (Exception e) { status.setText("Nachricht konnte nicht vorbereitet werden."); }
    }

    private String compatibilitySummary() {
        StringBuilder result = new StringBuilder();
        result.append("Android ").append(Build.VERSION.RELEASE).append(" (API ").append(Build.VERSION.SDK_INT).append(")");
        result.append(" · ").append(Build.MANUFACTURER).append(" ").append(Build.MODEL);
        if (Build.VERSION.SDK_INT >= 21 && Build.SUPPORTED_ABIS.length > 0) result.append(" · ").append(Build.SUPPORTED_ABIS[0]);
        ActivityManager manager = (ActivityManager) getSystemService(ACTIVITY_SERVICE);
        if (manager != null) {
            ActivityManager.MemoryInfo memory = new ActivityManager.MemoryInfo(); manager.getMemoryInfo(memory);
            result.append(" · RAM ").append(Math.max(1, memory.totalMem / (1024L * 1024L * 1024L))).append(" GB");
        }
        if (Build.VERSION.SDK_INT >= 26) {
            try {
                PackageInfo web = WebView.getCurrentWebViewPackage();
                result.append(" · WebView ").append(web == null ? "nicht erkannt" : web.versionName);
            } catch (RuntimeException e) { result.append(" · WebView-Status nicht ermittelbar"); }
        } else result.append(" · WebView-Version wird von Android nicht gemeldet");
        result.append(" · NSD ").append(getSystemService(NSD_SERVICE) == null ? "nicht verfügbar" : "bereit");
        ConnectivityManager cm = (ConnectivityManager) getSystemService(CONNECTIVITY_SERVICE);
        NetworkInfo active = cm == null ? null : cm.getActiveNetworkInfo();
        result.append(" · Netzwerk ").append(active != null && active.isConnected() ? "verbunden" : "offline");
        boolean multicastPermission = getPackageManager().checkPermission("android.permission.CHANGE_WIFI_MULTICAST_STATE", getPackageName()) == PackageManager.PERMISSION_GRANTED;
        result.append(" · Multicast ").append(multicastPermission ? "freigegeben" : "nicht freigegeben");
        return result.toString();
    }

    private void addCard(LinearLayout parent, String title, String description, String action, View.OnClickListener listener) {
        LinearLayout card = new LinearLayout(this); card.setOrientation(LinearLayout.VERTICAL); card.setPadding(dp(16), dp(14), dp(16), dp(12)); card.setBackgroundColor(CARD);
        LinearLayout.LayoutParams lp = new LinearLayout.LayoutParams(-1, -2); lp.bottomMargin = dp(12); parent.addView(card, lp);
        card.addView(text(title, 18, TEXT, true)); TextView body = text(description, 13, MUTED, false); body.setPadding(0, dp(5), 0, dp(9)); card.addView(body);
        Button button = button(action); button.setOnClickListener(listener); card.addView(button);
    }

    private Button button(String label) { Button b = new Button(this); b.setText(label); b.setTextColor(TEXT); b.setAllCaps(false); b.setBackgroundTintList(android.content.res.ColorStateList.valueOf(Color.rgb(37, 82, 66))); return b; }
    private TextView text(String value, int size, int color, boolean bold) { TextView t = new TextView(this); t.setText(value); t.setTextSize(size); t.setTextColor(color); if (bold) t.setTypeface(Typeface.DEFAULT, Typeface.BOLD); return t; }
    private void space(LinearLayout parent, int height) { View v = new View(this); parent.addView(v, new LinearLayout.LayoutParams(1, dp(height))); }
    private int dp(int value) { return (int) (value * getResources().getDisplayMetrics().density + 0.5f); }

    @Override public void onRooms(List<LocalNetworkCoordinator.Room> values) {
        rooms.clear(); rooms.addAll(values); if (roomList == null) return; roomList.removeAllViews();
        if (values.isEmpty()) { roomList.addView(text("Noch keine Räume gefunden. Tippe auf „Räume suchen“.", 12, MUTED, false)); return; }
        for (LocalNetworkCoordinator.Room room : values) { Button b = button(room.toString()); b.setOnClickListener(v -> promptJoin(room)); roomList.addView(b); }
    }

    @Override public void onStatus(String message) {
        runOnUiThread(() -> { status.setText(message); if (code != null && network.isHosting() && network.getRoomCode() != null) code.setText(network.getRoomCode() + " · Gastgeber"); });
    }

    @Override public void onPlayers(List<String> values) {
        runOnUiThread(() -> { if (playerList == null) return; playerList.removeAllViews(); if (network.isHosting()) playerList.addView(text("Du (Koordinator)", 13, TEXT, false)); for (String name : values) playerList.addView(text("• " + name, 13, TEXT, false)); });
    }

    @Override public void onEvent(JSONObject event) {
        String type = event.optString("type"); JSONObject payload = event.optJSONObject("payload");
        if (payload == null) return;
        if ("chat".equals(type) && chatLog != null) {
            String line = event.optString("senderName", "Mitspieler") + ": " + payload.optString("text", "");
            runOnUiThread(() -> chatLog.setText(line));
            return;
        }
        if (network == null) return;
        if ("move".equals(type)) {
            if (!network.isHosting()) return;
            String anfrage = payload.optString("anfrage");
            String fehler = galgen == null ? "Kein Spiel offen."
                    : galgen.aktion(event.optString("sender"), event.optString("senderName"),
                            payload.optString("aktion"), payload.optJSONObject("koerper"));
            try {
                JSONObject antwort = new JSONObject();
                antwort.put("anfrage", anfrage);
                if (fehler != null) {
                    antwort.put("fehler", fehler);
                } else {
                    antwort.put("game", galgen == null || galgen.istLeer() ? JSONObject.NULL : galgen.json());
                }
                network.sendeAllen("state", antwort);
            } catch (JSONException ignoriert) {
                // nichts zu retten; der Mitspieler bekommt keine Antwort und die Seite meldet es
            }
            return;
        }
        if ("state".equals(type)) {
            if (payload.has("game")) {
                try {
                    JSONObject wrapper = new JSONObject();
                    wrapper.put("game", payload.opt("game"));
                    gastStand = wrapper.toString();
                } catch (JSONException ignoriert) {
                    return;
                }
            }
            if (payload.has("anfrage")) antworte(payload);
        }
    }

    @Override public void onBackPressed() {
        if (onHomeScreen) { finish(); return; }
        if (spielView != null) { showLocal(); return; }
        if (onOnlinePage && onlineView != null && onlineView.canGoBack()) { onlineView.goBack(); return; }
        if (network != null) network.stop(); showHome();
    }

    @Override protected void onDestroy() {
        if (network != null) network.destroy();
        schliesseSpiel();
        if (onlineView != null) { onlineView.destroy(); onlineView = null; }
        if (wifiLock != null && wifiLock.isHeld()) wifiLock.release();
        super.onDestroy();
    }
}

