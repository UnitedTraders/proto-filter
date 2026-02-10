# CLI Interface Contract: Annotation Include/Exclude Filtering Modes

**Feature**: 007-annotation-include-exclude
**Date**: 2026-02-09

## CLI Flags

No new CLI flags. The existing `--config` flag is used to pass the YAML config file containing the annotation include/exclude configuration.

## Config YAML Contract

### New Structured Format

```yaml
annotations:
  include:            # Optional: list of annotation names to keep
    - "AnnotationName"
  exclude:            # Optional: list of annotation names to remove
    - "AnnotationName"
```

### Old Flat Format (backward compatible)

```yaml
annotations:          # Treated as exclude list
  - "AnnotationName"
```

### Validation Rules

| Condition | Behavior |
|-----------|----------|
| `annotations.include` non-empty AND `annotations.exclude` non-empty | Exit code 2, error on stderr |
| Old flat `annotations` AND new structured `annotations.include`/`annotations.exclude` | Not possible (YAML node is either list or map) |
| Neither include nor exclude | No annotation filtering applied |
| Only `annotations.include` | Keep only matching services/methods |
| Only `annotations.exclude` | Remove matching services/methods |
| Old flat `annotations` list | Same as `annotations.exclude` |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Input/output/parse error |
| 2 | Config error (including mutual exclusivity violation) |

## Verbose Output

When `--verbose` is set and annotation filtering is active, the existing stderr output format is used:

```
proto-filter: removed N services by annotation, N methods by annotation, N orphaned definitions
```

No changes to verbose output format. The filtering direction (include vs exclude) does not change the output message.
