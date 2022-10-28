BINARY := cronflux
PKG    := ./...
CMD    := ./cmd/cronflux
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X github.com/Fzgt/cronflux/internal/buildinfo.Version=$(VERSION)

.PHONY: all build run test race cover lint fmt fmtcheck vet tidy clean

all: fmtcheck vet lint test

build:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) $(CMD)

run:
	go run $(CMD)

test:
	go test $(PKG)

race:
	go test -race $(PKG)

cover:
	go test -coverprofile=coverage.out $(PKG)
	go tool cover -func=coverage.out | tail -n 1

lint:
	golangci-lint run

fmt:
	gofmt -w .

fmtcheck:
	@out=$$(gofmt -l .); if [ -n "$$out" ]; then echo "gofmt needed:"; echo "$$out"; exit 1; fi

vet:
	go vet $(PKG)

tidy:
	go mod tidy

clean:
	rm -rf bin dist coverage.out coverage.html
