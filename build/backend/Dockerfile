FROM golang:bookworm AS builder


WORKDIR /app

COPY ./go.mod .
RUN go mod tidy
RUN go install github.com/pressly/goose/v3/cmd/goose@latest
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o ./bin/backend ./cmd/backend

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /go/bin/goose          /usr/local/bin/goose
COPY --from=builder /app/migrations        /app/migrations
COPY --from=builder /app/bin/backend       /app/backend
COPY .env .

ENTRYPOINT ["/entrypoint.sh"]
