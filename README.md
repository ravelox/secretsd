# secretsd (standalone secrets manager - local Postgres wiring)

## Quick start (local)
```bash
docker compose up -d db
export KEK_FILE=./.dev/kek.bin
export PG_DSN=postgres://secretsd:secretsd@127.0.0.1:5432/secretsd?sslmode=disable
go run ./cmd/secretsd
```

Put/Get:
```bash
curl -sS -X POST localhost:8080/v1/put -H 'content-type: application/json' -d '{"path":"db/password","value":"Tr1cky"}'
curl -sS 'http://localhost:8080/v1/get?key=db/password'
```

CSI smoke test:
```bash
go run ./cmd/csi-smoketest --key db/password --out ./out/DB_PASSWORD
```

## Helm chart (install via local repo checkout)
```bash
helm upgrade --install secretsd ./charts/secretsd -n secretsd --create-namespace   --set image.repository=ghcr.io/OWNER/secretsd   --set image.tag=latest
```

## CI/CD
- `.github/workflows/ci.yaml` builds & pushes images to GHCR and lints chart.
- `.github/workflows/release-chart.yaml` packages the chart on tags, pushes to GHCR (OCI), and signs (Helm provenance + Cosign).

See `charts/secretsd/NOTES.txt` for smoke test commands after install.


## Development setup

The `Makefile` includes a `proto` target with a message:

```
(stub) add protoc in your env to generate gRPC stubs
```

If you plan to implement gRPC APIs (e.g. CSI provider, KMS plugin), you need the Protocol Buffers compiler and Go plugins:

```bash
# Install protoc (ensure version >= 3.20)
# On macOS (Homebrew):
brew install protobuf

# On Ubuntu/Debian:
sudo apt-get install -y protobuf-compiler

# Verify installation
protoc --version

# Install Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Make sure $GOPATH/bin is in your PATH
export PATH="$PATH:$(go env GOPATH)/bin"
```

Then you can run:

```bash
make proto
```

This will invoke `protoc` once `.proto` files are added under `api/` and generate Go stubs.


> **Note:** This repo uses a placeholder module path `github.com/yourname/secretsd`. Update `go.mod` to your actual repo path (e.g. `github.com/<you>/secretsd`) and run `go mod tidy`.
