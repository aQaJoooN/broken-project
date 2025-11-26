FROM golang:1.21-alpine3.20 AS builder

WORKDIR /app

COPY go.mod ./
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o api-server .

FROM alpine:3.22.2

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/api-server .

EXPOSE 8080

CMD ["./api-server"]
