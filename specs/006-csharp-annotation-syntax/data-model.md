# Data Model: C#-Style Annotation Syntax Support

**Date**: 2026-02-10
**Branch**: `006-csharp-annotation-syntax`

## No New Entities

This feature modifies a single regex pattern and the extraction logic
within `ExtractAnnotations`. No new data types, configuration fields,
or production structures are introduced.

## Changed Components

### annotationRegex (internal/filter/filter.go)

**Before**:
```go
var annotationRegex = regexp.MustCompile(`@(\w[\w.]*)`)
```

**After**:
```go
var annotationRegex = regexp.MustCompile(`@(\w[\w.]*)|\[(\w[\w.]*)(?:\([^)]*\))?\]`)
```

The regex gains a second alternation branch for bracket syntax. Capture
group 1 matches `@Name`, capture group 2 matches `[Name]` or `[Name(value)]`.

### ExtractAnnotations (internal/filter/filter.go)

The function body changes to check both capture groups from each match:
- If group 1 is non-empty → `@Name` annotation found
- If group 2 is non-empty → `[Name]` annotation found

The function signature and return type remain unchanged. All downstream
code (`FilterServicesByAnnotation`, `FilterMethodsByAnnotation`) works
without modification.

## Processing Pipeline (unchanged)

No changes to the processing pipeline. The annotation name extraction
is transparent to all consumers — they receive plain name strings
regardless of which syntax was used in the source proto comments.
