# Stage 1: Build
FROM golang:1.24-alpine AS builder

# Enable Go modules and install Git
ENV GO111MODULE=on

WORKDIR /app

# Install git (needed for private/go-gettable modules)
RUN apk add --no-cache git

# Copy go mod files first to cache dependencies
COPY go.mod go.sum ./
RUN go mod tidy && go mod download

# Copy the rest of the code
COPY . .

# Run tests
RUN go test ./...

# Build binary
RUN go build -o web-analyzer ./cmd/webanalyzer

# Stage 2: Production image
FROM alpine:latest

WORKDIR /app

# Copy the compiled binary from builder
COPY --from=builder /app/web-analyzer .

EXPOSE 8080

CMD ["./web-analyzer"]