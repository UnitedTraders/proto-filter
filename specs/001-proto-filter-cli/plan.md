# Implementation Plan: Proto Filter CLI

**Branch**: `001-proto-filter-cli` | **Date**: 2026-02-08 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/001-proto-filter-cli/spec.md`

## Summary

Build a Go CLI tool that takes an input directory of `.proto` files,
parses them into a structural AST using `emicklei/proto`, optionally
applies semantic filtering based on YAML configuration with glob
pattern matching, resolves transitive dependencies, and generates
filtered `.proto` files in an output directory. Comments on all
definitions are preserved through the parse-generate pipeline.

## Technical Context

**Language/Version**: Go 1.23+
**Primary Dependencies**:
- `github.com/emicklei/proto` — proto file parsing to AST
- `github.com/emicklei/proto-contrib/pkg/protofmt` — AST to proto source
- `gopkg.in/yaml.v3` — YAML config parsing
**Storage**: Filesystem only (input/output directories)
**Testing**: `go test ./...` with `-race` in CI
**Target Platform**: Cross-platform (Linux, macOS, Windows);
  `CGO_ENABLED=0` for static binaries
**Project Type**: Single CLI tool
**Performance Goals**: Process 100 proto files in under 5 seconds
**Constraints**: No CGo, no external tool invocations (no protoc)
**Scale/Scope**: Typical proto repos with 10–500 files

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Evidence |
|-----------|--------|----------|
| I. CLI-First Tool | PASS | Flags for input/output/config/verbose; stdout reserved; errors to stderr; exit codes 0/1/2 |
| II. Proto-Aware Filtering | PASS | Parses proto structure (packages, services, messages, enums); declarative YAML config; output is valid protoc input |
| III. Correctness Over Convenience | PASS | Transitive dependency resolution; conflicting rules produce errors; no silent drops |
| IV. Test-Driven Development | PASS | Unit tests per package; integration tests with proto fixtures in testdata/ |
| V. Simplicity and Minimal Dependencies | PASS | 3 dependencies total (proto parser, proto formatter, YAML); standard library for flags, glob, filesystem; flat structure with internal packages |

No violations. Complexity tracking not needed.

## Project Structure

### Documentation (this feature)

```text
specs/001-proto-filter-cli/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── cli-interface.md
└── tasks.md
```

### Source Code (repository root)

```text
.
├── main.go
├── go.mod
├── go.sum
├── internal/
│   ├── parser/
│   │   ├── parser.go
│   │   └── parser_test.go
│   ├── filter/
│   │   ├── filter.go
│   │   └── filter_test.go
│   ├── deps/
│   │   ├── deps.go
│   │   └── deps_test.go
│   ├── writer/
│   │   ├── writer.go
│   │   └── writer_test.go
│   └── config/
│       ├── config.go
│       └── config_test.go
└── testdata/
    ├── simple/
    ├── nested/
    ├── imports/
    ├── comments/
    └── filter/
```

**Structure Decision**: Single Go project. One `main.go` at root for
CLI entry point. Five internal packages under `internal/` organized by
concern: parser, filter, deps, writer, config. Test fixtures in
`testdata/` at root. This follows constitution principle V (simplicity)
while keeping each responsibility isolated and testable.
