GO    := go
GOFMT := gofmt

GIT_TAG    := $(shell git describe --tags --always 2> /dev/null)
GIT_COMMIT := $(shell git rev-parse HEAD)
BUILD_DATE := $(shell date '+%FT%T')

MODULE := $(shell $(GO) list -m)

LDFLAGS := -ldflags "-s -w -X $(MODULE)/version.Version=$(GIT_TAG) -X $(MODULE)/version.Commit=$(GIT_COMMIT) -X $(MODULE)/version.Date=$(BUILD_DATE)"

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
