SHELL:=/bin/bash

.PHONY: all deps proto server csi ctl tidy
# Generate stubs first so imports exist, then download deps, then build
all: proto deps server csi ctl

# Only download modules here; run `make tidy` after proto if you want to re-prune
deps:
	@echo "==> Downloading Go deps (without scanning missing imports)"
	@go mod download

tidy:
	@echo "==> go mod tidy"
	@go mod tidy

proto:
	rm -rf api/gen
	mkdir -p api/gen
	protoc -I api --go_out=api/gen --go-grpc_out=api/gen api/secrets.proto

server:
	@go build -o bin/secretsd ./cmd/secretsd

csi:
	@go build -o bin/csi-provider ./cmd/csi-provider

ctl:
	@go build -o bin/secretsctl ./cmd/secretsctl
