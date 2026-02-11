# Feature Specification: Annotation Error Locations

**Feature Branch**: `009-annotation-error-locations`
**Created**: 2026-02-11
**Status**: Draft
**Input**: User description: "Current reporting on unsubstituted annotations look like 'proto-filter: error: unsubstituted annotations found: Deprecated, Max, Min, SupportWindow'. Please add capability to show exact location of annotation in the source file."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - See Source Locations for Unsubstituted Annotations (Priority: P1)

A user runs proto-filter with `strict_substitutions: true` and an incomplete substitution mapping. The tool currently reports only a flat list of missing annotation names (e.g., `proto-filter: error: unsubstituted annotations found: Deprecated, Max, Min, SupportWindow`). The user must then manually search across all proto files to find where each annotation appears. Instead, the error output should list the exact file and line number for each unsubstituted annotation occurrence, so the user can immediately navigate to the source and decide whether to add a mapping or fix the proto file.

**Why this priority**: This is the core value of the feature. Without source locations, strict mode errors in large codebases require tedious manual searching. With locations, the user can jump directly to each occurrence.

**Independent Test**: Run proto-filter with `strict_substitutions: true` and an incomplete mapping against proto files with known annotation locations. Verify the error output includes file paths and line numbers for each unsubstituted annotation.

**Acceptance Scenarios**:

1. **Given** a proto file `orders.proto` with `@Deprecated` on line 7 and `@SupportWindow` on line 12, and a config with `strict_substitutions: true` and no mapping for these annotations, **When** the user runs proto-filter, **Then** stderr includes per-occurrence location lines like `  orders.proto:7: @Deprecated` and `  orders.proto:12: @SupportWindow`, and the tool exits with code 2.
2. **Given** multiple proto files each containing unsubstituted annotations, **When** the user runs proto-filter with strict mode, **Then** all occurrences across all files are listed in the error output, ordered by file path then line number.
3. **Given** a proto file where the same annotation name appears on multiple methods, **When** the user runs proto-filter with strict mode, **Then** each occurrence is listed separately with its own line number.
4. **Given** a config with `strict_substitutions: true` and a complete mapping for all annotations, **When** the user runs proto-filter, **Then** the tool succeeds with exit code 0 (no change to success behavior).

---

### User Story 2 - Preserve Summary Line for Backward Compatibility (Priority: P2)

In addition to detailed per-occurrence locations, the error output retains the existing summary line listing the unique annotation names that are missing mappings. This preserves the current behavior for scripts or users who grep for specific annotation names, while adding detailed location lines below.

**Why this priority**: Maintaining backward compatibility of the summary line ensures existing CI pipelines and scripts that parse the error output continue to work. The summary is a quick overview; the detailed locations below serve humans navigating to source.

**Independent Test**: Run strict mode with missing annotations. Verify the output contains both the summary line (existing format) and the per-location detail lines that follow it.

**Acceptance Scenarios**:

1. **Given** unsubstituted annotations in the input, **When** strict mode fails, **Then** stderr contains the summary line `proto-filter: error: unsubstituted annotations found: Name1, Name2` followed by per-occurrence location lines.
2. **Given** only one unsubstituted annotation at one location, **When** strict mode fails, **Then** the summary line lists one name, and exactly one location line follows.

---

### Edge Cases

- What happens when an annotation appears in both `@Name` and `[Name]` syntax in the same file? Each occurrence is reported with its own line number and shown in the syntax it appeared in.
- What happens when an annotation appears on a service-level comment vs. a method-level comment? Both are reported with correct line numbers.
- What happens when annotation filtering removes elements before strict checking? Only annotations on surviving elements are reported (consistent with current strict mode scope).
- What happens with an empty input (no proto files)? No error, exit code 0 (unchanged).
- What happens when the same annotation name appears on multiple lines of the same comment (e.g., a multi-line comment block)? Each line containing the annotation is reported separately.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: When strict substitution check fails, the tool MUST output a location line for each unsubstituted annotation occurrence on stderr, in the format `  <relative-file-path>:<line-number>: <annotation-token>` (indented with two spaces).
- **FR-002**: The line number MUST reference the position in the original source file as parsed, using 1-based numbering.
- **FR-003**: The summary line listing unique unsubstituted annotation names MUST still be printed (preserving backward compatibility with existing error format).
- **FR-004**: Location lines MUST appear after the summary line on stderr.
- **FR-005**: Location lines MUST be ordered by file path (alphabetically), then by line number (ascending) within each file.
- **FR-006**: The annotation token in the location line MUST show the full annotation expression as it appears in the source (e.g., `@Deprecated`, `[SupportWindow]`, `@HasAnyRole({"ADMIN"})`).
- **FR-007**: The collection of annotation locations MUST only consider annotations on elements that survive include/exclude filtering (consistent with current strict mode scope).
- **FR-008**: The exit code (2) and the condition for failure (any annotation without a substitution mapping) MUST remain unchanged.
- **FR-009**: When strict mode succeeds (all annotations have mappings), there MUST be no change in behavior or output.
- **FR-010**: Each annotation occurrence MUST be reported individually, even if the same annotation name appears multiple times in the same or different files.

### Key Entities

- **AnnotationLocation**: Represents a single annotation occurrence in source — contains the relative file path, the line number in the original source, the full annotation token as it appears in the comment, and the annotation name (for lookup against the substitution map).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Every unsubstituted annotation occurrence in the input is listed with its file path and line number in the error output — zero occurrences are missed.
- **SC-002**: Line numbers in error output match the actual line positions in the original source files (verified by golden file comparison or manual inspection).
- **SC-003**: The existing summary line format is preserved — existing tests for the summary message continue to pass without modification.
- **SC-004**: All existing tests pass without modification (backward compatibility).

## Assumptions

- Line numbers are 1-based (matching standard editor conventions and the proto parser's position tracking).
- The relative file path uses the same format as other proto-filter messages (relative to input directory).
- The annotation token shown in the location line is the full matched expression from the source comment (e.g., `@Name(...)` or `[Name(...)]`), not just the bare annotation name.
- The proto parser preserves source position information on `Comment` structs that can be used to compute per-line line numbers.
- Comment position tracking provides the starting line of the comment; individual line offsets within the comment can be computed by adding the line index.
