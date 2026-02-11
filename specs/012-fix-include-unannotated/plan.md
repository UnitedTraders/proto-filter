# Implementation Plan: Fix Include Filter Keeping Unannotated Services

**Branch**: `012-fix-include-unannotated` | **Date**: 2026-02-11 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/012-fix-include-unannotated/spec.md`

## Summary

Bug fix: `IncludeServicesByAnnotation` currently keeps services that have no annotation comments, treating them as "undecided" and deferring to method-level filtering. Per the feature 011 spec, services without annotations should be treated as not matching the include list and removed. The fix is a one-line change in the function's `len(annots) == 0` branch, plus test updates to reflect the corrected behavior.

## Technical Context

**Language/Version**: Go 1.25+
**Primary Dependencies**: `github.com/emicklei/proto` v1.14.3, `gopkg.in/yaml.v3`
**Storage**: N/A (file-based CLI tool)
**Testing**: `go test -race ./...`
**Target Platform**: Cross-platform CLI (CGO_ENABLED=0)
**Project Type**: Single Go module
**Performance Goals**: N/A (bug fix, no performance change)
**Constraints**: Output `.proto` files must remain syntactically valid and compilable by `protoc`
**Scale/Scope**: One function fix + 2 test updates + 1-2 golden file updates

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. CLI-First Tool | PASS | No CLI interface changes |
| II. Proto-Aware Filtering | PASS | Fix improves filtering correctness |
| III. Correctness Over Convenience | PASS | This fix enforces the documented include-gate behavior |
| IV. Test-Driven Development | PASS | Tests updated alongside implementation |
| V. Simplicity and Minimal Dependencies | PASS | One-line fix, no new dependencies |

All gates pass. No violations to justify.

## Project Structure

### Documentation (this feature)

```text
specs/012-fix-include-unannotated/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── spec.md              # Feature specification
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
internal/filter/
├── filter.go            # IncludeServicesByAnnotation fix (line ~280)
└── filter_test.go       # TestIncludeServicesByAnnotation update

testdata/annotations/
├── include_service.proto          # Input fixture (needs service-level annotation)
└── expected/
    └── include_service.proto      # Golden file (already updated by feature 011)

main.go                            # No changes needed
main_test.go                       # TestIncludeAnnotationFilteringCLI update
```

**Structure Decision**: Existing single Go module structure. Changes confined to `internal/filter/` and `testdata/`.

## Change Analysis

### Root Cause

In `internal/filter/filter.go:280-285`, `IncludeServicesByAnnotation` has:

```go
if len(annots) == 0 {
    // No service-level annotations: keep the service,
    // let method-level filtering decide its fate.
    filtered = append(filtered, elem)
    continue
}
```

This keeps unannotated services. The fix: remove the `continue` early-return so unannotated services fall through to the `hasMatch` check, which will be `false` (no annotations to match), and the service will be removed.

### Fix (1 file)

**`internal/filter/filter.go`** — `IncludeServicesByAnnotation` function (~line 280):

Remove the special case for `len(annots) == 0`. Unannotated services should fall through to the match check and be removed (since they have no matching annotation). The simplest fix is to delete the `if len(annots) == 0` block entirely — when `annots` is empty, the `hasMatch` loop will not execute, `hasMatch` stays `false`, and the service is correctly excluded.

### Test Updates (2 files)

**`internal/filter/filter_test.go`** — `TestIncludeServicesByAnnotation` (~line 1544):
- Change expected removed count from 1 to 2
- Remove assertion that `UnannotatedService` should remain
- Add assertion that `UnannotatedService` should be removed
- Update comments to reflect corrected behavior

**`main_test.go`** — `TestIncludeAnnotationFilteringCLI` (~line 756):
- The test uses `testdata/annotations/include_service.proto` which has an **unannotated** `OrderService` with method-level `@Public` and `[Public]` annotations
- After the fix, `IncludeServicesByAnnotation` will remove the entire service (no service-level annotation)
- Fix: Add a `[Public]` service-level annotation to the input fixture `testdata/annotations/include_service.proto`
- This way the service is kept by the include gate, and method-level filtering still applies within it
- The existing golden file (`testdata/annotations/expected/include_service.proto`) already has the correct expected output for this case

### Golden File Updates

**`testdata/annotations/include_service.proto`** (input fixture, NOT golden):
- Add `// [Public]` annotation comment above `service OrderService {`
- This makes the service match the include list at the service level

**`testdata/annotations/expected/include_service.proto`** (golden file):
- Already correct from feature 011 updates (annotations stripped, methods with `@Public`/`[Public]` kept)
- No changes needed

### No Changes Needed

- `main.go` — The calling code in main.go already calls `IncludeServicesByAnnotation` correctly; the fix is entirely within the function
- `testdata/combined/` — The combined test fixtures already have proper service-level annotations
- All exclude-only tests — Unaffected (they don't call `IncludeServicesByAnnotation`)
