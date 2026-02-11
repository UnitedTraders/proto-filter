# Feature Specification: Combined Include and Exclude Annotation Filters

**Feature Branch**: `011-combined-include-exclude`
**Created**: 2026-02-09
**Status**: Draft
**Input**: User description: "Remove mutual exclusivity constraint between annotation include and exclude. When both are specified, first apply inclusion logic, then apply exclusion logic. Also support exclude-style annotation filtering on message fields."

## Clarifications

### Session 2026-02-09

- Q: Should field-level annotation filtering apply automatically in exclude-only mode, or only when both include and exclude are configured? → A: Field filtering always applies when exclude annotations are configured (include-only, exclude-only, or combined).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Combined Include + Exclude Filtering (Priority: P1)

As a user, I want to specify both `annotations.include` and `annotations.exclude` in the same config so that I can first select a subset of services/methods via include annotations, and then further refine the result by removing elements marked with exclude annotations. This lets me express "give me everything marked as PublicApi, but remove anything marked Deprecated" in a single config.

**Why this priority**: This is the core feature request. The current mutual exclusivity constraint prevents users from composing include and exclude filters together, forcing workarounds.

**Independent Test**: Create a proto file with a service annotated `[PublicApi]` containing methods and messages. Mark one message field as `[Deprecated]`. Configure `annotations.include: [PublicApi]` and `annotations.exclude: [Deprecated]`. Run proto-filter and verify the service is included (because of `PublicApi`), but the deprecated field is removed.

**Acceptance Scenarios**:

1. **Given** a proto file with a service annotated `[PublicApi]` containing two RPCs, and a message with a field annotated `[Deprecated]`, **When** the config has `annotations.include: [PublicApi]` and `annotations.exclude: [Deprecated]`, **Then** the service and both RPCs are included, the deprecated field is removed from the message, and all required messages remain.
2. **Given** a proto file with two services — one annotated `[PublicApi]` and one annotated `[Internal]`, **When** the config has `annotations.include: [PublicApi]` and `annotations.exclude: [Internal]`, **Then** only the `PublicApi` service remains (the `Internal` service was not included in the first place).
3. **Given** a service annotated `[PublicApi]` with a method annotated `[Deprecated]`, **When** the config has `annotations.include: [PublicApi]` and `annotations.exclude: [Deprecated]`, **Then** the service is included but the deprecated method is removed.
4. **Given** a config with both `annotations.include` and `annotations.exclude` populated, **When** proto-filter runs, **Then** no mutual exclusivity error is produced (the old validation error is removed).

---

### User Story 2 - Message Field Annotation Filtering (Priority: P2)

As a user, I want annotation-based exclude filtering to also apply to individual message fields, so that I can remove specific fields (e.g., deprecated ones) from message definitions without removing the entire message.

**Why this priority**: The user's example explicitly shows a message field annotated with `[Deprecated]` being removed from the output. Currently, annotation filtering only applies to services and RPC methods. This extends filtering to message fields, which is essential for the combined filtering use case to work as demonstrated.

**Independent Test**: Create a proto file with a message containing a field annotated `[Deprecated]`, configure `annotations.exclude: [Deprecated]`, run proto-filter, and verify the deprecated field is removed while the rest of the message remains.

**Acceptance Scenarios**:

1. **Given** a message with three fields where one is annotated `// [Deprecated]`, **When** the config has `annotations.exclude: [Deprecated]`, **Then** the annotated field is removed and the other two fields remain.
2. **Given** a message where all fields are annotated with a filtered annotation, **When** exclude filtering is applied, **Then** the message remains but with no fields (an empty message body).
3. **Given** a message field annotated with `// @Deprecated` (at-sign syntax), **When** the config has `annotations.exclude: [Deprecated]`, **Then** the field is also removed (both annotation syntaxes work).
4. **Given** a message with an annotated field but no annotation filtering configured, **When** proto-filter runs, **Then** the field is kept (annotation comments are not touched without filtering).

---

### User Story 3 - Include-Only and Exclude-Only Backward Compatibility (Priority: P3)

As a user, when I specify only `annotations.include` or only `annotations.exclude`, the system must continue to behave identically to the current implementation. Only the mutual exclusivity constraint is removed; all other behavior remains unchanged.

