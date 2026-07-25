.PHONY: build test test-e2e coverage run clean docker-build docker-up docker-down

build:
	go build -o bin/proxy ./cmd/proxy/
	go build -o bin/node ./cmd/node/

test:
	go test ./... -v -race -count=1 -cover

test-e2e:
	go test ./tests/e2e/... -v -count=1 -timeout=300s

coverage:
	go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out -o coverage.html

run-proxy:
	go run ./cmd/proxy/

run-node:
	go run ./cmd/node/ -addr=:8081 -nodes=node-1

clean:
	rm -rf bin/ coverage.out coverage.html

docker-build:
	docker compose build

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down -v

docker-logs:
	docker compose logs -f
