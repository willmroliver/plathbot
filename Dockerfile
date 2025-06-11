# Dockerfile
FROM golang:1.24-alpine AS builder

ARG APP_DIR
ARG ENV_FILE

# Install GCC and related build tools
RUN apk add --no-cache build-base

WORKDIR $APP_DIR

ENV GOPRIVATE=github.com/go-telegram-bot-api/telegram-bot-api

# Copy source code
COPY go.mod go.sum ./
COPY ./lib ./lib

# Install dependencies for libraries in /lib
RUN for lib in ./lib/*; do \
        if [ -d "$lib" ]; then \
            cd "$lib" && go mod tidy && go mod download && cd -; \
        fi \
    done

COPY ./src/db ./src/db
COPY ./src/ds ./src/ds 
COPY ./src/model ./src/model
COPY ./src/util ./src/util

# Download dependencies
RUN go mod tidy && go mod download && \
    go build -o /dev/null ./...

COPY ./src ./src
COPY ./$ENV_FILE ./.env
COPY ./build.sh ./build.sh

# Build the application
RUN ./build.sh

# Final stage
FROM alpine:latest

ARG APP_DIR

# Install required dependencies for SQLite
RUN apk add --no-cache libc6-compat sqlite-libs

WORKDIR $APP_DIR

# Copy the binary from builder
COPY --from=builder $APP_DIR/main .

# Set the binary as the entrypoint
ENTRYPOINT ["./main"]
