export const BRANDS = {
  android: { label: "Android TV / Google TV · Remote v2", manufacturer: "Unbekannt", system: "Android TV / Google TV", protocol: "Android TV Remote v2 über TLS", adapter: "androidtv", pair: "Kopplung wird nur versucht, wenn das Gerät den Android-TV-Remote-v2-Dienst im lokalen Netz anbietet. Der TV zeigt einen sechsstelligen HEX-Code. Nicht jedes Android-basierte TV-System bietet diesen Dienst.", pairing: true },
  samsung: { label: "Samsung Tizen", manufacturer: "Samsung", system: "Tizen", protocol: "Samsung-Fernbedienungsprotokoll (nicht aktiviert)", adapter: "samsung", pair: "Kein aktivierter Drittanbieter-Adapter. Samsung-Freigaben und Verfahren sind modell-/generationsabhängig.", pairing: false },
  lg: { label: "LG webOS", manufacturer: "LG", system: "webOS", protocol: "SSAP-WebSocket (nicht aktiviert)", adapter: "webos", pair: "Kein aktivierter SSAP-Adapter. Lokale Verfahren sind modellabhängig und hier nicht als sichere Drittanbieter-API verifiziert.", pairing: false },
  sony: { label: "Sony BRAVIA · proprietäre Plattform", manufacturer: "Sony", system: "BRAVIA TV OS (model-dependent)", protocol: "BRAVIA IP Control (nicht aktiviert)", adapter: "sony", pair: "Kein aktivierter Adapter für BRAVIA-Modelle ohne Android TV. IP Control/Auth hängen vom Modell ab.", pairing: false },
  sony_android: { label: "Sony · Android / Google TV", manufacturer: "Sony", system: "Android TV / Google TV", protocol: "Android TV Remote v2 über TLS", adapter: "androidtv", pair: "Legacy-Profil. Nur Modelle mit Android-TV-Remote-v2-Dienst.", pairing: true },
  philips: { label: "Philips · JointSPACE", manufacturer: "Philips", system: "Philips Smart TV OS (model-dependent)", protocol: "JointSPACE (nicht aktiviert)", adapter: "philips", pair: "Kein aktivierter Adapter. JointSPACE-Version, TLS und Authentisierung unterscheiden sich nach Modell.", pairing: false },
  philips_android: { label: "Philips · Android / Google TV", manufacturer: "Philips", system: "Android TV / Google TV", protocol: "Android TV Remote v2 über TLS", adapter: "androidtv", pair: "Legacy-Profil. Nur Modelle mit Android-TV-Remote-v2-Dienst.", pairing: true },
  vidaa: { label: "VIDAA OS", manufacturer: "VIDAA TV", system: "VIDAA", protocol: "VIDAA Mobile Control (nicht aktiviert)", adapter: "vidaa", pair: "Kein aktivierter Adapter. Das offizielle VIDAA-Mobile-Control ist TV-/App- und generationsabhängig; ein öffentliches generisches Drittanbieterprotokoll wird hier nicht behauptet.", pairing: false },
  dial: { label: "DIAL · nur Erkennung", manufacturer: "Streaminggerät", system: "DIAL-fähiges Streaminggerät", protocol: "DIAL über SSDP (nur Erkennung)", adapter: "dial", pair: "DIAL findet Geräte und mögliche App-Dienste, bietet aber keine allgemeine Fernbedienung.", pairing: false },
  roku: { label: "Roku OS · nur Erkennung", manufacturer: "Roku", system: "Roku OS", protocol: "ECP/SSDP (nur Erkennung; Steuerung deaktiviert)", adapter: "roku", pair: "Roku-Geräte werden per SSDP erkannt. ECP-Befehle sind in dieser mobilen App deaktiviert.", pairing: false },
  tcl_roku: { label: "TCL · Roku OS (nur Erkennung)", manufacturer: "TCL", system: "Roku OS", protocol: "ECP/SSDP (nur Erkennung; Steuerung deaktiviert)", adapter: "roku", pair: "Roku-Geräte werden per SSDP erkannt. ECP-Befehle sind in dieser mobilen App deaktiviert.", pairing: false },
  firetv: { label: "Fire TV / Fire OS", manufacturer: "Amazon", system: "Fire OS / Vega OS", protocol: "ADB (Entwicklungszugang, deaktiviert)", adapter: "firetv", pair: "Kein generischer Adapter. ADB ist ein Entwicklungs-/Debugzugang und wird nicht automatisch aktiviert.", pairing: false },
  harmony: { label: "HarmonyOS TV", manufacturer: "Huawei Vision", system: "HarmonyOS TV", protocol: "Huawei Distributed Remote (Systemdienst, nicht aktiviert)", adapter: "harmony", pair: "HarmonyOS Distributed Remote benötigt HarmonyOS-Geräte, Systemdienste und geräteübergreifende Berechtigungen. Es ist kein generisches Android-LAN-Protokoll; daher hier nicht aktiviert.", pairing: false },
  titan: { label: "Titan OS", manufacturer: "Herstellerabhängig", system: "Titan OS", protocol: "Kein verifiziertes lokales Drittanbieterprotokoll (inaktiv)", adapter: "titan", pair: "Kein öffentlich verifiziertes generisches lokales Drittanbieter-Fernbedienungsprotokoll in dieser App.", pairing: false },
  whale: { label: "Whale TV OS", manufacturer: "Herstellerabhängig", system: "Whale TV OS", protocol: "Kein verifiziertes allgemeines Mobilprotokoll (inaktiv)", adapter: "whale", pair: "Die verfügbare Entwicklerdokumentation beschreibt TV-App-/Tastenintegration, nicht ein allgemeines Handy-zu-TV-Remote-Protokoll.", pairing: false },
  coolita: { label: "Coolita OS", manufacturer: "Skyworth / Coocaa", system: "Coolita OS", protocol: "Kein verifiziertes lokales Drittanbieterprotokoll (inaktiv)", adapter: "coolita", pair: "Kein öffentlich verifiziertes generisches lokales Drittanbieter-Fernbedienungsprotokoll in dieser App.", pairing: false },
  unknown: { label: "Unbekannte / andere TV-Plattform", manufacturer: "Unbekannt", system: "Unbekannt", adapter: "unknown", pair: "Nur Metadaten und manuelle Adresse. Die Plattform wird nicht geraten; ohne erkannten Adapter ist keine Steuerung verfügbar.", pairing: false },
  tcl: { label: "TCL · Android / Google TV", manufacturer: "TCL", system: "Android TV / Google TV", protocol: "Android TV Remote v2 über TLS", adapter: "androidtv", pair: "Legacy-Profil. Nur Modelle mit Android-TV-Remote-v2-Dienst.", pairing: true },
  hisense_android: { label: "Hisense · Android / Google TV", manufacturer: "Hisense", system: "Android TV / Google TV", protocol: "Android TV Remote v2 über TLS", adapter: "androidtv", pair: "Legacy-Profil. Nur Modelle mit Android-TV-Remote-v2-Dienst.", pairing: true },
  hisense_vidaa: { label: "Hisense · VIDAA", manufacturer: "Hisense", system: "VIDAA", protocol: "VIDAA Mobile Control (nicht aktiviert)", adapter: "vidaa", pair: "Kein aktivierter VIDAA-Adapter.", pairing: false }
};

export const MANUFACTURER_SUGGESTIONS = [
  "Xiaomi", "Mi TV", "Redmi TV", "Skyworth", "Coocaa", "Changhong", "Haier", "KONKA",
  "Huawei Vision", "Hisense", "Vidda", "Metz", "Grundig", "Thomson", "Telefunken", "JVC",
  "Toshiba", "Sharp", "Blaupunkt", "Aiwa", "RCA", "Vestel", "Philips", "Sony", "TCL",
  "Samsung", "LG", "Roku", "Amazon Fire TV", "Andere / unbekannt"
];

export const LEGACY_PROFILE = { tcl: "tcl", hisense: "hisense_vidaa", androidtv: "android", googletv: "android", webos: "lg" };
export const profileFor = id => BRANDS[id] ? BRANDS[id] : BRANDS[LEGACY_PROFILE[id]] || null;
export const adapterFor = id => profileFor(id)?.adapter || null;
export const brandLabel = id => profileFor(id)?.label || "Unbekannte Plattform";
