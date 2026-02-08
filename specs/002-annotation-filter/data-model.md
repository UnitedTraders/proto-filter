# Data Model: Annotation-Based Method Filtering

**Date**: 2026-02-08
**Branch**: `002-annotation-filter`

## New Entities

### Annotation

A metadata token extracted from an RPC method's comment.

| Field | Type | Description |
|-------|------|-------------|
| Name | string | Annotation name without `@` prefix (e.g., `HasAnyRole`) |
| SourceLine | string | Original comment line containing the annotation |

**Extraction rules**:
- Matched via regex `@(\w[\w.]*)` against each line of the RPC's `Comment.Lines`
- Multiple annotations per comment block are supported (one per line or multiple per line)
- Arguments in parentheses are ignored for matching purposes
- Only comments attached to `*proto.RPC` elements are scanned

### MethodInfo

Metadata about an individual RPC method within a service, used during annotation filtering.

| Field | Type | Description |
|-------|------|-------------|
| ServiceName | string | Name of the containing service (e.g., `OrderService`) |
| MethodName | string | Name of the RPC method (e.g., `CreateOrder`) |
| Annotations | []string | List of annotation names found in the method's comment |
| RequestType | string | Input message type name |
| ReturnsType | string | Output message type name |

## Extended Entities

### FilterConfig (extended)

The existing `FilterConfig` is extended with an `Annotations` field.

| Field | Type | Description |
|-------|------|-------------|
| Include | []string | *(existing)* Glob patterns for definitions to include |
| Exclude | []string | *(existing)* Glob patterns for definitions to exclude |
| **Annotations** | **[]string** | ***(new)*** List of annotation names; methods carrying these are removed |

**YAML schema** (extended):

```yaml
include:
  - "my.package.OrderService"
exclude:
  - "my.package.internal.*"
annotations:
  - "HasAnyRole"
  - "Internal"
```

**Semantics**:
- If `annotations` is empty or absent, no method-level filtering occurs (backward compatible)
- If `annotations` is non-empty, RPC methods whose comments contain any listed annotation are removed from their services
- Annotation filtering is applied after name-based include/exclude filtering
- The `IsPassThrough()` check must account for the `annotations` field

## Processing Pipeline (updated)

```
Input Directory
    → Discover .proto files
    → Parse each file into ProtoFile
    → Build DependencyGraph from all ProtoFiles
    → Load FilterConfig (if --config provided)
    → Apply name-based filter rules + resolve transitive deps
    → [NEW] Filter RPC methods by annotation within kept services
    → [NEW] Remove services that became empty after method filtering
    → [NEW] Remove orphaned messages/enums no longer referenced
    → [NEW] Skip files with no remaining definitions
    → Generate .proto files via formatter
    → Write to output directory
```

## Orphan Detection

After method filtering, orphan detection determines which messages
and enums are no longer needed.

**Algorithm** (iterative reference counting):

1. Collect all type references from:
   - Remaining RPC methods (request/response types)
   - Remaining message fields (field types)
2. Build a set of "referenced FQNs"
3. Mark any message/enum not in the referenced set as orphaned
4. Remove orphaned definitions from the AST
5. Repeat steps 1-4 until no new orphans are found (handles transitive case)

**External imports** (e.g., `google.protobuf.Timestamp`) are always
considered referenced and never orphaned.
