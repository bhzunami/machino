BINARY_NAME=machino-api
GO=go

.PHONY: build test run web-build fmt install

build:
	$(GO) build -o bin/$(BINARY_NAME) ./cmd/api

test:
	$(GO) test ./...

run:
	$(GO) run ./cmd/api

web-build:
	cd web && npm run build

fmt:
	$(GO) fmt ./...

install:
	$(GO) mod download
	cd web && npm install
