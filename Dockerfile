FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o amp-free-proxy

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/amp-free-proxy .
COPY config.example.yaml config.yaml

EXPOSE 8318
CMD ["./amp-free-proxy"]
