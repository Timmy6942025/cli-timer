const fs = require("fs");
const os = require("os");
const path = require("path");
const { spawnSync } = require("child_process");
const figlet = require("figlet");

const PROJECT_ROOT = path.join(__dirname, "..");
const CONFIG_DIR = path.join(os.homedir(), ".cli-timer");
const CONFIG_PATH = path.join(CONFIG_DIR, "config.json");
const SETTINGS_STATE_PATH = path.join(CONFIG_DIR, "settings-state.json");
const DEFAULT_FONT = "Standard";
const TIMER_SAMPLE_TEXT = "00:00:00";

const MIN_TICK_RATE_MS = 50;
const MAX_TICK_RATE_MS = 1000;

const DEFAULT_KEYBINDINGS = Object.freeze({
  pauseKey: "p",
  pauseAltKey: "space",
  restartKey: "r",
  exitKey: "q",
  exitAltKey: "e"
});

const LEGACY_DEFAULT_KEYBINDINGS = Object.freeze({
  pauseKey: "p",
  pauseAltKey: "space",
  restartKey: "r",
  exitKey: "s",
  exitAltKey: "e"
});

const DEFAULT_CONFIG = Object.freeze({
  font: DEFAULT_FONT,
  centerDisplay: true,
  showHeader: true,
  showControls: true,
  tickRateMs: 100,
  completionMessage: "Time is up!",
  keybindings: { ...DEFAULT_KEYBINDINGS }
});

let allFontsCache = null;
let compatibleFontsSlowCache = null;

function clearScreen() {
  process.stdout.write("\x1b[2J\x1b[H");
}

function hideCursor() {
  process.stdout.write("\x1b[?25l");
}

function showCursor() {
  process.stdout.write("\x1b[?25h");
}

function formatHms(totalSeconds) {
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;
  return [hours, minutes, seconds].map((n) => String(n).padStart(2, "0")).join(":");
}

function getAllFonts() {
  if (allFontsCache) {
    return allFontsCache;
  }

  try {
    allFontsCache = figlet.fontsSync().slice().sort((a, b) => a.localeCompare(b));
  } catch (_error) {
    allFontsCache = [DEFAULT_FONT];
  }

  return allFontsCache;
}

function hasVisibleGlyphs(text) {
  return typeof text === "string" && /[^\s]/.test(text);
}

function renderWithFont(text, fontName) {
  try {
    return figlet.textSync(text, {
      font: fontName,
      horizontalLayout: "fitted",
      verticalLayout: "default",
      width: process.stdout.columns || 120,
      whitespaceBreak: true
    });
  } catch (_error) {
    return "";
  }
}

function isTimerCompatibleFont(fontName) {
  return hasVisibleGlyphs(renderWithFont(TIMER_SAMPLE_TEXT, fontName));
}

function getTimerCompatibleFontsSlow() {
  if (compatibleFontsSlowCache) {
    return compatibleFontsSlowCache;
  }

  const fonts = getAllFonts();
  compatibleFontsSlowCache = fonts.filter((fontName) => isTimerCompatibleFont(fontName));
  if (compatibleFontsSlowCache.length === 0) {
    compatibleFontsSlowCache = [DEFAULT_FONT];
  }
  return compatibleFontsSlowCache;
}

function normalizeFontName(input) {
  const fonts = getAllFonts();
  const exact = fonts.find((name) => name === input);
  if (exact) {
    return exact;
  }
  const lowerInput = input.toLowerCase();
  return fonts.find((name) => name.toLowerCase() === lowerInput) || null;
}

function sanitizeTickRate(raw) {
  if (!Number.isFinite(raw)) {
    return DEFAULT_CONFIG.tickRateMs;
  }

  const value = Math.floor(raw);
  if (value < MIN_TICK_RATE_MS) {
    return MIN_TICK_RATE_MS;
  }
  if (value > MAX_TICK_RATE_MS) {
    return MAX_TICK_RATE_MS;
  }
  return value;
}

