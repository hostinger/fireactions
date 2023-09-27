GO    := go
GOFMT := gofmt

.PHONY: build
build:
	@ GOOS=linux GOARCH=amd64 $(GO) build -v -o fireactions .

.PHONY: clean
clean:
	@ rm -f bin

.PHONY: fmt
fmt:
	@ $(GOFMT) -w -s .

.PHONY: test
test:
	@ $(GO) test -v ./...
