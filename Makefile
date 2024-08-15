MAIN_PACKAGE_PATH := ./cmd
BINARY_NAME := mem-db

.PHONY: default
default: run

.PHONY: tidy
tidy:
	go fmt ./...
	go mod tidy -v

.PHONY: build
build: tidy
	go build -o=./${BINARY_NAME} ${MAIN_PACKAGE_PATH}

.PHONY: run
run: build
	./${BINARY_NAME} cmd/config/config.json

.PHONY: test
test: build
	go test -v ./...

.PHONY: clean
clean: build
	rm ${BINARY_NAME}