function normalizeCompletionMessage(raw) {
  if (typeof raw !== "string") {
    return DEFAULT_CONFIG.completionMessage;
  }

  return raw.replace(/\r/g, "").replace(/\n/g, " ").slice(0, 240);
}

function normalizeKeyToken(raw, fallback) {
  if (typeof raw !== "string") {
    return fallback;
  }

  const value = raw.trim().toLowerCase();
  if (value === "space") {
    return "space";
  }

  if (/^[!-~]$/.test(value)) {
    return value;
  }

  return fallback;
}

function normalizeKeybindings(raw) {
  const next = { ...DEFAULT_KEYBINDINGS };
  if (!raw || typeof raw !== "object") {
    return next;
  }

  next.pauseKey = normalizeKeyToken(raw.pauseKey, next.pauseKey);
  next.pauseAltKey = normalizeKeyToken(raw.pauseAltKey, next.pauseAltKey);
  next.restartKey = normalizeKeyToken(raw.restartKey, next.restartKey);
  next.exitKey = normalizeKeyToken(raw.exitKey, next.exitKey);
  next.exitAltKey = normalizeKeyToken(raw.exitAltKey, next.exitAltKey);

  if (
    next.pauseKey === LEGACY_DEFAULT_KEYBINDINGS.pauseKey &&
    next.pauseAltKey === LEGACY_DEFAULT_KEYBINDINGS.pauseAltKey &&
    next.restartKey === LEGACY_DEFAULT_KEYBINDINGS.restartKey &&
    next.exitKey === LEGACY_DEFAULT_KEYBINDINGS.exitKey &&
    next.exitAltKey === LEGACY_DEFAULT_KEYBINDINGS.exitAltKey
  ) {
    return { ...DEFAULT_KEYBINDINGS };
  }

  return next;
}

function ensureConfigDir() {
  if (!fs.existsSync(CONFIG_DIR)) {
    fs.mkdirSync(CONFIG_DIR, { recursive: true });
  }
}

function normalizeConfig(raw) {
  const next = {
    font: DEFAULT_FONT,
    centerDisplay: DEFAULT_CONFIG.centerDisplay,
    showHeader: DEFAULT_CONFIG.showHeader,
    showControls: DEFAULT_CONFIG.showControls,
    tickRateMs: DEFAULT_CONFIG.tickRateMs,
    completionMessage: DEFAULT_CONFIG.completionMessage,
    keybindings: { ...DEFAULT_KEYBINDINGS }
  };

  if (raw && typeof raw === "object") {
    if (typeof raw.centerDisplay === "boolean") {
      next.centerDisplay = raw.centerDisplay;
    }
    if (typeof raw.showHeader === "boolean") {
      next.showHeader = raw.showHeader;
    }
    if (typeof raw.showControls === "boolean") {
      next.showControls = raw.showControls;
    }
    if (typeof raw.tickRateMs === "number") {
      next.tickRateMs = sanitizeTickRate(raw.tickRateMs);
    }
    if (typeof raw.completionMessage === "string") {
      next.completionMessage = normalizeCompletionMessage(raw.completionMessage);
    }
    next.keybindings = normalizeKeybindings(raw.keybindings);
    if (typeof raw.font === "string") {
      const normalizedFont = normalizeFontName(raw.font);
      if (normalizedFont && isTimerCompatibleFont(normalizedFont)) {
        next.font = normalizedFont;
      }
    }
  }

  return next;
}

function readConfig() {
  try {
    if (!fs.existsSync(CONFIG_PATH)) {
      return normalizeConfig({});
    }
    const text = fs.readFileSync(CONFIG_PATH, "utf8");
    const parsed = JSON.parse(text);
    return normalizeConfig(parsed);
  } catch (_error) {
    return normalizeConfig({});
  }
}

function writeConfig(config) {
  ensureConfigDir();
  const normalized = normalizeConfig(config);
  fs.writeFileSync(CONFIG_PATH, `${JSON.stringify(normalized, null, 2)}\n`, "utf8");
}

