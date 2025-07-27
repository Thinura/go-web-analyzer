APP_NAME := webanalyzer
PKG := ./...
BIN_DIR := bin
SRC := $(shell find . -type f -name '*.go')
COVER_FILE := coverage.out

all: build

build:
	go build -o $(BIN_DIR)/$(APP_NAME) ./cmd/server

test:
	go test -v -cover -coverprofile=$(COVER_FILE) $(PKG)

cover:
	go tool cover -html=$(COVER_FILE)

lint:
	golangci-lint run

docker:
	docker build -t $(APP_NAME):latest .

run-docker:
	docker run -p 8080:8080 $(APP_NAME):latest

clean:
	rm -rf $(BIN_DIR) $(COVER_FILE)

.PHONY: all build test cover lint docker run-docker clean