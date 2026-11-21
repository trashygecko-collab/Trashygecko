package de.trashygecko.tmahjong;

import android.content.Context;
import android.net.nsd.NsdManager;
import android.net.nsd.NsdServiceInfo;
import android.net.wifi.WifiManager;
import android.os.Handler;
import android.os.Looper;
import android.util.Base64;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import java.io.BufferedReader;
import java.io.BufferedWriter;
import java.io.IOException;
import java.io.InputStreamReader;
import java.io.OutputStreamWriter;
import java.net.InetAddress;
import java.net.ServerSocket;
import java.net.Socket;
import java.net.SocketException;
import java.net.UnknownHostException;
import java.security.SecureRandom;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.UUID;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.ScheduledFuture;
import java.util.concurrent.atomic.AtomicLong;

/**
 * Small LAN room relay for app-to-app play. A room host only relays authenticated,
 * ordered JSON events; game rules and state remain in the game. Discovery uses
 * Android NSD so it can follow Wi-Fi and hotspot interface address changes.
 */
public final class LocalNetworkCoordinator {
    private static final String SERVICE_TYPE = "_trashygecko._tcp.";
    private static final int PROTOCOL_VERSION = 1;
    private static final long PLAYER_GRACE_MS = 120_000L;

    public interface Listener {
        void onRooms(List<Room> rooms);
        void onStatus(String message);
        void onPlayers(List<String> players);
        void onEvent(JSONObject event);
    }

    public static final class Room {
        public final String name;
        public final String game;
        public final String host;
        public final int port;
        public final String serviceName;

        Room(String name, String game, String host, int port, String serviceName) {
            this.name = name;
            this.game = game;
            this.host = host;
            this.port = port;
            this.serviceName = serviceName;
        }

        @Override public String toString() {
            return name + (game.isEmpty() ? "" : " · " + game) + " · " + host;
        }
    }

    private static final class Peer {
        final String id;
        final String name;
        final String token;
        volatile Socket socket;
        volatile BufferedWriter writer;
        volatile long lastSeen;

        Peer(String id, String name, String token, Socket socket, BufferedWriter writer) {
            this.id = id;
            this.name = name;
            this.token = token;
            this.socket = socket;
            this.writer = writer;
            this.lastSeen = System.currentTimeMillis();
        }
    }

    private final Context context;
    private final NsdManager nsd;
    private final Listener listener;
    private final Handler main = new Handler(Looper.getMainLooper());
    private final ExecutorService io = Executors.newCachedThreadPool();
    private final ScheduledExecutorService heartbeat = Executors.newSingleThreadScheduledExecutor();
    private final Map<String, NsdServiceInfo> found = new ConcurrentHashMap<>();
    private final Map<String, Room> rooms = new ConcurrentHashMap<>();
    private final Map<String, Peer> peers = new ConcurrentHashMap<>();
    private final Map<String, String> playerTokens = new ConcurrentHashMap<>();
    private final Set<String> receivedIds = Collections.newSetFromMap(new ConcurrentHashMap<String, Boolean>());
    private final AtomicLong sequence = new AtomicLong();
    private final String playerId = UUID.randomUUID().toString();
    private final String playerName;

    private WifiManager.MulticastLock multicastLock;
    private NsdManager.DiscoveryListener discoveryListener;
    private NsdManager.RegistrationListener registrationListener;
    private volatile ServerSocket server;
    private Socket client;
    private volatile BufferedWriter clientWriter;
    private volatile ScheduledFuture<?> heartbeatTask;
    private volatile boolean closed;
    private volatile boolean hosting;
    private volatile String roomCode;
    private volatile String roomName;
    private volatile String gameName;
    private volatile String resumeToken;
    private volatile Room joinedRoom;

    public LocalNetworkCoordinator(Context context, String playerName, Listener listener) {
        this.context = context.getApplicationContext();
        this.playerName = playerName == null || playerName.trim().isEmpty() ? "Spieler" : playerName.trim();
        this.listener = listener;
        this.nsd = (NsdManager) context.getSystemService(Context.NSD_SERVICE);
    }

