# Implementation Plan: Annotation Include/Exclude Filtering Modes

**Branch**: `007-annotation-include-exclude` | **Date**: 2026-02-09 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/007-annotation-include-exclude/spec.md`

## Summary

Add annotation include mode (keep only matching services/methods) alongside the existing exclude mode (remove matching). Restructure the config from a flat `annotations: [list]` to a nested `annotations.include`/`annotations.exclude` format, with backward compatibility for the old flat format via custom YAML unmarshaling. Add mutual exclusivity validation. Implement inverse filter functions (`IncludeServicesByAnnotation`, `IncludeMethodsByAnnotation`) for include mode.

## Technical Context

**Language/Version**: Go 1.25+ (existing project)
**Primary Dependencies**: `github.com/emicklei/proto` v1.14.3, `gopkg.in/yaml.v3`
**Storage**: N/A (file-based CLI tool)
**Testing**: `go test ./...` with `-race` flag; table-driven tests + golden file comparison
**Target Platform**: Cross-platform (CGO_ENABLED=0 static binary)
**Project Type**: Single Go project with `internal/` packages
**Performance Goals**: N/A (negligible performance impact)
**Constraints**: Backward compatible — old flat `annotations` config format must continue to work
**Scale/Scope**: ~80 lines of production code changed/added across 3 files; ~150 lines of new tests

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. CLI-First Tool | PASS | No CLI flag changes; config format extended with backward compat. Errors on stderr with non-zero exit codes. |
| II. Proto-Aware Filtering | PASS | Extends annotation filtering to support both include and exclude modes. Declarative configuration. |
| III. Correctness Over Convenience | PASS | Mutual exclusivity validation rejects ambiguous configs rather than guessing. Old+new format coexistence rejected with error. |
| IV. Test-Driven Development | PASS | Unit tests for config parsing, include functions, validation; golden file tests; CLI integration tests. |
| V. Simplicity and Minimal Dependencies | PASS | No new dependencies. Custom YAML unmarshaler is ~20 lines. Two new filter functions mirror existing ones. |

**Post-Phase 1 Re-check**: All gates PASS. No violations.

## Project Structure

### Documentation (this feature)

```text
specs/007-annotation-include-exclude/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── cli-interface.md # Config contract (no CLI flag changes)
└── tasks.md             # Phase 2 output (created by /speckit.tasks)
```

### Source Code (files to modify/create)

```text
.
├── internal/
│   ├── config/
│   │   ├── config.go           # Modify: AnnotationConfig struct, custom UnmarshalYAML, Validate()
│   │   └── config_test.go      # Modify: tests for new config format, backward compat, validation
│   └── filter/
│       ├── filter.go           # Modify: add IncludeServicesByAnnotation, IncludeMethodsByAnnotation
│       └── filter_test.go      # Modify: tests for include functions + golden file tests
├── main.go                     # Modify: config validation call, include/exclude dispatch logic
├── main_test.go                # Modify: CLI integration tests for include mode, validation errors
└── testdata/
    └── annotations/
        ├── include_service.proto          # New: fixture for include mode testing
        ├── expected/
        │   └── include_service.proto      # New: golden file for include mode
        └── (existing fixtures unchanged)
```

**Structure Decision**: Reuse existing project structure. Changes span 3 production files (`config.go`, `filter.go`, `main.go`) and their corresponding test files. One new test fixture added to `testdata/annotations/`.

## Complexity Tracking

No constitution violations. No complexity justification needed.
