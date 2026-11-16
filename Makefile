VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build test vet fmt install clean

build:
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o golarm .

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -l -w .

install:
	CGO_ENABLED=0 go install -trimpath -ldflags "$(LDFLAGS)" .

clean:
	rm -f golarm
