GO ?= go
VERSION ?= 0.0.1
COMPILER=$(shell go version)

GO_LDFLAGS = \
	-ldflags "-X main.BuildTime=`date -u +%Y.%m.%d-%H:%M:%S` -X main.Version=$(VERSION)-`git rev-parse --short HEAD` -X 'main.CompilerVersion=$(COMPILER)'"
build:
	$(GO) build $(GO_LDFLAGS)

install:
	$(GO) install $(GO_LDFLAGS)

test:
	$(GO) test -v -cover . 

cover: coverage
	$(GO) tool cover -func=coverage.out

htmlcover: coverage
	$(GO) tool cover -html=coverage.out
