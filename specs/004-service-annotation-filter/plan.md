# Implementation Plan: Service-Level Annotation Filtering

**Branch**: `004-service-annotation-filter` | **Date**: 2026-02-09 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/004-service-annotation-filter/spec.md`

## Summary

Extend the proto-filter annotation filtering to inspect service-level comments. When a service's comment contains an annotation matching the filter config, the entire service (including all its RPC methods) is removed from the output. Implemented as a new `FilterServicesByAnnotation` function that runs before the existing method-level `FilterMethodsByAnnotation` in the filtering pipeline. Reuses existing `ExtractAnnotations`, orphan cleanup, and empty-file handling. No config format changes — the existing `annotations` YAML key applies to both service and method levels.

## Technical Context

**Language/Version**: Go 1.25+ (existing project)
**Primary Dependencies**: `github.com/emicklei/proto` v1.14.3, `github.com/emicklei/proto-contrib` v0.18.3, `gopkg.in/yaml.v3`
**Storage**: N/A (file-based CLI tool)
**Testing**: `go test ./...` with `-race` flag; golden file comparison pattern
**Target Platform**: Cross-platform (CGO_ENABLED=0 static binary)
**Project Type**: Single Go project with `internal/` packages
**Performance Goals**: N/A (batch CLI tool)
**Constraints**: No new dependencies; no CGo; backward compatible
**Scale/Scope**: Adds ~1 new function to `internal/filter/`; ~50 lines of new code; updates `main.go` pipeline; new test fixtures

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. CLI-First Tool | PASS | No CLI changes; same `--config` flag and `annotations` YAML key |
| II. Proto-Aware Filtering | PASS | Extends filtering to service-level granularity; output remains valid proto |
| III. Correctness Over Convenience | PASS | Orphan detection ensures no dangling references after service removal |
| IV. Test-Driven Development | PASS | Tests written before/alongside implementation per plan |
| V. Simplicity and Minimal Dependencies | PASS | No new dependencies; new function added to existing `filter` package; no new packages |

## Project Structure

### Documentation (this feature)

```text
specs/004-service-annotation-filter/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── cli-interface.md # Extended CLI contract
└── tasks.md             # Phase 2 output (created by /speckit.tasks)
```

### Source Code (files to modify/create)

```text
.
├── main.go                          # Wire service annotation filtering into pipeline
├── internal/
│   ├── filter/
│   │   ├── filter.go                # Add FilterServicesByAnnotation function
│   │   └── filter_test.go           # Tests for service-level annotation filtering
│   ├── config/                      # No changes
│   ├── parser/                      # No changes
│   ├── deps/                        # No changes
│   └── writer/                      # No changes
└── testdata/
    └── annotations/                 # New test fixture
        └── service_annotated.proto  # Service with annotations at service level
```

**Structure Decision**: Reuse existing project structure. New function goes into existing `internal/filter/` package. New test fixture goes into existing `testdata/annotations/`. No new packages.

## Testing Strategy

- **Unit tests**: Test `FilterServicesByAnnotation` with various annotation scenarios (matching, non-matching, multiple annotations, mixed service/method annotations)
- **Golden file tests**: Create expected output files and compare byte-for-byte
- **CLI integration tests**: Run the binary with service-annotated proto files and verify output
- **Backward compatibility**: Verify all existing tests continue to pass unchanged
