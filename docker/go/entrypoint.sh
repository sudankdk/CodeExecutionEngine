#!/bin/sh
set -e

GO_FILE="$1"
STDIN_FILE="$2"

# Compile
go build -o program "$GO_FILE"

# Run with stdin if provided
if [ -n "$STDIN_FILE" ] && [ -f "$STDIN_FILE" ]; then
    ./program < "$STDIN_FILE"
else
    ./program
fi
