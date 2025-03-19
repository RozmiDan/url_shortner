FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o url_shortener ./cmd/app/main.go

RUN ls -lah /app

FROM debian:latest

WORKDIR /app

COPY --from=builder /app/url_shortener /app/url_shortener

COPY config/config.yaml /app/config.yaml

ENV CONFIG_PATH="/app/config.yaml"

EXPOSE 8080

CMD ["/app/url_shortener"]
