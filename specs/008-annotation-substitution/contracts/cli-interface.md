# Contract: CLI Interface — Annotation Substitution

**Feature**: 008-annotation-substitution

## Config YAML Schema

No new CLI flags. All configuration is via the YAML config file (existing `--config` flag).

### New Config Keys

```yaml
# Annotation substitution mapping (optional)
# Key: annotation name (without @ or [] syntax)
# Value: replacement text (empty string = remove annotation)
substitutions:
  HasAnyRole: "Requires authentication"
  Internal: ""

# Strict substitution mode (optional, default: false)
# When true, fail if any annotation in processed files lacks a mapping
strict_substitutions: true
```

### Full Config Example (all features combined)

```yaml
# FQN-based include/exclude (existing)
include:
  - "my.package.*"
exclude:
  - "my.package.internal.*"

# Annotation-based include/exclude (existing)
annotations:
  exclude:
    - "Internal"

# Annotation substitution (new)
substitutions:
  HasAnyRole: "Requires authentication"
  Public: "Available to all users"

# Strict substitution enforcement (new)
strict_substitutions: false
```

## Behavior Contract

### Substitution Processing

1. Substitution runs after annotation filtering and block comment conversion
2. Each annotation token in a comment line is matched and replaced in-place
3. Surrounding text on the same line is preserved
4. Empty substitution values remove the annotation token; if the line becomes empty, it is removed
5. If all comment lines are removed, the comment is removed from the element

### Strict Mode Behavior

1. When `strict_substitutions: true`:
   - The tool scans all annotations on surviving elements across all processed files
   - If any annotation name has no entry in `substitutions`, the tool fails
   - Error message lists all unique unsubstituted annotation names
   - Exit code: 2
   - No output files are written
2. When `strict_substitutions: false` (default):
   - Annotations without mappings are left unchanged in the output

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Input/output error (missing dirs, parse failure) |
| 2 | Configuration error (including strict substitution violations) |

### Backward Compatibility

- Configs without `substitutions` or `strict_substitutions` keys behave identically to before
- Both keys have sensible defaults: empty map and `false` respectively
- No interaction with existing `IsPassThrough()` logic — substitution-only configs still write all files
