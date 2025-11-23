#!/bin/sh

CODE_FILE=${1:-/run/code/main.py}
INPUT_FILE=${2:-/run/stdin/input.txt}

if [ -f "$INPUT_FILE" ]; then
    python "$CODE_FILE" < "$INPUT_FILE"
else
    python "$CODE_FILE"
fi
