FROM golang:1.24-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

RUN go build -o order-service ./cmd/order-service/main.go

FROM debian:bullseye-slim
WORKDIR /app

RUN apt-get update && apt-get install -y postgresql-client make && rm -rf /var/lib/apt/lists/*
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY --from=builder /app/order-service .
COPY config ./config
COPY entrypoint.sh /app/entrypoint.sh
COPY .env /app/.env
COPY Makefile /app/Makefile
COPY migrations /app/migrations
RUN chmod +x /app/entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["./order-service"]