# CLI Timer

A simple and customizable, easy to setup/use timer and stopwatch that runs in your terminal.

## Install

Install globally (this is all you need):

```bash
npm i -g @timmy6942025/cli-timer
# or
bun add -g @timmy6942025/cli-timer
```

Then use commands anywhere:

```bash
timer 5 min
stopwatch
```

## Usage

### Stopwatch

To use the stopwatch, simply run the following command:

```bash
stopwatch
```

### Timer

To use the timer, run the following command with the desired duration:

```bash
timer <number> <hr/hrs/min/sec> [<number> <hr/hrs/min/sec> ...]
```

For example:

```bash
timer 5 min 2 sec
```

By default, timer and stopwatch output is centered in the terminal.

### Update CLI Timer

To update globally and verify you are on latest:

```bash
timer update
```

`timer update` auto-detects whether your current install came from Bun, npm, or pnpm, runs the matching global update command, then checks your installed version against npm `latest`.

## Controls

- `p` or `Spacebar`: Pause/Resume
- `r`: Restart
- `f`: Random style/font
- `q`, `e` or `Ctrl+C`: Exit

## Font Styles

You can view and set the ASCII font style used for the timer and stopwatch display.

To list all available fonts:

```bash
timer style
```

This is instant and lists all figlet fonts.
`timer style --all` is supported as an explicit alias.

To list only timer-compatible fonts (fonts that render `01:23:45` visibly):

```bash
timer style --compatible
```

To set a random font:

```bash
timer style random
```

To set your preferred font:

```bash
timer style <font>
```

Replace `<font>` with any font name from the list shown by `timer style`.
If a font is missing native timer digits, CLI Timer now substitutes close glyphs so the display stays stylized instead of falling back to plain text.

## Settings UI

Open interactive settings UI:

```bash
timer settings
```

`timer settings` ships with prebuilt Bubble Tea binaries for:

- Linux x64 / arm64
- macOS x64 / arm64
- Windows x64 / arm64

If your platform is unsupported, it falls back to running the Go source (`go run`) when Go is installed.

This launches a Bubble Tea based screen where you can change:

- Font
- Center display
- Show header
- Show controls
- Tick rate (50-1000 ms)
- Completion message
- System notification on completion (default On)
- Completion sound/alarm on completion (default Off)
- Pause key / pause alt key
- Restart key
- Style key
- Exit key / exit alt key

When completion sound/alarm is enabled, it plays 5 terminal bell beeps.

Controls in settings UI:

- `Enter`: select/toggle
- `Ctrl+S`: save and exit
- `/`: filter fonts in font picker
- `/`: filter keys in key picker
- `Esc`/`q`: back/cancel

Note for macOS: If system notifications are inconsistent with built-in AppleScript notifications, install `terminal-notifier` (`brew install terminal-notifier`) for improved reliability.

Notification notes by platform:
- Windows: notifications depend on OS notification permissions and whether your desktop session allows toasts/balloons.
- Linux: notifications require a running desktop notification daemon (`notify-send` talks to that daemon over D-Bus).
