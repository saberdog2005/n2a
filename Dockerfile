# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o note2anki main.go

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1000 note2anki && \
    adduser -D -u 1000 -G note2anki note2anki

# Set working directory
WORKDIR /home/note2anki

# Copy binary from builder
COPY --from=builder /app/note2anki /usr/local/bin/note2anki

# Make binary executable
RUN chmod +x /usr/local/bin/note2anki

# Switch to non-root user
USER note2anki

# Set environment variable for API key (to be overridden)
ENV ANTHROPIC_API_KEY=""

# Entry point
ENTRYPOINT ["note2anki"]

# Default command (show help)
CMD ["-help"]