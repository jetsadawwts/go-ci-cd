# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install git for go mod download
RUN apk add --no-cache git

# Copy dependency files first (layer cache)
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /server ./cmd/api

# Runtime stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary and migrations
COPY --from=builder /server .
COPY --from=builder /app/migrations ./migrations

# Non-root user
RUN adduser -D -u 10001 appuser
USER appuser

EXPOSE 8080

ENTRYPOINT ["./server"]
