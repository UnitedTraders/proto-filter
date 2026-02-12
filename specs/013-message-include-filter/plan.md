# Implementation Plan: Include Annotation Filtering for Top-Level Messages

**Branch**: `013-message-include-filter` | **Date**: 2026-02-11 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/013-message-include-filter/spec.md`

## Summary

When `annotations.include` is configured, top-level messages and enums currently pass through unfiltered — only services are gated by include annotations. This feature adds `IncludeMessagesByAnnotation` to filter top-level messages/enums by include annotations, then relies on existing orphan removal to preserve their transitive dependencies. The orphan removal gate in `main.go` must also include message removal counts.

## Technical Context

**Language/Version**: Go 1.25+
**Primary Dependencies**: `github.com/emicklei/proto` v1.14.3, `gopkg.in/yaml.v3`
**Storage**: N/A (file-based CLI tool)
**Testing**: `go test -race ./...`
**Target Platform**: Cross-platform CLI (CGO_ENABLED=0)
**Project Type**: Single Go module
**Performance Goals**: N/A (feature addition, no performance change)
**Constraints**: Output `.proto` files must remain syntactically valid and compilable by `protoc`
**Scale/Scope**: One new function + main.go pipeline update + tests

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. CLI-First Tool | PASS | No CLI interface changes |
| II. Proto-Aware Filtering | PASS | Extends filtering to messages, consistent with proto structure awareness |
| III. Correctness Over Convenience | PASS | Preserves referential integrity — included messages keep their dependencies |
| IV. Test-Driven Development | PASS | Unit tests + integration tests + golden file tests |
| V. Simplicity and Minimal Dependencies | PASS | Follows existing `IncludeServicesByAnnotation` pattern, no new deps |

All gates pass.

## Project Structure

### Documentation (this feature)

```text
specs/013-message-include-filter/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── spec.md              # Feature specification
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
internal/filter/
├── filter.go            # IncludeMessagesByAnnotation function
└── filter_test.go       # Unit tests for IncludeMessagesByAnnotation

main.go                  # Pipeline update: call IncludeMessagesByAnnotation, update orphan gate
main_test.go             # CLI integration tests for message include filtering

testdata/messages/                    # New test fixture directory
├── messages_only.proto               # Input: messages only, no services
├── messages_annotated.proto          # Input: one annotated message with deps
└── expected/
    ├── messages_annotated.proto      # Golden: annotated message + deps
    └── (messages_only.proto absent)  # Empty output = file not written
```

**Structure Decision**: Existing single Go module. New `testdata/messages/` directory for message-only test fixtures.

## Change Analysis

### Design Approach

**Add `IncludeMessagesByAnnotation`** following the exact pattern of `IncludeServicesByAnnotation`:

1. Walk `def.Elements`
2. For each `*proto.Message` and `*proto.Enum`, extract annotations from comment
3. If annotations match the include list → keep
4. If annotations don't match (or no annotations) → remove and increment counter
5. Non-message/non-enum elements (services, syntax, package, imports, options) pass through unchanged

After this function runs, existing `RemoveOrphanedDefinitions` handles dependencies:
- `CollectReferencedTypes` (line 461) already collects references from **message fields** (NormalField, MapField, OneOfField via `collectMessageRefs`)
- So if `[PublishedApi] Parameters` references `Params1` and `Params2`, and we keep `Parameters`, orphan removal sees `Parameters` still references `Params1` and `Params2` and keeps them
- Messages not referenced by anything are removed as orphans

### New Function (1 file)

**`internal/filter/filter.go`** — `IncludeMessagesByAnnotation`:

```
func IncludeMessagesByAnnotation(def *proto.Proto, annotations []string) int
```

- Follows `IncludeServicesByAnnotation` pattern exactly
- Filters `*proto.Message` and `*proto.Enum` elements
- Uses `ExtractAnnotations(msg.Comment)` — already works on any `*proto.Comment`
- Returns count of removed messages/enums

### Pipeline Update (1 file)

**`main.go`** — annotation filtering block (~line 194):

1. Add call to `filter.IncludeMessagesByAnnotation(pf.def, cfg.Annotations.Include)` inside the `if cfg.HasAnnotationInclude()` block
2. Track result in a new variable (e.g., `msgr`) or add to existing counter
3. Update orphan removal gate from `if sr > 0 || mr > 0 || fr > 0` to also include message removal count

Current:
```go
if cfg.HasAnnotationInclude() {
    sr += filter.IncludeServicesByAnnotation(pf.def, cfg.Annotations.Include)
    ...
}
...
if sr > 0 || mr > 0 || fr > 0 {
    orphansRemoved += filter.RemoveOrphanedDefinitions(pf.def, pf.pkg)
}
```

After:
```go
if cfg.HasAnnotationInclude() {
    sr += filter.IncludeServicesByAnnotation(pf.def, cfg.Annotations.Include)
    msgr += filter.IncludeMessagesByAnnotation(pf.def, cfg.Annotations.Include)
    ...
}
...
if sr > 0 || mr > 0 || fr > 0 || msgr > 0 {
    orphansRemoved += filter.RemoveOrphanedDefinitions(pf.def, pf.pkg)
}
```

4. Add verbose output for message removals

### Annotation Stripping

The existing `StripAnnotations` call at line 219-221 already strips include annotations from all element types (services, messages, enums, fields) via `SubstituteAnnotations`. No additional changes needed for stripping.

### Test Plan

**Unit tests** (`internal/filter/filter_test.go`):
- `TestIncludeMessagesByAnnotation` — basic: annotated message kept, unannotated removed, enum handled
- `TestIncludeMessagesByAnnotation_NoAnnotations` — all messages removed when none match
- `TestIncludeMessagesByAnnotation_EmptyList` — empty annotation list returns 0

**Golden file tests** (`internal/filter/filter_test.go`):
- `TestGoldenFileMessagesAnnotated` — annotated message + deps kept, unreferenced removed

**CLI integration tests** (`main_test.go`):
- `TestIncludeMessageOnlyCLI` — messages-only file, no matches → empty output
- `TestIncludeAnnotatedMessageCLI` — user's exact example: `[PublishedApi] Parameters` + `Params1` + `Params2`
- `TestIncludeMessageMixedCLI` — mixed services + messages, unreferenced message removed

### No Changes Needed

- `internal/config/` — no config changes, `HasAnnotationInclude()` already works
- `RemoveOrphanedDefinitions` — already handles message field references correctly
- `StripAnnotations` / `SubstituteAnnotations` — already process message comments
- `CollectAnnotationLocations` — already collects from messages
- Exclude-only tests — unaffected (no message include gating in exclude mode)
