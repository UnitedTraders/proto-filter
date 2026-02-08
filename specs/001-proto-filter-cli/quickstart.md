# Quickstart: Proto Filter CLI

## Prerequisites

- Go 1.23 or later

## Build

```bash
go build -o proto-filter .
```

## Basic Usage: Pass-Through (No Filtering)

Parse and regenerate all `.proto` files:

```bash
./proto-filter --input ./protos --output ./out
```

This discovers all `*.proto` files in `./protos` (recursively),
parses each into a structural representation, and writes regenerated
`.proto` files to `./out` preserving the directory structure.

## Filtered Usage

Create a filter configuration file `filter.yaml`:

```yaml
include:
  - "myapp.orders.*"
  - "myapp.common.*"
exclude:
  - "myapp.orders.InternalService"
```

Run with the config:

```bash
./proto-filter --input ./protos --output ./out --config filter.yaml
```

Only definitions matching the include patterns (minus excludes) and
their transitive dependencies will appear in the output.

## Verbose Mode

See what the tool processed:

```bash
./proto-filter --input ./protos --output ./out --config filter.yaml --verbose
```

Output (stderr):

```
proto-filter: processed 12 files, 45 definitions
proto-filter: included 8 definitions, excluded 37
proto-filter: wrote 5 files to ./out
```

## Verify Output

Check that the output compiles:

```bash
protoc --proto_path=./out --descriptor_set_out=/dev/null ./out/**/*.proto
```

## Run Tests

```bash
go test ./...
```

With race detection:

```bash
go test -race ./...
```
