.PHONY: build run test extract docker

build:
	go build -o bin/server ./cmd/server
	go build -o bin/extract ./cmd/extract

run:
	set -a; [ -f .env ] && . ./.env; set +a; go run ./cmd/server

test:
	go test ./...

extract:
	set -a; [ -f .env ] && . ./.env; set +a; go run ./cmd/extract $(FILES)

docker:
	docker build -t ai-data-extractor .
