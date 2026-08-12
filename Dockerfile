# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build all files in the cmd package — not just main.go.
RUN CGO_ENABLED=0 GOOS=linux go build -a -trimpath -ldflags="-s -w" -o server ./cmd/

# Final stage — minimal image
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=builder /app/server .

# Run as non-root
USER app

EXPOSE 8080

CMD ["./server"]
