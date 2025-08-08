import time
import sys
import threading

size = 2
paused = False
reset_flag = False
exit_flag = False

def display_time(seconds):
    global size
    mins, secs = divmod(seconds, 60)
    hours, mins = divmod(mins, 60)
    spacing = " " * size
    sys.stdout.write(f"\r{spacing}{hours:02}:{mins:02}:{secs:02}{spacing}")
    sys.stdout.flush()

def key_listener():
    global paused, reset_flag, exit_flag, size
    while True:
        key = sys.stdin.read(1)
        if key == "p":
            paused = not paused
        elif key == "r":
            reset_flag = True
        elif key == "e":
            exit_flag = True
            break
        elif key == "+":
            size += 1
        elif key == "-":
            size = max(1, size - 1)

def run_timer(total_seconds):
    global paused, reset_flag, exit_flag
    paused = False
    reset_flag = False
    exit_flag = False

    threading.Thread(target=key_listener, daemon=True).start()

    while total_seconds >= 0:
        if exit_flag:
            break
        if reset_flag:
            reset_flag = False
            total_seconds = original_seconds
        if not paused:
            display_time(total_seconds)
            time.sleep(1)
            total_seconds -= 1
        else:
            time.sleep(0.1)
    print("\nDone!")
