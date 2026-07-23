.PHONY: build test test-e2e coverage run docker-build docker-run clean

build:
	go build -o bin/server ./cmd/server

test:
	go test ./... -v -race -count=1

test-e2e:
	go test ./tests/e2e/... -v -count=1

coverage:
	go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out -o coverage.html

run:
	go run ./cmd/server/ -config=config.yaml

docker-build:
	docker build -t rate-limiter:latest .

docker-run:
	docker run -p 8080:8080 -v $(PWD)/config.yaml:/app/config.yaml rate-limiter:latest

clean:
	rm -rf bin/ coverage.out coverage.html