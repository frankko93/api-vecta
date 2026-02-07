#!/bin/bash
# Start API

cd "$(dirname "$0")/.."

# Kill any existing API
killall -9 api 2>/dev/null
lsof -ti:3080 | xargs kill -9 2>/dev/null

# Build
echo "🔨 Building API..."
go build -o bin/api ./cmd/go8/

# Start
echo "🚀 Starting API on http://localhost:3080..."
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

./bin/api

