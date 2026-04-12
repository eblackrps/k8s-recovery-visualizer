import { spawn } from "node:child_process";
import { setTimeout as delay } from "node:timers/promises";
import path from "node:path";
import process from "node:process";
import { chromium } from "playwright";

const frontendDir = process.cwd();
const repoRoot = path.resolve(frontendDir, "..", "..");
const previewUrl = "http://127.0.0.1:4173";

const pages = [
  { url: `${previewUrl}/?view=home`, file: path.join(repoRoot, "images", "gui-dashboard.png") },
  { url: `${previewUrl}/?view=scan`, file: path.join(repoRoot, "images", "gui-scan-wizard.png") },
  { url: `${previewUrl}/?view=live`, file: path.join(repoRoot, "images", "gui-live-run.png") },
  { url: `${previewUrl}/?view=results&tab=Findings`, file: path.join(repoRoot, "images", "gui-results-findings.png") },
  { url: `${previewUrl}/?view=results&tab=Compare`, file: path.join(repoRoot, "images", "gui-compare.png") },
];

const previewCommand =
  process.platform === "win32"
    ? { command: "cmd.exe", args: ["/c", "npm", "run", "preview"] }
    : { command: "npm", args: ["run", "preview"] };

const preview = spawn(previewCommand.command, previewCommand.args, {
  cwd: frontendDir,
  stdio: "inherit",
});

const stopPreview = () => {
  if (!preview.killed) {
    preview.kill();
  }
};

process.on("exit", stopPreview);
process.on("SIGINT", () => {
  stopPreview();
  process.exit(130);
});

async function waitForServer() {
  for (let attempt = 0; attempt < 50; attempt += 1) {
    try {
      const response = await fetch(previewUrl, { method: "GET" });
      if (response.ok) {
        return;
      }
    } catch {
      // ignore until the preview server is ready
    }
    await delay(500);
  }
  throw new Error("Vite preview did not start in time.");
}

async function main() {
  await waitForServer();
  const browser = await chromium.launch();
  const page = await browser.newPage({ viewport: { width: 1440, height: 980 } });
  for (const shot of pages) {
    await page.goto(shot.url, { waitUntil: "networkidle" });
    await page.screenshot({ path: shot.file, fullPage: true });
  }
  await browser.close();
}

main()
  .then(() => {
    stopPreview();
  })
  .catch((error) => {
    console.error(error);
    stopPreview();
    process.exitCode = 1;
  });
