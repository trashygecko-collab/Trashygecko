/** Async WebView-to-native transport. Network credentials never pass through localStorage. */
(function installTrashyRemoteBridge() {
  if (window.TrashyRemoteBridge || !window.TGRemoteNative || typeof window.TGRemoteNative.request !== "function") return;
  const pending = new Map();
  let sequence = 0;
  window.__trashyRemoteDeliver = (id, result) => {
    const entry = pending.get(String(id));
    if (!entry) return;
    pending.delete(String(id));
    clearTimeout(entry.timer);
    entry.resolve(result && typeof result === "object" ? result : { ok: false, message: "Ungültige native Antwort." });
  };
  function request(method, payload) {
    return new Promise((resolve, reject) => {
      const id = Date.now().toString(36) + "-" + (++sequence).toString(36);
      const timer = setTimeout(() => { pending.delete(id); reject(new Error("Zeitüberschreitung der Android-Bridge.")); }, method === "discover" ? 18000 : 12000);
      pending.set(id, { resolve, reject, timer });
      try { window.TGRemoteNative.request(id, method, JSON.stringify(payload || {})); }
      catch (error) { clearTimeout(timer); pending.delete(id); reject(error); }
    });
  }
  window.TrashyRemoteBridge = {
    version: 2,
    discover: options => request("discover", options),
    stopDiscovery: () => request("stopDiscovery", {}),
    connect: device => request("connect", { device }),
    disconnect: device => request("disconnect", { device }),
    send: (device, command, args) => request("send", { device, command, args: args || {} }),
    pair: (device, code) => request("pair", { device, code: String(code || "").slice(0, 128) }),
    sendText: (device, text) => request("sendText", { device, text: String(text || "").slice(0, 256) }),
    listApps: device => request("listApps", { device }),
    launchApp: (device, appId) => request("launchApp", { device, appId: String(appId || "").slice(0, 80) }),
    saveDevice: device => request("saveDevice", { device }),
    removeDevice: device => request("removeDevice", { device }),
    getStoredDevices: () => request("getStoredDevices", {})
  };
})();

export function getBridge() {
  return window.TrashyRemoteBridge || null;
}

export function bridgeVersion() {
  const version = Number(getBridge()?.version);
  return version === 1 || version === 2 ? version : 0;
}

export function isNativeReady() {
  const bridge = getBridge();
  return bridgeVersion() > 0 && ["discover", "connect", "send", "pair"].every(name => typeof bridge[name] === "function");
}

export function bridgeHas(name) {
  return isNativeReady() && typeof getBridge()[name] === "function";
}

export function bridgeMode() {
  return isNativeReady() ? "Android-Bridge v" + bridgeVersion() : "Browser-Modus";
}

export function normalizeBridgeResult(value) {
  if (!value || typeof value !== "object") return { ok: false, connected: false, message: "Der Netzwerkdienst lieferte keine Antwort." };
  const ok = value.ok === true || value.connected === true;
  return { ...value, ok, connected: value.connected === true || (ok && value.connected !== false) };
}
