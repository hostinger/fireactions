GO     := go
GOFMT  := gofmt
GOOS   := linux
GOARCH := amd64

GIT_TAG    := $(shell git describe --tags --always --match "fireactions-v*" 2> /dev/null | sed "s/fireactions-//")
GIT_COMMIT := $(shell git rev-parse HEAD)
BUILD_DATE := $(shell date -u '+%FT%TZ')

LDFLAGS := -ldflags "-s -w -X github.com/hostinger/fireactions.Version=$(GIT_TAG) -X github.com/hostinger/fireactions.Commit=$(GIT_COMMIT) -X github.com/hostinger/fireactions.Date=$(BUILD_DATE)"

.PHONY: build
build:
	@ GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build $(LDFLAGS) -o fireactions ./cmd/fireactions

.PHONY: clean
clean:
	@ rm -rf dist

.PHONY: fmt
fmt:
	@ $(GOFMT) -w -s .

.PHONY: test
test:
	@ $(GO) test -v ./...
