import { ProtocolAdapter } from "./base.js";


class NativeOnlyAdapter extends ProtocolAdapter {
  constructor(id, explanation) { super(id); this.explanation = explanation; }
  async connect() { return { ok: false, connected: false, message: this.explanation }; }
}

const adapters = {
  roku: new NativeOnlyAdapter("roku", "Roku-ECP-Steuerung ist in dieser mobilen Anwendung deaktiviert. Roku-Geräte werden nur per SSDP erkannt; lokale HTTP-Steuerung wird nicht aufgerufen."),
  dial: new NativeOnlyAdapter("dial", "DIAL dient hier nur zur Erkennung. Es gibt keine generische Fernsteuerung über DIAL."),
  samsung: new NativeOnlyAdapter("samsung", "Samsung Remote-/Smart View-Verfahren sind modell- und generationsabhängig. Discovery, Freigabe, TLS und Befehle brauchen die native Android-Brücke."),
  webos: new NativeOnlyAdapter("webos", "LG webOS SSAP verlangt eine WebSocket-Registrierung und typischerweise eine Bestätigung am TV. Die WebView kann diese lokale Verbindung nicht selbst aufbauen."),
  sony: new NativeOnlyAdapter("sony", "BRAVIA IP Control muss am Fernseher aktiviert sein. Authentisierung, IRCC/JSON-RPC und Power-Unterstützung hängen vom Modell ab."),
  androidtv: new NativeOnlyAdapter("androidtv", "Android TV/Google TV setzt einen kompatiblen Remote-Dienst und PIN-Kopplung voraus. Gerätegenerationen und Drittanbieter-Android-Boxen unterscheiden sich."),
  philips: new NativeOnlyAdapter("philips", "Philips JointSPACE-API-Version, Ports, TLS und PIN-Verfahren unterscheiden sich nach Modell und Betriebssystem."),
  vidaa: new NativeOnlyAdapter("vidaa", "Für VIDAA gibt es keine hier zugesicherte, generationsübergreifende öffentliche Fernbedienungsschnittstelle. Modell/Firmware müssen nativ geprüft werden."),
  firetv: new NativeOnlyAdapter("firetv", "Nur bestimmte Fire OS-Generationen können über aktiviertes und bestätigtes ADB angesprochen werden. Amazon Vega OS wird von diesem ADB-Verfahren nicht abgedeckt."),
  harmony: new NativeOnlyAdapter("harmony", "HarmonyOS TV verlangt Huawei-Distributed-Device-Manager/Systemdienste und passende HarmonyOS-Berechtigungen. Eine allgemeine Android-LAN-Verbindung ist nicht implementiert."),
  titan: new NativeOnlyAdapter("titan", "Titan OS ist als Plattform erfasst, aber ein verifiziertes lokales Drittanbieter-Fernbedienungsprotokoll ist nicht aktiviert."),
  whale: new NativeOnlyAdapter("whale", "Whale TV OS wird erkannt/auswählbar geführt; die verfügbare Entwicklerdokumentation ist keine allgemeine Handy-Fernbedienungs-API."),
  coolita: new NativeOnlyAdapter("coolita", "Coolita OS ist als Skyworth/Coocaa-Plattform erfasst, aber kein allgemeiner lokaler Drittanbieteradapter ist implementiert."),
  unknown: new NativeOnlyAdapter("unknown", "Hersteller oder TV-Plattform sind nicht sicher erkannt. Es wird kein Netzwerkprotokoll geraten.")
};

export function adapterFor(profile) {
  return adapters[profile?.adapter] || null;
}
