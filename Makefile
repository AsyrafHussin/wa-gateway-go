BINARY_NAME=wa-gateway-go
BUILD_DIR=.

.PHONY: build build-linux run dev clean tidy fmt lint docker docker-down docker-logs setup

build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) .

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME) .

run: build
	./$(BINARY_NAME)

dev:
	go run .

clean:
	rm -f $(BUILD_DIR)/$(BINARY_NAME)
	go clean

tidy:
	go mod tidy

fmt:
	goimports -w .

lint:
	golangci-lint run ./...

setup:
	@sh setup.sh

docker:
	docker compose up -d --build

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f
