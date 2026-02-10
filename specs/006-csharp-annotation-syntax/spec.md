# Feature Specification: C#-Style Annotation Syntax Support

**Feature Branch**: `006-csharp-annotation-syntax`
**Created**: 2026-02-10
**Status**: Draft
**Input**: User description: "User can define annotation using C# style syntax like '[Name]' or '[Name(value)]'."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Recognize C#-Style Annotations in Proto Comments (Priority: P1)

As a developer using proto-filter, I want to annotate my proto service and method comments using C#-style square bracket syntax (`[Name]` or `[Name(value)]`) in addition to the existing Java-style `@Name` syntax, so that I can use whichever convention my team prefers.

For example, in a proto file:
```
// [Internal]
// Administrative operations.
service AdminService { ... }

// [HasAnyRole("ADMIN")]
// Creates a new order.
rpc CreateOrder(...) returns (...);
```

When the filter config specifies `annotations: ["Internal", "HasAnyRole"]`, services and methods annotated with `[Internal]` or `[HasAnyRole("ADMIN")]` should be filtered out, just as they would be with `@Internal` or `@HasAnyRole("ADMIN")`.

**Why this priority**: This is the core capability — without recognizing the new syntax, no other features work.

**Independent Test**: Annotate a proto service or method with `[Name]` syntax, configure annotation filtering, and verify the annotated element is removed from output.

**Acceptance Scenarios**:

1. **Given** a proto method comment containing `[HasAnyRole]`, **When** filtering with annotation `HasAnyRole`, **Then** the method is removed from output.
2. **Given** a proto method comment containing `[HasAnyRole("ADMIN")]`, **When** filtering with annotation `HasAnyRole`, **Then** the method is removed (arguments are ignored, matching is name-only).
3. **Given** a proto service comment containing `[Internal]`, **When** filtering with annotation `Internal`, **Then** the entire service is removed from output.
4. **Given** a proto comment containing `[com.example.Secure]`, **When** filtering with annotation `com.example.Secure`, **Then** the annotated element is removed (dotted names are supported).

---

### User Story 2 - Mixed Annotation Styles in Same Project (Priority: P2)

As a developer working on a project where different team members use different annotation conventions, I want proto-filter to recognize both `@Name` and `[Name]` styles in the same project (even in the same file), so that I don't have to enforce a single convention.

**Why this priority**: Real projects often have mixed conventions. Supporting both simultaneously makes the tool practical for teams migrating between styles or with mixed preferences.

**Independent Test**: Create a proto file with some annotations using `@Name` and others using `[Name]`, run the filter, and verify both styles are correctly matched and filtered.

**Acceptance Scenarios**:

1. **Given** a proto file where one method uses `@HasAnyRole` and another uses `[HasAnyRole]`, **When** filtering with annotation `HasAnyRole`, **Then** both methods are removed.
2. **Given** a proto file where a service uses `[Internal]` and its method uses `@HasAnyRole`, **When** filtering with annotations `Internal` and `HasAnyRole`, **Then** the service is removed entirely and the method would also have been filtered.

---

### Edge Cases

- What happens when square brackets appear in comments but are not annotations (e.g., `// See [RFC 7231]` or `// Returns [error code]`)? The tool should only match the annotation pattern `[Name]` or `[Name(...)]` where Name follows the same naming rules as @-style annotations (starts with a letter or underscore, contains word characters and dots). Plain English text in brackets containing spaces should not be treated as an annotation.
- What happens with nested brackets like `[Name([inner])]`? Only the outermost bracket pair with its name should be matched; the content inside parentheses is ignored for matching purposes.
- What happens with empty brackets `[]`? They should not be treated as an annotation.
- What happens with whitespace inside brackets like `[ Name ]`? It should not be treated as a valid annotation (no leading/trailing whitespace around the name).
- What happens with a bracket annotation at the start of a comment line vs. mid-line (e.g., `// text [Name] more text`)? The annotation should be recognized regardless of its position in the comment line, consistent with how `@Name` is recognized anywhere in a comment line.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The tool MUST recognize annotations in proto comments using C#-style square bracket syntax: `[Name]` and `[Name(value)]`.
- **FR-002**: Annotation name extraction from `[Name]` and `[Name(value)]` MUST follow the same naming rules as the existing `@Name` syntax (word characters and dots, starting with a word character).
- **FR-003**: Annotation matching MUST be name-only — any arguments inside parentheses within the brackets MUST be ignored, consistent with existing `@Name(value)` behavior.
- **FR-004**: The tool MUST support both `@Name` and `[Name]` styles simultaneously. Both styles MUST be recognized in any combination within the same file or across files.
- **FR-005**: The configuration format MUST remain unchanged — the `annotations` YAML key specifies annotation names without any syntax prefix (no `@` or `[]`), and the same name matches both styles.
- **FR-006**: Existing behavior for `@Name` annotations MUST be fully preserved — this is a backward-compatible addition.
- **FR-007**: The tool MUST NOT treat plain English text in brackets as annotations. Only bracket content matching the annotation naming pattern (no spaces in the name) should be recognized.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Proto methods and services annotated with `[Name]` or `[Name(value)]` in comments are correctly filtered when the annotation name matches the filter config.
- **SC-002**: Both `@Name` and `[Name]` styles work in the same file and are matched by the same config entry.
- **SC-003**: All existing tests continue to pass unchanged (backward compatibility).
- **SC-004**: All new tests pass with `go test -race ./...`.
- **SC-005**: Plain English text in brackets (e.g., `[RFC 7231]`, `[error code]`) is not falsely matched as an annotation.
