.PHONY: fmt vet build
default: build

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

build: vet
	go build -o build/chronos_bot main.go