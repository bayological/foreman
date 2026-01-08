#!/bin/bash
cd "$(dirname "$0")/.."

echo "🔨 Ralph starting..."
echo "📁 Directory: $(pwd)"
echo ""

while :; do
    echo "=== $(date) ==="
    echo "⏳ Claude is thinking (this may take 1-2 minutes)..."
    
    # Use script to force TTY and show real-time output
    script -q /dev/null -c "cat .ralph/PROMPT.md | claude -p --dangerously-skip-permissions"
    
    echo ""
    echo "✅ Iteration complete"
    echo "⏳ Next iteration in 5 seconds..."
    sleep 5
done