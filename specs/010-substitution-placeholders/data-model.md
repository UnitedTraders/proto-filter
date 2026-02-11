# Data Model: Substitution Placeholders

**Feature**: 010-substitution-placeholders
**Date**: 2026-02-11

## Entities

### AnnotationArgument (conceptual — not a new struct)

The annotation argument is the text content between parentheses in an annotation expression. It is extracted at substitution time from the regex capture groups and interpolated into the replacement value.

| Attribute | Type | Source | Description |
|-----------|------|--------|-------------|
| Content | `string` | Regex capture group 2 or 4 | Text between outermost parentheses, excluding parens themselves |

**Identity**: Derived from the annotation match at substitution time. Not stored independently.

**Lifecycle**: Exists only during `substituteInComment` execution. Extracted from regex match, used for `%s` replacement, then discarded.

## Modified Data Structures

### substitutionRegex (modified)

**Before** (2 capture groups):
```
Group 0: Full match (e.g., @Min(3) or [Tag(x)])
Group 1: @-style annotation name (e.g., Min)
Group 2: [-style annotation name (e.g., Tag)
```

**After** (4 capture groups):
```
Group 0: Full match (e.g., @Min(3) or [Tag(x)])
Group 1: @-style annotation name (e.g., Min)
Group 2: @-style annotation argument (e.g., 3) — empty string if no parens
Group 3: [-style annotation name (e.g., Tag)
Group 4: [-style annotation argument (e.g., x) — empty string if no parens
```

### Substitution map (unchanged)

`map[string]string` — keys are annotation names, values are replacement text that may now contain `%s` placeholder.

## Relationships

```
Config YAML substitutions map
    └── key: annotation name (string)
    └── value: replacement text, optionally containing %s (string)

Annotation in proto comment
    └── name: extracted from regex group 1 or 3
    └── argument: extracted from regex group 2 or 4 (may be empty)

Substitution output = strings.Replace(value, "%s", argument, 1)
```

## Validation Rules

- No validation on argument content — inserted literally (FR-006)
- No validation on `%s` count — only first replaced (FR-007)
- Empty argument (no parens or empty parens) → `%s` replaced with empty string (FR-003, FR-008)
- No `%s` in value → argument ignored (FR-004)
