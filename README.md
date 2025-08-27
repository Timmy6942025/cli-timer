# CLI Timer

A simple, easy-to-use command-line timer and stopwatch that runs in your terminal.

## Installation

```bash
# Recommended (single command — requires git):
sh -c 'tmpdir=$(mktemp -d) && git clone https://github.com/Timmy6942025/cli-timer.git "$tmpdir/cli-timer" && (cd "$tmpdir/cli-timer" && ./install.sh) && rm -rf "$tmpdir"'

# Alternative (one-liner, works if git is not available — tries curl then wget):
sh -c 'URL="https://github.com/Timmy6942025/cli-timer/archive/refs/heads/master.tar.gz"; tmpdir=$(mktemp -d) && cd "$tmpdir" || exit 1; if command -v curl >/dev/null 2>&1; then curl -fsSL "$URL" | tar xz --strip-components=1; elif command -v wget >/dev/null 2>&1; then wget -qO- "$URL" | tar xz --strip-components=1; else echo "Error: install curl, wget, or git and re-run." >&2; exit 1; fi; ./install.sh && cd - >/dev/null 2>&1 || true; rm -rf "$tmpdir"'

This uses a temporary directory so you can paste the command from any folder — it will fetch the repo, run `install.sh`, and clean up the temporary files. The installer will attempt to ensure Python3/pip and the `pyfiglet` package are available and will create `timer` and `stopwatch` executables.
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

## Controls

-   **p** or **Spacebar**: Pause/Resume
-   **r**: Restart
-   **s**, **e** or **Ctrl+C**: Exit

---

## Font Styles

You can view and set the ASCII font style used for the timer and stopwatch display.

- To list all available fonts:

```bash
timer style
```

- To set your preferred font:

```bash
timer style <font>
```

Replace `<font>` with any font name from the list shown by `timer style`.
