SHELL:=/bin/bash

.PHONY: all proto server csi esowebhook ctl
all: proto server csi ctl

proto:
	@echo "(stub) add protoc in your env to generate gRPC stubs"
	@echo "proto generation skipped in this minimal bundle"

server:
	@go build -o bin/secretsd ./cmd/secretsd

csi:
	@go build -o bin/csi-provider ./cmd/csi-provider

ctl:
	@go build -o bin/secretsctl ./cmd/secretsctl
