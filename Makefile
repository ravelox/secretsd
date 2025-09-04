SHELL:=/bin/bash

.PHONY: all proto server csi esowebhook ctl
all: proto server csi ctl

proto:
	protoc -I api --go_out=api/gen --go-grpc_out=api/gen api/secrets.proto


server:
	@go build -o bin/secretsd ./cmd/secretsd

csi:
	@go build -o bin/csi-provider ./cmd/csi-provider

ctl:
	@go build -o bin/secretsctl ./cmd/secretsctl
