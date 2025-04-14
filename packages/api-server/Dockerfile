# Build stage
FROM golang:1.23.7-alpine3.21 AS builder

# Build arguments for version information
ARG VERSION=dev
ARG COMMIT_ID=unknown
ARG BUILD_TIME=unknown
ARG MODULE_PREFIX=github.com/babelcloud/gbox
ARG VERSION_PKG=${MODULE_PREFIX}/packages/api-server/internal/misc/service

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with version information
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-X ${VERSION_PKG}.Version=${VERSION} \
    -X ${VERSION_PKG}.BuildTime=${BUILD_TIME} \
    -X ${VERSION_PKG}.CommitID=${COMMIT_ID}" \
    -o api-server ./cmd/app/main.go

# Final stage
FROM alpine:3.21

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