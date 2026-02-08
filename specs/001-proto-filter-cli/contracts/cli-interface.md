# CLI Interface Contract: proto-filter

**Date**: 2026-02-08

## Command

```
proto-filter --input <dir> --output <dir> [--config <file>] [--verbose]
```

## Flags

| Flag | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `--input` | string | Yes | — | Path to directory containing source `.proto` files |
| `--output` | string | Yes | — | Path to directory where filtered `.proto` files are written |
| `--config` | string | No | — | Path to YAML filter configuration file |
| `--verbose` | bool | No | false | Print processing summary to stderr |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Runtime error (invalid input, parse failure, I/O error) |
| 2 | Configuration error (invalid YAML, conflicting rules) |

## Output Behavior

- **stdout**: Reserved for future use (no output in current version)
- **stderr**: Error messages, warnings, and verbose summary
- **filesystem**: Generated `.proto` files in output directory

## Verbose Output Format (stderr)

```
proto-filter: processed 12 files, 45 definitions
proto-filter: included 8 definitions, excluded 37
proto-filter: wrote 5 files to ./output
```

## Error Message Format (stderr)

```
proto-filter: error: <message>
proto-filter: warning: <message>
```

## Configuration File Schema

```yaml
# filter.yaml
include:
  - "my.package.OrderService"
  - "my.package.common.*"
exclude:
  - "my.package.internal.*"
```

### Validation Rules

- File MUST be valid YAML
- `include` and `exclude` are optional lists of strings
- Each string is a glob pattern matched against fully qualified names
- A definition matching both an include and exclude pattern is an error
- Empty file (no include, no exclude) means pass-through (no filtering)
