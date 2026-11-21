package de.trashygecko.tmahjong.remote;

import java.util.HashMap;
import java.util.Locale;
import java.util.Map;

/** Protocol registry. Only Android TV Remote v2 is currently a control adapter; other profiles fail closed. */
public final class RemoteAdapterRegistry {
    private final Map<String,TvRemoteAdapter> adapters=new HashMap<>();
    public RemoteAdapterRegistry(RokuEcpAdapter roku,AndroidTvRemoteAdapter androidTv){
        adapters.put("roku",roku);adapters.put("android",androidTv);adapters.put("google",androidTv);adapters.put("googletv",androidTv);
        adapters.put("tcl",androidTv);adapters.put("hisense_android",androidTv);adapters.put("sony_android",androidTv);adapters.put("philips_android",androidTv);
    }
    public TvRemoteAdapter forBrand(String brand){String id=brand==null?"":brand.trim().toLowerCase(Locale.ROOT);if("tcl_roku".equals(id))id="roku";return adapters.get(id);}
    public static String unsupportedMessage(String brand){
        String id=brand==null?"":brand.trim().toLowerCase(Locale.ROOT);
        if("roku".equals(id)||"tcl_roku".equals(id))return "Roku erlaubt laut aktueller ECP-Dokumentation keine ECP-Befehle von Drittanbieter-Mobilplattformen. Das Gerät kann per SSDP erscheinen, wird aber nicht gesteuert.";
        if("samsung".equals(id))return "Für Samsung Tizen ist hier kein öffentlich dokumentierter, generationsübergreifender lokaler Drittanbieter-Fernbedienungsdienst verifiziert. SmartThings-/Smart-View-Funktionen werden nicht vorgetäuscht.";
        if("lg".equals(id)||"webos".equals(id))return "LG SSAP ist in der veröffentlichten webOS-TV-API nicht als Drittanbieter-Remote-API zugesichert; der verbreitete lokale WebSocket-Dienst ist modellabhängig und unverschlüsselt. Daher nicht aktiviert.";
        if("sony".equals(id))return "BRAVIA-IP-Control gilt modellabhängig; die verfügbaren Consumer-Verfahren benötigen häufig lokales HTTP/PSK. Ohne enge, dynamische Android-Netzwerkregel wird es nicht aktiviert. Android-/Google-TV-BRAVIA kann als Android-TV-Profil gekoppelt werden.";
        if("philips".equals(id))return "JointSPACE variiert nach Generation. Ältere Varianten verwenden HTTP; API v6 nutzt TLS mit modellabhängigem Zertifikat und PIN. Nur Philips-Android-TV-Modelle mit Android TV Remote v2 können über das Android-TV-Profil gekoppelt werden.";
        if("vidaa".equals(id)||"hisense_vidaa".equals(id))return "VIDAA-Remote-Kommunikation ist firmware- und modellabhängig; ein öffentliches, generationsübergreifendes lokales Drittanbieterprotokoll ist nicht verifiziert.";
        if("firetv".equals(id))return "Amazon dokumentiert ADB für Entwicklung und Debugging, nicht als allgemeine Fernbedienungs-API. ADB-Debugging wird aus Sicherheitsgründen nicht automatisch aktiviert oder angesprochen. Vega OS ist ebenfalls nicht abgedeckt.";
        if("dial".equals(id))return "DIAL-SSDP weist ein Streaminggerät nach, bietet aber keine generische Fernbedienung. Die Suche ruft keine unverschlüsselten DIAL-Endpunkte ab.";
        if("harmony".equals(id))return "Huawei Vision/HarmonyOS verwendet geräteübergreifende HarmonyOS-Systemdienste und Berechtigungen. Das ist kein generisches Android-LAN-Remote-Protokoll und wird hier nicht aktiviert.";
        if("titan".equals(id))return "Titan OS wird als Plattform angezeigt, aber ein öffentlich verifiziertes allgemeines lokales Drittanbieter-Remote-Protokoll ist hier nicht implementiert.";
        if("whale".equals(id))return "Whale TV OS Entwicklerunterlagen beschreiben TV-App-/Tastenintegration; ein allgemeines Handy-zu-TV-Steuerprotokoll ist hier nicht implementiert.";
        if("coolita".equals(id))return "Coolita OS wird als Skyworth/Coocaa-Plattform angezeigt; ein öffentlich verifiziertes allgemeines lokales Drittanbieter-Remote-Protokoll ist hier nicht implementiert.";
        if("unknown".equals(id))return "Hersteller/Plattform sind nicht erkannt. Bitte Modell und Plattformangabe des TVs prüfen; es wird kein Protokoll geraten.";
        return "Für dieses TV-Profil gibt es in dieser Version keinen verifizierten Steueradapter. Discovery allein bestätigt keine Steuerfähigkeit.";
    }
}