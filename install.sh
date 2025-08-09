#!/bin/bash

# Check if pyfiglet is installed
if ! python3 -c "import pyfiglet" &> /dev/null
then
    echo "pyfiglet could not be found. Please install it with one of the following commands:"
    echo "sudo apt-get install python3-pyfiglet"
    echo "pipx install pyfiglet"
    exit
fi

# Make the main script executable
chmod +x cli_timer.py

# Create symbolic links to the executables
sudo rm /usr/local/bin/timer /usr/local/bin/stopwatch /usr/local/bin/cli-timer-style
sudo ln -s "$(pwd)/cli_timer.py" /usr/local/bin/timer
sudo ln -s "$(pwd)/cli_timer.py" /usr/local/bin/stopwatch
sudo ln -s "$(pwd)/cli_timer.py" /usr/local/bin/cli-timer-style

# Print a success message
echo "Timer and stopwatch installed successfully!"
echo "You can now use the 'timer' and 'stopwatch' commands."
echo "You can change the style with the 'cli-timer-style' command."
