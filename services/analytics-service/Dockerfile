# --- STAGE 1: Build Stage ---
FROM golang:1.21-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Install git for fetching Go dependencies
RUN apk add --no-cache git

# Copy dependency files first for better caching
COPY go.mod go.sum* ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
# Standardizing the path to ./cmd/analytics-api/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/analytics-api/main.go

# --- STAGE 2: Runtime Stage ---
FROM alpine:latest

# Install runtime dependencies
# mysql-client is required for the migrate.sh script
RUN apk --no-cache add ca-certificates mysql-client

# Set working directory
WORKDIR /root/

# Copy the compiled binary from builder stage
COPY --from=builder /app/main .

# Copy database migrations
COPY --from=builder /app/migrations ./migrations

# FIXED: Instead of Heredoc, we copy a physical script file
# Make sure you have created 'migrate.sh' in the same folder as this Dockerfile
COPY migrate.sh ./migrate.sh

# Ensure the script is executable
RUN chmod +x ./migrate.sh

# Expose Analytics Service port
EXPOSE 8081

# Run migration script then start the application
# Using ./main to ensure it points to the file in /root/
# CMD ["sh", "-c", "./migrate.sh && ./main"]
CMD ["./migrate.sh", "&&", "./main"]
