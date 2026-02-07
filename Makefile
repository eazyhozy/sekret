APP_NAME := sekret

build:
	go build -o $(APP_NAME) .

test:
	go test ./...

lint:
	golangci-lint run

clean:
	rm -f $(APP_NAME)

.PHONY: build test lint clean
