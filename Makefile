.PHONY: build test lint vet tidy clean run

BINARY   := bin/server
PKG      := ./...
MAIN     := ./cmd/server

build:
	go build -o $(BINARY) $(MAIN)

test:
	go test -race -count=1 -coverprofile=coverage.out $(PKG)
	go tool cover -func=coverage.out

lint:
	golangci-lint run $(PKG)

vet:
	go vet $(PKG)

tidy:
	go mod tidy

clean:
	rm -rf bin/ coverage.out

run: build
	./$(BINARY)

# One-shot CI target: tidy check + vet + test + lint
ci: tidy vet test lint
