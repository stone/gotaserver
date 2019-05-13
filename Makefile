all: deps test build

deps:
	go get -d -v ./...

test: deps
	go test -timeout=3s -v ./...

build: deps
	go build

.PHONY: all deps test
