# CLI Interface Contract: Annotation Filtering

**Date**: 2026-02-08
**Branch**: `002-annotation-filter`

## No CLI Flag Changes

Annotation filtering is configured entirely through the YAML config
file (`--config`). No new CLI flags are added.

## Extended YAML Config Schema

```yaml
# Existing keys (unchanged)
include:
  - "my.package.OrderService"     # glob: keep this service
  - "my.package.common.*"         # glob: keep all in package
exclude:
  - "my.package.internal.*"       # glob: remove from included set

# New key
annotations:                       # list of annotation names
  - "HasAnyRole"                  # methods with @HasAnyRole are removed
  - "Internal"                    # methods with @Internal are removed
```

### Annotation Name Format

- Names are specified **without** the `@` prefix
- Names are matched **case-sensitively**
- Names may contain dots (e.g., `com.example.Secure`)
- Arguments in source (e.g., `({"ADMIN"})`) are ignored during matching

## Extended Verbose Output

When `--verbose` is enabled and annotations are configured:

```
proto-filter: processed 12 files, 45 definitions
proto-filter: included 8 definitions, excluded 37
proto-filter: removed 5 methods by annotation, 3 orphaned definitions
proto-filter: wrote 5 files to ./out
```

New line format: `proto-filter: removed N methods by annotation, M orphaned definitions`

## Exit Codes (unchanged)

| Code | Meaning |
|------|---------|
| 0    | Success |
| 1    | Runtime error (missing directory, parse failure, I/O error) |
| 2    | Configuration error (invalid YAML, conflicting filter rules) |

## Behavioral Rules

1. If `annotations` is empty or absent: no method-level filtering (backward compatible)
2. If `annotations` is present: methods with matching annotations are removed from output
3. If all methods in a service are removed: the service itself is removed
4. If a message/enum is only referenced by removed methods: it is removed (orphaned)
5. If a file has no remaining definitions after filtering: the file is not written
6. Annotation filtering is applied **after** name-based include/exclude filtering
