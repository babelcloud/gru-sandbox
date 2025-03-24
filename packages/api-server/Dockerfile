# Build stage
FROM golang:1.21.13-alpine3.20 AS builder

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o api-server

# Final stage
FROM alpine:3.20

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Set timezone
ENV TZ=Asia/Shanghai

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/api-server .

# Copy config files if any
COPY --from=builder /app/config ./config

# Expose port
EXPOSE 28080

# Run the application
CMD ["./api-server"] 