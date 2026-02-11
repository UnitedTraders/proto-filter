# Data Model: Annotation Error Locations

**Feature**: 009-annotation-error-locations

## Entities

### AnnotationLocation

Represents a single annotation occurrence found in a proto source file.

| Field | Type   | Description |
|-------|--------|-------------|
| File  | string | Relative file path (e.g., `orders.proto`, `sub/payments.proto`) |
| Line  | int    | 1-based line number in the original source file |
| Name  | string | Annotation name without syntax markers (e.g., `HasAnyRole`, `Deprecated`) — used for substitution map lookup |
| Token | string | Full annotation token as it appears in the source comment (e.g., `@HasAnyRole({"ADMIN"})`, `[Public]`) |

### Relationships

- **AnnotationLocation → FilterConfig.Substitutions**: The `Name` field is used as the lookup key against `cfg.Substitutions` to determine whether the annotation has a mapping. Locations whose `Name` has no entry in the map are considered "unsubstituted" and are reported in the error output.

### Lifecycle

1. **Created** during Pass 1 of the pipeline, after annotation filtering and block comment conversion, by `CollectAnnotationLocations()`.
2. **Aggregated** across all processed files into a single slice in `main.go`.
3. **Filtered** during the strict check: only locations whose `Name` is not in `cfg.Substitutions` are reported.
4. **Sorted** by `File` (alphabetically), then `Line` (ascending) before printing.
5. **Consumed** by the error formatter which prints each location as `  File:Line: Token`.
