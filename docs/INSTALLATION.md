# Installation

Obscura Scan is a single static binary. There is nothing to install beyond the
binary itself — no Python, no virtualenv, no database server.

## Requirements

- To **run**: nothing (the binary is fully static, `CGO_ENABLED=0`).
- To **build from source**: Go **1.25+**.
- Optional: Docker, for the container image.

## Option 1 — Download a release binary

Grab the binary for your platform from the
[Releases page](https://github.com/secuity-life-org/Obscura/releases):

```bash
# example (Linux x86-64)
chmod +x obscura-linux-amd64
./obscura-linux-amd64 --version
```

Pre-built targets: `linux/amd64`, `linux/arm64`, `windows/amd64`, `darwin/arm64`.

## Option 2 — Build from source

```bash
git clone https://github.com/secuity-life-org/Obscura.git
cd Obscura

make build                 # -> bin/obscura
# or directly:
CGO_ENABLED=0 go build -o obscura ./cmd/obscura
```

Cross-compile everything:

```bash
make build-all             # bin/obscura-{linux-amd64,linux-arm64,windows-amd64.exe,darwin-arm64}
```

Because the build is pure Go (SQLite via `modernc.org/sqlite`), cross-compilation
needs no C toolchain.

## Option 3 — Docker

A multi-stage build produces a tiny distroless image:

```bash
make docker                                  # builds obscurascan:9.0.0
# or:
docker build -t obscurascan:9.0.0 .

docker run --rm -p 8080:8080 \
  -v "$(pwd)/data:/app/data" \
  obscurascan:9.0.0
```

The container listens on `0.0.0.0:8080` and stores its database in the mounted
`/app/data` volume.

## First run

```bash
./bin/obscura
# INFO configuration loaded ... modules=43
# INFO database ready (WAL) path=obscura.db
# INFO Obscura Scan listening url=http://127.0.0.1:8080
```

Open <http://127.0.0.1:8080>. On first start a fresh `obscura.db` (SQLite, WAL) is
created in the working directory. If a legacy `aegis.db` is present alongside it, it
is imported automatically.

## Configuration (optional)

Every setting is optional and read from the environment or a `.env` file:

```bash
cp .env.example .env       # then fill in only what you need
```

See [Configuration](CONFIGURATION.md) for the full reference. Real OS environment
variables always take precedence over `.env`.

## Verify

```bash
./bin/obscura --version
curl -s http://127.0.0.1:8080/healthz
# {"status":"ok","version":"9.0.0","modules":43,"active_scans":0}
```

## Uninstall

Delete the binary and (if you no longer need scan history) the `obscura.db*` files.
There is nothing else to clean up.
