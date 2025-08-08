#!/bin/bash

# Make the main script executable
chmod +x cli_timer.py

# Create symbolic links to the executables
ln -s "$(pwd)/cli_timer.py" /usr/local/bin/timer
ln -s "$(pwd)/cli_timer.py" /usr/local/bin/stopwatch

# Print a success message
echo "Timer and stopwatch installed successfully!"
echo "You can now use the 'timer' and 'stopwatch' commands."
