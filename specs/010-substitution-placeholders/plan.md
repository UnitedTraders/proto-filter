# Implementation Plan: Substitution Placeholders

**Branch**: `010-substitution-placeholders` | **Date**: 2026-02-11 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/010-substitution-placeholders/spec.md`

## Summary

Enable `%s` placeholder interpolation in annotation substitution values. When a substitution value contains `%s` and the matched annotation has parenthesized arguments (e.g., `@Min(3)`), the `%s` is replaced with the argument content (e.g., `"Minimal value is %s"` → `"Minimal value is 3"`). This requires modifying `substitutionRegex` to capture the argument text as a group, and updating `substituteInComment` to perform the `%s` replacement when a captured argument is present.

## Technical Context

**Language/Version**: Go 1.25+
**Primary Dependencies**: `github.com/emicklei/proto` v1.14.3, `gopkg.in/yaml.v3`
**Storage**: N/A (file-based CLI tool)
**Testing**: `go test -race ./...`
**Target Platform**: Cross-platform CLI (`CGO_ENABLED=0`)
**Project Type**: Single Go module
**Performance Goals**: N/A (string replacement, negligible overhead)
**Constraints**: Backward compatibility with all existing substitution tests
**Scale/Scope**: 2 source files modified (`internal/filter/filter.go`, no changes to `main.go`)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. CLI-First Tool | PASS | No CLI interface changes; feature is internal to substitution pipeline |
| II. Proto-Aware Filtering | PASS | Operates on annotation comments within proto AST |
| III. Correctness Over Convenience | PASS | Backward compatible (FR-004, FR-009, FR-010); only first `%s` replaced (FR-007) |
| IV. Test-Driven Development | PASS | Tests written first per TDD mandate |
| V. Simplicity and Minimal Dependencies | PASS | No new dependencies; change is ~15 lines in existing function + regex tweak |

All gates pass. No violations.

## Project Structure

### Documentation (this feature)

```text
specs/010-substitution-placeholders/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── cli-interface.md # No CLI changes (placeholder only)
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
internal/filter/
├── filter.go            # Modified: substitutionRegex + substituteInComment
└── filter_test.go       # Modified: new placeholder interpolation tests

testdata/substitution/
├── substitution_service.proto          # Existing fixture (already has args)
├── placeholder_service.proto           # New fixture for placeholder tests
└── expected/
    ├── substitution_replaced.proto     # Existing golden (unchanged)
    ├── substitution_removed.proto      # Existing golden (unchanged)
    └── placeholder_replaced.proto      # New golden for placeholder output
```

**Structure Decision**: Single Go module, flat internal package structure. Changes are confined to `internal/filter/filter.go` (regex + substitution logic) and `internal/filter/filter_test.go` (new tests). No `main.go` changes needed — the config already passes `map[string]string` substitutions, and the interpolation happens inside `substituteInComment`.

## Complexity Tracking

No violations to justify.
