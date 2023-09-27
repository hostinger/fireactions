GO    := go
GOFMT := gofmt

GIT_TAG    := $(shell git describe --tags --always --match "fireactions-v*" 2> /dev/null | sed "s/fireactions-//")
GIT_COMMIT := $(shell git rev-parse HEAD)
BUILD_DATE := $(shell date '+%FT%T')

LDFLAGS := -ldflags "-s -w -X github.com/hostinger/fireactions/version.Version=$(GIT_TAG) -X github.com/hostinger/fireactions/version.Commit=$(GIT_COMMIT) -X github.com/hostinger/fireactions/version.Date=$(BUILD_DATE)"

.PHONY: build
build:
	@ $(GO) build $(LDFLAGS) -o fireactions ./cmd/fireactions

.PHONY: clean
clean:
	@ rm -rf dist

.PHONY: fmt
fmt:
	@ $(GOFMT) -w -s .

.PHONY: test
test:
	@ $(GO) test -v ./...
