SHELL:=/bin/bash

.PHONY: all deps proto server csi ctl
all: deps proto server csi ctl

deps:
	@echo "==> Ensuring Go deps (go.mod/go.sum)"
	@go mod tidy
	@go mod download

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
