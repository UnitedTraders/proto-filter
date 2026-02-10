# Implementation Plan: C#-Style Annotation Syntax Support

**Branch**: `006-csharp-annotation-syntax` | **Date**: 2026-02-10 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/006-csharp-annotation-syntax/spec.md`

## Summary

Add support for C#-style bracket annotation syntax (`[Name]` and `[Name(value)]`) in proto comments, alongside the existing Java-style `@Name` syntax. The implementation extends a single regex pattern and its extraction logic in `ExtractAnnotations`. No changes to configuration, CLI, or downstream filtering functions. Both styles are recognized simultaneously and matched by the same config entries.

## Technical Context

**Language/Version**: Go 1.25+ (existing project)
**Primary Dependencies**: `github.com/emicklei/proto` v1.14.3, `github.com/emicklei/proto-contrib` v0.18.3, `gopkg.in/yaml.v3`
**Storage**: N/A (file-based CLI tool)
**Testing**: `go test ./...` with `-race` flag; table-driven tests + golden file comparison
**Target Platform**: Cross-platform (CGO_ENABLED=0 static binary)
**Project Type**: Single Go project with `internal/` packages
**Performance Goals**: N/A (regex change has negligible performance impact)
**Constraints**: Backward compatible — all existing tests must pass unchanged
**Scale/Scope**: ~10 lines of production code changed; ~50 lines of new tests

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. CLI-First Tool | PASS | No CLI changes; existing interface preserved |
| II. Proto-Aware Filtering | PASS | Extends annotation recognition to support additional syntax |
| III. Correctness Over Convenience | PASS | Regex carefully designed to reject false positives (plain English in brackets) |
| IV. Test-Driven Development | PASS | New unit tests for regex, integration tests for pipeline |
| V. Simplicity and Minimal Dependencies | PASS | No new packages or dependencies; single regex change |

## Project Structure

### Documentation (this feature)

```text
specs/006-csharp-annotation-syntax/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── cli-interface.md # No CLI changes
└── tasks.md             # Phase 2 output (created by /speckit.tasks)
```

### Source Code (files to modify/create)

```text
.
├── internal/
│   └── filter/
│       ├── filter.go            # Modify: annotationRegex + ExtractAnnotations
│       └── filter_test.go       # Modify: new test cases for bracket syntax
├── main_test.go                 # Modify: integration test with bracket annotations
└── testdata/
    └── annotations/
        ├── bracket_service.proto       # New: service with [Name] annotations
        └── expected/
            └── bracket_service.proto   # New: golden file for bracket filtering
```

**Structure Decision**: Reuse existing project structure. New test fixture added to existing `testdata/annotations/` directory alongside other annotation test fixtures. No new packages or directories beyond the test fixture.

## Complexity Tracking

No constitution violations. No complexity justification needed.