**Why this priority**: Existing users must not experience regressions. This is a safety story ensuring backward compatibility.

**Independent Test**: Run existing test suites for include-only and exclude-only configs. All must pass without modification.

**Acceptance Scenarios**:

1. **Given** a config with only `annotations.include: [Public]`, **When** proto-filter runs, **Then** behavior is identical to the current include-only mode.
2. **Given** a config with only `annotations.exclude: [Internal]`, **When** proto-filter runs, **Then** behavior is identical to the current exclude-only mode.
3. **Given** a config using the old flat `annotations: [Internal]` format, **When** proto-filter runs, **Then** it works as before (treated as exclude list).

---

### Edge Cases

- What happens when include and exclude lists contain the same annotation name? The inclusion pass includes elements with that annotation, but the exclusion pass then removes them. The net result is those elements are removed.
- What happens when a service is included by `annotations.include` but one of its methods is excluded by `annotations.exclude`? The service is included, but the specific method is removed. If all methods are removed, the service is also removed (empty service pruning).
- What happens when exclude filtering removes all fields from a message? The message remains with an empty body. This is consistent with how services are treated — empty services are pruned, but messages may legitimately have no fields.
- What happens when a message field has an inline comment (after the field) with an annotation? The inline comment annotation is recognized and the field is filtered accordingly.
- What happens with combined filtering and orphan removal? After both include and exclude passes complete, orphaned type definitions are removed as usual.
- What happens with field filtering and substitution? Field-level annotation filtering runs before substitution. Substitution only applies to annotations that survive filtering.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST accept configs where both `annotations.include` and `annotations.exclude` are non-empty, removing the current mutual exclusivity validation error.
- **FR-002**: When both include and exclude are configured, the system MUST first apply inclusion logic (keeping only elements matching the include list), then apply exclusion logic (removing elements matching the exclude list) from the included result.
- **FR-003**: System MUST support annotation-based exclude filtering on individual message fields. Fields whose comments (preceding or inline) contain a matching exclude annotation are removed from the message.
- **FR-004**: Field-level annotation filtering MUST support both `@Name` and `[Name]` annotation syntaxes.
- **FR-005**: When only `annotations.include` is configured (no exclude), behavior MUST be identical to the current include-only mode.
- **FR-006**: When only `annotations.exclude` is configured (no include), behavior MUST be identical to the current exclude-only mode for services and methods, and MUST additionally apply field-level annotation filtering to message fields. This is not a regression — it is a deliberate extension of exclude filtering to a new element type.
- **FR-007**: The old flat `annotations: [list]` config format MUST continue to work as exclude-only mode.
- **FR-008**: After combined filtering, the system MUST remove orphaned type definitions that are no longer referenced by any remaining service method, consistent with current behavior.
- **FR-009**: Empty services (all methods removed) MUST be pruned after filtering, consistent with current behavior.
- **FR-010**: Verbose output MUST report field-level removals in addition to service and method removals.

### Key Entities

- **AnnotationConfig**: Updated to allow both `include` and `exclude` to be non-empty simultaneously.
- **Message Field**: Individual field within a protobuf message definition, now subject to annotation-based exclude filtering.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Combined include+exclude filtering produces correct output matching the user's example (service included via `[PublicApi]`, deprecated field removed via `[Deprecated]`), verified by golden file tests.
- **SC-002**: All existing tests pass without modification (backward compatibility).
- **SC-003**: Field-level annotation filtering correctly removes annotated fields while preserving non-annotated fields, verified by unit and integration tests.
- **SC-004**: Processing order is deterministic: include first, then exclude, then orphan cleanup.

### Assumptions

- Field-level annotation filtering applies whenever exclude annotations are configured (exclude-only, combined include+exclude, or include-only with exclude). Include mode itself continues to operate at the service and method level only, since the user's description specifies that inclusion logic applies to services and methods.
- The `annotations.include` annotation on a service acts as a gate: if a service has a matching include annotation, all its methods are included (unless individually excluded). Services without annotations and without include matches are excluded.
- Enum fields within enum definitions are not subject to annotation-based field filtering in this feature (only message fields).
