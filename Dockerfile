# Dockerfile
# Stage 1: Build the Go app
FROM golang:1.23-alpine AS builder

# Set up environment
ENV CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64

# Install required build tools
RUN apk --no-cache add gcc g++ make git

# Set the working directory inside the container
WORKDIR /app

# Copy the Go modules and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
RUN go build -o the-keeper ./cmd/bot/main.go

# Stage 2: Run the Go app in a lightweight container
FROM alpine:3.18

# Install CA certificates for SSL, SQLite, gettext for envsubst
RUN apk --no-cache add ca-certificates tzdata sqlite-libs gettext

# Set working directory in the second stage
WORKDIR /app

# Create a directory for the database files
RUN mkdir -p /app/data2

# Set environment variable for Railway volume mount path
ENV RAILWAY_VOLUME_MOUNT_PATH="/app/data2"

# Copy the compiled binary from the builder stage
COPY --from=builder /app/the-keeper .

# Copy the entire configs directory
COPY configs ./configs

# Replace environment variables in config.template.yaml at runtime
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Expose the port the app will run on
EXPOSE 8080

# Define the entry point for the container to run the bot
ENTRYPOINT ["/entrypoint.sh"]