function updateConfig(patch) {
  const current = readConfig();
  const merged = { ...current, ...patch };
  writeConfig(merged);
  return readConfig();
}

function getFontFromConfig() {
  return readConfig().font || DEFAULT_FONT;
}

function setFontInConfig(requestedFont) {
  const normalized = normalizeFontName(requestedFont);
  if (!normalized) {
    return { ok: false, reason: "unknown", font: null };
  }
  if (!isTimerCompatibleFont(normalized)) {
    return { ok: false, reason: "incompatible", font: null };
  }
  const updated = updateConfig({ font: normalized });
  return { ok: true, reason: null, font: updated.font };
}

function parseDurationArgs(args) {
  if (args.length === 0 || args.length % 2 !== 0) {
    return { ok: false, error: "Duration must be in <number> <unit> pairs." };
  }

  const unitToSeconds = {
    h: 3600,
    hr: 3600,
    hrs: 3600,
    hour: 3600,
    hours: 3600,
    m: 60,
    min: 60,
    mins: 60,
    minute: 60,
    minutes: 60,
    s: 1,
    sec: 1,
    secs: 1,
    second: 1,
    seconds: 1
  };

  let totalSeconds = 0;

  for (let index = 0; index < args.length; index += 2) {
    const numberText = args[index];
    const unitText = args[index + 1];
    const value = Number(numberText);

    if (!Number.isFinite(value) || value < 0) {
      return { ok: false, error: `Invalid duration number: ${numberText}` };
    }

    if (!Number.isInteger(value)) {
      return { ok: false, error: `Duration number must be an integer: ${numberText}` };
    }

    const multiplier = unitToSeconds[unitText.toLowerCase()];
    if (!multiplier) {
      return { ok: false, error: `Unknown unit: ${unitText}` };
    }

    totalSeconds += value * multiplier;
  }

  if (totalSeconds <= 0) {
    return { ok: false, error: "Total duration must be greater than zero." };
  }

  return { ok: true, totalSeconds };
}

function renderTimeAscii(timeText, fontName) {
  const preferred = renderWithFont(timeText, fontName);
  if (hasVisibleGlyphs(preferred)) {
    return preferred;
  }

  const fallback = renderWithFont(timeText, DEFAULT_FONT);
  if (hasVisibleGlyphs(fallback)) {
    return fallback;
  }

  return `${timeText}\n`;
}

function keyTokenToLabel(token) {
  if (token === "space") {
    return "Spacebar";
  }
  return token;
}

function controlsHelpLine(keybindings) {
  const pause = `${keyTokenToLabel(keybindings.pauseKey)}/${keyTokenToLabel(keybindings.pauseAltKey)}`;
  const restart = keyTokenToLabel(keybindings.restartKey);
  const exit = `${keyTokenToLabel(keybindings.exitKey)}/${keyTokenToLabel(keybindings.exitAltKey)}/Ctrl+C`;
  return `Controls: ${pause} Pause-Resume | ${restart} Restart | ${exit} Exit`;
}

function keyTokenFromInput(chunk) {
  if (chunk === " ") {
    return "space";
  }
  if (chunk.length !== 1) {
    return null;
  }
  return chunk.toLowerCase();
}

function toDisplayLines(text) {
  const lines = String(text).replace(/\r/g, "").split("\n");
  while (lines.length > 0 && lines[lines.length - 1] === "") {
    lines.pop();
  }
  return lines.length > 0 ? lines : [""];
}

function writeCenteredBlock(lines) {
  const safeLines = lines.length > 0 ? lines : [""];
  const terminalWidth = process.stdout.columns || 120;
  const terminalHeight = process.stdout.rows || safeLines.length;
  const blockWidth = safeLines.reduce((max, line) => Math.max(max, line.length), 0);

  const padLeft = Math.max(0, Math.floor((terminalWidth - blockWidth) / 2));
  const padTop = Math.max(0, Math.floor((terminalHeight - safeLines.length) / 2));
  const leftPrefix = " ".repeat(padLeft);

  let output = "";
  if (padTop > 0) {
    output += "\n".repeat(padTop);
  }
  output += safeLines.map((line) => `${leftPrefix}${line}`).join("\n");
  output += "\n";
  process.stdout.write(output);
}

