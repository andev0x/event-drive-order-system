# --- STAGE 1: Build Stage ---
# Use golang:alpine for a small build footprint
FROM golang:1.21-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Install git, required for fetching Go dependencies
RUN apk add --no-cache git

# Copy dependency files first to leverage Docker layer caching
COPY go.mod go.sum* ./
RUN go mod download

# Copy the entire service source code
COPY . .

# Build the Go application
# Pointing to the standardized cmd entrypoint
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/order-api/main.go

# --- STAGE 2: Runtime Stage ---
# Use a minimal alpine image for the final executable
FROM alpine:latest

# Install runtime dependencies
# ca-certificates: for secure communication
# mysql-client: required by the migration script to talk to the DB
RUN apk --no-cache add ca-certificates mysql-client

# Set working directory to /root/ where the binary will live
WORKDIR /root/

# Copy the compiled binary from the builder stage
COPY --from=builder /app/main .

# Copy the database migrations folder
COPY --from=builder /app/migrations ./migrations

# Copy the migration shell script (make sure this file exists in your host folder)
COPY migrate.sh ./migrate.sh

# Ensure the migration script has execution permissions
RUN chmod +x ./migrate.sh

# The port the service listens on
EXPOSE 8080

# Run the migration script first, then start the main application
# This ensures the schema is ready before the app starts accepting requests
# CMD ["sh", "-c", "./migrate.sh && ./main"]MD sh -c "/root/migrate.sh && /root/main"
CMD ["./migrate.sh", "&&", "./main"]
