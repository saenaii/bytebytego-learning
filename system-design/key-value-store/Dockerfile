FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/proxy ./cmd/proxy/ && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/node ./cmd/node/

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata curl
WORKDIR /app
COPY --from=builder /app/proxy /app/proxy
COPY --from=builder /app/node /app/node
RUN adduser -D -h /app kvstore && chown -R kvstore:kvstore /app
USER kvstore
EXPOSE 8080 8081 8082 8083 8084 8085