function writeCenteredBlockWithTop(topLines, centerLines) {
  const safeTop = topLines.length > 0 ? topLines : [];
  const safeCenter = centerLines.length > 0 ? centerLines : [""];
  const terminalWidth = process.stdout.columns || 120;
  const terminalHeight = process.stdout.rows || (safeTop.length + safeCenter.length);
  const blockWidth = safeCenter.reduce((max, line) => Math.max(max, line.length), 0);

  const padLeft = Math.max(0, Math.floor((terminalWidth - blockWidth) / 2));
  const availableHeight = Math.max(0, terminalHeight - safeTop.length);
  const padTop = Math.max(0, Math.floor((availableHeight - safeCenter.length) / 2));
  const leftPrefix = " ".repeat(padLeft);

  let output = "";
  if (safeTop.length > 0) {
    output += `${safeTop.join("\n")}\n`;
  }
  if (padTop > 0) {
    output += "\n".repeat(padTop);
  }
  output += safeCenter.map((line) => `${leftPrefix}${line}`).join("\n");
  output += "\n";
  process.stdout.write(output);
}

function drawFrame({ mode, seconds, paused, config, done }) {
  clearScreen();

  const topLines = [];
  const centerLines = [];
  const title = mode === "timer" ? "Timer" : "Stopwatch";

  if (config.showHeader) {
    topLines.push(`${title} | Font: ${config.font}`);
  }
  if (config.showControls) {
    topLines.push(controlsHelpLine(config.keybindings));
  }
  if (topLines.length > 0) {
    topLines.push("");
  }

  centerLines.push(...toDisplayLines(renderTimeAscii(formatHms(seconds), config.font)));

  if (done) {
    if (config.completionMessage) {
      centerLines.push("");
      centerLines.push(config.completionMessage);
    }
  } else if (paused) {
    centerLines.push("");
    centerLines.push("Paused");
  }

  if (config.centerDisplay) {
    writeCenteredBlockWithTop(topLines, centerLines);
  } else {
    const lines = [...topLines, ...centerLines];
    process.stdout.write(`${lines.join("\n")}\n`);
  }
}

function runNonInteractiveTimer(initialSeconds, tickRateMs) {
  const startedAt = Date.now();
  let lastSecond = null;

  function printRemaining() {
    const elapsed = Math.floor((Date.now() - startedAt) / 1000);
    const remaining = Math.max(0, initialSeconds - elapsed);
    if (remaining === lastSecond) {
      return remaining;
    }
    lastSecond = remaining;
    process.stdout.write(`${formatHms(remaining)}\n`);
    return remaining;
  }

  if (printRemaining() <= 0) {
    return;
  }

  const interval = setInterval(() => {
    const remaining = printRemaining();
    if (remaining <= 0) {
      clearInterval(interval);
    }
  }, tickRateMs);
}

