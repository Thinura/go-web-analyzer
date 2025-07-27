# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS builder

# Enable Go modules and install git
ENV CGO_ENABLED=0 \
    GO111MODULE=on

WORKDIR /app

# Install git (needed for some Go dependencies)
RUN apk add --no-cache git

# Copy go.mod and download dependencies (cached unless changed)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source
COPY . .

# Run unit tests
RUN go test -v ./...

# Build the app binary
RUN go build -o webanalyzer ./cmd/server

# Final minimal image
FROM alpine:latest
WORKDIR /app

# Copy built binary and static assets
COPY --from=builder /app/webanalyzer .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static
COPY --from=builder /app/config ./config

EXPOSE 8080

CMD ["./webanalyzer"]