#!/bin/bash
set -e

# Wait until Postgres is ready
until pg_isready -U "$POSTGRES_USER" > /dev/null 2>&1; do
  echo "⏳ Waiting for Postgres to start..."
  sleep 1
done

echo "🗄️ Creating databases..."
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    CREATE DATABASE keycloakdb;
    CREATE DATABASE geodb;
EOSQL

echo "✅ Databases created!"