function runClock({ mode, initialSeconds, config }) {
  const isTimer = mode === "timer";
  const tickRateMs = sanitizeTickRate(config.tickRateMs);

  if (!process.stdin.isTTY || !process.stdout.isTTY) {
    if (!isTimer) {
      process.stderr.write("Stopwatch requires an interactive terminal (TTY).\n");
      process.exitCode = 1;
      return;
    }
    runNonInteractiveTimer(initialSeconds, tickRateMs);
    return;
  }

  let paused = false;
  let done = false;
  const baseSeconds = initialSeconds;
  let anchorMs = Date.now();
  let elapsedWhilePaused = 0;
  let tick = null;
  let lastDrawState = "";
  let hasExited = false;

  const stdin = process.stdin;

  function getElapsedSeconds() {
    const elapsedMs = paused ? elapsedWhilePaused : elapsedWhilePaused + (Date.now() - anchorMs);
    return Math.floor(elapsedMs / 1000);
  }

  function getDisplaySeconds() {
    const elapsed = getElapsedSeconds();
    if (isTimer) {
      return Math.max(0, baseSeconds - elapsed);
    }
    return elapsed;
  }

  function refreshDoneState(displaySeconds) {
    if (isTimer && displaySeconds <= 0) {
      done = true;
      paused = true;
    }
  }

  function draw(force) {
    const displaySeconds = getDisplaySeconds();
    refreshDoneState(displaySeconds);
    const stateKey = `${displaySeconds}|${paused ? 1 : 0}|${done ? 1 : 0}|${process.stdout.columns || 0}|${process.stdout.rows || 0}`;
    if (!force && stateKey === lastDrawState) {
      return;
    }
    lastDrawState = stateKey;
    drawFrame({
      mode,
      seconds: displaySeconds,
      paused,
      config,
      done
    });
  }

  function restart() {
    paused = false;
    done = false;
    elapsedWhilePaused = 0;
    anchorMs = Date.now();
    lastDrawState = "";
    draw(true);
  }

  function onSignal() {
    cleanupAndExit(0);
  }

  function onResize() {
    lastDrawState = "";
    draw(true);
  }

  function cleanupAndExit(code) {
    if (hasExited) {
      return;
    }
    hasExited = true;
    if (tick !== null) {
      clearInterval(tick);
    }
    stdin.removeListener("data", onKeypress);
    process.removeListener("SIGINT", onSignal);
    process.removeListener("SIGTERM", onSignal);
    process.removeListener("SIGWINCH", onResize);
    if (stdin.isTTY && typeof stdin.setRawMode === "function") {
      stdin.setRawMode(false);
    }
    stdin.pause();
    showCursor();
    clearScreen();
    process.exit(code);
  }

  function togglePause() {
    if (paused) {
      paused = false;
      anchorMs = Date.now();
      return;
    }
    paused = true;
    elapsedWhilePaused += Date.now() - anchorMs;
  }

  function onKeypress(chunk) {
    const key = String(chunk);
    if (key === "\u0003") {
      cleanupAndExit(0);
      return;
    }

    const token = keyTokenFromInput(key);
    if (!token) {
      return;
    }

    if (token === config.keybindings.pauseKey || token === config.keybindings.pauseAltKey) {
      if (!done) {
        togglePause();
        draw(true);
      }
      return;
    }

    if (token === config.keybindings.restartKey) {
      restart();
      return;
    }

    if (token === config.keybindings.exitKey || token === config.keybindings.exitAltKey) {
      cleanupAndExit(0);
    }
  }

  process.on("SIGINT", onSignal);
  process.on("SIGTERM", onSignal);
  process.on("SIGWINCH", onResize);

  if (stdin.isTTY && typeof stdin.setRawMode === "function") {
    stdin.setRawMode(true);
  }
  stdin.resume();
  stdin.setEncoding("utf8");
  stdin.on("data", onKeypress);
  hideCursor();

  draw(true);
  tick = setInterval(() => draw(false), tickRateMs);
}

function runSettingsUI() {
  if (!process.stdin.isTTY || !process.stdout.isTTY) {
    process.stderr.write("`timer settings` requires an interactive terminal (TTY).\n");
    process.exitCode = 1;
    return;
  }

  ensureConfigDir();

  const goVersion = spawnSync("go", ["version"], { stdio: "ignore" });
  if (goVersion.error || goVersion.status !== 0) {
    process.stderr.write("Go is required for `timer settings` (Bubble Tea UI).\n");
    process.exitCode = 1;
    return;
  }

  const state = {
    configPath: CONFIG_PATH,
    config: readConfig(),
    fonts: getAllFonts()
  };

  fs.writeFileSync(SETTINGS_STATE_PATH, JSON.stringify(state), "utf8");

  const result = spawnSync("go", ["run", ".", "--state", SETTINGS_STATE_PATH], {
    cwd: path.join(PROJECT_ROOT, "settings-ui"),
    stdio: "inherit"
  });

  try {
    fs.unlinkSync(SETTINGS_STATE_PATH);
  } catch (_error) {
  }

  if (result.error) {
    process.stderr.write(`Failed to run settings UI: ${result.error.message}\n`);
    process.exitCode = 1;
    return;
  }

  if (result.status === 2) {
    return;
  }

  if (typeof result.status === "number" && result.status !== 0) {
    process.exitCode = result.status;
  }
}

