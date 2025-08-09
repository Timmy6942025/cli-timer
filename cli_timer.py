#!/usr/bin/env python3
import argparse
import os
import sys
import termios
import threading
import time
import tty
from pyfiglet import Figlet, FigletFont

def get_key():
    old_settings = termios.tcgetattr(sys.stdin)
    tty.setcbreak(sys.stdin.fileno())
    try:
        while True:
            b = os.read(sys.stdin.fileno(), 3).decode()
            if len(b) == 1:
                return b
    finally:
        termios.tcsetattr(sys.stdin, termios.TCSADRAIN, old_settings)


class Timer:
    def __init__(self, duration, font="standard"):
        self.duration = duration
        self.remaining = duration
        self.paused = False
        self.running = False
        self.thread = None
        self.fig = Figlet(font=font)
        self.lines = 0

    def start(self):
        self.running = True
        self.thread = threading.Thread(target=self.run)
        self.thread.start()

    def run(self):
        while self.remaining > 0 and self.running:
            if not self.paused:
                self.remaining -= 1
                self.display()
                time.sleep(1)
        if self.remaining == 0:
            print("Time's up!")
            self.running = False

    def display(self):
        if hasattr(self, 'lines') and self.lines > 0:
            sys.stdout.write(f'\x1b[{self.lines}A')
            sys.stdout.write('\r')
        mins, secs = divmod(self.remaining, 60)
        hours, mins = divmod(mins, 60)
        time_str = f"{hours:02d}:{mins:02d}:{secs:02d}"
        output = self.fig.renderText(time_str)
        sys.stdout.write(output)
        sys.stdout.flush()
        self.lines = output.count('\n')

    def toggle_pause(self):
        self.paused = not self.paused

    def restart(self):
        self.remaining = self.duration

    def stop(self):
        self.running = False


class Stopwatch:
    def __init__(self, font="standard"):
        self.elapsed = 0
        self.paused = False
        self.running = False
        self.thread = None
        self.fig = Figlet(font=font)
        self.lines = 0

    def start(self):
        self.running = True
        self.thread = threading.Thread(target=self.run)
        self.thread.start()

    def run(self):
        while self.running:
            if not self.paused:
                self.elapsed += 1
                self.display()
                time.sleep(1)

    def display(self):
        if hasattr(self, 'lines') and self.lines > 0:
            sys.stdout.write(f'\x1b[{self.lines}A')
            sys.stdout.write('\r')
        mins, secs = divmod(self.elapsed, 60)
        hours, mins = divmod(mins, 60)
        time_str = f"{hours:02d}:{mins:02d}:{secs:02d}"
        output = self.fig.renderText(time_str)
        sys.stdout.write(output)
        sys.stdout.flush()
        self.lines = output.count('\n')

    def toggle_pause(self):
        self.paused = not self.paused

    def restart(self):
        self.elapsed = 0

    def stop(self):
        self.running = False


def get_font():
    try:
        with open(os.path.expanduser("~/.cli-timer-font"), "r") as f:
            return f.read().strip()
    except FileNotFoundError:
        return "standard"


def main():
    parser = argparse.ArgumentParser(description="A simple command-line timer and stopwatch.")

    # Determine if invoked as 'timer' or 'stopwatch'
    script_name = os.path.basename(sys.argv[0])

    if script_name == "timer":
        parser.add_argument("duration", nargs="+", help="e.g., 5 min 2 sec")
        args = parser.parse_args(sys.argv[1:]) # Parse arguments excluding the script name
        args.command = "timer"

        duration = 0
        for i in range(0, len(args.duration), 2):
            value = int(args.duration[i])
            unit = args.duration[i + 1]
            if unit in ["hr", "hrs"]:
                duration += value * 3600
            elif unit == "min":
                duration += value * 60
            elif unit == "sec":
                duration += value

        timer = Timer(duration, font=get_font())
        timer.start()

        while timer.running:
            key = get_key()
            if key == "p":
                timer.toggle_pause()
            elif key == "r":
                timer.restart()
            elif key == "e" or key == "\x03":
                timer.stop()

    elif script_name == "stopwatch":
        args = parser.parse_args(sys.argv[1:]) # No arguments expected for stopwatch
        args.command = "stopwatch"

        stopwatch = Stopwatch(font=get_font())
        stopwatch.start()

        while stopwatch.running:
            key = get_key()
            if key == "p":
                stopwatch.toggle_pause()
            elif key == "r":
                stopwatch.restart()
            elif key == "e" or key == "\x03":
                stopwatch.stop()

    elif script_name == "cli-timer-style":
        parser.add_argument("font", nargs="?", help="The font to use.")
        args = parser.parse_args(sys.argv[1:])

        if args.font:
            if args.font in FigletFont.getFonts():
                with open(os.path.expanduser("~/.cli-timer-font"), "w") as f:
                    f.write(args.font)
                print(f"Font updated to {args.font} successfully!")
            else:
                print("Invalid font.")
        else:
            print("Available fonts:")
            for font in FigletFont.getFonts():
                print(font)

    else:
        # Fallback for direct execution or unknown invocation
        subparsers = parser.add_subparsers(dest="command")
        timer_parser = subparsers.add_parser("timer")
        timer_parser.add_argument("duration", nargs="+", help="e.g., 5 min 2 sec")
        subparsers.add_parser("stopwatch")
        args = parser.parse_args()

        if args.command == "timer":
            duration = 0
            for i in range(0, len(args.duration), 2):
                value = int(args.duration[i])
                unit = args.duration[i + 1]
                if unit in ["hr", "hrs"]:
                    duration += value * 3600
                elif unit == "min":
                    duration += value * 60
                elif unit == "sec":
                    duration += value

            timer = Timer(duration, font=get_font())
            timer.start()

            while timer.running:
                key = get_key()
                if key == "p":
                    timer.toggle_pause()
                elif key == "r":
                    timer.restart()
                elif key == "e" or key == "\x03":
                    timer.stop()

        elif args.command == "stopwatch":
            stopwatch = Stopwatch(font=get_font())
            stopwatch.start()

            while stopwatch.running:
                key = get_key()
                if key == "p":
                    stopwatch.toggle_pause()
                elif key == "r":
                    stopwatch.restart()
                elif key == "e" or key == "\x03":
                    stopwatch.stop()


if __name__ == "__main__":
    main()
