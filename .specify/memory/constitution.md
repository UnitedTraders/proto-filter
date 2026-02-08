<!--
  Sync Impact Report
  ==================
  Version change: N/A (initial) → 1.0.0
  Modified principles: N/A (initial creation)
  Added sections:
    - Core Principles (I–V)
    - Technical Constraints
    - Development Workflow
    - Governance
  Removed sections: None
  Templates requiring updates:
    - .specify/templates/plan-template.md ✅ no changes needed (generic)
    - .specify/templates/spec-template.md ✅ no changes needed (generic)
    - .specify/templates/tasks-template.md ✅ no changes needed (generic)
    - .specify/templates/checklist-template.md ✅ no changes needed (generic)
    - .specify/templates/agent-file-template.md ✅ no changes needed (generic)
  Follow-up TODOs: None
-->

# Proto-Filter Constitution

## Core Principles

### I. CLI-First Tool

Proto-filter is a command-line tool. It MUST accept input via
arguments/flags and produce output to stdout. Errors and diagnostics
MUST go to stderr. The tool MUST support non-interactive, scriptable
usage so it can be embedded in CI pipelines and build systems (e.g.,
Makefiles, Buf workflows, Bazel rules). Exit codes MUST distinguish
success (0) from failure (non-zero).

### II. Proto-Aware Filtering

The tool MUST operate on Protocol Buffer `.proto` files and understand
their structure: packages, services, methods, messages, and imports.
Filtering MUST be declarative—users specify what to keep or exclude
via configuration or flags, and the tool produces a valid, minimal
subset. Output `.proto` files MUST remain syntactically valid and
compilable by `protoc` without modification.

### III. Correctness Over Convenience

Filtered output MUST preserve referential integrity. If a service
method is included, all its request/response message types and their
transitive dependencies MUST be included. Missing imports or dangling
references are considered bugs. The tool MUST NOT silently drop
required definitions. When a conflict between inclusion and exclusion
rules arises, the tool MUST report an error rather than guess.

### IV. Test-Driven Development

Every public function and CLI behavior MUST have corresponding tests.
Unit tests cover parsing and filtering logic. Integration tests cover
end-to-end CLI invocations against real `.proto` fixtures. Tests MUST
be written before or alongside implementation—never deferred. The
`go test ./...` command MUST pass at all times on the main branch.

### V. Simplicity and Minimal Dependencies

The tool MUST use the Go standard library wherever sufficient. Third-
party dependencies MUST be justified by a clear need that the standard
library cannot meet (e.g., a proto parser). Avoid frameworks, ORMs,
and abstractions that add indirection without value. Prefer a flat
package structure until complexity demands otherwise. One `main`
package, one or two internal packages is the target ceiling.

## Technical Constraints

- **Language**: Go (latest stable release, currently 1.23+)
- **Build**: `go build ./...` from repository root
- **Test**: `go test ./...` with `-race` flag in CI
- **Lint**: `golangci-lint run` with default configuration
- **Formatting**: `gofmt` / `goimports` enforced; no unformatted
  code may be committed
- **Module**: Single `go.mod` at repository root
- **Proto parsing**: Use a Go-native proto parser (e.g.,
  `github.com/bufbuild/protocompile` or
  `google.golang.org/protobuf/compiler`) rather than shelling out
  to `protoc`
- **Output**: Filtered `.proto` files written to a specified output
  directory or stdout
- **No CGo**: The tool MUST compile with `CGO_ENABLED=0` for
  maximum portability

## Development Workflow

- **Branching**: Feature branches off `main`; merge via pull request
- **Commit messages**: Conventional Commits format
  (`feat:`, `fix:`, `test:`, `docs:`, `chore:`)
- **Code review**: All changes require at least one review before
  merge
- **CI gate**: `go test -race ./...` and `golangci-lint run` MUST
  pass before merge
- **Test fixtures**: Sample `.proto` files used in tests MUST live
  under `testdata/` directories adjacent to the test files that use
  them
- **Error handling**: Return errors; do not panic. Use `fmt.Errorf`
  with `%w` for wrapping. Log to stderr only when running as CLI

## Governance

This constitution is the authoritative source of project principles.
All design decisions, code reviews, and implementation plans MUST
be evaluated against these principles. Amendments require:

1. A written proposal describing the change and its rationale.
2. Update to this file with incremented version number.
3. Review and approval via pull request.
4. Propagation check: verify that `plan-template.md`,
   `spec-template.md`, and `tasks-template.md` remain consistent
   with updated principles.

Versioning follows semantic versioning:
- **MAJOR**: Principle removed or fundamentally redefined.
- **MINOR**: New principle or section added.
- **PATCH**: Wording clarification or typo fix.

**Version**: 1.0.0 | **Ratified**: 2026-02-08 | **Last Amended**: 2026-02-08
