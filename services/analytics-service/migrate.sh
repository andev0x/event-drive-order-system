#!/bin/sh
set -e

echo "Waiting for MySQL ($DB_HOST:$DB_PORT) to be ready..."
until mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" -e "SELECT 1" > /dev/null 2>&1; do
    sleep 2
done

echo "Running migrations..."
for file in /root/migrations/*.sql; do
    if [ -f "$file" ]; then
        echo "Executing $file..."
        mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" < "$file"
    fi
done
echo "Migrations completed successfully"
