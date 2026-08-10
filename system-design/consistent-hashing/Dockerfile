FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/proxy ./cmd/proxy/ && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/node ./cmd/node/ && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/traffic ./cmd/traffic/

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata curl
WORKDIR /app
COPY --from=builder /app/proxy /app/proxy
COPY --from=builder /app/node /app/node
COPY --from=builder /app/traffic /app/traffic
RUN adduser -D -h /app chdemo && chown -R chdemo:chdemo /app
USER chdemo
EXPOSE 8080 8081 8082 8083 8084 8085
