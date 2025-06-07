# Dockerfile
FROM golang:1.24-alpine AS builder

# Install GCC and related build tools
RUN apk add --no-cache build-base

WORKDIR /app

ENV GOPRIVATE=github.com/go-telegram-bot-api/telegram-bot-api

# Copy source code
COPY . .

# Install dependencies for libraries in /lib
RUN for lib in ./lib/*; do \
        if [ -d "$lib" ]; then \
            cd "$lib" && go mod tidy && go mod download && cd -; \
        fi \
    done

# Download dependencies
RUN go mod tidy && go mod download && go mod vendor

# Build the application
RUN ./build.sh

# Final stage
FROM alpine:latest

# Install required dependencies for SQLite
RUN apk add --no-cache libc6-compat sqlite-libs

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/main .

# Set the binary as the entrypoint
ENTRYPOINT ["/app/main"]
