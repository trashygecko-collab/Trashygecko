export class ProtocolAdapter {
  constructor(id) { this.id = id; }
  async discover() { return { devices: [], message: "Für die lokale Netzwerksuche wird die native Android-Brücke benötigt." }; }
  async connect() { return { ok: false, connected: false, message: "Für dieses TV-Protokoll ist die native Android-Netzwerkbrücke erforderlich." }; }
  async pair() { return { ok: false, message: "Die Kopplung muss vom nativen Host mit dem TV-Protokoll ausgeführt werden." }; }
  async send() { return { ok: false, message: "Dieser Befehl ist im aktuellen Verbindungstyp nicht verfügbar." }; }
}
