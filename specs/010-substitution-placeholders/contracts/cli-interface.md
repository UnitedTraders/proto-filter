# CLI Interface Contract: Substitution Placeholders

**Feature**: 010-substitution-placeholders
**Date**: 2026-02-11

## No CLI Interface Changes

This feature does not modify the CLI interface. All changes are internal to the substitution pipeline.

### Unchanged Interfaces

- **CLI flags**: No new flags. `--input`, `--output`, `--config`, `--verbose` remain unchanged.
- **Config YAML format**: The `substitutions` map format is unchanged. Values may now contain `%s` which is interpreted as a placeholder, but this is backward compatible — existing configs without `%s` behave identically.
- **Exit codes**: Unchanged (0=success, 1=input error, 2=config/strict error).
- **Verbose output**: The `substituted N annotations` count remains the same — placeholder interpolation is counted the same as static substitution.
- **Strict mode**: Unchanged — checks annotation names against substitution map keys, not argument presence (FR-009).
- **Location reporting**: Unchanged — shows full original token including arguments (FR-010).

### Config YAML Example (new capability)

```yaml
substitutions:
  Min: "Minimal value is %s"
  Max: "Maximum value is %s"
  HasAnyRole: "Requires roles: %s"
  Deprecated: "This method is deprecated"  # No %s — static replacement (unchanged behavior)
```

### Behavior with Placeholder

Given `// @Min(3)` and substitution `Min: "Minimal value is %s"`:
- Output comment: `// Minimal value is 3`

Given `// @Min` (no arguments) and substitution `Min: "Minimal value is %s"`:
- Output comment: `// Minimal value is` (empty string replaces `%s`)
