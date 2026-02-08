# Implementation Plan: Comment Style Conversion

**Branch**: `003-comment-conversion` | **Date**: 2026-02-09 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/003-comment-conversion/spec.md`

## Summary

Convert all C-style block comments (`/* ... */` and `/** ... */`) in proto files to consecutive single-line `//` comments. The `emicklei/proto` parser already strips asterisk prefixes from block comment lines, so the conversion only requires setting `Cstyle = false` on each `Comment` struct and cleaning any residual `*` prefixes from `Lines`. This is a post-parse, pre-write AST transformation applied unconditionally to all processed files.

## Technical Context

**Language/Version**: Go 1.25+ (existing project)
**Primary Dependencies**: `github.com/emicklei/proto` v1.14.3, `github.com/emicklei/proto-contrib` v0.18.3
**Storage**: N/A (file-based CLI tool)
**Testing**: `go test -race ./...`
**Target Platform**: Cross-platform CLI (CGO_ENABLED=0)
**Project Type**: Single Go module
**Performance Goals**: N/A (file transformation, no latency concerns)
**Constraints**: Must not modify comment text content; must produce valid parseable proto files
**Scale/Scope**: Operates on same file sets as existing filtering pipeline

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. CLI-First Tool | PASS | Comment conversion runs as part of existing CLI pipeline, no new flags needed |
| II. Proto-Aware Filtering | PASS | Operates on parsed proto AST, output remains valid proto |
| III. Correctness Over Convenience | PASS | Preserves all comment text content; no references affected |
| IV. Test-Driven Development | PASS | Tests for conversion function, AST walking, and integration planned |
| V. Simplicity and Minimal Dependencies | PASS | No new dependencies; reuses existing `proto.Walk` patterns; single new function in existing package |

## Project Structure

### Documentation (this feature)

```text
specs/003-comment-conversion/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── cli-interface.md
└── tasks.md
```

### Source Code (repository root)

```text
internal/
├── filter/
│   ├── filter.go          # Add ConvertBlockComments function
│   └── filter_test.go     # Add conversion tests
main.go                    # Wire conversion into pipeline
main_test.go               # Add CLI integration test
testdata/
└── comments/
    ├── commented.proto            # Existing: single-line comments
    ├── multiline.proto            # Existing: block comments
    ├── block_comments.proto       # New input fixture with block comments
    └── expected/                  # Golden files for output comparison
        ├── commented.proto        # Expected output for commented.proto (unchanged)
        ├── multiline.proto        # Expected output for multiline.proto (converted)
        └── block_comments.proto   # Expected output for block_comments.proto (converted)
```

**Structure Decision**: Existing single-project Go module structure. Comment conversion is a new function in `internal/filter/filter.go`, following the same pattern as `FilterMethodsByAnnotation` and `RemoveEmptyServices`. No new packages needed.

## Testing Strategy

### Golden File Comparison

Tests use a golden file pattern: each input `.proto` file in `testdata/comments/` has a corresponding expected output file in `testdata/comments/expected/`. Tests process the input file through the comment conversion pipeline, write the result to a temp directory, and compare the actual output byte-for-byte against the golden file.

**Pattern**:
1. Parse input file from `testdata/comments/<name>.proto`
2. Run `ConvertBlockComments()` on the AST
3. Write output via `writer.WriteProtoFile()` to a temp file
4. Read the golden file from `testdata/comments/expected/<name>.proto`
5. Compare actual output against golden file content
6. Fail with a diff if they don't match

**Golden files cover**:
- `commented.proto` → single-line comments pass through unchanged
- `multiline.proto` → block comments converted to `//` style
- `block_comments.proto` → new fixture with varied block comment patterns (Javadoc-style, inline, empty, mixed indentation)

**Updating golden files**: When intentional output changes are made, regenerate expected files by running the conversion on inputs and copying results to `expected/`.
