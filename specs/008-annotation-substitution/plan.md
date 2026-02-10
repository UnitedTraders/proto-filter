# Implementation Plan: Annotation Substitution

**Branch**: `008-annotation-substitution` | **Date**: 2026-02-09 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/008-annotation-substitution/spec.md`

## Summary

Add annotation substitution to proto-filter: a configurable mapping from annotation names to description text that replaces annotations in proto comments during output processing. Empty descriptions remove annotations entirely with clean line/comment cleanup. An optional strict mode detects and reports all unsubstituted annotations across all processed files before writing any output. Substitution operates independently from existing annotation include/exclude filtering.

## Technical Context

**Language/Version**: Go 1.25+ (existing project)
**Primary Dependencies**: `github.com/emicklei/proto` v1.14.3, `gopkg.in/yaml.v3`
**Storage**: N/A (file-based CLI tool)
**Testing**: `go test ./...` with `-race` flag; table-driven tests + golden file comparison
**Target Platform**: Cross-platform (CGO_ENABLED=0 static binary)
**Project Type**: Single Go project with `internal/` packages
**Performance Goals**: N/A (negligible performance impact — string replacement per comment line)
**Constraints**: Backward compatible — configs without `substitutions`/`strict_substitutions` must work identically
**Scale/Scope**: ~120 lines of production code added across 3 files; ~200 lines of new tests

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. CLI-First Tool | PASS | No new CLI flags. Config extends YAML format. Errors on stderr with exit code 2. Strict mode error is descriptive and lists all missing annotations. |
| II. Proto-Aware Filtering | PASS | Substitution modifies proto comment text on surviving AST elements. Output remains syntactically valid — comment content changes don't affect proto validity. |
| III. Correctness Over Convenience | PASS | Strict mode prevents accidental annotation leakage. Unsubstituted annotations are left unchanged (not silently dropped) in non-strict mode. |
| IV. Test-Driven Development | PASS | Unit tests for substitution logic, empty line cleanup, comment removal. Golden file tests for end-to-end output. CLI integration tests for strict mode errors. |
| V. Simplicity and Minimal Dependencies | PASS | No new dependencies. One new regex. ~4 new functions in filter.go. Two new fields in config struct. |

**Post-Phase 1 Re-check**: All gates PASS. No violations.

## Project Structure

### Documentation (this feature)

```text
specs/008-annotation-substitution/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── cli-interface.md # Config contract
└── tasks.md             # Phase 2 output (created by /speckit.tasks)
```

### Source Code (files to modify/create)

```text
.
├── internal/
│   ├── config/
│   │   ├── config.go           # Modify: add Substitutions map + StrictSubstitutions bool + HasSubstitutions() + IsPassThrough() update
│   │   └── config_test.go      # Modify: tests for new config fields
│   └── filter/
│       ├── filter.go           # Modify: add SubstituteAnnotations(), CollectAllAnnotations(), substitution regex
│       └── filter_test.go      # Modify: tests for substitution + strict mode + golden files
├── main.go                     # Modify: strict mode check + substitution call in pipeline
├── main_test.go                # Modify: CLI integration tests for substitution + strict mode
└── testdata/
    └── annotations/
        ├── substitution_service.proto          # New: fixture with various annotations for substitution testing
        ├── expected/
        │   ├── substitution_replaced.proto     # New: golden file — annotations replaced with descriptions
        │   └── substitution_removed.proto      # New: golden file — annotations removed (empty substitution)
        └── (existing fixtures unchanged)
```

**Structure Decision**: Reuse existing project structure. Changes span 3 production files (`config.go`, `filter.go`, `main.go`) and their corresponding test files. New test fixtures added to `testdata/annotations/`.

### Key Implementation Details

**New regex** (for substitution — matches full annotation expression including parameters):
```
@(\w[\w.]*)(?:\([^)]*\))?|\[(\w[\w.]*)(?:\([^)]*\))?\]
```

This differs from the existing `annotationRegex` by also matching `@Name(...)` parameters for `@`-style annotations. Group 0 = full token to replace, group 1 or 2 = annotation name for lookup.

**Pipeline insertion point** (in main.go per-file loop):
1. Annotation filtering (existing)
2. Block comment conversion (existing)
3. **Strict mode collection** (new — first pass: collect all annotation names from comments)
4. **Substitution** (new — second pass: replace annotation tokens)
5. Write output (existing)

Note: Strict mode must be a pre-pass across ALL files before any writing, not per-file. This means:
- First loop: filter + convert + collect annotations per file
- After loop: strict check (if enabled) — fail if any unsubstituted
- Second loop (or deferred write): substitute + write

**Config changes**:
- `Substitutions map[string]string` with `yaml:"substitutions"` tag
- `StrictSubstitutions bool` with `yaml:"strict_substitutions"` tag
- `HasSubstitutions() bool` helper
- `IsPassThrough()` does NOT consider substitutions (substitution-only config still writes all files)

## Complexity Tracking

No constitution violations. No complexity justification needed.
