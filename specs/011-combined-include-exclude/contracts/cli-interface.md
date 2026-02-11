# CLI Interface: Combined Include and Exclude Annotation Filters

## No CLI Changes

This feature does not modify the CLI interface. No new flags, arguments, or exit codes are introduced.

### Existing Behavior Preserved

- `--config` flag accepts YAML config with `annotations.include` and/or `annotations.exclude`
- Exit code 0 on success
- Exit code 1 on input errors
- Exit code 2 on config validation errors (reduced — mutual exclusivity no longer triggers this)
- `--verbose` flag now additionally reports field-level removals

### Config Format (unchanged structure, relaxed constraint)

```yaml
# All three combinations now valid:

# 1. Include-only (unchanged)
annotations:
  include:
    - "PublicApi"

# 2. Exclude-only (unchanged)
annotations:
  exclude:
    - "Internal"

# 3. Combined (NEW — previously rejected)
annotations:
  include:
    - "PublicApi"
  exclude:
    - "Deprecated"

# 4. Old flat format (unchanged — treated as exclude)
annotations:
  - "Internal"
```

### Verbose Output Change

**Before**: `proto-filter: removed N services by annotation, M methods by annotation, O orphaned definitions`

**After**: `proto-filter: removed N services by annotation, M methods by annotation, F fields by annotation, O orphaned definitions`
