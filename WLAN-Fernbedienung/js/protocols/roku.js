import { ProtocolAdapter } from "./base.js";

/** Roku ECP is discovery-only in the mobile app; never issue local cleartext HTTP here. */
export class RokuAdapter extends ProtocolAdapter {
  constructor() { super("roku-discovery-only"); }
  async connect() {
    return { ok: false, connected: false, message: "Roku-ECP-Steuerung ist in dieser mobilen Anwendung deaktiviert. Roku-Geräte werden nur per SSDP erkannt; lokale HTTP-Steuerung wird nicht aufgerufen." };
  }
  async send() {
    return { ok: false, connected: false, message: "Roku-ECP-Tastenbefehle sind in dieser mobilen Anwendung deaktiviert." };
  }
}
