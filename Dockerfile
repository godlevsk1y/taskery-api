# Stage 1: Builder
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the code
COPY . .

# Build the binary with static linking
RUN CGO_ENABLED=0 go build -o /app/taskery-api -ldflags="-s -w" ./cmd/taskery-api

# Stage 2: Final
FROM alpine:latest

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/taskery-api .
COPY --from=builder /app/config ./config
COPY --from=builder /app/migrations ./migrations

# Expose the required port
EXPOSE 23456

# Run the binary
CMD ["/app/taskery-api"]