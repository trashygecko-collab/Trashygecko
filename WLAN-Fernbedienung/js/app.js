import { BRANDS, MANUFACTURER_SUGGESTIONS, profileFor, brandLabel } from "./brands.js";
import { loadDevices, saveDevices, normalizeDiscovered } from "./device-store.js";
import { getBridge, bridgeHas, isNativeReady, bridgeMode, normalizeBridgeResult } from "./bridge.js";
import { adapterFor } from "./protocols/index.js";

const $ = selector => document.querySelector(selector);
const REPEATABLE = new Set(["up", "down", "left", "right", "volUp", "volDown", "channelUp", "channelDown"]);
const COMMAND_LABEL = {
  up: "Oben", down: "Unten", left: "Links", right: "Rechts", ok: "OK", back: "Zurück", home: "Home",
  menu: "Menü", settings: "Einstellungen", power: "Standby", volUp: "Lauter", volDown: "Leiser", mute: "Stumm",
  channelUp: "Kanal +", channelDown: "Kanal −", playPause: "Play/Pause", play: "Play", pause: "Pause",
  rewind: "Zurückspulen", fastForward: "Vorspulen", info: "Info", guide: "Guide", source: "Quelle", text: "Tastatur"
};
const state = {
  devices: loadDevices(), found: [], activeId: null, room: "all",
  retryTimer: 0, retries: 0, retryDelay: 1000, pairingDevice: null, editId: ""
};
let toastTimer = 0;
let repeatTimer = 0;
let repeatDelay = 0;

