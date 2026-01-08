#!/bin/bash
# Foreman Ralph Runner
# Usage: ./ralph.sh

cd "$(dirname "$0")/.."

echo "🔨 Starting Ralph for Foreman SpecKit integration..."
echo "📁 Working directory: $(pwd)"
echo "📝 Prompt: .ralph/PROMPT.md"
echo ""
echo "Press Ctrl+C to stop"
echo "================================"

while :; do
    echo ""
    echo "🔄 Starting iteration at $(date)"
    echo "--------------------------------"
    
    cat .ralph/PROMPT.md | claude -p --dangerously-skip-permissions
    
    echo ""
    echo "✅ Iteration complete at $(date)"
    echo "⏳ Waiting 5 seconds before next iteration..."
    sleep 5
done