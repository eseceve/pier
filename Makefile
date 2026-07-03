BINARY := pier

.PHONY: build test lint fmt tidy run clean install-tools

build:
	go build -o bin/$(BINARY) .

test:
	go test ./... -race -cover

lint:
	golangci-lint run

fmt:
	gofmt -w .

tidy:
	go mod tidy

run:
	go run . $(ARGS)

clean:
	rm -rf bin dist

install-tools:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	go install github.com/goreleaser/goreleaser/v2@latest
	go install github.com/evilmartians/lefthook@latest
