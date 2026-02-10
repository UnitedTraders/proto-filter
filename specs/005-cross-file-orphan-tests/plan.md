# Implementation Plan: Cross-File Orphan Detection Tests

**Branch**: `005-cross-file-orphan-tests` | **Date**: 2026-02-09 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/005-cross-file-orphan-tests/spec.md`

## Summary

Add test coverage for cross-file orphan detection when annotation filtering is applied. When service methods reference message types from separate "common" proto files, the filtering pipeline must preserve shared types still referenced by surviving methods and remove types only referenced by filtered-out services. This feature adds test fixtures and tests only — no production code changes.

The tests verify behavior at two levels:
1. The dependency graph's cross-file transitive resolution (via `TransitiveDeps` and `RequiredFiles`)
2. The per-file `RemoveOrphanedDefinitions` behavior after annotation filtering

## Technical Context

**Language/Version**: Go 1.25+ (existing project)
**Primary Dependencies**: `github.com/emicklei/proto` v1.14.3, `github.com/emicklei/proto-contrib` v0.18.3, `gopkg.in/yaml.v3`
**Storage**: N/A (file-based CLI tool)
**Testing**: `go test ./...` with `-race` flag; golden file comparison pattern
**Target Platform**: Cross-platform (CGO_ENABLED=0 static binary)
**Project Type**: Single Go project with `internal/` packages
**Performance Goals**: N/A (test-only feature)
**Constraints**: No production code changes; no new dependencies
**Scale/Scope**: New test fixtures in `testdata/crossfile/`; new test functions in `internal/filter/filter_test.go` and `main_test.go`

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. CLI-First Tool | PASS | No CLI changes; tests verify existing CLI behavior |
| II. Proto-Aware Filtering | PASS | Tests verify cross-file referential integrity is preserved |
| III. Correctness Over Convenience | PASS | Tests ensure shared types are not incorrectly removed as orphans |
| IV. Test-Driven Development | PASS | This feature IS the tests — adding coverage for cross-file scenarios |
| V. Simplicity and Minimal Dependencies | PASS | No new dependencies; tests use existing test infrastructure |

## Project Structure

### Documentation (this feature)

```text
specs/005-cross-file-orphan-tests/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── cli-interface.md # No CLI changes (test-only)
└── tasks.md             # Phase 2 output (created by /speckit.tasks)
```

### Source Code (files to create)

```text
.
├── internal/
│   └── filter/
│       └── filter_test.go           # New cross-file unit tests
├── main_test.go                     # New cross-file CLI integration tests
└── testdata/
    └── crossfile/                   # New test fixture directory
        ├── common.proto             # Shared message types (no services)
        ├── orders.proto             # Service referencing common types
        ├── payments.proto           # Service with @Internal, referencing common types
        └── expected/               # Golden file outputs
            ├── common.proto         # Expected common.proto after filtering
            ├── orders.proto         # Expected orders.proto (unchanged)
            └── payments.proto       # Expected payments.proto (service removed)
```

**Structure Decision**: Reuse existing project structure. New test fixtures go into a dedicated `testdata/crossfile/` directory to isolate cross-file tests from existing single-file annotation tests. No new packages or production code.

## Testing Strategy

- **Unit tests** (`internal/filter/filter_test.go`): Test that the filter pipeline correctly handles multi-file scenarios where messages are defined in a common file and referenced from service files
- **Golden file tests**: Create expected output files for the cross-file scenario and compare byte-for-byte
- **CLI integration tests** (`main_test.go`): Run the binary end-to-end with cross-file fixtures and annotation config, verify correct output files are produced with correct content
- **Key scenarios**:
  - Shared type referenced by surviving service → preserved
  - Shared type referenced only by removed service → removed
  - Shared type referenced by multiple services, one removed → preserved
  - Service-level annotation filtering with cross-file references
  - Common file with only messages (no services) → types preserved if referenced

## Complexity Tracking

No constitution violations. No complexity justification needed.
