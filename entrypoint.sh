#!/bin/sh

echo "Waiting for PostgreSQL to be ready..."
until pg_isready -h postgres -p 5432 -U postgres; do
  sleep 2
done

echo "Running migrations..."
make migrate-up

echo "Starting application..."
exec "$@"