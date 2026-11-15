FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/url-shortener ./cmd/proxy

FROM alpine:3.19
RUN apk add --no-cache ca-certificates && adduser -D -h /app urlshortener
WORKDIR /app
COPY --from=builder /app/url-shortener .
COPY --from=builder /app/config.json .
RUN chown -R urlshortener:urlshortener /app
USER urlshortener
EXPOSE 8080
ENTRYPOINT ["./url-shortener"]
CMD ["-config", "/app/config.json"]
