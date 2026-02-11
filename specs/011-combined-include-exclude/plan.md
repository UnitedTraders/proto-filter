# Implementation Plan: Combined Include and Exclude Annotation Filters

**Branch**: `011-combined-include-exclude` | **Date**: 2026-02-09 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/011-combined-include-exclude/spec.md`

## Summary

Remove the mutual exclusivity constraint between `annotations.include` and `annotations.exclude`. When both are configured, apply inclusion first (service/method level), then exclusion (service/method/field level). Add new capability: annotation-based exclude filtering on individual message fields. This requires changes to config validation, filter orchestration in main.go, and a new `FilterFieldsByAnnotation` function in filter.go.

## Technical Context

**Language/Version**: Go 1.25+ (existing project)
**Primary Dependencies**: `github.com/emicklei/proto` v1.14.3, `github.com/emicklei/proto-contrib` v0.18.3, `gopkg.in/yaml.v3`
**Storage**: N/A (file-based CLI tool)
**Testing**: `go test -race ./...`
**Target Platform**: Cross-platform CLI (CGO_ENABLED=0)
**Project Type**: Single Go module
**Performance Goals**: N/A (batch file processing)
**Constraints**: Output .proto files must remain syntactically valid and compilable by protoc
**Scale/Scope**: Additive feature change — modifies 3 existing files, adds 1 new function

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Evidence |
|---|---|---|
| I. CLI-First Tool | PASS | No CLI interface changes; same flags, same exit codes |
| II. Proto-Aware Filtering | PASS | Extends filtering to message fields — deeper proto awareness; output remains valid |
| III. Correctness Over Convenience | PASS | Include-then-exclude ordering is deterministic; orphan removal still runs |
| IV. Test-Driven Development | PASS | Tests written before implementation per spec |
| V. Simplicity and Minimal Dependencies | PASS | No new dependencies; single new function follows existing patterns |

## Project Structure

### Documentation (this feature)

```text
specs/011-combined-include-exclude/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
└── tasks.md             # Phase 2 output (created by /speckit.tasks)
```

### Source Code (repository root)

```text
internal/
├── config/
│   ├── config.go        # MODIFY: Remove mutual exclusivity validation
│   └── config_test.go   # MODIFY: Update TestValidateMutualExclusivity, add combined config test
├── filter/
│   ├── filter.go        # MODIFY: Add FilterFieldsByAnnotation function
│   └── filter_test.go   # MODIFY: Add field filtering unit tests, combined filtering tests
├── deps/
├── parser/
└── writer/
main.go                  # MODIFY: Change annotation filtering orchestration (if/else → sequential)
main_test.go             # MODIFY: Update TestMutualExclusivityErrorCLI, add combined CLI tests
testdata/
└── combined/            # NEW: Test fixtures for combined include+exclude scenarios
    ├── combined_service.proto
    └── expected/
        └── combined_filtered.proto
```

**Structure Decision**: Existing flat package structure is sufficient. The new `FilterFieldsByAnnotation` function follows the established pattern of `FilterMethodsByAnnotation` / `FilterServicesByAnnotation` in `internal/filter/filter.go`. No new packages needed.
