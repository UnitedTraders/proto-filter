# Feature Specification: Substitution Placeholders

**Feature Branch**: `010-substitution-placeholders`
**Created**: 2026-02-11
**Status**: Draft
**Input**: User description: "Annotation substitution argument can be used as placeholder in substitution value. For example, substitution mapping 'Min: \"Minimal value is %s\"' for source value \"Min(3)\" becomes \"Minimal value is 3\"."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Interpolate Annotation Arguments into Substitution Text (Priority: P1)

A user configures a substitution mapping with a `%s` placeholder in the replacement value. When proto-filter encounters an annotation that has arguments (e.g., `@Min(3)`, `@SupportWindow({duration: "6M"})`), the tool extracts the argument content from between the parentheses and inserts it into the replacement text at the `%s` position. This lets users produce human-readable documentation from structured annotation metadata — for example, turning `@Min(3)` into `Minimal value is 3` or `@SupportWindow({duration: "6M"})` into `Supported for {duration: "6M"}`.

**Why this priority**: This is the entire scope of the feature. Without placeholder support, annotations with arguments lose their parameter values during substitution; users must either keep raw annotations or write static replacement text that ignores the dynamic argument.

**Independent Test**: Configure a substitution mapping with `%s` placeholder, run proto-filter on a proto file containing annotations with arguments, and verify the output comments contain the replacement text with the argument value interpolated.

**Acceptance Scenarios**:

1. **Given** a config with `Min: "Minimal value is %s"` and a proto file containing `// @Min(3)`, **When** the user runs proto-filter, **Then** the output comment contains `Minimal value is 3`.
2. **Given** a config with `SupportWindow: "Supported for %s"` and a proto file containing `// @SupportWindow({duration: "6M"})`, **When** the user runs proto-filter, **Then** the output comment contains `Supported for {duration: "6M"}`.
3. **Given** a config with `Max: "Maximum value is %s"` and a proto file containing `// @Max(100)`, **When** the user runs proto-filter, **Then** the output comment contains `Maximum value is 100`.
4. **Given** a config with `HasAnyRole: "Requires roles: %s"` and a proto file containing `// @HasAnyRole({"ADMIN", "MANAGER"})`, **When** the user runs proto-filter, **Then** the output comment contains `Requires roles: {"ADMIN", "MANAGER"}`.
5. **Given** a config with `Deprecated: "This method is deprecated"` (no `%s` placeholder) and a proto file containing `// @Deprecated`, **When** the user runs proto-filter, **Then** the output comment contains `This method is deprecated` (unchanged behavior — no interpolation needed).

---

### User Story 2 - Graceful Handling of Missing Arguments (Priority: P2)

When a substitution value contains a `%s` placeholder but the annotation in the source has no arguments (e.g., `@Min` without parentheses), the tool must handle this gracefully. The `%s` placeholder is replaced with an empty string so the user still gets readable output rather than a literal `%s` in the result.

**Why this priority**: This is a common edge case — users may define a substitution template with a placeholder for annotations that sometimes appear with arguments and sometimes without. Graceful degradation ensures the output is always usable.

**Independent Test**: Configure a substitution mapping with `%s` placeholder, run proto-filter on a proto file where the annotation appears without arguments, and verify the `%s` is replaced with an empty string.

**Acceptance Scenarios**:

1. **Given** a config with `Min: "Minimal value is %s"` and a proto file containing `// @Min` (no arguments), **When** the user runs proto-filter, **Then** the output comment contains `Minimal value is` (with `%s` replaced by empty string).
2. **Given** a config with `Tag: "Tagged: %s"` and a proto file containing `// [Tag]` (bracket style, no arguments), **When** the user runs proto-filter, **Then** the output comment contains `Tagged:` (with `%s` replaced by empty string).

---

### Edge Cases

- What happens when the substitution value contains multiple `%s` placeholders? Only the first `%s` is replaced with the annotation argument; subsequent `%s` markers are left as literal text.
- What happens when the annotation argument contains special characters like `%` or `\`? The argument is inserted literally, with no further interpretation of escape sequences.
- What happens when the annotation argument is empty parentheses `@Min()`? The `%s` is replaced with an empty string.
- What happens when the substitution value has no `%s` placeholder but the annotation has arguments? The arguments are ignored — the entire annotation token (including arguments) is replaced with the static substitution text. This is unchanged behavior.
- What happens with bracket-style annotations with arguments, e.g., `[Min(3)]`? The same interpolation applies — `%s` is replaced with `3`.
- What happens when the argument contains nested parentheses? The argument is everything between the outermost parentheses, as already matched by the existing annotation regex.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: When a substitution value contains `%s` and the matched annotation has arguments (text between parentheses), the tool MUST replace `%s` in the substitution value with the argument content.
- **FR-002**: The argument content MUST be the text between the outermost parentheses of the annotation, excluding the parentheses themselves (e.g., `@Min(3)` yields argument `3`, `@HasAnyRole({"ADMIN"})` yields argument `{"ADMIN"}`).
- **FR-003**: When a substitution value contains `%s` but the annotation has no arguments, the `%s` MUST be replaced with an empty string.
- **FR-004**: When a substitution value does NOT contain `%s`, the behavior MUST remain unchanged — the annotation token is replaced with the static substitution text regardless of whether the annotation has arguments.
- **FR-005**: Placeholder interpolation MUST work for both `@Name(args)` and `[Name(args)]` annotation styles.
- **FR-006**: The argument content MUST be inserted literally into the substitution value — no escape processing or format-string interpretation beyond the single `%s` replacement.
- **FR-007**: Only the first `%s` occurrence in the substitution value MUST be replaced. Any additional `%s` markers in the same value are left as literal text.
- **FR-008**: Empty argument parentheses (e.g., `@Min()`) MUST be treated as an empty string argument — `%s` is replaced with an empty string.
- **FR-009**: The strict substitution check (`strict_substitutions: true`) MUST continue to work unchanged — it checks annotation names against the substitution map keys, not argument presence.
- **FR-010**: The annotation location reporting (feature 009) MUST continue to show the full original annotation token including arguments, unaffected by placeholder substitution.

### Key Entities

- **AnnotationArgument**: The text content between parentheses in an annotation expression (e.g., `3` from `@Min(3)`, `{"ADMIN", "MANAGER"}` from `@HasAnyRole({"ADMIN", "MANAGER"})`). Extracted during substitution and interpolated into the replacement value at the `%s` position.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Every annotation with arguments and a `%s`-containing substitution value produces output with the argument correctly interpolated — zero instances of literal `%s` remaining in output when the annotation had arguments.
- **SC-002**: Annotations without arguments paired with `%s`-containing substitution values produce output with `%s` replaced by an empty string.
- **SC-003**: All existing substitution tests pass without modification (backward compatibility).
- **SC-004**: The feature works identically for both `@Name(args)` and `[Name(args)]` annotation styles.

## Assumptions

- The `%s` placeholder follows printf-style convention as a familiar pattern. Only `%s` is supported — no other format verbs like `%d`, `%v`, etc.
- Only the first `%s` in a substitution value is replaced. This is the simplest, most predictable behavior and avoids ambiguity when a substitution value intentionally contains `%s` as literal text after the first placeholder.
- The argument content is taken verbatim from between the outermost parentheses as already captured by the existing annotation regex. No trimming of inner whitespace is performed.
- The existing regex already matches the parenthesized portion of annotations; the change is to capture it as a group rather than discarding it.
- Placeholder interpolation is purely a string replacement within the substitution value — it does not affect annotation filtering, strict mode checking, or location reporting.
