# Data Model: Service-Level Annotation Filtering

**Date**: 2026-02-09
**Branch**: `004-service-annotation-filter`

## No New Entities

This feature does not introduce new data types or configuration fields.
It extends the behavioral scope of existing entities.

## Existing Entities (behavior extended)

### FilterConfig (unchanged)

The existing `FilterConfig.Annotations` field already carries the list
of annotation names for filtering. Service-level filtering uses the
same annotation list — no separate config key is needed.

```yaml
annotations:
  - "Internal"
  - "HasAnyRole"
```

When `HasAnnotations()` returns true, both service-level and method-level
annotation filtering are applied.

### Annotation Extraction (scope extended)

The existing `ExtractAnnotations(comment *proto.Comment)` function is
now applied to two scopes:

| Scope | AST Field | Existing | New |
|-------|-----------|----------|-----|
| RPC method | `proto.RPC.Comment` | Yes (002-annotation-filter) | Unchanged |
| Service | `proto.Service.Comment` | No | **Yes** |

Matching semantics remain identical: regex `@(\w[\w.]*)` extracts
annotation names; any match against the configured list triggers removal.

## Processing Pipeline (updated)

```
Input Directory
    → Discover .proto files
    → Parse each file into ProtoFile
    → Build DependencyGraph from all ProtoFiles
    → Load FilterConfig (if --config provided)
    → Apply name-based filter rules + resolve transitive deps
    → [NEW] Filter entire services by annotation (service-level comments)
    → Filter RPC methods by annotation within remaining services
    → Remove services that became empty after method filtering
    → Remove orphaned messages/enums no longer referenced
    → Skip files with no remaining definitions
    → Convert block comments to single-line style
    → Generate .proto files via formatter
    → Write to output directory
```

The new step runs **before** method-level filtering. This means:
- A service removed at the service level will not have its methods
  checked individually (FR-004 is satisfied — method-level filtering
  only applies to non-removed services)
- Orphan detection runs after both service and method removal, cleaning
  up all unreferenced types in one pass
