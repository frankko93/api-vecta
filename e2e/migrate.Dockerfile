FROM postgres:17

WORKDIR /app

# Copy SQL schemas
COPY database/schema.sql ./database/
COPY database/config_schema.sql ./database/
COPY database/data_schema.sql ./database/

# Script to run migrations
COPY e2e/run-migrations.sh ./
RUN chmod +x ./run-migrations.sh

CMD ["./run-migrations.sh"]
