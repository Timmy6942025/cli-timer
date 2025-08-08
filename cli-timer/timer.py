import time
import sys
import threading

size = 2
paused = False
reset_flag = False
exit_flag = False

# Your display_time and key_listener functions remain the same.
def display_time(seconds):
    # ... (code as is)

def key_listener():
    # ... (code as is)

def start_timer(total_seconds):
    """This function starts the countdown timer."""
    global paused, reset_flag, exit_flag
    original_seconds = total_seconds

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
