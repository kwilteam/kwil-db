.phony all: build

build:
	@echo "Building wasm..."
	@GOOS=js GOARCH=wasm go build -o kl.wasm main.go