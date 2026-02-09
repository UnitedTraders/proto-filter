# CLI Interface Contract: Service-Level Annotation Filtering

**Date**: 2026-02-09
**Branch**: `004-service-annotation-filter`

## No CLI Changes

This feature does not add or modify any CLI flags, arguments, or exit
codes. Service-level annotation filtering uses the existing `--config`
flag and the existing `annotations` key in the YAML configuration.

## Existing CLI Interface (unchanged)

```
proto-filter --input <dir> --output <dir> [--config <file>] [--verbose]
```

## Configuration (unchanged)

```yaml
annotations:
  - "Internal"
  - "HasAnyRole"
```

The same annotation list now applies to both service-level and
method-level comments.

## Verbose Output (extended)

When `--verbose` is set, a new counter is included:

```
proto-filter: processed N files, N definitions
proto-filter: included N definitions, excluded N
proto-filter: removed N services by annotation, N methods by annotation, N orphaned definitions
proto-filter: wrote N files to ./out
```

The services-by-annotation count is added before the existing method
count in the verbose annotation line.
