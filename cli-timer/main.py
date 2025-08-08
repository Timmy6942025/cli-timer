import sys
from .timer import run_timer
from .stopwatch import run_stopwatch

def main():
    if len(sys.argv) == 0:
        print("Usage: timer <minutes> [seconds] or stopwatch")
        return

    prog = sys.argv[0]
    if prog.endswith("timer"):
        args = sys.argv[1:]
        if not args:
            print("Usage: timer <minutes> [seconds]")
            return
        # Convert args to seconds
        total_seconds = 0
        for arg in args:
            if arg.isdigit():
                total_seconds += int(arg) * 60 if "min" in args else int(arg)
            elif arg.endswith("min"):
                total_seconds += int(arg[:-3]) * 60
            elif arg.endswith("sec"):
                total_seconds += int(arg[:-3])
        run_timer(total_seconds)

    elif prog.endswith("stopwatch"):
        run_stopwatch()
