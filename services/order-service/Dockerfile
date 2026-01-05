# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install necessary build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates mysql-client

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/main .

# Copy migrations
COPY --from=builder /app/migrations ./migrations

# Expose port
EXPOSE 8080

# Add migration script
COPY <<'EOF' /root/migrate.sh
#!/bin/sh
echo "Waiting for MySQL to be ready..."
until mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" -e "SELECT 1" > /dev/null 2>&1; do
    sleep 2
done

echo "Running migrations..."
for file in /root/migrations/*.sql; do
    echo "Executing $file..."
    mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" < "$file"
done
echo "Migrations completed"
EOF

RUN chmod +x /root/migrate.sh

# Run migrations and start app
CMD sh -c "/root/migrate.sh && /root/main"
