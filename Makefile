DEPS = $(go list -f '{{range .TestImports}}{{.}} {{end}}' ./...)

all: deps test

deps:
	go get -d -v ./...

test: deps
	go test -timeout=3s -v ./...

build: deps
	go build

.PHONY: all deps test
