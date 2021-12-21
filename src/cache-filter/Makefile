goimports := golang.org/x/tools/cmd/goimports@v0.1.5
golangci_lint := github.com/golangci/golangci-lint/cmd/golangci-lint@v1.42.0

.PHONY: build-wasm
build:
	@find . -type f -name "main.go" | xargs -Ip tinygo build -o p.wasm -scheduler=none -target=wasi p