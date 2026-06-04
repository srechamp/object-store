.PHONY: build test run lint

build:
	go build -o bin/object-store ./main.go

test:
	go test ./... -race -v

run:
	go run main.go --port=8080

lint:
	go vet ./...