function esc(value) {
  return String(value == null ? "" : value).replace(/[&<>"']/g, char => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", "\"": "&quot;", "'": "&#39;" }[char]));
}
function activeDevice() { return state.devices.find(device => device.id === state.activeId) || null; }
function profile(device) { return profileFor(device?.brand); }
function caps(device) { return Array.isArray(device?.capabilities) ? device.capabilities : []; }
function showToast(message) {
  const element = $("#toast");
  element.textContent = message;
  element.classList.add("show");
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => element.classList.remove("show"), 4200);
}
function setDiscoveryMessage(message) { $("#discoveryStatus").textContent = message; }
function persist() {
  if (!saveDevices(state.devices)) showToast("Geräteliste konnte lokal nicht gespeichert werden.");
  if (bridgeHas("saveDevice")) for (const device of state.devices) getBridge().saveDevice({ ...device }).catch(() => {});
}
function validHost(host) {
  if (!host) return true;
  if (host.length > 253 || /[\s/\\?#@]/.test(host)) return false;
  const ipv4 = host.split(".");
  if (ipv4.length === 4 && ipv4.every(part => /^\d{1,3}$/.test(part) && Number(part) <= 255)) return true;
  return /^(?=.{1,253}$)[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?(?:\.[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?)*$/.test(host);
}
function statusLabel(device) {
  if (device.status === "connecting") return "Verbinde …";
  if (device.status === "pairing") return "Kopplung nötig";
  if (device.status === "error") return "Nicht erreichbar";
  if (device.connected) return "Verbunden";
  return "Getrennt";
}
function statusClass(device) {
  if (device.connected) return "on";
  if (device.status === "connecting" || device.status === "pairing") return "busy";
  if (device.status === "error") return "err";
  return "";
}
function clearRetry() { clearTimeout(state.retryTimer); state.retryTimer = 0; }
function resetRetry() { clearRetry(); state.retries = 0; state.retryDelay = 1000; }

/* ---------- Darstellung ---------- */

function renderChips() {
  const wrap = $("#chips");
  wrap.replaceChildren();
  for (const device of state.devices) {
    const chip = document.createElement("button");
    chip.type = "button";
    chip.className = "chip " + statusClass(device) + (device.id === state.activeId ? " sel" : "");
    chip.dataset.id = device.id;
    chip.title = statusLabel(device) + " – " + (device.host || "ohne IP");
    chip.innerHTML = '<span class="dot" aria-hidden="true"></span><span>' + esc(device.name) + '</span>';
    chip.addEventListener("click", () => selectDevice(device.id));
    wrap.append(chip);
  }
}
function renderRemote() {
  const device = activeDevice();
  const card = $("#remoteCard");
  card.classList.toggle("hidden", !device);
  $("#emptyState").classList.toggle("hidden", state.devices.length > 0);
  if (!device) return;
  $("#remoteTitle").textContent = device.name;
  $("#connectionDetail").textContent = [device.manufacturer, device.model, device.room, device.host]
    .filter(value => value && value !== "Unbekannt").join(" · ") || brandLabel(device.brand);
  const connection = $("#connection");
  connection.className = "connection " + (device.connected ? "online" : device.status === "error" ? "error" : device.status === "connecting" || device.status === "pairing" ? "busy" : "");
  $("#connectionText").textContent = statusLabel(device);
  $("#reconnectBtn").textContent = device.connected ? "Neu verbinden" : "Verbinden";
  $("#reconnectBtn").setAttribute("aria-disabled", device.status === "connecting" ? "true" : "false");
  $("#pairBtn").hidden = !profile(device)?.pairing || (device.paired && device.connected);

  const list = caps(device);
  $("#capabilityNote").textContent = !device.connected
    ? (device.status === "error" ? "Letzter Fehler: " + (device.lastError || "Verbindung nicht hergestellt.") : "Noch nicht verbunden – die Tasten werden nach dem Verbinden freigeschaltet.")
    : list.length ? list.length + " Tastenfunktionen vom Gerät bestätigt." : "Verbunden, aber das Gerät meldet keine Tastaturfunktionen.";

  document.querySelectorAll("#remoteCard [data-command]").forEach(button => {
    const command = button.dataset.command;
    const supported = device.connected && (command === "text" ? list.includes("text") : list.includes(command) || (command === "playPause" && list.includes("play") && list.includes("pause")));
    button.classList.toggle("off", !supported);
    button.setAttribute("aria-disabled", supported ? "false" : "true");
    button.title = supported ? "Senden: " + (COMMAND_LABEL[command] || command) : "Dieses Gerät hat die Taste nicht gemeldet";
  });
}
function renderManage() {
  $("#manageCount").textContent = state.devices.length + (state.devices.length === 1 ? " Gerät" : " Geräte");
  const rooms = Array.from(new Set(state.devices.map(device => device.room || "Zuhause"))).sort((a, b) => a.localeCompare(b, "de"));
  const filter = $("#roomFilter");
  const old = state.room;
  filter.replaceChildren(new Option("Alle Räume", "all"));
  for (const room of rooms) filter.add(new Option(room, room));
  filter.value = rooms.includes(old) || old === "all" ? old : "all";
  $("#roomOptions").replaceChildren(...rooms.map(room => new Option(room, room)));

  const list = $("#manageList");
  list.replaceChildren();
  const visible = state.devices.filter(device => state.room === "all" || device.room === state.room);
  if (!state.devices.length) {
    const note = document.createElement("p");
    note.className = "message-line";
    note.textContent = "Noch kein Gerät gespeichert.";
    list.append(note);
    return;
  }
  if (!visible.length) {
    const note = document.createElement("p");
    note.className = "message-line";
    note.textContent = "In diesem Raum ist noch kein Gerät gespeichert.";
    list.append(note);
    return;
  }
  for (const device of visible) {
    const row = document.createElement("div");
    row.className = "device-row" + (device.id === state.activeId ? " sel" : "");
    row.innerHTML = '<span class="glyph" aria-hidden="true">▣</span><div class="main"><strong>' + esc(device.name) + '</strong>'
      + '<small>' + esc(statusLabel(device)) + ' · ' + esc(device.room || "Zuhause")
      + (device.model ? ' · ' + esc(device.model) : "") + (device.host ? ' · ' + esc(device.host) : " · keine IP") + '</small></div>'
      + '<div class="acts">'
      + '<button class="btn ghost" type="button" data-act="select" data-id="' + esc(device.id) + '">' + (device.id === state.activeId ? "Aktiv" : "Wählen") + '</button>'
      + '<button class="btn ghost" type="button" data-act="edit" data-id="' + esc(device.id) + '">Bearbeiten</button>'
      + (profile(device)?.pairing && !device.paired ? '<button class="btn ghost" type="button" data-act="pair" data-id="' + esc(device.id) + '">Koppeln</button>' : "")
      + '<button class="btn ghost danger" type="button" data-act="remove" data-id="' + esc(device.id) + '">Entfernen</button>'
      + '</div>';
    list.append(row);
  }
}
function renderFound() {
  const list = $("#foundDevices");
  list.replaceChildren();
  if (!state.found.length) {
    const note = document.createElement("p");
    note.className = "message-line";
    note.textContent = isNativeReady() ? "Noch kein Gerät gefunden. Starte eine Suche, während der Fernseher eingeschaltet ist." : "Die Netzwerksuche läuft nur in der Android-App.";
    list.append(note);
    return;
  }
  for (const device of state.found) {
    const saved = state.devices.some(item => item.id === device.id || (device.host && item.host === device.host));
    const row = document.createElement("div");
    row.className = "found-row";
    row.innerHTML = '<div class="main"><strong>' + esc(device.name) + '</strong><small>' + esc(device.system || brandLabel(device.brand))
      + (device.host ? ' · ' + esc(device.host) : "") + '</small></div>'
      + '<button class="btn ghost" type="button" data-found="' + esc(device.id) + '"' + (saved ? " aria-disabled=\"true\"" : "") + '>' + (saved ? "Gespeichert" : "Hinzufügen") + '</button>';
    list.append(row);
  }
}
function render() {
  renderChips();
  renderRemote();
  renderManage();
  renderFound();
  const badge = $("#bridgeBadge");
  badge.innerHTML = '<i></i>' + esc(isNativeReady() ? "Gerät verbunden" : "Browser-Modus");
  badge.classList.toggle("native", isNativeReady());
  $("#scanBtn").disabled = !isNativeReady();
  $("#scanBtn").title = isNativeReady() ? "Geräte im WLAN suchen" : "Nur in der Android-App verfügbar";
  $("#modeNote").textContent = isNativeReady()
    ? ""
    : "Im Browser lassen sich Geräte verwalten, aber nicht steuern. Suche und Tasten brauchen die TrashyGecko-App auf dem Handy.";
}

/* ---------- Verbindung ---------- */

function clearConnection(device) { device.connected = false; device.capabilities = []; }
function scheduleReconnect(device) {
  clearRetry();
  if (!device || device.id !== state.activeId || !isNativeReady() || state.retries >= 6) return;
  const delay = Math.min(state.retryDelay, 30000);
  state.retryTimer = setTimeout(() => {
    state.retries += 1;
    state.retryDelay = Math.min(state.retryDelay * 2, 30000);
    connectDevice(device, { automatic: true });
  }, delay);
}
async function connectDevice(device, options) {
  const automatic = options?.automatic === true;
  if (!device) return;
  if (!device.host) {
    device.status = "error";
    device.lastError = "Keine IP-Adresse hinterlegt.";
    clearConnection(device);
    render();
    showToast("Für dieses Gerät fehlt die IP-Adresse. Bearbeite den Eintrag und ergänze sie.");
    return;
  }
  if (!isNativeReady()) {
    device.status = "error";
    device.lastError = "Steuerung nur in der TrashyGecko-App möglich.";
    clearConnection(device);
    render();
    showToast("Tasten und Suche laufen nur in der Android-App – im Browser lässt sich nichts senden.");
    return;
  }
  if (!automatic) resetRetry();
  device.status = "connecting";
  clearConnection(device);
  render();
  try {
    const adapter = adapterFor(profile(device));
    const raw = isNativeReady() ? await getBridge().connect({ ...device }) : adapter ? await adapter.connect(device) : null;
    const result = normalizeBridgeResult(raw);
    if (result.requiresPairing === true || result.pairingRequired === true || result.state === "pairing") {
      device.status = "pairing";
      device.pairMessage = result.message || "Bestätige die Verbindung am Fernseher.";
      render();
      openPairDialog(device, device.pairMessage);
      showToast(device.pairMessage);
      return;
    }
    if (!result.connected) {
      device.status = "error";
      device.lastError = result.message || "Verbindung nicht hergestellt.";
      render();
      if (!automatic) showToast(device.lastError);
      scheduleReconnect(device);
      return;
    }
    device.connected = true;
    device.status = "online";
    device.capabilities = Array.isArray(result.capabilities) ? result.capabilities.filter(item => typeof item === "string") : [];
    if (result.host) device.host = String(result.host);
    if (result.model) device.model = String(result.model).slice(0, 80);
    if (result.manufacturer && result.manufacturer !== "Unbekannt") device.manufacturer = String(result.manufacturer).slice(0, 60);
    if (typeof result.paired === "boolean") device.paired = result.paired;
    device.lastSeen = Date.now();
    device.lastError = "";
    persist();
    render();
    resetRetry();
    if (!automatic) showToast(device.capabilities.length ? "Verbunden. " + device.capabilities.length + " Tasten sind freigeschaltet." : "Verbunden, aber ohne gemeldete Tastenfunktionen.");
  } catch (error) {
    device.status = "error";
    clearConnection(device);
    device.lastError = error.message || "Verbindungsfehler";
    render();
    if (!automatic) showToast("Verbindung fehlgeschlagen: " + device.lastError);
    scheduleReconnect(device);
  }
}
function selectDevice(id) {
  const previous = activeDevice();
  if (previous && previous.id !== id && bridgeHas("disconnect")) getBridge().disconnect(previous).catch(() => {});
  state.activeId = id;
  resetRetry();
  render();
  const device = activeDevice();
  if (device) connectDevice(device, { automatic: false });
}
async function sendCommand(command, args) {
  const device = activeDevice();
  if (!device) return showToast("Wähle zuerst ein Gerät aus.");
  if (!device.connected) return showToast("Gerät ist getrennt. Tippe auf „Verbinden“.");
  const list = caps(device);
  const supported = list.includes(command) || (command === "playPause" && list.includes("play") && list.includes("pause"));
  if (!supported) return showToast("Dieses Gerät hat " + (COMMAND_LABEL[command] || command) + " nicht gemeldet.");
  navigator.vibrate?.(12);
  try {
    const adapter = adapterFor(profile(device));
    const raw = isNativeReady() ? await getBridge().send({ ...device }, command, args || {}) : adapter ? await adapter.send(device, command, args || {}) : null;
    const result = normalizeBridgeResult(raw);
    if (!result.ok) {
      if (result.connected === false && result.connectionLost) {
        device.status = "error";
        clearConnection(device);
        render();
        scheduleReconnect(device);
      }
      showToast(result.message || "Befehl wurde nicht bestätigt.");
    }
  } catch (error) {
    showToast("Befehl fehlgeschlagen: " + (error.message || "Netzwerkfehler"));
  }
}
async function pairDevice(device, code) {
  if (!isNativeReady()) return showToast("Kopplung braucht die Android-App.");
  const secret = String(code || "").trim().toUpperCase();
  if (!/^[0-9A-F]{6}$/.test(secret)) return showToast("Der Code muss aus sechs Zeichen (0–9, A–F) bestehen.");
  try {
    const result = normalizeBridgeResult(await getBridge().pair({ ...device }, secret));
    if (result.ok) {
      device.paired = true;
      device.status = "offline";
      persist();
      $("#pairDialog").close();
      render();
      showToast(result.message || "Kopplung erfolgreich.");
      connectDevice(device, { automatic: false });
    } else {
      $("#pairPrompt").textContent = result.message || "Kopplung nicht bestätigt. Prüfe den Code auf dem Fernseher.";
      showToast(result.message || "Kopplung nicht abgeschlossen.");
    }
  } catch (error) {
    showToast("Kopplung fehlgeschlagen: " + (error.message || "Netzwerkfehler"));
  }
}

/* ---------- Dialoge ---------- */

function populateBrands() {
  const groups = [
    ["Steuerbar", ["android", "tcl", "hisense_android", "sony_android", "philips_android"]],
    ["Nur erkannt / nicht steuerbar", ["samsung", "lg", "sony", "philips", "vidaa", "hisense_vidaa", "roku", "tcl_roku", "dial", "firetv", "harmony", "titan", "whale", "coolita", "unknown"]]
  ];
  const select = $("#deviceBrand");
  select.replaceChildren();
  $("#manufacturerOptions").replaceChildren(...MANUFACTURER_SUGGESTIONS.map(name => new Option(name, name)));
  for (const [label, ids] of groups) {
    const group = document.createElement("optgroup");
    group.label = label;
    for (const id of ids) {
      if (!BRANDS[id]) continue;
      const option = new Option(BRANDS[id].label, id);
      group.append(option);
    }
    select.append(group);
  }
}
function updatePairHint() {
  const option = profile($("#deviceBrand").value);
  $("#pairHint").textContent = option?.pair || "";
}
function openDeviceDialog(device) {
  $("#deviceForm").reset();
  state.editId = device?.id || "";
  $("#dialogTitle").textContent = device ? "Gerät bearbeiten" : "Gerät hinzufügen";
  $("#deviceName").value = device?.name || "";
  $("#deviceBrand").value = device?.brand || "android";
  $("#deviceManufacturer").value = device?.manufacturer && device.manufacturer !== "Unbekannt" ? device.manufacturer : "";
  $("#deviceModel").value = device?.model || "";
  $("#deviceRoom").value = device?.room || "Zuhause";
  $("#deviceHost").value = device?.host || "";
  updatePairHint();
  $("#deviceDialog").showModal();
  setTimeout(() => $("#deviceName").focus(), 80);
}
function openPairDialog(device, message) {
  state.pairingDevice = device;
  $("#pairPrompt").textContent = message || profile(device)?.pair || "Gib den Code ein, der auf dem Fernseher angezeigt wird.";
  $("#pairCode").value = "";
  if (!$("#pairDialog").open) $("#pairDialog").showModal();
  setTimeout(() => $("#pairCode").focus(), 80);
}

/* ---------- Suche ---------- */

async function discover() {
  if (!isNativeReady()) return showToast("Die Suche läuft nur in der Android-App.");
  $("#discoverPanel").open = true;
  $("#scanBtn").disabled = true;
  $("#discoveryState").textContent = "Suche …";
  setDiscoveryMessage("Suche nach Android-TV-Remote- und Streamingdiensten im WLAN …");
  try {
    const raw = await Promise.race([
      getBridge().discover({ timeoutMs: 12000 }),
      new Promise((_, reject) => setTimeout(() => reject(new Error("Suche dauerte zu lange.")), 15000))
    ]);
    const list = Array.isArray(raw) ? raw : Array.isArray(raw?.devices) ? raw.devices : [];
    state.found = list.map(normalizeDiscovered).filter(Boolean);
    $("#discoveryState").textContent = state.found.length ? state.found.length + " gefunden" : "nichts gefunden";
    const warnings = Array.isArray(raw?.warnings) ? raw.warnings : [];
    setDiscoveryMessage(warnings.length ? warnings.join(" ") : state.found.length ? "Gerät antippen, um es zu speichern." : "Kein Gerät geantwortet. Prüfe WLAN, Gastnetz-Isolation und ob der Fernseher eingeschaltet ist.");
  } catch (error) {
    $("#discoveryState").textContent = "Suche fehlgeschlagen";
    setDiscoveryMessage(error.message || "Netzwerksuche fehlgeschlagen.");
  } finally {
    $("#scanBtn").disabled = false;
    renderFound();
  }
}
function addFound(id) {
  const found = state.found.find(item => item.id === id);
  if (!found) return;
  const existing = state.devices.find(item => item.id === found.id || (found.host && item.host === found.host));
  if (existing) {
    existing.host = found.host || existing.host;
    existing.model = found.model || existing.model;
    existing.lastSeen = Date.now();
    persist();
    render();
    return showToast("Gerät aktualisiert.");
  }
  const device = { ...found, name: found.name || brandLabel(found.brand), room: "Zuhause", paired: false, connected: false, capabilities: [], status: "offline" };
  state.devices.unshift(device);
  state.activeId = device.id;
  persist();
  render();
  if (device.host) connectDevice(device, { automatic: false });
}

/* ---------- Bedienung ---------- */

function stopRepeat() {
  clearTimeout(repeatTimer);
  clearInterval(repeatDelay);
  repeatTimer = 0;
  repeatDelay = 0;
  document.querySelectorAll(".key.pressed").forEach(key => key.classList.remove("pressed"));
}
function handleKeyPress(button) {
  const command = button.dataset.command;
  const device = activeDevice();
  if (!device) return;
  if (button.getAttribute("aria-disabled") === "true") {
    if (!device.connected) return showToast("Gerät ist getrennt. Tippe auf „Verbinden“.");
    return showToast("Dieses Gerät hat „" + (COMMAND_LABEL[command] || command) + "“ nicht gemeldet.");
  }
  if (command === "source") return $("#sourceDialog").showModal();
  if (command === "text") return $("#textDialog").showModal();
  sendCommand(command);
}
function bindKey(button) {
  button.addEventListener("click", () => handleKeyPress(button));
  const command = button.dataset.command;
  if (!REPEATABLE.has(command)) return;
  const start = () => {
    stopRepeat();
    button.classList.add("pressed");
    repeatTimer = setTimeout(() => {
      repeatDelay = setInterval(() => sendCommand(command), 160);
      sendCommand(command);
    }, 460);
  };
  button.addEventListener("pointerdown", event => { if (event.button === 0) start(); });
  for (const type of ["pointerup", "pointercancel", "pointerleave"]) button.addEventListener(type, stopRepeat);
  button.addEventListener("keydown", event => { if (event.key === " " || event.key === "Enter") return; });
}
const KEYBOARD = {
  ArrowUp: "up", ArrowDown: "down", ArrowLeft: "left", ArrowRight: "right",
  Enter: "ok", Escape: "back", Backspace: "back",
  h: "home", m: "menu", p: "playPause", s: "settings", i: "info", g: "guide",
  "+": "volUp", "=": "volUp", "-": "volDown", "_": "volDown", "*": "mute"
};
document.addEventListener("keydown", event => {
  if (event.target.closest("input, select, textarea") || document.querySelector("dialog[open]")) return;
  if (event.ctrlKey || event.altKey || event.metaKey) return;
  const device = activeDevice();
  if (!device?.connected) return;
  let command = KEYBOARD[event.key];
  if (!command && /^[0-9]$/.test(event.key)) command = "digit" + event.key;
  if (!command) return;
  event.preventDefault();
  const list = caps(device);
  if (!list.includes(command)) return showToast("Dieses Gerät hat „" + (COMMAND_LABEL[command] || command) + "“ nicht gemeldet.");
  sendCommand(command);
});

/* ---------- Aufbau ---------- */

populateBrands();
updatePairHint();
render();
for (let digit = 1; digit <= 9; digit += 1) {
  const button = document.createElement("button");
  button.type = "button";
  button.className = "key";
  button.textContent = String(digit);
  button.dataset.command = "digit" + digit;
  bindKey(button);
  $("#numberGrid").append(button);
}
const zero = document.createElement("button");
zero.type = "button";
zero.className = "key";
zero.textContent = "0";
zero.dataset.command = "digit0";
bindKey(zero);
$("#numberGrid").append(zero);

document.querySelectorAll("#remoteCard [data-command]").forEach(bindKey);
$("#scanBtn").addEventListener("click", discover);
$("#addDeviceBtn").addEventListener("click", () => openDeviceDialog());
$("#emptyAdd").addEventListener("click", () => openDeviceDialog());
$("#closeDialog").addEventListener("click", () => $("#deviceDialog").close());
$("#deviceBrand").addEventListener("change", updatePairHint);
$("#roomFilter").addEventListener("change", event => { state.room = event.target.value; renderManage(); });
$("#reconnectBtn").addEventListener("click", () => { const device = activeDevice(); if (device) connectDevice(device, { automatic: false }); });
$("#pairBtn").addEventListener("click", () => { const device = activeDevice(); if (device) connectDevice(device, { automatic: false }); });
$("#manageList").addEventListener("click", event => {
  const button = event.target.closest("button[data-act]");
  if (!button) return;
  const device = state.devices.find(item => item.id === button.dataset.id);
  if (!device) return;
  const action = button.dataset.act;
  if (action === "select") selectDevice(device.id);
  if (action === "edit") openDeviceDialog(device);
  if (action === "pair") { state.activeId = device.id; render(); connectDevice(device, { automatic: false }); }
  if (action === "remove") {
    if (!window.confirm("„" + device.name + "“ aus der Geräteliste entfernen?")) return;
    if (bridgeHas("removeDevice")) getBridge().removeDevice({ ...device }).catch(() => {});
    state.devices = state.devices.filter(item => item.id !== device.id);
    if (state.activeId === device.id) { state.activeId = state.devices[0]?.id || null; clearRetry(); }
    persist();
    render();
  }
});
$("#foundDevices").addEventListener("click", event => {
  const button = event.target.closest("button[data-found]");
  if (button && button.getAttribute("aria-disabled") !== "true") addFound(button.dataset.found);
});
$("#pairForm").addEventListener("submit", event => {
  event.preventDefault();
  const device = state.pairingDevice;
  const code = $("#pairCode").value;
  $("#pairCode").value = "";
  if (device && code) pairDevice(device, code);
});
$("#closePair").addEventListener("click", () => { $("#pairCode").value = ""; state.pairingDevice = null; $("#pairDialog").close(); });
$("#textForm").addEventListener("submit", async event => {
  event.preventDefault();
  const device = activeDevice();
  const text = $("#remoteText").value.trim();
  if (!device || !text) return;
  try {
    const result = normalizeBridgeResult(await getBridge().sendText({ ...device }, text));
    showToast(result.message || (result.ok ? "Text gesendet." : "Text wurde nicht bestätigt."));
  } catch (error) {
    showToast("Texteingabe fehlgeschlagen: " + (error.message || "Netzwerkfehler"));
  }
  $("#remoteText").value = "";
  $("#textDialog").close();
});
$("#closeText").addEventListener("click", () => $("#textDialog").close());
$("#closeSource").addEventListener("click", () => $("#sourceDialog").close());
$("#sourceForm").addEventListener("submit", event => {
  event.preventDefault();
  const source = $("#sourceChoice").value;
  $("#sourceDialog").close();
  sendCommand("source", { source });
});
$("#deviceForm").addEventListener("submit", async event => {
  event.preventDefault();
  const name = $("#deviceName").value.trim();
  const brand = $("#deviceBrand").value;
  const host = $("#deviceHost").value.trim();
  if (!name) return showToast("Bitte einen Namen eintragen.");
  if (!validHost(host)) { showToast("IP-Adresse oder Hostname ist ungültig."); $("#deviceHost").focus(); return; }
  const old = state.devices.find(item => item.id === state.editId);
  const changedEndpoint = old && (old.host !== host || old.brand !== brand);
  const device = {
    ...(old || {}),
    id: old?.id || ("tv-" + Date.now().toString(36) + "-" + Math.random().toString(36).slice(2, 7)),
    name, brand,
    manufacturer: $("#deviceManufacturer").value.trim().slice(0, 60) || profileFor(brand)?.manufacturer || "Unbekannt",
    system: profileFor(brand)?.system || "Unbekannt",
    protocol: profileFor(brand)?.protocol || "",
    model: $("#deviceModel").value.trim(),
    room: $("#deviceRoom").value.trim() || "Zuhause",
    host,
    paired: changedEndpoint ? false : old?.paired === true,
    connected: false, capabilities: [], status: "offline"
  };
  if (old) state.devices[state.devices.findIndex(item => item.id === old.id)] = device;
  else state.devices.unshift(device);
  state.activeId = device.id;
  persist();
  $("#deviceDialog").close();
  render();
  if (host) connectDevice(device, { automatic: false });
  else showToast("Gespeichert. Ergänze eine IP-Adresse oder suche das Gerät im WLAN.");
});
document.addEventListener("visibilitychange", () => {
  if (document.hidden) return stopRepeat();
  const device = activeDevice();
  if (device && !device.connected && device.host && state.retries < 6) scheduleReconnect(device);
});

function selectInitialDevice() {
  if (!state.activeId && state.devices.length) state.activeId = state.devices[0].id;
  render();
  const device = activeDevice();
  if (device?.host && !device.connected && device.status !== "connecting") setTimeout(() => connectDevice(device, { automatic: true }), 250);
}
selectInitialDevice();
if (bridgeHas("getStoredDevices")) {
  getBridge().getStoredDevices().then(result => {
    const stored = Array.isArray(result?.devices) ? result.devices : [];
    let added = false;
    for (const entry of stored) {
      const device = normalizeDiscovered(entry);
      if (!device || state.devices.some(item => item.id === device.id || (item.host && item.host === device.host))) continue;
      state.devices.push(device);
      added = true;
    }
    if (added) { persist(); selectInitialDevice(); }
  }).catch(() => {});
}
