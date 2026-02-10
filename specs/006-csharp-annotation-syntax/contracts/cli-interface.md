# CLI Interface Contract: C#-Style Annotation Syntax Support

**Date**: 2026-02-10
**Branch**: `006-csharp-annotation-syntax`

## No CLI Changes

This feature adds no new CLI flags, arguments, or config fields.

The existing `annotations` YAML key already specifies annotation names
without syntax prefix. The same config matches both `@Name` and `[Name]`
styles transparently.

## Config Format (unchanged)

```yaml
# No changes — same format as before
annotations:
  - "HasAnyRole"
  - "Internal"
```

## Command Usage (unchanged)

```bash
proto-filter -input ./protos -output ./filtered -config filter.yaml
```
