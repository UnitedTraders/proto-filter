# Data Model: Annotation Include/Exclude Filtering Modes

**Feature**: 007-annotation-include-exclude
**Date**: 2026-02-09

## Entities

### AnnotationConfig (new)

Replaces the flat `Annotations []string` field in `FilterConfig`.

| Field | Type | Description |
|-------|------|-------------|
| Include | list of strings | Annotation names to include (keep matching). Empty = not set. |
| Exclude | list of strings | Annotation names to exclude (remove matching). Empty = not set. |

**Constraints**:
- At most one of `Include` or `Exclude` may be non-empty at a time.
- Both empty = no annotation filtering.

**YAML representations** (all valid):

```yaml
# New structured format (include mode):
annotations:
  include:
    - "Public"

# New structured format (exclude mode):
annotations:
  exclude:
    - "HasAnyRole"

# Old flat format (backward compat, treated as exclude):
annotations:
  - "HasAnyRole"
```

### FilterConfig (modified)

Current structure:

| Field | Type | YAML Key |
|-------|------|----------|
| Include | list of strings | `include` |
| Exclude | list of strings | `exclude` |
| Annotations | list of strings | `annotations` |

New structure:

| Field | Type | YAML Key |
|-------|------|----------|
| Include | list of strings | `include` |
| Exclude | list of strings | `exclude` |
| Annotations | AnnotationConfig | `annotations` |

The `Annotations` field changes from `[]string` to `AnnotationConfig`. Custom YAML unmarshaling handles both old and new formats transparently.

## Validation Rules

1. `AnnotationConfig.Include` and `AnnotationConfig.Exclude` cannot both be non-empty → error: "annotations.include and annotations.exclude are mutually exclusive"
2. Old flat `annotations: [...]` format is deserialized into `AnnotationConfig.Exclude` — no coexistence with structured format possible (the YAML node is either a list or a map, never both).

## State Transitions

No state machines. Config is immutable after loading. Pipeline is sequential: load config → validate → filter.

## Relationships

- `FilterConfig` → contains one `AnnotationConfig`
- `AnnotationConfig` → used by `main.go` to determine which filter functions to call
- `IncludeServicesByAnnotation` / `IncludeMethodsByAnnotation` (new) → inverse of existing `Filter*ByAnnotation` functions