    public boolean isHosting() { return hosting; }
    public String getRoomCode() { return roomCode; }
    public String getPlayerId() { return playerId; }
    public String getPlayerName() { return playerName; }
    /** Gast: steht die Verbindung zum Gastgeber? */
    public boolean isVerbunden() { return clientWriter != null; }
    public Room getRoom(String serviceName) { return rooms.get(serviceName); }

    public void discover() {
        if (nsd == null) { status("Android-Netzwerksuche ist nicht verfügbar."); return; }
        found.clear();
        rooms.clear();
        acquireMulticastLock();
        discoveryListener = new NsdManager.DiscoveryListener() {
            @Override public void onDiscoveryStarted(String type) { status("Suche im lokalen WLAN und Hotspot …"); }
            @Override public void onServiceFound(NsdServiceInfo service) {
                if (SERVICE_TYPE.equals(service.getServiceType())) resolve(service);
            }
            @Override public void onServiceLost(NsdServiceInfo service) {
                rooms.remove(service.getServiceName());
                publishRooms();
            }
            @Override public void onDiscoveryStopped(String type) { }
            @Override public void onStartDiscoveryFailed(String type, int error) {
                status("Netzwerksuche fehlgeschlagen (" + error + ").");
                stopDiscovery();
            }
            @Override public void onStopDiscoveryFailed(String type, int error) { status("Netzwerksuche konnte nicht beendet werden."); }
        };
        try { nsd.discoverServices(SERVICE_TYPE, NsdManager.PROTOCOL_DNS_SD, discoveryListener); }
        catch (RuntimeException e) { status("Netzwerksuche nicht möglich: " + e.getMessage()); stopDiscovery(); }
    }

    private void resolve(NsdServiceInfo service) {
        synchronized (found) {
            if (found.containsKey(service.getServiceName())) return;
            found.put(service.getServiceName(), service);
        }
        try {
            nsd.resolveService(service, new NsdManager.ResolveListener() {
                @Override public void onResolveFailed(NsdServiceInfo s, int error) { found.remove(s.getServiceName()); }
                @Override public void onServiceResolved(NsdServiceInfo resolved) {
                    InetAddress address = resolved.getHost();
                    if (address == null || resolved.getPort() <= 0) return;
                    String game = "";
                    try { game = new String(resolved.getAttributes().get("game"), "UTF-8"); }
                    catch (Exception ignored) { }
                    rooms.put(resolved.getServiceName(), new Room(resolved.getServiceName(), game,
                            address.getHostAddress(), resolved.getPort(), resolved.getServiceName()));
                    publishRooms();
                }
            });
        } catch (RuntimeException e) { found.remove(service.getServiceName()); }
    }

    public void stopDiscovery() {
        if (nsd != null && discoveryListener != null) {
            try { nsd.stopServiceDiscovery(discoveryListener); } catch (RuntimeException ignored) { }
            discoveryListener = null;
        }
        releaseMulticastLock();
    }

    public void createRoom(String name, String game) {
        stop();
        closed = false;
        roomName = clean(name, "Spieleabend", 50);
        gameName = clean(game, "TrashyGecko", 40);
        roomCode = randomCode();
        peers.clear();
        playerTokens.clear();
        receivedIds.clear();
        sequence.set(0);
        hosting = true;
        io.execute(() -> {
            try {
                server = new ServerSocket(0, 32, InetAddress.getByName("0.0.0.0"));
                registerService(server.getLocalPort());
                main.post(() -> { status("Raum bereit · Beitrittscode " + roomCode); publishPlayers(); });
                while (!closed && !server.isClosed()) {
                    Socket socket = server.accept();
                    io.execute(() -> serve(socket));
                }
            } catch (IOException e) {
                if (!closed) status("Raum konnte nicht gestartet werden: " + e.getMessage());
                stop();
            }
        });
    }

    private void registerService(int port) {
        NsdServiceInfo info = new NsdServiceInfo();
        info.setServiceName("TG-" + roomName.replaceAll("[^A-Za-z0-9-]", "-") + "-" + roomCode.substring(0, 4));
        info.setServiceType(SERVICE_TYPE);
        info.setPort(port);
        info.setAttribute("v", String.valueOf(PROTOCOL_VERSION));
        info.setAttribute("game", gameName);
        registrationListener = new NsdManager.RegistrationListener() {
            @Override public void onServiceRegistered(NsdServiceInfo service) { roomName = service.getServiceName(); }
            @Override public void onRegistrationFailed(NsdServiceInfo service, int error) { status("Raum läuft, aber die automatische Bekanntgabe schlug fehl."); }
            @Override public void onServiceUnregistered(NsdServiceInfo service) { }
            @Override public void onUnregistrationFailed(NsdServiceInfo service, int error) { }
        };
        try { nsd.registerService(info, NsdManager.PROTOCOL_DNS_SD, registrationListener); }
        catch (RuntimeException e) { status("Automatische Bekanntgabe nicht möglich: " + e.getMessage()); }
    }

