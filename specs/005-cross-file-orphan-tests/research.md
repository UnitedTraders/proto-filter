# Research: Cross-File Orphan Detection Tests

**Date**: 2026-02-09
**Branch**: `005-cross-file-orphan-tests`

## Decision 1: Understanding the Cross-File Orphan Detection Mechanism

**Decision**: The tests must verify behavior at two levels: (1) the
dependency graph's cross-file transitive resolution, and (2) the
per-file `RemoveOrphanedDefinitions` behavior after annotation filtering.

**Rationale**: The pipeline processes files in two phases:

1. **Name-based filtering** (graph-aware): `ApplyFilter` → `TransitiveDeps`
   → `keepFQNs` → `PruneAST`. This phase correctly resolves cross-file
   references via the dependency graph. A common file's message types are
   included in `keepFQNs` if any service's method references them.

2. **Annotation filtering** (per-file): `FilterServicesByAnnotation` →
   `FilterMethodsByAnnotation` → `RemoveEmptyServices` →
   `RemoveOrphanedDefinitions`. This phase operates on each file's AST
   independently. `RemoveOrphanedDefinitions` only sees references
   within the current file — it has no cross-file visibility.

**Key insight**: For a common file containing only messages and no
services, `RemoveOrphanedDefinitions` will find zero references within
that file (since there are no services or methods to reference anything).
However, since the common file has no services, annotation filtering
has no effect on it — no services or methods are removed, so
`RemoveOrphanedDefinitions` sees the same state as after `PruneAST`.

For a service file, after annotation filtering removes methods, some of
its locally-defined messages may become orphaned. But messages from
external files are not in this file's AST at all — they live in their
own file and are handled independently.

**Conclusion**: The cross-file orphan concern is addressed by:
- The dependency graph ensuring common files are included in the output
- Per-file orphan detection only removing locally-defined types
- Common files (message-only) being unaffected by annotation filtering

Tests should verify all three aspects.

## Decision 2: Test Fixture Strategy

**Decision**: Create a dedicated `testdata/crossfile/` directory with
a multi-file setup: one common/shared proto file and one or more service
proto files that reference types from the common file.

**Rationale**: Existing `testdata/annotations/` fixtures use a single
package but each file is relatively self-contained (messages and services
in the same file, or `shared.proto` which has both services and messages).
None of the existing fixtures test the pure "common file with only
messages, referenced from another file" pattern.

A dedicated directory keeps cross-file tests isolated and makes the
fixture structure self-documenting.

**Alternatives considered**:

| Approach | Verdict | Why rejected |
|----------|---------|--------------|
| Add to existing `testdata/annotations/` | Rejected | Would clutter existing fixtures and risk breaking existing tests that enumerate files in the directory |
| Use programmatic AST construction | Rejected | CLI integration tests need real files; fixtures better represent real-world usage |

## Decision 3: Test Levels

**Decision**: Write both unit-level tests (in `internal/filter/filter_test.go`)
and CLI integration tests (in `main_test.go`).

**Rationale**: Unit tests verify the filter functions' behavior in
isolation. CLI integration tests verify the full pipeline including
file discovery, graph construction, filtering, and output writing.
FR-004 explicitly requires CLI-level tests.

**Alternatives considered**:

| Approach | Verdict | Why rejected |
|----------|---------|--------------|
| CLI tests only | Rejected | Would miss per-file orphan detection behavior details |
| Unit tests only | Rejected | Would not verify the full pipeline (FR-004) |
