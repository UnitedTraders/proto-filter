# Data Model: Annotation Substitution

**Feature**: 008-annotation-substitution

## Entities

### SubstitutionMapping (config field)

A key-value mapping from annotation name to replacement description text.

| Attribute | Type | Constraints | Description |
|-----------|------|-------------|-------------|
| Key | string | Non-empty, matches annotation names (e.g., `HasAnyRole`, `Internal`) | The annotation name without `@` or `[]` prefix/suffix |
| Value | string | Can be empty string | The replacement text. Empty string means "remove the annotation entirely" |

**YAML representation**:
```yaml
substitutions:
  HasAnyRole: "Requires authentication"
  Internal: ""
  Public: "Available to all users"
```

### StrictSubstitutions (config field)

A boolean flag controlling whether unsubstituted annotations cause an error.

| Attribute | Type | Default | Description |
|-----------|------|---------|-------------|
| strict_substitutions | bool | false | When true, the tool scans all annotations in surviving comments and fails if any lack a substitution mapping |

### FilterConfig (modified)

Existing config struct extended with two new fields:

| Field | Type | YAML Key | Description |
|-------|------|----------|-------------|
| Include | []string | `include` | FQN glob patterns to include (existing) |
| Exclude | []string | `exclude` | FQN glob patterns to exclude (existing) |
| Annotations | AnnotationConfig | `annotations` | Include/exclude annotations (existing) |
| **Substitutions** | map[string]string | `substitutions` | **New**: annotation name → description mapping |
| **StrictSubstitutions** | bool | `strict_substitutions` | **New**: require all annotations to have mappings |

## Relationships

- `Substitutions` operates independently of `Annotations` (include/exclude). Both can coexist in the same config.
- `StrictSubstitutions` only has meaning when used with `Substitutions` or on its own (to enforce no annotations exist).
- Substitution keys match the same annotation names that `ExtractAnnotations` returns (the name portion only, not `@`/`[]` syntax).

## State Transitions

No state transitions — substitution is a pure transformation applied once per file during output processing.
