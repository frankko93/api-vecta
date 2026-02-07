#!/bin/bash
# Restart API (stop, rebuild, start)

cd "$(dirname "$0")"

echo "🔄 Restarting API..."

# Stop
./api-stop.sh

# Build and start in background
cd ..

echo "🔨 Rebuilding..."
go build -o bin/api ./cmd/go8/

echo "🚀 Starting in background..."
export DB_DRIVER=pgx
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASS=password
export DB_NAME=vecta_db
export DB_SSL_MODE=disable
export DB_SSLMODE=disable
export API_HOST=0.0.0.0
export API_PORT=3080
export CORS_ALLOWED_ORIGINS="*"

nohup ./bin/api > api.log 2>&1 &

sleep 3

# Test
echo ""
echo "Testing..."
curl -s http://localhost:3080/version && echo ""
curl -s http://localhost:3080/api/health/readiness && echo ""

echo ""
echo "✅ API restarted on http://localhost:3080"
echo "📝 Logs: tail -f api.log"

