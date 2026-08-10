.PHONY: build test test-e2e coverage run-proxy run-node run-traffic clean docker-build docker-up docker-down docker-logs

build:
	go build -o bin/proxy ./cmd/proxy/
	go build -o bin/node ./cmd/node/
	go build -o bin/traffic ./cmd/traffic/

test:
	go test ./... -v -race -count=1 -cover

test-e2e:
	go test ./tests/e2e/... -v -count=1 -timeout=300s

coverage:
	go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out -o coverage.html

run-proxy:
	go run ./cmd/proxy/

run-node:
	go run ./cmd/node/ -addr=:8081 -id=node-1 -proxy=http://localhost:8080

run-traffic:
	go run ./cmd/traffic/ -proxy=http://localhost:8080 -rate=100 -keyspace=10000

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
