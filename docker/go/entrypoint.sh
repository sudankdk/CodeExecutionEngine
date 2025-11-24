#!/bin/sh
GO_FILE="$1"
STDIN_FILE="$2"

if [ -n "$STDIN_FILE" ] && [ -f "$STDIN_FILE" ]; then
    go run "$GO_FILE" < "$STDIN_FILE"
else
    go run "$GO_FILE"
fi
