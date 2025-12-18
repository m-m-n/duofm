.PHONY: build test clean install run fmt vet lint deps

BINARY_NAME=duofm
BINARY_PATH=./cmd/duofm
BUILD_DIR=.
GO=go

build:
	$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) $(BINARY_PATH)

test:
	$(GO) test -v ./...

test-coverage:
	$(GO) test -v -cover ./...
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

clean:
	rm -f $(BUILD_DIR)/$(BINARY_NAME)
	rm -f coverage.out coverage.html

install: build
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/

run: build
	./$(BINARY_NAME)

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

lint:
	golangci-lint run

deps:
	$(GO) mod download
	$(GO) mod tidy

.DEFAULT_GOAL := build
