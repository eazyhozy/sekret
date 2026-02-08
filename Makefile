APP_NAME := sekret
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X github.com/eazyhozy/sekret/cmd.version=$(VERSION)

build:
	go build -ldflags "$(LDFLAGS)" -o $(APP_NAME) .

test:
	go test ./...

lint:
	golangci-lint run

clean:
	rm -f $(APP_NAME)

.PHONY: build test lint clean
