# Feature Specification: Annotation Substitution

**Feature Branch**: `008-annotation-substitution`
**Created**: 2026-02-09
**Status**: Draft
**Input**: User description: "Add substitution feature that allows user to configure annotation -> description mapping. These substitutions should be used when processing comments. Empty description is allowed. User can set an option to check if any annotation is not substituted and stop with descriptive error about all the un-substituted annotations found in the input."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Replace Annotations with Descriptions (Priority: P1)

A developer configuring proto-filter wants to replace annotation markers in proto comments with human-readable descriptions. For example, `@HasAnyRole({"ADMIN", "MANAGER"})` in a comment should be replaced with a readable note like "Requires role: ADMIN or MANAGER", or simply removed if the description is set to empty string. This makes the output proto files suitable for external consumers who should not see internal annotation markers.

**Why this priority**: This is the core value of the feature — without substitution, annotations either remain as-is (confusing to external consumers) or get removed entirely (losing context). Substitution bridges both needs.

**Independent Test**: Configure a substitution mapping `HasAnyRole: "Requires authentication"`, run proto-filter on a file containing `@HasAnyRole({"ADMIN"})` comments, verify the output comment reads `Requires authentication` instead of the annotation.

**Acceptance Scenarios**:

1. **Given** a config with `substitutions: { HasAnyRole: "Requires authentication" }` and a proto file with comment `// @HasAnyRole({"ADMIN", "MANAGER"})`, **When** proto-filter runs, **Then** the output comment reads `// Requires authentication` (the annotation and its parameters are replaced by the description).
2. **Given** a config with `substitutions: { HasAnyRole: "" }` (empty description) and a proto file with comment `// @HasAnyRole`, **When** proto-filter runs, **Then** the annotation line is removed from the comment entirely.
3. **Given** a config with `substitutions: { Internal: "For internal use only" }` and a proto file with comment `// [Internal] Deletes an order.`, **When** proto-filter runs, **Then** the output comment reads `// For internal use only Deletes an order.` (the `[Internal]` token is replaced in-place with the description text, preserving surrounding text on the same line).
4. **Given** a config with `substitutions: { HasAnyRole: "Restricted" }` and a proto file with comment `// @HasAnyRole({"ADMIN"})\n// Creates an order.`, **When** proto-filter runs, **Then** the output comment reads `// Restricted\n// Creates an order.` (multi-line comments preserve non-annotation lines).
5. **Given** a proto file with a comment containing an annotation that has no substitution mapping configured, **When** proto-filter runs without strict mode, **Then** the annotation is left unchanged in the output.

---

### User Story 2 - Remove Annotations via Empty Substitution (Priority: P2)

A developer wants to strip specific annotations from output without leaving any replacement text. By mapping an annotation to an empty string, the annotation (and its parameters) is completely removed from the comment. If removing the annotation leaves the comment line empty, the empty line is cleaned up.

**Why this priority**: This is a specialized but common use case of substitution — stripping internal markers from public-facing proto files without leaving empty comment lines behind.

**Independent Test**: Configure `substitutions: { HasAnyRole: "" }` on a proto file where a method has `// @HasAnyRole({"ADMIN"})\n// Creates an order.`, verify the output contains only `// Creates an order.` with no empty lines.

**Acceptance Scenarios**:

1. **Given** a config with `substitutions: { HasAnyRole: "" }` and a comment `// @HasAnyRole({"ADMIN", "MANAGER"})\n// Creates a new order.`, **When** proto-filter runs, **Then** the output comment is `// Creates a new order.` (annotation line removed, descriptive line preserved).
2. **Given** a config with `substitutions: { HasAnyRole: "" }` and a comment that consists of only `// @HasAnyRole`, **When** proto-filter runs, **Then** the comment is removed entirely (no empty comment left on the element).
3. **Given** a config with `substitutions: { HasAnyRole: "", Internal: "" }` and a comment `// @HasAnyRole\n// [Internal]\n// Some description`, **When** proto-filter runs, **Then** the output comment is `// Some description` (multiple annotations removed, only non-annotation content remains).

---

### User Story 3 - Strict Mode: Detect Unsubstituted Annotations (Priority: P3)

A team lead wants to enforce that every annotation in the proto files has a corresponding substitution mapping. When strict mode is enabled, proto-filter should scan all comments across all processed files, collect any annotations that don't have a substitution entry, and fail with a clear error listing all unsubstituted annotation names. This prevents accidentally leaking internal annotations in output files.

**Why this priority**: This is a safety net for teams who need guarantees that no unmapped annotations slip through. It's complementary to the core substitution feature but not required for basic usage.

**Independent Test**: Enable strict mode with an incomplete substitution map, run proto-filter on a file containing an unmapped annotation, verify the tool exits with an error listing the missing annotation names.

**Acceptance Scenarios**:

