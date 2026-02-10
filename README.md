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

### Annotation filtering

Filter services and methods based on annotations in their comments. Annotations use `@Name` or `[Name]` syntax.

```yaml
# Exclude mode: remove elements with these annotations
exclude:
  - "Internal"
  - "HasAnyRole"

# Include mode: keep only elements with these annotations
include:
  - "Public"
```

`exclude` and `include` are mutually exclusive.

### Annotation substitution

Replace annotation markers in comments with human-readable descriptions. This is useful for producing documentation-friendly proto files where implementation annotations are replaced with descriptive text.

```yaml
substitutions:
  HasAnyRole: "Requires authentication"
  Internal: "For internal use only"
  Public: "Available to all users"
```

Given a proto file:

```protobuf
service OrderService {
  // @HasAnyRole({"ADMIN", "MANAGER"})
  // Creates a new order in the system.
  rpc CreateOrder(CreateOrderRequest) returns (CreateOrderResponse);

  // @Internal
  rpc DeleteOrder(DeleteOrderRequest) returns (DeleteOrderResponse);

  // [Public] Lists all orders.
  rpc ListOrders(ListOrdersRequest) returns (ListOrdersResponse);
}
```

The output will be:

```protobuf
service OrderService {
  // Requires authentication
  // Creates a new order in the system.
  rpc CreateOrder(CreateOrderRequest) returns (CreateOrderResponse);

  // For internal use only
  rpc DeleteOrder(DeleteOrderRequest) returns (DeleteOrderResponse);

  // Available to all users Lists all orders.
  rpc ListOrders(ListOrdersRequest) returns (ListOrdersResponse);
}
```

Both `@Name(...)` and `[Name(...)]` syntax are supported. Surrounding text on the same line is preserved.

**Empty substitutions** remove annotations entirely. Use an empty string to strip annotation markers without replacing them:

```yaml
substitutions:
  HasAnyRole: ""
  Internal: ""
```

Annotation-only comment lines are removed. If all lines in a comment are removed, the comment is dropped from the element.

**Strict mode** enforces that every annotation in the input has a substitution mapping. Enable it to catch annotations you forgot to map:

```yaml
substitutions:
  HasAnyRole: "Requires authentication"
strict_substitutions: true
```

If any annotation in the processed files lacks a mapping, the tool exits with code 2 and lists the missing annotations on stderr. No output files are written.

**Combined with annotation filtering**: substitution and annotation include/exclude work together. Filtering removes elements first, then substitution replaces annotations on surviving elements:

```yaml
annotations:
  exclude:
    - "Internal"
substitutions:
  HasAnyRole: "Requires authentication"
```

Methods annotated with `@Internal` are removed; `@HasAnyRole` on remaining methods is replaced with the description text.

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
| 2 | Configuration error (invalid YAML, conflicting filter rules, unsubstituted annotations in strict mode) |

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
