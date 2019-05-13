FROM golang:1-alpine
RUN apk add --no-cache make
RUN mkdir -p /go/src/github.com/stone/gotaserver
WORKDIR /go/src/github.com/stone/gotaserver
