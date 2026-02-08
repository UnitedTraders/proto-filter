# proto-filter

A CLI tool for semantically filtering Protocol Buffer `.proto` files. It parses proto files into a structural AST, applies declarative include/exclude rules, resolves transitive dependencies, and generates valid output `.proto` files. Comments are preserved through the pipeline.

## Install

```bash
go install github.com/unitedtraders/proto-filter@latest
```

Or build from source:

```bash
git clone https://github.com/unitedtraders/proto-filter.git
cd proto-filter
go build -o proto-filter .
```

Or pull the Docker image:

```bash
docker pull ghcr.io/unitedtraders/proto-filter:latest
```

## Usage

### Pass-through (no filtering)

Copy all `.proto` files, re-generating them from parsed form:

```bash
proto-filter --input ./protos --output ./out
```

### Filtered output

Create a filter configuration file:

```yaml
# filter.yaml
include:
  - "myapp.orders.OrderService"
  - "myapp.common.*"
exclude:
  - "myapp.orders.InternalService"
```

Run with the config:

```bash
proto-filter --input ./protos --output ./out --config filter.yaml
```

Only definitions matching the include patterns (minus excludes) and their transitive dependencies will appear in the output.

### Docker

The image defaults to `--input /input --output /output --config /filter.yaml`. Mount your directories to these paths:

```bash
docker run --rm \
  -v ./protos:/input \
  -v ./out:/output \
  -v ./filter.yaml:/filter.yaml \
  ghcr.io/unitedtraders/proto-filter:latest
```

Without a config file (pass-through), override the defaults:

```bash
docker run --rm \
  -v ./protos:/input \
  -v ./out:/output \
  ghcr.io/unitedtraders/proto-filter:latest \
  --input /input --output /output
```

### Verbose mode

```bash
proto-filter --input ./protos --output ./out --config filter.yaml --verbose
```

```
proto-filter: processed 12 files, 45 definitions
proto-filter: included 8 definitions, excluded 37
proto-filter: wrote 5 files to ./out
```

## Flags

| Flag | Required | Description |
|------|----------|-------------|
| `--input` | Yes | Source directory containing `.proto` files |
| `--output` | Yes | Destination directory for generated files |
| `--config` | No | Path to YAML filter configuration file |
| `--verbose` | No | Print processing summary to stderr |

## Filter configuration

The config file is YAML with two optional keys:

```yaml
include:
  - "my.package.*"         # glob: all definitions in my.package
  - "other.Specific"       # exact match
exclude:
  - "my.package.Internal"  # remove from included set
```

**Pattern matching** uses glob syntax where `*` matches a single package segment:

- `my.package.*` matches `my.package.Foo` but not `my.package.sub.Bar`
- `*.OrderService` matches `any.package.OrderService`

**Semantics:**

- `include` only: allowlist mode, only matching definitions kept
- `exclude` only: denylist mode, matching definitions removed
- Both: include applied first, then exclude removes from the result
- Neither: pass-through, all definitions kept
- A definition matching both include and exclude is an error

**Transitive dependencies** are resolved automatically. If you include a service, all its request/response message types and their dependencies are included.

## How it works

1. Recursively discovers all `*.proto` files in the input directory
2. Parses each file into a structural AST (packages, services, messages, enums, imports, comments)
3. Builds a dependency graph across all definitions
4. Applies filter rules and resolves transitive dependencies
5. Prunes ASTs to keep only matching definitions
6. Generates output files via formatter, preserving comments and directory structure

External imports (e.g., `google/protobuf/timestamp.proto`) are passed through as-is.

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Runtime error (missing directory, parse failure, I/O error) |
| 2 | Configuration error (invalid YAML, conflicting filter rules) |

## Development

```bash
# Run tests
go test ./...

# Run tests with race detection
go test -race ./...

# Build static binary
CGO_ENABLED=0 go build -o proto-filter .
```

## License

See [LICENSE](LICENSE) for details.
