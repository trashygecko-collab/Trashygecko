// Playwright-Konfiguration fuer die Spiele-Suite.
//
// Die Spiele laden gemeinsame Dateien absolut als /spiele/... (lob.js, netz.js
// und Bilder).  Der HTTP-Server laeuft deshalb im Elternordner "Homepage",
// nicht in "spiele" - sonst waeren diese Pfade 404 und die Tests wuerden
// Fehler melden, die es auf dem echten Server nicht gibt.
//
// Berichte und Testreste gehoeren nicht in die Homepage: sie landen im
// Arbeitsordner unter werkzeuge/testberichte (..\..\arbeit).
const path = require("path");
const { defineConfig } = require("@playwright/test");

const PORT = 4173;
const WURZEL = path.resolve(__dirname, "..");
const BERICHTE = path.resolve(WURZEL, "..", "arbeit", "werkzeuge", "testberichte");

module.exports = defineConfig({
    testDir: "./tests",
    // Referenzbilder ohne Plattform-Zusatz: Playwright haengt sonst -win32 bzw.
    // -linux an, und eine andere Maschine findet ihre eigenen Bilder nicht.
    snapshotPathTemplate: "{testDir}/{testFilePath}-snapshots/{arg}-{projectName}{ext}",
    outputDir: path.join(BERICHTE, "test-results"),
    timeout: 30000,
    fullyParallel: true,
    workers: 4,
    retries: 0,
    reporter: [["list"], ["html", { open: "never", outputFolder: path.join(BERICHTE, "playwright-report") }]],
    use: {
        baseURL: `http://127.0.0.1:${PORT}/spiele/`,
        browserName: "chromium",
        headless: true,
        trace: "retain-on-failure",
        video: "off",
    },
    webServer: {
        command: `python -m http.server ${PORT}`,
        cwd: WURZEL,
        port: PORT,
        reuseExistingServer: false,
        timeout: 30000,
    },
    projects: [{ name: "chromium" }],
});
