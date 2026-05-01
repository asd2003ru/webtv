APP_NAME := webtv
CMD_PATH := ./cmd/webtv
BIN_DIR := bin
BIN := $(BIN_DIR)/$(APP_NAME)
WEB_DIR := web
WEB_DIST_DIR := internal/webui/dist

.PHONY: all install deps web-install web-build build run run-debug test clean

all: build

install: deps web-install

deps:
	go mod tidy

web-install:
	npm --prefix $(WEB_DIR) install

web-build:
	npm --prefix $(WEB_DIR) run build

build: web-build
	mkdir -p $(BIN_DIR)
	go build -o $(BIN) $(CMD_PATH)

run: web-build
	go run $(CMD_PATH)

run-debug: web-build
	WEBTV_LOG_LEVEL=debug go run $(CMD_PATH)

test:
	go test ./...

clean:
	rm -rf $(BIN_DIR)
	rm -rf $(WEB_DIST_DIR)
	mkdir -p $(WEB_DIST_DIR)
	touch $(WEB_DIST_DIR)/.keep
