#!/bin/bash
# Reset everything (DB + API)

cd "$(dirname "$0")"

echo "🔄 Resetting everything..."

# Stop API
./api-stop.sh

# Stop and remove PostgreSQL
cd ..
docker-compose down -v

echo ""
echo "✅ Everything cleaned"
echo ""
echo "To start fresh:"
echo "1. ./scripts/db-setup.sh"
echo "2. ./scripts/api-restart.sh"

