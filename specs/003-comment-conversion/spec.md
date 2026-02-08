# Feature Specification: Comment Style Conversion

**Feature Branch**: `003-comment-conversion`
**Created**: 2026-02-09
**Status**: Draft
**Input**: User description: "Java-style multiline comments with asterisks should be converted to single-line `//` comments"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Convert Block Comments to Single-Line Comments (Priority: P1)

When proto files are processed through the tool, any Java-style multiline block comments (`/** ... */` or `/* ... */`) should be automatically converted to consecutive single-line comments (`// ...`). This ensures consistent comment style in the output.

For example, a block comment like:
```
/**
 * Returns updates of price for all symbols from symbolSpec
 * @StartsWithSnapshot
 * @SupportWindow
 */
```
should become:
```
// Returns updates of price for all symbols from symbolSpec
// @StartsWithSnapshot
// @SupportWindow
```

**Why this priority**: This is the core feature. Without it, block comments pass through unchanged.

**Independent Test**: Can be fully tested by processing a proto file containing block comments and verifying the output contains only single-line `//` comments with the same text content.

**Acceptance Scenarios**:

1. **Given** a proto file with a `/** ... */` block comment above an RPC method, **When** the file is processed, **Then** the output contains the same comment text as consecutive `//` lines.
2. **Given** a proto file with a `/* ... */` block comment above a message, **When** the file is processed, **Then** the output contains the same comment text as consecutive `//` lines.
3. **Given** a proto file with only `//` comments, **When** the file is processed, **Then** the comments are unchanged.

---

### User Story 2 - Preserve Comment Content During Conversion (Priority: P1)

During conversion, the meaningful text content of the comment must be preserved exactly. Leading asterisks (`*`) and extra whitespace from the block comment formatting must be stripped, but the actual descriptive text and annotations must remain intact.

**Why this priority**: Losing or corrupting comment content during conversion would make the feature harmful rather than helpful.

**Independent Test**: Can be tested by comparing the text content of comments before and after conversion, verifying annotations like `@HasAnyRole` and descriptive text are preserved verbatim.

**Acceptance Scenarios**:

1. **Given** a block comment with leading ` * ` on each line, **When** converted, **Then** each output line contains the text after the ` * ` prefix without leading/trailing whitespace artifacts.
2. **Given** a block comment containing `@Annotation` markers, **When** converted, **Then** the annotations appear in the output `//` lines unchanged.
3. **Given** a block comment with an empty line (just ` * `), **When** converted, **Then** the output contains an empty `//` comment line.

---

### Edge Cases

- What happens when a block comment has no text content (just `/* */`)? The comment should be converted to an empty `//` line.
- What happens when a block comment line has no leading asterisk (unusual but valid)? The line content should be preserved as-is.
- What happens when inline block comments exist (e.g., `/* comment */` after a field)? These should also be converted to `//` style.
- What happens when the block comment contains lines with mixed indentation? The relative indentation within the comment text should be preserved.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The tool MUST convert all C-style block comments (`/* ... */` and `/** ... */`) in proto files to consecutive single-line `//` comments.
- **FR-002**: The tool MUST strip leading asterisk prefixes (` * `, ` *`) from each line of the block comment during conversion.
- **FR-003**: The tool MUST preserve all meaningful text content of comments during conversion, including annotation markers (e.g., `@HasAnyRole`, `@StartsWithSnapshot`).
- **FR-004**: The tool MUST leave existing single-line `//` comments unchanged.
- **FR-005**: The tool MUST apply comment conversion to all comment positions: leading comments on services, messages, enums, fields, RPCs, and inline comments.
- **FR-006**: The tool MUST apply comment conversion regardless of whether filtering is configured (including pass-through mode).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of block comments in processed proto files are converted to single-line `//` comments in the output.
- **SC-002**: Comment text content is preserved exactly after conversion, with no loss of annotations or descriptive text.
- **SC-003**: Existing single-line comments pass through unchanged.
- **SC-004**: Output proto files remain valid and parseable after comment conversion.
