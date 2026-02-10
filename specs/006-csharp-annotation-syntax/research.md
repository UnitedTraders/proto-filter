# Research: C#-Style Annotation Syntax Support

**Date**: 2026-02-10
**Branch**: `006-csharp-annotation-syntax`

## Decision 1: Regex Design for Bracket Annotation Syntax

**Decision**: Extend `annotationRegex` to a combined pattern that matches
both `@Name` and `[Name]`/`[Name(value)]` syntax in a single pass.

The new regex pattern:
```
@(\w[\w.]*)|(?:^|[^\\])\[(\w[\w.]*)(?:\([^)]*\))?\]
```

Simplified to two separate regexes composed together, or more practically:

```go
var annotationRegex = regexp.MustCompile(`@(\w[\w.]*)|\[(\w[\w.]*)(?:\([^)]*\))?\]`)
```

This captures:
- `@HasAnyRole` → group 1: `HasAnyRole`
- `@HasAnyRole({"ADMIN"})` → group 1: `HasAnyRole` (parens not part of `@` regex)
- `[HasAnyRole]` → group 2: `HasAnyRole`
- `[HasAnyRole("ADMIN")]` → group 2: `HasAnyRole`
- `[com.example.Secure]` → group 2: `com.example.Secure`
- `[]` → no match (name required: `\w[\w.]*` needs at least one char)
- `[ Name ]` → no match (space before Name prevents `\w` from matching immediately after `[`)
- `[RFC 7231]` → no match (space in "RFC 7231" breaks the pattern)
- `[error code]` → no match (space breaks the pattern)

The `ExtractAnnotations` function takes whichever capture group is non-empty.

**Rationale**: A single combined regex is simpler than running two
separate regexes per line. The alternation `|` cleanly separates the
two syntax families. The `\w[\w.]*` naming rule is shared between both
branches, ensuring FR-002 compliance. The optional `(?:\([^)]*\))?`
in the bracket branch handles `[Name(value)]` — it matches parenthesized
content within the brackets and ignores it (FR-003).

Note: for `@Name(value)`, the existing regex `@(\w[\w.]*)` already
ignores parenthesized arguments because `(` is not a word character,
so the capture stops at the name. The bracket syntax needs explicit
optional group `(?:\([^)]*\))?` because without it, `[Name(value)]`
would fail to match — the `]` would need to come right after the name.

**Alternatives considered**:

| Approach | Verdict | Why rejected |
|----------|---------|--------------|
| Two separate regexes, run sequentially | Viable but unnecessary | One combined regex is simpler and faster; both patterns share the same naming rules |
| Preprocess: convert `[Name]` to `@Name` before regex | Fragile | Would require careful bracket parsing to avoid converting non-annotation brackets |
| Full parser with state machine | Overkill | Regex is sufficient for the defined syntax; no nesting needs to be handled |

## Decision 2: ExtractAnnotations Function Changes

**Decision**: Modify `ExtractAnnotations` to use the updated regex and
check both capture groups (group 1 for `@Name`, group 2 for `[Name]`).
No changes to the function signature.

**Rationale**: `ExtractAnnotations` is the single point where annotation
names are extracted from comment lines. All downstream consumers
(`FilterServicesByAnnotation`, `FilterMethodsByAnnotation`) work with
annotation name strings and are completely agnostic to syntax. This
means only the regex and the extraction logic inside `ExtractAnnotations`
need to change — no other functions are affected.

The change is minimal:
```go
for _, m := range matches {
    if m[1] != "" {
        annotations = append(annotations, m[1])  // @Name match
    } else if m[2] != "" {
        annotations = append(annotations, m[2])  // [Name] match
    }
}
```

**Alternatives considered**:

| Approach | Verdict | Why rejected |
|----------|---------|--------------|
| Create a separate `ExtractBracketAnnotations` function | Rejected | Would require changing all callers to call both functions and merge results; unnecessary complexity |
| Return annotation with syntax metadata (struct) | Rejected | Downstream only needs the name; adding metadata violates YAGNI |

## Decision 3: No Config Changes

**Decision**: No changes to the configuration format. The `annotations`
YAML key continues to accept annotation names without syntax prefix.
The same name matches both `@Name` and `[Name]` styles.

**Rationale**: Per FR-005, the config format must remain unchanged. Users
specify annotation names (e.g., `HasAnyRole`, `Internal`), and the tool
matches regardless of which syntax was used in the proto comments. This
is the natural behavior since `ExtractAnnotations` already strips the
syntax prefix when extracting names.

## Decision 4: Test Strategy

**Decision**: Add new test cases to the existing `TestExtractAnnotations`
table-driven test and create new test fixtures with bracket-style
annotations. Reuse existing golden file patterns.

**Rationale**: The feature is a backward-compatible extension. Existing
tests must continue to pass (SC-003). New tests cover:
1. Bracket annotation extraction (unit tests in `filter_test.go`)
2. Mixed `@Name` + `[Name]` in the same comment
3. False positive rejection (`[RFC 7231]`, `[]`, `[ Name ]`)
4. Integration test with bracket-annotated proto fixture
5. Mixed-style filtering (both syntaxes in same file)

No new test fixture directories needed — extend existing `testdata/annotations/`
with a new proto file using bracket syntax.

**Alternatives considered**:

| Approach | Verdict | Why rejected |
|----------|---------|--------------|
| Separate test directory for bracket tests | Rejected | Adds unnecessary structure; bracket tests are annotation tests and belong with them |
| Only unit tests, no integration | Rejected | Integration test ensures the full pipeline works with the new syntax |
