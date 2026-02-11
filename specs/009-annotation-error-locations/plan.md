# Implementation Plan: Annotation Error Locations

**Branch**: `009-annotation-error-locations` | **Date**: 2026-02-11 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/009-annotation-error-locations/spec.md`

## Summary

Enhance strict substitution error reporting to include exact source file locations (file path and line number) for each unsubstituted annotation. Currently, the tool outputs only a flat list of missing annotation names. The new output adds per-occurrence location lines after the existing summary line, enabling users to jump directly to each annotation in their editor. Requires replacing `CollectAllAnnotations` (which returns `map[string]bool`) with a new `CollectAnnotationLocations` function that returns a slice of location structs containing file path, line number, annotation name, and full annotation token.

## Technical Context

**Language/Version**: Go 1.25+ (existing project)
**Primary Dependencies**: `github.com/emicklei/proto` v1.14.3, `gopkg.in/yaml.v3`
**Storage**: N/A (file-based CLI tool)
**Testing**: `go test ./...` with `-race` flag; table-driven tests + golden file comparison
**Target Platform**: Cross-platform (CGO_ENABLED=0 static binary)
**Project Type**: Single Go project with `internal/` packages
**Performance Goals**: N/A (error path only — negligible overhead)
**Constraints**: Backward compatible — existing summary line format preserved; existing tests must pass without modification
**Scale/Scope**: ~40 lines of production code changed across 2 files; ~60 lines of new tests

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. CLI-First Tool | PASS | Enhanced error output goes to stderr. Exit code 2 unchanged. Location format uses standard `file:line:` convention familiar to compilers and editors. |
| II. Proto-Aware Filtering | PASS | No change to filtering logic. Location collection reads comment positions from the proto AST. Output files remain valid proto. |
| III. Correctness Over Convenience | PASS | Every unsubstituted annotation occurrence is reported with its exact location. Zero occurrences missed. |
| IV. Test-Driven Development | PASS | Unit tests for the new collection function. CLI integration tests for the location output format. Existing strict mode tests continue to pass. |
| V. Simplicity and Minimal Dependencies | PASS | No new dependencies. One new struct type. One new function replacing an existing one. Minor change to main.go error formatting. |

**Post-Phase 1 Re-check**: All gates PASS. No violations.

## Project Structure

### Documentation (this feature)

```text
specs/009-annotation-error-locations/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── cli-interface.md # Error output contract
└── tasks.md             # Phase 2 output (created by /speckit.tasks)
```

### Source Code (files to modify)

```text
.
├── internal/
│   └── filter/
│       ├── filter.go           # Modify: add AnnotationLocation struct, CollectAnnotationLocations() function
│       └── filter_test.go      # Modify: tests for CollectAnnotationLocations
├── main.go                     # Modify: use CollectAnnotationLocations, format location lines in strict error
├── main_test.go                # Modify: CLI integration tests for location output
└── testdata/
    └── substitution/
        └── substitution_service.proto  # Existing fixture — reuse for location tests
```

**Structure Decision**: Reuse existing project structure. Changes span 2 production files (`filter.go`, `main.go`) and their corresponding test files. No new test fixtures needed — the existing `substitution_service.proto` has annotations at known line numbers.

### Key Implementation Details

**New struct** (in `filter.go`):
```go
type AnnotationLocation struct {
    File  string // relative file path
    Line  int    // 1-based line number in source
    Name  string // annotation name (for substitution map lookup)
    Token string // full annotation token as it appears in source
}
```

**New function** `CollectAnnotationLocations(def *proto.Proto, relPath string) []AnnotationLocation`:
- Walks all elements in the proto AST (services, RPCs, messages, enums, fields)
- For each comment, iterates lines and uses `substitutionRegex` to find annotation tokens
- Computes line number: `comment.Position.Line + lineIndex` (for `//` style single-line comments, `Position.Line` points to the first `//` line, and each `Lines[i]` corresponds to `Position.Line + i`)
- Returns a slice of `AnnotationLocation` for all annotations found

**Pipeline change** (in `main.go`):
- Replace `allAnnotations map[string]bool` with `allLocations []filter.AnnotationLocation`
- Replace `filter.CollectAllAnnotations(pf.def)` call with `filter.CollectAnnotationLocations(pf.def, pf.rel)`
- In the strict check block: extract unique missing names from locations, print summary line (unchanged), then print sorted location lines for unsubstituted annotations only

**Error output format**:
```
proto-filter: error: unsubstituted annotations found: Deprecated, SupportWindow
  orders.proto:7: @Deprecated
  orders.proto:12: @SupportWindow({duration: "6M"})
```

**Backward compatibility**: The existing `CollectAllAnnotations` function can remain as-is (it's used only by main.go in the strict check path, but keeping it avoids breaking any external callers). However, since it's an internal package, we can simply replace its usage in main.go with the new function.

## Complexity Tracking

No constitution violations. No complexity justification needed.
