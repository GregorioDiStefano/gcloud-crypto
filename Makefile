GO ?= go
VERSION ?= 0.0.1
COMPILER=$(shell go version)
BUILD_OUTPUT_DIR=build
BUILD_OUTPUT_FILENAME=gcloud-crypto

GO_LDFLAGS = \
	-ldflags "-X main.BuildTime=`date -u +%Y.%m.%d-%H:%M:%S` -X main.Version=$(VERSION)-`git rev-parse --short HEAD` -X 'main.CompilerVersion=$(COMPILER)'"
build:
	$(GO) build -o $(BUILD_OUTPUT_DIR)/$(BUILD_OUTPUT_FILENAME) $(GO_LDFLAGS)

install:
	$(GO) install $(GO_LDFLAGS)

test:
	$(GO) test -v -cover . 

cover: coverage
	$(GO) tool cover -func=coverage.out

htmlcover: coverage
	$(GO) tool cover -html=coverage.out

clean:
	-rm -rf $(BUILD_OUTPUT_DIR)

$(BUILD_OUTPUT_DIR)/$(BUILD_OUTPUT_FILENAME):   .FORCE

.PHONY: build
clean:
    $(build)