    private void serve(Socket socket) {
        String id = null;
        try {
            socket.setSoTimeout(45_000);
            BufferedReader in = new BufferedReader(new InputStreamReader(socket.getInputStream(), "UTF-8"));
            BufferedWriter out = new BufferedWriter(new OutputStreamWriter(socket.getOutputStream(), "UTF-8"));
            String first = in.readLine();
            if (first == null || first.length() > 16_384) { socket.close(); return; }
            JSONObject hello = new JSONObject(first);
            if (!"join".equals(hello.optString("type")) || !roomCode.equals(hello.optString("code"))) {
                write(out, packet("error", "Raumcode ist ungültig.")); socket.close(); return;
            }
            id = clean(hello.optString("id"), "", 80);
            String name = clean(hello.optString("name"), "Mitspieler", 40);
            String providedToken = hello.optString("token", "");
            if (id.isEmpty()) { socket.close(); return; }
            String token = playerTokens.get(id);
            if (token != null && !token.equals(providedToken)) {
                write(out, packet("error", "Wiederverbindung abgelehnt: Gerätekennung stimmt nicht.")); socket.close(); return;
            }
            if (token == null) { token = randomToken(); playerTokens.put(id, token); }
            Peer peer = new Peer(id, name, token, socket, out);
            peers.put(id, peer);
            JSONObject welcome = packet("welcome", "Verbunden"); welcome.put("token", token); welcome.put("room", roomName); welcome.put("sequence", sequence.get());
            write(out, welcome);
            sendPlayers();
            readPeer(peer, in);
        } catch (Exception e) {
            if (!closed && !(e instanceof SocketException)) status("Ein Gerät hat die Verbindung verloren.");
        } finally {
            if (id != null) {
                Peer p = peers.get(id);
                if (p != null && p.socket == socket) { p.lastSeen = System.currentTimeMillis(); p.socket = null; p.writer = null; }
                publishPlayers();
            }
            try { socket.close(); } catch (IOException ignored) { }
        }
    }

    private void readPeer(Peer peer, BufferedReader in) throws IOException, JSONException {
        String line;
        while (!closed && (line = in.readLine()) != null) {
            if (line.length() > 256_000) continue;
            JSONObject event = new JSONObject(line);
            if (!peer.id.equals(event.optString("sender"))) continue;
            String type = event.optString("type");
            if ("ping".equals(type)) { peer.lastSeen = System.currentTimeMillis(); write(peer.writer, packet("pong", "")); continue; }
            if (!("chat".equals(type) || "move".equals(type) || "state".equals(type) || "leave".equals(type))) continue;
            String eventId = event.optString("eventId", "");
            if (eventId.isEmpty() || !receivedIds.add(eventId)) continue;
            peer.lastSeen = System.currentTimeMillis();
            event.put("sequence", sequence.incrementAndGet());
            event.put("senderName", peer.name);
            if ("state".equals(type) && !hosting) continue;
            broadcast(event, peer.id);
            if ("leave".equals(type)) { peers.remove(peer.id); break; }
        }
    }

    public void joinRoom(Room room, String code, String name, String savedToken) {
        stop();
        closed = false;
        hosting = false;
        joinedRoom = room;
        roomCode = code;
        resumeToken = savedToken == null ? "" : savedToken;
        io.execute(() -> connectClient(room, name));
    }

