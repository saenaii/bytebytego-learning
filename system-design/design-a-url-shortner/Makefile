.PHONY: build test vet run docker-build up down clean

build:
	go build -o bin/url-shortener ./cmd/proxy

test:
	go test -race ./...

vet:
	go vet ./...

run:
	go run ./cmd/proxy -config=config.json

docker-build:
	docker compose build

up:
	docker compose up --build

down:
	docker compose down

clean:
	rm -rf bin
