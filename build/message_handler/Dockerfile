FROM golang:bookworm AS builder


WORKDIR /app

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY . .
RUN go mod tidy
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o ./bin/message_handler ./cmd/message_handler

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /app/.env ./.env
COPY --from=builder /app/bin/message_handler ./message_handler

CMD ["/app/message_handler"]