function printUsage() {
  process.stdout.write("Usage\n\n");
  process.stdout.write("Stopwatch\n");
  process.stdout.write("  stopwatch\n\n");
  process.stdout.write("Timer\n");
  process.stdout.write("  timer <number> <hr/hrs/min/sec> [<number> <hr/hrs/min/sec> ...]\n");
  process.stdout.write("  Example: timer 5 min 2 sec\n\n");
  process.stdout.write("Settings\n");
  process.stdout.write("  timer settings\n\n");
  process.stdout.write("Controls\n");
  process.stdout.write("  Defaults: p/Space Pause-Resume | r Restart | q/e/Ctrl+C Exit\n");
  process.stdout.write("  Keybindings are customizable in `timer settings`.\n\n");
  process.stdout.write("Font Styles\n");
  process.stdout.write("  timer style\n");
  process.stdout.write("  timer style --compatible\n");
  process.stdout.write("  timer style <font>\n");
}

function runStopwatch() {
  const config = readConfig();
  runClock({ mode: "stopwatch", initialSeconds: 0, config });
}

function runTimer(args) {
  if (args.length === 0) {
    printUsage();
    process.exitCode = 1;
    return;
  }

  if (args[0] === "settings") {
    runSettingsUI();
    return;
  }

  if (args[0] === "style") {
    if (args.length === 1 || (args.length === 2 && args[1] === "--all")) {
      const currentFont = getFontFromConfig();
      const fonts = getAllFonts();
      process.stdout.write(`Current font: ${currentFont}\n\n`);
      process.stdout.write("Available fonts:\n");
      for (const font of fonts) {
        process.stdout.write(`${font}\n`);
      }
      process.stdout.write("\nTip: Some fonts do not support timer digits.\n");
      process.stdout.write("Use `timer style <font>` to validate and set safely.\n");
      return;
    }

    if (args.length === 2 && args[1] === "--compatible") {
      process.stdout.write("Checking font compatibility for timer digits...\n\n");
      const currentFont = getFontFromConfig();
      const fonts = getTimerCompatibleFontsSlow();
      process.stdout.write(`Current font: ${currentFont}\n\n`);
      process.stdout.write("Timer-compatible fonts:\n");
      for (const font of fonts) {
        process.stdout.write(`${font}\n`);
      }
      return;
    }

    const requestedFont = args.slice(1).join(" ");
    const result = setFontInConfig(requestedFont);
    if (!result.ok) {
      if (result.reason === "incompatible") {
        process.stderr.write(`Font is incompatible with timer digits: ${requestedFont}\n`);
      } else {
        process.stderr.write(`Unknown font: ${requestedFont}\n`);
      }
      process.stderr.write("Run `timer style` to list fonts.\n");
      process.stderr.write("Run `timer style --compatible` to list only compatible fonts.\n");
      process.exitCode = 1;
      return;
    }
    process.stdout.write(`Font set to: ${result.font}\n`);
    return;
  }

  const parsed = parseDurationArgs(args);
  if (!parsed.ok) {
    process.stderr.write(`${parsed.error}\n\n`);
    printUsage();
    process.exitCode = 1;
    return;
  }

  const config = readConfig();
  runClock({ mode: "timer", initialSeconds: parsed.totalSeconds, config });
}

module.exports = {
  runTimer,
  runStopwatch,
  runSettingsUI,
  printUsage,
  parseDurationArgs,
  getAllFonts,
  getTimerCompatibleFontsSlow,
  getFontFromConfig,
  readConfig,
  setFontInConfig,
  DEFAULT_FONT,
  DEFAULT_CONFIG
};
