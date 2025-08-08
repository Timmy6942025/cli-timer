import sys
import cli_timer.timer
import cli_timer.stopwatch

def main():
    command = sys.argv[0].split('/')[-1]
    
    if command == "timer":
        parse_timer_args()
    elif command == "stopwatch":
        cli_timer.stopwatch.start_stopwatch()
    else:
        print("Unknown command. Use 'timer' or 'stopwatch'.")

def parse_timer_args():
    if len(sys.argv) < 3:
        print("Usage: timer <time> <unit>")
        sys.exit(1)

    try:
        value = int(sys.argv[1])
        unit = sys.argv[2].lower()
    except (ValueError, IndexError):
        print("Invalid time format. Use 'timer <number> <unit>'.")
        sys.exit(1)

    total_seconds = 0
    if unit in ["s", "sec", "second", "seconds"]:
        total_seconds = value
    elif unit in ["m", "min", "minute", "minutes"]:
        total_seconds = value * 60
    elif unit in ["h", "hr", "hour", "hours"]:
        total_seconds = value * 3600
    else:
        print("Invalid unit. Use 's', 'min', or 'hr'.")
        sys.exit(1)
    
    cli_timer.timer.start_timer(total_seconds)

if __name__ == "__main__":
    main()
