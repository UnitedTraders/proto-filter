# Implementation Plan: Annotation-Based Method Filtering

**Branch**: `002-annotation-filter` | **Date**: 2026-02-08 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/002-annotation-filter/spec.md`

## Summary

Extend the proto-filter CLI to support annotation-based RPC method filtering. Methods whose comments contain specified Java-style annotations (e.g., `@HasAnyRole`) are removed from the output. After method removal, orphaned messages/enums are cleaned up and empty services are dropped. Configured via a new `annotations` key in the existing YAML config file. Implemented as a separate filtering pass after the existing name-based filter pipeline, keeping backward compatibility.

## Technical Context

**Language/Version**: Go 1.25+ (existing project)
**Primary Dependencies**: `github.com/emicklei/proto` v1.14.3, `github.com/emicklei/proto-contrib` v0.18.3, `gopkg.in/yaml.v3`
**Storage**: N/A (file-based CLI tool)
**Testing**: `go test ./...` with `-race` flag
**Target Platform**: Cross-platform (CGO_ENABLED=0 static binary)
**Project Type**: Single Go project with `internal/` packages
**Performance Goals**: N/A (batch CLI tool)
**Constraints**: No new dependencies; no CGo; backward compatible
**Scale/Scope**: Extends 5 existing internal packages; ~200 lines of new code

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. CLI-First Tool | PASS | No CLI changes; annotation config via existing `--config` flag |
| II. Proto-Aware Filtering | PASS | Extends filtering to method-level granularity; output remains valid proto |
| III. Correctness Over Convenience | PASS | Orphan detection ensures no dangling references; transitive cleanup prevents incomplete output |
| IV. Test-Driven Development | PASS | Tests written before implementation per plan |
| V. Simplicity and Minimal Dependencies | PASS | No new dependencies; new functions added to existing `filter` and `config` packages; no new packages |

**Post-Phase 1 re-check**: All gates still PASS. No new packages added, no new dependencies, annotation logic fits naturally into existing `internal/filter/` package.

## Project Structure

### Documentation (this feature)

```text
specs/002-annotation-filter/
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
├── main.go                          # Wire annotation filtering into pipeline
├── internal/
│   ├── config/
│   │   ├── config.go                # Add Annotations field to FilterConfig
│   │   └── config_test.go           # Test annotation config loading
│   ├── filter/
│   │   ├── filter.go                # Add annotation parsing, method filtering, orphan detection
│   │   └── filter_test.go           # Tests for all new filter functions
│   ├── parser/
│   │   └── parser_test.go           # Integration test for annotation pipeline
│   ├── deps/                        # No changes
│   └── writer/                      # No changes
└── testdata/
    └── annotations/                 # New test fixtures
        ├── service.proto            # Service with annotated and non-annotated methods
        ├── shared.proto             # Shared messages referenced across methods
        └── internal_only.proto      # Service where all methods are annotated
```

**Structure Decision**: Reuse existing project structure. New logic goes into existing packages (`config`, `filter`). New test fixtures go into `testdata/annotations/`. No new internal packages.
