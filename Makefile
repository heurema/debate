BIN := debate
PKG := ./...

.PHONY: build vet test check

build:
	go build -o $(BIN) ./cmd/debate

vet:
	go vet $(PKG)

test:
	go test $(PKG)

check: vet test build
