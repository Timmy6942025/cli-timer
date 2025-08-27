#!/usr/bin/env bash
set -euo pipefail

# Minimal installer that tries to make this project work with one pasted command.
# It will try to ensure python3, pip, and pyfiglet are available and then create
# 'timer' and 'stopwatch' links in /usr/local/bin (or fallback to ~/.local/bin).

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "Installing CLI Timer from: $ROOT_DIR"

has_cmd() { command -v "$1" >/dev/null 2>&1; }

detect_pkg_manager() {
    if has_cmd apt-get; then echo apt-get
    elif has_cmd dnf; then echo dnf
    elif has_cmd yum; then echo yum
    elif has_cmd pacman; then echo pacman
    elif has_cmd apk; then echo apk
    elif has_cmd brew; then echo brew
    else echo ""; fi
}

PKG_MANAGER=$(detect_pkg_manager)

ensure_python() {
    if ! has_cmd python3; then
        if [ -n "$PKG_MANAGER" ] && has_cmd sudo; then
            echo "python3 not found — installing via $PKG_MANAGER (requires sudo)."
            case "$PKG_MANAGER" in
                apt-get) sudo apt-get update && sudo apt-get install -y python3 python3-venv python3-pip ;; 
                dnf) sudo dnf install -y python3 python3-venv python3-pip ;; 
                yum) sudo yum install -y python3 python3-venv python3-pip ;; 
                pacman) sudo pacman -Sy --noconfirm python python-pip ;; 
                apk) sudo apk add --no-cache python3 py3-pip ;; 
                brew) brew install python ;; 
            esac
        else
            echo "python3 is required but not found. Please install Python 3 and re-run this script." >&2
            return 1
        fi
    fi
}

ensure_pip() {
    if ! python3 -m pip --version >/dev/null 2>&1; then
        if [ -n "$PKG_MANAGER" ] && has_cmd sudo; then
            echo "pip not found — attempting to install pip via $PKG_MANAGER (requires sudo)."
            case "$PKG_MANAGER" in
                apt-get) sudo apt-get update && sudo apt-get install -y python3-pip ;; 
                dnf) sudo dnf install -y python3-pip ;; 
                yum) sudo yum install -y python3-pip ;; 
                pacman) sudo pacman -Sy --noconfirm python-pip ;; 
                apk) sudo apk add --no-cache py3-pip ;; 
                brew) brew postinstall python ;; 
            esac
        else
            echo "pip for python3 not found. Will attempt to bootstrap pip via get-pip.py."
            curl -sS https://bootstrap.pypa.io/get-pip.py -o /tmp/get-pip.py
            python3 /tmp/get-pip.py --user
            rm -f /tmp/get-pip.py
        fi
    fi
}

install_pyfiglet() {
    if ! python3 -c "import pyfiglet" >/dev/null 2>&1; then
        echo "pyfiglet not found — installing via pip (user install)."
        python3 -m pip install --user pyfiglet
    else
        echo "pyfiglet already installed."
    fi
}

create_links() {
    TARGET_BIN="/usr/local/bin"
    USE_SUDO="true"

    if [ ! -w "$TARGET_BIN" ] || ! has_cmd sudo; then
        TARGET_BIN="$HOME/.local/bin"
        USE_SUDO="false"
    fi

    mkdir -p "$TARGET_BIN"

    echo "Installing executables to: $TARGET_BIN"

    # Remove existing links if present
    if [ "$USE_SUDO" = "true" ]; then
        sudo rm -f "$TARGET_BIN/timer" "$TARGET_BIN/stopwatch"
        sudo ln -sf "$ROOT_DIR/cli_timer.py" "$TARGET_BIN/timer"
        sudo ln -sf "$ROOT_DIR/cli_timer.py" "$TARGET_BIN/stopwatch"
    else
        rm -f "$TARGET_BIN/timer" "$TARGET_BIN/stopwatch" || true
        ln -sf "$ROOT_DIR/cli_timer.py" "$TARGET_BIN/timer"
        ln -sf "$ROOT_DIR/cli_timer.py" "$TARGET_BIN/stopwatch"
    fi

    chmod +x "$ROOT_DIR/cli_timer.py"

    echo "Created 'timer' and 'stopwatch' in $TARGET_BIN"

    if ! echo ":$PATH:" | grep -q ":$TARGET_BIN:"; then
        echo
        echo "Note: $TARGET_BIN is not in your PATH. To use 'timer' and 'stopwatch' directly add it to your PATH.";
        echo "For bash/zsh, you can run:";
        echo "  echo 'export PATH=\"\$PATH:$TARGET_BIN\"' >> \$HOME/.profile";
        echo "Then reopen your terminal or run: source ~/.profile";
    fi
}

main() {
    ensure_python || exit 1
    ensure_pip || true
    install_pyfiglet
    create_links

    echo
    echo "Installation complete. You can now run:"
    echo "  timer 5 min"
    echo "  stopwatch"
    echo "Change the font with: timer style <font> or list fonts: timer style"
}

main "$@"

