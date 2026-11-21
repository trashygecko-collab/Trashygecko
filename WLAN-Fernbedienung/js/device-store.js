import { profileFor, LEGACY_PROFILE } from "./brands.js";

const KEY = "trashyremote.devices";
const SAFE_FIELDS = ["id", "name", "brand", "manufacturer", "system", "protocol", "model", "host", "room", "paired", "lastSeen"];

function normalize(device) {
  if (!device || typeof device !== "object") return null;
  const rawBrand = String(device.brand || device.profile || "").trim();
  const brand = LEGACY_PROFILE[rawBrand] || rawBrand;
  if (!profileFor(brand)) return null;
  const host = String(device.host || device.ip || "").trim();
  return {
    id: String(device.id || ((brand + "-" + host) || "tv-" + Date.now())).slice(0, 180),
    name: String(device.name || device.model || profileFor(brand).manufacturer).trim().slice(0, 40),
    brand,
    manufacturer: String(device.manufacturer || profileFor(brand).manufacturer).slice(0, 60),
    system: String(device.system || profileFor(brand).system).slice(0, 80),
    protocol: String(device.protocol || profileFor(brand).protocol || "").slice(0, 100),
    model: String(device.model || "").slice(0, 80),
    host: host.slice(0, 255),
    room: String(device.room || "Zuhause").trim().slice(0, 40) || "Zuhause",
    paired: device.paired === true,
    lastSeen: Number.isFinite(Number(device.lastSeen)) ? Number(device.lastSeen) : null,
    connected: false,
    capabilities: []
  };
}

export function loadDevices() {
  try {
    const values = JSON.parse(localStorage.getItem(KEY) || "[]");
    return Array.isArray(values) ? values.map(normalize).filter(Boolean) : [];
  } catch {
    return [];
  }
}

export function saveDevices(devices) {
  try {
    const safe = devices.map(normalize).filter(Boolean).map(device => {
      const stored = {};
      for (const field of SAFE_FIELDS) stored[field] = device[field];
      return stored;
    });
    localStorage.setItem(KEY, JSON.stringify(safe));
    return true;
  } catch {
    return false;
  }
}

export function normalizeDiscovered(value) {
  if (!value || typeof value !== "object") return null;
  const profileValue = String(value.profile || value.brand || value.protocol || "").toLowerCase();
  let brand = profileValue;
  if (!profileFor(brand)) {
    if (profileValue.includes("roku")) brand = "roku";
    else if (profileValue.includes("webos")) brand = "lg";
    else if (profileValue.includes("samsung") || profileValue.includes("tizen")) brand = "samsung";
    else if (profileValue.includes("sony") || profileValue.includes("bravia")) brand = "sony";
    else if (profileValue.includes("philips") || profileValue.includes("jointspace")) brand = "philips";
    else if (profileValue.includes("vida")) brand = "vidaa";
    else if (profileValue.includes("fire")) brand = "firetv";
    else if (profileValue.includes("tcl")) brand = "tcl";
    else if (profileValue.includes("hisense")) brand = "hisense_android";
    else if (profileValue.includes("android") || profileValue.includes("google")) brand = "android";
  }
  if (!profileFor(brand)) return null;
  const normalized = normalize({ ...value, brand, host: value.host || value.ip || "" });
  if (!normalized) return null;
  normalized.id = String(value.id || normalized.id || (brand + "-" + normalized.host)).slice(0, 180);
  normalized.name = String(value.name || value.userDeviceName || value.model || profileFor(brand).manufacturer).slice(0, 40);
  normalized.model = String(value.model || value.modelName || "").slice(0, 80);
  normalized.lastSeen = Date.now();
  return normalized;
}
