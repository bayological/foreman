#!/bin/bash
cd "$(dirname "$0")/.."

while :; do
    cat .ralph/PROMPT.md | claude -p --dangerously-skip-permissions
    sleep 5
done