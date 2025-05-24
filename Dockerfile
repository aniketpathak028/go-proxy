FROM docker.io/library/golang:1.24-alpine AS builder

WORKDIR /app
COPY . .

RUN go mod download
RUN go build -o proxy-go ./cmd/main.go

FROM docker.io/library/alpine:latest

WORKDIR /app
COPY --from=builder /app/proxy-go .

EXPOSE 8080

CMD ["./proxy-go"]