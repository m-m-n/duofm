.PHONY: build test clean install run fmt vet lint deps test-e2e test-e2e-build dpkg clean-dpkg

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

clean: clean-dpkg
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

# E2E Tests with Docker (isolated environment with permission tests)
E2E_IMAGE=duofm-e2e-test

test-e2e-build:
	docker build -t $(E2E_IMAGE) -f test/e2e/Dockerfile .

test-e2e: test-e2e-build
	docker run --rm $(E2E_IMAGE)

# Debian package creation
dpkg:
	@if [ ! -f scripts/build-dpkg.sh ]; then \
		echo "Error: scripts/build-dpkg.sh not found"; \
		echo "Please run /make-dpkg command first"; \
		exit 1; \
	fi
	@bash scripts/build-dpkg.sh

clean-dpkg:
	rm -f *.deb
	rm -rf build/dpkg

.DEFAULT_GOAL := build
