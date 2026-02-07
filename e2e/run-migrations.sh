#!/bin/bash
set -e

echo "Waiting for PostgreSQL..."
until PGPASSWORD=$DB_PASS psql -h postgres -U $DB_USER -d $DB_NAME -c '\q'; do
  sleep 1
done

echo "Running migrations..."
PGPASSWORD=$DB_PASS psql -h postgres -U $DB_USER -d $DB_NAME -f /app/database/schema.sql
PGPASSWORD=$DB_PASS psql -h postgres -U $DB_USER -d $DB_NAME -f /app/database/config_schema.sql
PGPASSWORD=$DB_PASS psql -h postgres -U $DB_USER -d $DB_NAME -f /app/database/data_schema.sql

echo "✓ Migrations completed"