    private void connectClient(Room room, String name) {
        for (int attempt = 0; attempt <= 12 && !closed; attempt++) {
            try {
                Socket s = new Socket();
                s.connect(new java.net.InetSocketAddress(room.host, room.port), 5000);
                s.setSoTimeout(45_000);
                BufferedReader in = new BufferedReader(new InputStreamReader(s.getInputStream(), "UTF-8"));
                BufferedWriter out = new BufferedWriter(new OutputStreamWriter(s.getOutputStream(), "UTF-8"));
                JSONObject hello = new JSONObject(); hello.put("type", "join"); hello.put("code", roomCode); hello.put("id", playerId); hello.put("name", playerName); hello.put("token", resumeToken);
                write(out, hello);
                String response = in.readLine();
                if (response == null) throw new IOException("Der Raum hat die Verbindung beendet.");
                JSONObject welcome = new JSONObject(response);
                if ("error".equals(welcome.optString("type"))) { status(welcome.optString("message")); try { s.close(); } catch (IOException ignored) { } return; }
                resumeToken = welcome.optString("token", resumeToken);
                client = s; clientWriter = out; startHeartbeat(out);
                status("Im Raum „" + welcome.optString("room", room.name) + "“ verbunden.");
                while (!closed && !s.isClosed()) {
                    String line = in.readLine(); if (line == null) break;
                    if (line.length() > 256_000) continue;
                    JSONObject event = new JSONObject(line);
                    if ("players".equals(event.optString("type"))) publishPlayerJson(event.optJSONArray("players"));
                    else if ("event".equals(event.optString("type"))) publishEvent(event.optJSONObject("event"));
                    else if ("error".equals(event.optString("type"))) status(event.optString("message"));
                }
                closeClient();
                if (!closed) status("Verbindung unterbrochen. Wiederverbindung läuft …");
            } catch (Exception e) {
                closeClient();
                if (!closed) status("Verbindung fehlgeschlagen: " + e.getMessage());
            }
            if (attempt < 12 && !closed) {
                try { Thread.sleep(Math.min(15_000, (attempt + 1) * 1_500L)); }
                catch (InterruptedException e) { Thread.currentThread().interrupt(); return; }
            }
        }
        if (!closed) status("Wiederverbindung nicht möglich. Suche den Raum erneut.");
    }

    public void send(String type, JSONObject payload) {
        if (closed || (!hosting && clientWriter == null)) { status("Kein verbundener Spielraum."); return; }
        try {
            JSONObject event = new JSONObject(); event.put("type", type); event.put("sender", playerId); event.put("eventId", UUID.randomUUID().toString()); event.put("payload", payload == null ? new JSONObject() : payload);
            if (hosting) { event.put("sequence", sequence.incrementAndGet()); event.put("senderName", playerName); broadcast(event, playerId); }
            else write(clientWriter, event);
        } catch (Exception e) { status("Nachricht konnte nicht gesendet werden."); }
    }

    /**
     * Gastgeber schickt ein Ereignis an alle Mitspieler – auch an den, der es
     * ausgelöst hat. Nur so bekommt der Fragende die Antwort auf seinen Zug.
     */
    public void sendeAllen(String type, JSONObject payload) {
        if (closed || !hosting) return;
        try {
            JSONObject event = new JSONObject();
            event.put("type", type);
            event.put("sender", playerId);
            event.put("senderName", playerName);
            event.put("eventId", UUID.randomUUID().toString());
            event.put("sequence", sequence.incrementAndGet());
            event.put("payload", payload == null ? new JSONObject() : payload);
            JSONObject packet = new JSONObject();
            packet.put("type", "event");
            packet.put("event", event);
            for (Peer peer : peers.values()) {
                if (peer.writer == null) continue;
                try { write(peer.writer, packet); } catch (IOException verloren) { peer.lastSeen = System.currentTimeMillis(); }
            }
        } catch (JSONException fehler) {
            status("Antwort konnte nicht gesendet werden.");
        }
    }

    private void broadcast(JSONObject event, String excludeId) {
        JSONObject packet = new JSONObject();
        try { packet.put("type", "event"); packet.put("event", event); } catch (JSONException ignored) { return; }
        for (Peer p : peers.values()) {
            if (p.id.equals(excludeId) || p.writer == null) continue;
            try { write(p.writer, packet); } catch (IOException e) { p.lastSeen = System.currentTimeMillis(); }
        }
        publishEvent(event);
    }

