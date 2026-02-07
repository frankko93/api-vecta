#!/bin/bash
# Stop API

echo "🛑 Stopping API..."

killall -9 api 2>/dev/null
pkill -9 -f "go8/main.go" 2>/dev/null
lsof -ti:3080 | xargs kill -9 2>/dev/null

echo "✅ API stopped"

