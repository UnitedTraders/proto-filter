# CLI Interface Contract: Annotation Error Locations

**Feature**: 009-annotation-error-locations

## Strict Substitution Error Output

### Current Format (preserved)

```
proto-filter: error: unsubstituted annotations found: <Name1>, <Name2>, ...
```

- Printed to **stderr**
- Exit code: **2**
- Annotation names are sorted alphabetically, comma-separated

### Enhanced Format (new)

```
proto-filter: error: unsubstituted annotations found: <Name1>, <Name2>, ...
  <file-path>:<line-number>: <annotation-token>
  <file-path>:<line-number>: <annotation-token>
  ...
```

#### Summary Line

| Field | Format | Description |
|-------|--------|-------------|
| Prefix | `proto-filter: error: unsubstituted annotations found: ` | Fixed prefix (unchanged) |
| Names | `Name1, Name2` | Unique unsubstituted annotation names, sorted alphabetically, comma-separated (unchanged) |

#### Location Lines

Each location line is indented with **two spaces** and follows the summary line.

| Field | Format | Description |
|-------|--------|-------------|
| Indent | `  ` (two spaces) | Leading indentation |
| File path | `<relative-path>` | Relative to input directory (e.g., `orders.proto`, `sub/payments.proto`) |
| Separator | `:` | Between file path and line number |
| Line number | `<int>` | 1-based line number in the original source file |
| Separator | `: ` | Between line number and annotation token (colon + space) |
| Annotation token | `<token>` | Full annotation expression as it appears in source (e.g., `@Deprecated`, `@HasAnyRole({"ADMIN"})`, `[Public]`) |

#### Ordering

Location lines are ordered by:
1. File path (alphabetical, ascending)
2. Line number (numeric, ascending) within each file

#### Examples

Single file, single annotation:
```
proto-filter: error: unsubstituted annotations found: Deprecated
  orders.proto:4: @Deprecated
```

Multiple files, multiple annotations:
```
proto-filter: error: unsubstituted annotations found: Deprecated, Internal
  orders.proto:4: @Deprecated
  payments.proto:5: @Internal
  payments.proto:11: @Deprecated
```

### Unchanged Behaviors

| Aspect | Behavior |
|--------|----------|
| Exit code | 2 (unchanged) |
| Output stream | stderr (unchanged) |
| Success case | No output, exit code 0 (unchanged) |
| Non-strict mode | No strict check performed (unchanged) |
| Summary line format | Preserved exactly (backward compatible) |
