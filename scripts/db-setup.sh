#!/bin/bash
# Setup database (run once or to reset)

echo "🗄️  Setting up database..."

cd "$(dirname "$0")/.."

# Start PostgreSQL
docker-compose up -d

# Wait for PostgreSQL to be ready
echo "⏳ Waiting for PostgreSQL..."
sleep 5

# Run schemas
echo "📊 Running schemas..."
docker exec -i vecta_postgres_local psql -U postgres -d vecta_db < database/schema.sql > /dev/null 2>&1
docker exec -i vecta_postgres_local psql -U postgres -d vecta_db < database/config_schema.sql > /dev/null 2>&1
docker exec -i vecta_postgres_local psql -U postgres -d vecta_db < database/data_schema.sql > /dev/null 2>&1

echo "✅ Database ready!"
echo ""
echo "Test connection:"
docker exec vecta_postgres_local psql -U postgres -d vecta_db -c "SELECT 'Connected!' as status"

