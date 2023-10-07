GO    := go
GOFMT := gofmt

GIT_TAG    := $(shell git describe --tags --always 2> /dev/null)
GIT_COMMIT := $(shell git rev-parse HEAD)
BUILD_DATE := $(shell date '+%FT%T')

MODULE := $(shell $(GO) list -m)

LDFLAGS := -ldflags "-X $(MODULE)/build.GitTag=$(GIT_TAG) -X $(MODULE)/build.GitCommit=$(GIT_COMMIT) -X $(MODULE)/build.BuildDate=$(BUILD_DATE)"

.PHONY: build
build:
	@ GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o ./dist/fireactions .

.PHONY: clean
clean:
	@ rm -f dist

.PHONY: fmt
fmt:
	@ $(GOFMT) -w -s .

.PHONY: test
test:
	@ $(GO) test -v ./...