    private void sendPlayers() {
        JSONArray arr = new JSONArray();
        for (Peer p : peers.values()) { JSONObject item = new JSONObject(); try { item.put("id", p.id); item.put("name", p.name); item.put("online", p.socket != null && !p.socket.isClosed()); } catch (JSONException ignored) { } arr.put(item); }
        JSONObject packet = new JSONObject(); try { packet.put("type", "players"); packet.put("players", arr); } catch (JSONException ignored) { }
        for (Peer p : peers.values()) if (p.writer != null) try { write(p.writer, packet); } catch (IOException ignored) { }
        publishPlayerJson(arr);
    }

    private void publishPlayers() { main.post(this::sendPlayers); }
    private void publishPlayerJson(JSONArray players) {
        List<String> names = new ArrayList<>();
        if (players != null) for (int i = 0; i < players.length(); i++) names.add(players.optJSONObject(i) == null ? "Mitspieler" : players.optJSONObject(i).optString("name", "Mitspieler"));
        main.post(() -> listener.onPlayers(names));
    }
    private void publishEvent(JSONObject event) { main.post(() -> listener.onEvent(event)); }
    private void publishRooms() { List<Room> snapshot = new ArrayList<>(rooms.values()); main.post(() -> listener.onRooms(snapshot)); }
    private void status(String text) { main.post(() -> listener.onStatus(text)); }

    public void stop() {
        closed = true;
        if (registrationListener != null && nsd != null) {
            try { nsd.unregisterService(registrationListener); } catch (RuntimeException ignored) { }
            registrationListener = null;
        }
        stopDiscovery();
        try { if (server != null) server.close(); } catch (IOException ignored) { }
        server = null;
        closeClient();
        for (Peer p : peers.values()) if (p.socket != null) try { p.socket.close(); } catch (IOException ignored) { }
        peers.clear();
        hosting = false;
        roomCode = null;
    }

    private void closeClient() {
        if (heartbeatTask != null) { heartbeatTask.cancel(false); heartbeatTask = null; }
        try { if (client != null) client.close(); } catch (IOException ignored) { }
        client = null; clientWriter = null;
    }
    private void startHeartbeat(BufferedWriter writer) {
        if (heartbeatTask != null) heartbeatTask.cancel(false);
        heartbeatTask = heartbeat.scheduleAtFixedRate(() -> {
            if (closed || clientWriter != writer) return;
            try { JSONObject ping = new JSONObject(); ping.put("type", "ping"); ping.put("sender", playerId); write(writer, ping); }
            catch (Exception e) { closeClient(); }
        }, 15, 15, java.util.concurrent.TimeUnit.SECONDS);
    }
    public void destroy() { stop(); io.shutdownNow(); heartbeat.shutdownNow(); }

    private void acquireMulticastLock() {
        try {
            WifiManager wifi = (WifiManager) context.getApplicationContext().getSystemService(Context.WIFI_SERVICE);
            if (wifi != null && (multicastLock == null || !multicastLock.isHeld())) { multicastLock = wifi.createMulticastLock("trashygecko-nsd"); multicastLock.setReferenceCounted(false); multicastLock.acquire(); }
        } catch (SecurityException ignored) { status("Android erlaubt derzeit keine lokale Netzwerksuche."); }
    }
    private void releaseMulticastLock() { try { if (multicastLock != null && multicastLock.isHeld()) multicastLock.release(); } catch (RuntimeException ignored) { } multicastLock = null; }

    private static void write(BufferedWriter writer, JSONObject object) throws IOException {
        synchronized (writer) { writer.write(object.toString()); writer.write("\n"); writer.flush(); }
    }
    private static JSONObject packet(String type, String message) throws JSONException { JSONObject p = new JSONObject(); p.put("type", type); p.put("message", message); return p; }
    private static String clean(String value, String fallback, int max) { String s = value == null ? "" : value.replaceAll("[\\p{Cntrl}]", "").trim(); if (s.isEmpty()) s = fallback; return s.length() > max ? s.substring(0, max) : s; }
    private static String randomCode() { SecureRandom r = new SecureRandom(); return String.format(java.util.Locale.US, "%08d", r.nextInt(100_000_000)); }
    private static String randomToken() { byte[] b = new byte[24]; new SecureRandom().nextBytes(b); return Base64.encodeToString(b, Base64.URL_SAFE | Base64.NO_WRAP | Base64.NO_PADDING); }
}
