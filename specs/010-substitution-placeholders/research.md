# Research: Substitution Placeholders

**Feature**: 010-substitution-placeholders
**Date**: 2026-02-11

## Research Decision 1: Regex modification strategy

**Decision**: Add capturing groups for annotation arguments to `substitutionRegex` while preserving backward compatibility.

**Current regex** (`filter.go:15`):
```
@(\w[\w.]*)(?:\([^)]*\))?|\[(\w[\w.]*)(?:\([^)]*\))?\]
```

This uses non-capturing groups `(?:\([^)]*\))?` for the parenthesized arguments — meaning the argument text is matched but discarded.

**New regex**:
```
@(\w[\w.]*)(?:\(([^)]*)\))?|\[(\w[\w.]*)(?:\(([^)]*)\))?\]
```

Changes:
- `(?:\([^)]*\))?` → `(?:\(([^)]*)\))?` — convert inner non-capturing group to capturing group
- This adds two new capture groups (group 2 for `@`-style args, group 4 for `[`-style args)
- Capture groups shift: `@`-style name = group 1, `@`-style args = group 2, `[`-style name = group 3, `[`-style args = group 4
- Total groups: 4 (was 2)

**Impact on existing code**: All code using `submatch[1]` and `submatch[2]` for annotation names must update to `submatch[1]` and `submatch[3]`. This affects:
- `substituteInComment` (`filter.go:626-629`)
- `collectLocationsFromComment` (`filter.go:702-705`)

**Rationale**: Minimal change — only the regex pattern and group index references change. No new regex needed.

**Alternatives considered**:
- **Separate regex for argument extraction**: Would require a second regex pass only during substitution. More complex, duplicates pattern logic.
- **Named capture groups**: Go `regexp` supports `(?P<name>...)` syntax but adds verbosity for no real benefit in this small codebase.

## Research Decision 2: Argument extraction during substitution

**Decision**: Extract the annotation argument from the new capture groups in `substituteInComment`'s `ReplaceAllStringFunc` callback and use `strings.Replace` with count=1 to replace the first `%s` in the substitution value.

**Implementation approach**:
```go
// Inside the ReplaceAllStringFunc callback:
submatch := substitutionRegex.FindStringSubmatch(match)
name := submatch[1]
args := submatch[2]  // argument for @-style
if name == "" {
    name = submatch[3]
    args = submatch[4]  // argument for [-style
}
if replacement, ok := substitutions[name]; ok {
    count++
    if strings.Contains(replacement, "%s") {
        return strings.Replace(replacement, "%s", args, 1)
    }
    return replacement
}
```

**Key behaviors**:
- `args` is empty string when annotation has no parentheses (group doesn't match → empty string in Go regex)
- `strings.Replace(replacement, "%s", args, 1)` replaces only the first `%s` (FR-007)
- When `args` is empty and `%s` is present, `%s` is replaced with empty string (FR-003, FR-008)
- When no `%s` in replacement, argument is ignored (FR-004)
- Literal `%s` after the first one is preserved (FR-007)

**Rationale**: `strings.Replace` with count=1 is the simplest, most idiomatic Go approach. No `fmt.Sprintf` needed (which would interpret other `%` verbs).

**Alternatives considered**:
- **`fmt.Sprintf`**: Would interpret `%d`, `%v` etc. in the replacement string — violates FR-006 (literal insertion).
- **`regexp.MustCompile("%s")` replacement**: More complex, no benefit over `strings.Replace`.

## Research Decision 3: Impact on CollectAnnotationLocations

**Decision**: Update `collectLocationsFromComment` to use the new group indices (1→1, 2→3) for annotation name extraction. The Token field (full match, `m[0]`) is unaffected.

**Impact**: The `Token` field already uses `m[0]` which is the full match string — this includes parenthesized arguments and is unaffected by capture group changes. Only the name extraction indices change.

**Rationale**: FR-010 requires location reporting to show the full original annotation token. Since `m[0]` is always the full match, this works correctly with the new regex.

## Research Decision 4: Impact on annotationRegex

**Decision**: `annotationRegex` (`filter.go:14`) is NOT modified. It is used only for annotation filtering (include/exclude by name) and does not need argument capture.

**Current annotationRegex**:
```
@(\w[\w.]*)|\[(\w[\w.]*)(?:\([^)]*\))?\]
```

This regex is separate from `substitutionRegex` and serves a different purpose (filtering, not substitution). No change needed.

## Research Decision 5: Test fixture strategy

**Decision**: Create a new test fixture `testdata/substitution/placeholder_service.proto` with annotations containing arguments (`@Min(3)`, `@Max(100)`, `@HasAnyRole({"ADMIN", "MANAGER"})`, bracket-style `[Tag(important)]`) and a corresponding golden file for verification. Also add targeted unit tests using programmatic AST construction for edge cases (empty args, multiple `%s`, no `%s`).

**Rationale**: Existing `substitution_service.proto` has `@HasAnyRole({"ADMIN", "MANAGER"})` with arguments, but the substitution values in existing tests don't use `%s`. A new dedicated fixture makes test intent clearer and avoids coupling with existing golden files.

**Alternatives considered**:
- **Modify existing fixture**: Would break existing golden file tests (SC-003 violation).
- **Use only programmatic AST tests**: Misses integration-level validation of the full pipeline.
