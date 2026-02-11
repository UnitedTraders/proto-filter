# Research: Annotation Error Locations

**Feature**: 009-annotation-error-locations

## Decision 1: Comment Line Number Computation

**Decision**: Use `comment.Position.Line + lineIndex` to compute the source line number for each annotation occurrence within a comment.

**Rationale**: The `emicklei/proto` parser stores a `scanner.Position` on each `Comment` struct. For `//`-style comments (which is what we work with after `ConvertBlockComments`), `Position.Line` gives the line number of the first `//` line, and `Lines[i]` corresponds to `Position.Line + i`. This has been verified by inspecting the parser source.

**Alternatives considered**:
- Re-parsing the source file to find annotation positions: Rejected — unnecessarily complex and slow; the AST already has position info.
- Using element positions (e.g., `rpc.Position.Line`) instead of comment positions: Rejected — element positions point to the element declaration, not the comment lines.

## Decision 2: Annotation Token Extraction for Location Lines

**Decision**: Use `substitutionRegex.FindAllString(line, -1)` to extract the full annotation token as it appears in the source (e.g., `@HasAnyRole({"ADMIN"})`, `[Public]`). Use `substitutionRegex.FindAllStringSubmatch` to also get the annotation name (group 1 or 2) for lookup against the substitution map.

**Rationale**: The `substitutionRegex` already captures the full annotation expression including parameters. Using `FindAllString` gives the matched token directly. Using `FindAllStringSubmatch` gives both the full match (group 0) and the name (group 1 or 2).

**Alternatives considered**:
- Using `annotationRegex` instead: Rejected — `annotationRegex` doesn't capture `@Name(...)` parameters for `@`-style, so the token would be incomplete.
- Showing only the annotation name without parameters: Rejected — the spec requires showing the full token as it appears in source (FR-006).

## Decision 3: Where to Perform Sorting

**Decision**: Sort the location lines in `main.go` before printing, not in the collection function.

**Rationale**: The collection function returns locations per-file. In `main.go`, locations from all files are aggregated, then sorted by file path + line number before printing. This follows the existing pattern where `main.go` handles cross-file concerns.

**Alternatives considered**:
- Sorting in the collection function: Rejected — the function only sees one file at a time, so cross-file sorting isn't possible there.

## Decision 4: Handling of CollectAllAnnotations

**Decision**: Keep `CollectAllAnnotations` in `filter.go` as-is (don't delete it). Add the new `CollectAnnotationLocations` alongside it. Update `main.go` to use the new function.

**Rationale**: `CollectAllAnnotations` is simple and may be useful for other callers. Since it's in an `internal/` package, there's no external compatibility concern, but keeping it costs nothing and avoids unnecessary churn.

**Alternatives considered**:
- Removing `CollectAllAnnotations`: Acceptable but unnecessary; it's 20 lines and may serve other use cases.
- Making `CollectAnnotationLocations` return both the location slice and a `map[string]bool`: Rejected — over-engineering; the caller can trivially derive the set from the slice.

## Decision 5: Line Number for Block Comments After Conversion

**Decision**: Collect annotation locations *after* `ConvertBlockComments` runs, using the comment's `Position.Line` which was set during parsing and is preserved through conversion. The conversion changes `Cstyle` to `false` and cleans up lines but does not modify `Position`.

**Rationale**: The existing pipeline runs `ConvertBlockComments` before annotation collection. After conversion, comments are in `//` style with `Position` preserved from the original parse. Each `Lines[i]` maps to `Position.Line + i` because the conversion preserves line-to-line correspondence (it strips `*` prefixes and trims leading/trailing empty lines from block comments, but the `Position.Line` still points to the original `/*` line). For `//` comments, `Position.Line` was already set to the first `//` line.

**Note**: For block comments that had leading empty lines trimmed during conversion, the computed line number may be off by the number of trimmed lines. However, the existing test fixtures and typical proto files don't use block comments with leading empty lines for annotations. This is acceptable given that block comment conversion happens before annotation collection — the position is as accurate as the parser provides.

**Alternatives considered**:
- Collecting locations before `ConvertBlockComments`: Would require handling both C-style and single-line comment formats. Rejected — adds complexity for marginal accuracy gain on an uncommon edge case.
