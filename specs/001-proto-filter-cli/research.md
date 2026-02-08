# Research: Proto Filter CLI

**Date**: 2026-02-08
**Branch**: `001-proto-filter-cli`

## Decision 1: Proto Parsing Library

**Decision**: Use `github.com/emicklei/proto` for parsing and
`github.com/emicklei/proto-contrib/pkg/protofmt` for regeneration.

**Rationale**: Only library that provides both a structural AST from
raw `.proto` source AND a working formatter to write the AST back.
Comments are first-class fields on every element (`Comment`,
`InlineComment`). Supports parsing single files without resolving
imports. Visitor pattern via `proto.Walk` with typed handlers
simplifies filtering. Actively maintained (last commit Feb 2026,
~609 stars, MIT license).

**Alternatives considered**:

| Library | Verdict | Why rejected |
|---------|---------|--------------|
| `bufbuild/protocompile` | Best parser, no printer | Comments live on tokens not elements; no write-back; oriented toward compiled descriptors |
| `jhump/protoreflect` | Has printer, but descriptor-based | Requires all imports to be resolvable; deprecated parser; works on compiled descriptors not structural AST |
| `yoheimuta/go-protoparser` | Good parser, no printer | No write-back capability; less actively maintained (last commit May 2025) |

## Decision 2: YAML Configuration Library

**Decision**: Use `gopkg.in/yaml.v3` for parsing YAML config files.

**Rationale**: Standard Go YAML library. Stable, well-tested, minimal
API. No need for a heavier config framework (Viper, Koanf) since the
tool reads a single config file with a simple schema.

**Alternatives considered**:

| Library | Verdict | Why rejected |
|---------|---------|--------------|
| `github.com/spf13/viper` | Too heavy | Adds env vars, watch, remote config — none needed for a CLI tool with a single config file |
| `github.com/knadh/koanf` | Unnecessary | Same reasoning as Viper |

## Decision 3: CLI Flag Parsing

**Decision**: Use Go standard library `flag` package.

**Rationale**: Constitution principle V (Simplicity and Minimal
Dependencies) mandates standard library where sufficient. The tool
has only 4 flags (`--input`, `--output`, `--config`, `--verbose`).
No need for cobra, urfave/cli, or kong.

**Alternatives considered**:

| Library | Verdict | Why rejected |
|---------|---------|--------------|
| `github.com/spf13/cobra` | Overkill | Single command, 4 flags — no subcommands needed |
| `github.com/urfave/cli` | Overkill | Same reasoning |

## Decision 4: Glob Pattern Matching for Filter Rules

**Decision**: Use `path.Match` from Go standard library for glob
pattern matching on fully qualified proto names.

**Rationale**: `path.Match` supports `*` (match any sequence except
separator) and `?` wildcards. Since proto FQNs use `.` as separator,
this naturally matches at the package boundary level. For example,
`my.package.*` matches `my.package.MyService` but not
`my.package.sub.Other`.

**Alternatives considered**:

| Library | Verdict | Why rejected |
|---------|---------|--------------|
| `github.com/gobwas/glob` | More powerful but unnecessary | Double-star and character classes not needed for FQN matching |
| Full regex | Too permissive | Higher risk of user errors; glob is more intuitive for this domain |

## Decision 5: Project Structure

**Decision**: Flat Go project with `main.go` at root and an
`internal/` package for core logic.

```text
.
├── main.go                  # CLI entry point, flag parsing
├── go.mod
├── go.sum
├── internal/
│   ├── parser/              # Proto file discovery and parsing
│   │   ├── parser.go
│   │   └── parser_test.go
│   ├── filter/              # Filter rule loading and application
│   │   ├── filter.go
│   │   └── filter_test.go
│   ├── deps/                # Dependency graph building and resolution
│   │   ├── deps.go
│   │   └── deps_test.go
│   ├── writer/              # Proto file generation/writing
│   │   ├── writer.go
│   │   └── writer_test.go
│   └── config/              # YAML config parsing
│       ├── config.go
│       └── config_test.go
└── testdata/                # Shared test fixtures
    ├── simple/              # Basic proto files
    ├── nested/              # Nested directory structure
    ├── imports/             # Files with cross-file imports
    ├── comments/            # Files with various comment styles
    └── filter/              # Files for filter testing
```

**Rationale**: Constitution principle V caps at "one main package, one
or two internal packages." Using sub-packages under `internal/` keeps
the code organized by concern while staying within a single module.
All packages are internal — no public API surface.