1. **Given** a config with `substitutions: { HasAnyRole: "Auth required" }` and `strict_substitutions: true`, and a proto file with both `@HasAnyRole` and `@Internal` annotations, **When** proto-filter runs, **Then** it exits with a non-zero exit code and an error message listing `Internal` as an unsubstituted annotation.
2. **Given** a config with `substitutions: { HasAnyRole: "Auth required", Internal: "" }` and `strict_substitutions: true`, and a proto file with only `@HasAnyRole` and `@Internal` annotations, **When** proto-filter runs, **Then** it succeeds (all annotations have mappings, including empty-string mappings).
3. **Given** a config with `strict_substitutions: true` but no `substitutions` map, and a proto file containing `@HasAnyRole` and `@Public`, **When** proto-filter runs, **Then** it exits with a non-zero exit code and an error message listing both `HasAnyRole` and `Public` as unsubstituted annotations.
4. **Given** a config with `strict_substitutions: true` and complete substitution mappings, and proto files across multiple directories containing various annotations, **When** proto-filter runs, **Then** the error message (if any) lists all unique unsubstituted annotation names found across all files, not just the first one encountered.

---

### Edge Cases

- What happens when a comment line contains multiple annotations (e.g., `// @HasAnyRole @Internal`)? Each annotation is substituted independently according to its mapping. If both map to non-empty descriptions, both descriptions appear. If both map to empty, the line is removed.
- What happens when the substitution description contains special characters (e.g., quotes, newlines)? The description is inserted as-is into the comment text. Newlines in descriptions are not supported (description is a single line of text).
- What happens when an annotation appears in a non-comment context (e.g., inside a string field default)? Only annotations found in proto comment blocks are subject to substitution. Other content is untouched.
- What happens when `substitutions` is configured but `annotations.include` or `annotations.exclude` is also set? Both features operate independently — annotation include/exclude controls which services/methods are kept or removed, while substitutions modify comment text on the elements that remain.
- What happens when `strict_substitutions` is enabled but no annotations exist in the input files? The tool succeeds — there are no unsubstituted annotations to report.
- What happens when the same annotation name appears with different parameter values (e.g., `@HasAnyRole({"ADMIN"})` and `@HasAnyRole({"USER"})`)? Both are substituted with the same description text since substitution is keyed by annotation name only, not by parameters.

## Clarifications

### Session 2026-02-09

- Q: How should substitution behave when an annotation appears inline with other text on the same comment line? → A: Replace the annotation token with description text in-place, preserving surrounding text on the same line.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST support a `substitutions` configuration key that maps annotation names to description strings.
- **FR-002**: System MUST replace annotation tokens (the `@Name(...)` or `[Name(...)]` expression) in proto comments with their configured description text in-place, preserving any surrounding text on the same line.
- **FR-003**: System MUST support empty string as a valid substitution value, causing the annotation (and its parameters) to be removed from the comment.
- **FR-004**: When a substitution removes an annotation and leaves a comment line empty, the system MUST remove the empty line from the comment.
- **FR-005**: When all lines of a comment are removed by substitution, the system MUST remove the comment entirely from the element.
- **FR-006**: Substitution MUST work with both `@Name` and `[Name]` annotation syntaxes, including annotations with parameters (e.g., `@Name(...)` and `[Name(...)]`).
- **FR-007**: System MUST support a `strict_substitutions` boolean configuration option (default: false).
- **FR-008**: When `strict_substitutions` is true, the system MUST scan all annotations in processed proto files and report an error listing every annotation name that lacks a substitution mapping.
- **FR-009**: The strict mode error MUST list all unique unsubstituted annotation names (not just the first one found), collected across all processed files.
- **FR-010**: The strict mode error MUST cause the tool to exit with a non-zero exit code and not write any output files.
- **FR-011**: Annotations without a substitution mapping MUST be left unchanged in the output when strict mode is disabled.
- **FR-012**: Substitution MUST operate on comments of services, RPC methods, messages, enums, and fields — all element types that can carry comments.
- **FR-013**: Substitution processing MUST occur after annotation-based include/exclude filtering, so that only comments on surviving elements are modified.

### Key Entities

- **Substitution Mapping**: A named pair of annotation name (key) and description text (value). The annotation name matches without the `@` or `[]` prefix/suffix. The description is a plain text string that replaces the full annotation expression (name + parameters) in the comment.
- **Strict Mode Flag**: A boolean setting that controls whether unsubstituted annotations cause an error. When enabled, the tool collects all annotation names found in comments across all processed files and compares them against the substitution mapping keys.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All configured annotation substitutions are correctly applied to every occurrence in every processed proto file's comments.
- **SC-002**: Empty substitutions result in clean removal of annotation lines with no leftover empty comment lines or orphaned comment markers.
- **SC-003**: Strict mode detects 100% of unsubstituted annotations across all processed files and reports them in a single error message.
- **SC-004**: Substitution feature does not interfere with existing annotation include/exclude filtering — both features can be used together in the same configuration.
- **SC-005**: Existing configurations without `substitutions` or `strict_substitutions` keys continue to work identically (full backward compatibility).
- **SC-006**: The tool processes substitutions without measurable performance degradation for typical proto file sets (hundreds of files).
