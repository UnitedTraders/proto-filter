# Research: Annotation Include/Exclude Filtering Modes

**Feature**: 007-annotation-include-exclude
**Date**: 2026-02-09

## Decision 1: Config Structure for Annotation Include/Exclude

**Decision**: Use a nested YAML struct with `annotations.include` and `annotations.exclude` sub-keys, while keeping the old flat `annotations` key for backward compatibility.

**Rationale**: The current `FilterConfig` has `Annotations []string` for the flat list. The new config needs to support both include and exclude modes. A nested struct maps cleanly to YAML:

```yaml
# New structured format:
annotations:
  include:
    - "Public"
  # OR
  exclude:
    - "HasAnyRole"

# Old flat format (still works as exclude):
annotations:
  - "HasAnyRole"
```

Go's `gopkg.in/yaml.v3` library supports custom `UnmarshalYAML` for polymorphic parsing — the `Annotations` field can detect whether it receives a list (old format) or a map (new format).

**Alternatives considered**:
- Separate top-level keys (`annotation_include`, `annotation_exclude`): Rejected — clutters config namespace and breaks grouping.
- Single `annotations` key with a `mode` field: Rejected — more complex YAML structure for users.

## Decision 2: YAML Unmarshaling Strategy

**Decision**: Replace the `Annotations []string` field with an `AnnotationConfig` struct that implements `UnmarshalYAML`. The custom unmarshaler detects the YAML node type:
- If it's a sequence (list), treat as the old flat `annotations: [...]` format → populate `Exclude` field.
- If it's a mapping (object), parse as `{include: [...], exclude: [...]}`.

**Rationale**: This is the standard Go/YAML approach for backward-compatible config evolution. The `yaml.v3` library exposes `yaml.Node` in `UnmarshalYAML`, allowing inspection of node kind before parsing. No reflection hacks or multiple parse attempts needed.

**Alternatives considered**:
- Two-pass parsing (try list first, fall back to struct): Rejected — fragile and harder to maintain.
- Keep `Annotations []string` and add separate `AnnotationInclude`/`AnnotationExclude` fields: Rejected — the old flat list and new structured list would both be populated, requiring complex precedence logic.

## Decision 3: Include Mode Implementation Approach

**Decision**: Add two new functions `IncludeServicesByAnnotation` and `IncludeMethodsByAnnotation` in `internal/filter/filter.go` that are the inverse of the existing `FilterServicesByAnnotation`/`FilterMethodsByAnnotation`. For include mode, services/methods WITHOUT a matching annotation are removed.

**Rationale**: The existing `Filter*ByAnnotation` functions have clear "remove matching" semantics. Adding a separate pair of "keep matching" functions maintains clear naming and avoids confusing boolean parameters. The caller in `main.go` selects which pair to call based on the config mode.

**Alternatives considered**:
- Add a `mode` parameter to existing functions: Rejected — makes the API harder to read and test; violates single responsibility.
- Invert the annotation set logic internally: Rejected — same function doing opposite things based on a flag is error-prone.

## Decision 4: Validation and Error Handling

**Decision**: Add a `Validate() error` method to `FilterConfig` that checks:
1. `AnnotationConfig.Include` and `AnnotationConfig.Exclude` are not both non-empty.
2. The old flat `Annotations` field and new structured `AnnotationConfig` are not both populated (when old flat format is detected, it's already migrated into the struct, so this is enforced by the unmarshaler).

Call `Validate()` in `main.go` immediately after `LoadConfig()`.

**Rationale**: Validation at config load time gives users immediate, clear errors before any processing begins. The error message explicitly states "annotations.include and annotations.exclude are mutually exclusive".

**Alternatives considered**:
- Validate at filter time: Rejected — delays error reporting; user already waited for file discovery/parsing.
- Validate in `LoadConfig`: Rejected — mixes parsing with business logic; harder to test independently.

## Decision 5: Impact on Existing Tests

**Decision**: Existing tests that use `cfg.Annotations` will be updated to use `cfg.AnnotationConfig.Exclude` after the migration. Tests using the old flat YAML format (`annotations: [...]`) will continue to work via the custom unmarshaler.

**Rationale**: The internal code change (struct field rename) requires updating direct struct literals in tests, but YAML-based tests remain unchanged. This aligns with SC-005 (all existing tests pass without modification) at the YAML level, while acknowledging that Go struct literals need updating.

**Alternatives considered**:
- Keep both `Annotations []string` and `AnnotationConfig`: Rejected — dual representation is a maintenance burden and source of bugs.
