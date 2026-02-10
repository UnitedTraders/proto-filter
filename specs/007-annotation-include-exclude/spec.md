# Feature Specification: Annotation Include/Exclude Filtering Modes

**Feature Branch**: `007-annotation-include-exclude`
**Created**: 2026-02-09
**Status**: Draft
**Input**: User description: "Current version can filter by annotation i.e. all the methods and services WITH annotations from the 'annotations' list will be removed from the result. Please add opposite option to LEAVE only the services and methods that HAVE annotation from 'include' list from config. Current 'annotations' config section should be renamed to 'exclude'. Single config should have either 'include' or 'exclude', not both of them."

## Clarifications

### Session 2026-02-09

- Q: How should the system behave when both the old flat `annotations` key and the new structured `annotations.exclude`/`annotations.include` key are present? → A: Reject with error (same as include+exclude conflict).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Include Mode: Keep Only Annotated Elements (Priority: P1)

As a user, I want to specify an `annotations.include` list in my config so that only services and methods that have a matching annotation are kept in the output, and everything else is removed. This is the opposite of the current behavior.

**Why this priority**: This is the core new capability — the entire feature request is about adding this "include" mode. Without it, there is nothing to deliver.

**Independent Test**: Create a proto file with annotated and unannotated methods, configure `annotations.include` with the annotation name, run proto-filter, verify only annotated methods remain in output.

**Acceptance Scenarios**:

1. **Given** a proto file with methods annotated `@Public` and unannotated methods, **When** the config has `annotations.include: [Public]`, **Then** only the `@Public` methods remain in the output and unannotated methods are removed.
2. **Given** a proto file with a service annotated `@Public` and a service without annotations, **When** the config has `annotations.include: [Public]`, **Then** only the annotated service remains in the output.
3. **Given** a proto file where no methods or services match the include list, **When** filtering is applied, **Then** the file is omitted from output (no definitions remain).
4. **Given** a proto file with methods annotated using bracket syntax `[Public]`, **When** the config has `annotations.include: [Public]`, **Then** bracket-annotated methods are also kept (both syntaxes work with include mode).

---

### User Story 2 - Rename Config Key: `annotations` to `annotations.exclude` (Priority: P2)

As a user, I want the existing flat `annotations` config key to be renamed to a structured `annotations.exclude` key, so the config clearly communicates the filtering direction. The old flat `annotations` format should continue to work for backward compatibility.

**Why this priority**: This restructures the config to accommodate both modes. It must work before include mode is usable, but the rename itself builds on existing exclude behavior.

**Independent Test**: Take an existing config with `annotations: [HasAnyRole]`, run proto-filter, verify it still works identically. Then change config to `annotations.exclude: [HasAnyRole]` and verify identical output.

**Acceptance Scenarios**:

1. **Given** a config file using the old format `annotations: [HasAnyRole]`, **When** proto-filter runs, **Then** it behaves identically to before (backward compatibility).
2. **Given** a config file using the new format `annotations:\n  exclude: [HasAnyRole]`, **When** proto-filter runs, **Then** annotated methods/services are removed (same as old behavior).
3. **Given** a config file using the new `annotations.exclude` format, **When** verbose mode is enabled, **Then** the output reflects the exclude operation.

---

### User Story 3 - Mutual Exclusivity Validation (Priority: P3)

As a user, if I accidentally specify both `annotations.include` and `annotations.exclude` in the same config, I want a clear error message telling me to choose one or the other.

**Why this priority**: This is a safety guard. The core functionality works without it, but it prevents confusing behavior from ambiguous configs.

**Independent Test**: Create a config with both `annotations.include` and `annotations.exclude` populated, run proto-filter, verify it exits with a clear error message.

**Acceptance Scenarios**:

1. **Given** a config with both `annotations.include: [Public]` and `annotations.exclude: [Internal]`, **When** proto-filter runs, **Then** it exits with a non-zero status and prints an error message indicating that include and exclude are mutually exclusive.
2. **Given** a config with `annotations.include: [Public]` and `annotations.exclude: []` (empty), **When** proto-filter runs, **Then** it succeeds (empty list is treated as absent).
3. **Given** a config with neither include nor exclude annotations, **When** proto-filter runs, **Then** it succeeds with no annotation filtering applied.

---

### Edge Cases

- What happens when the include list contains an annotation name that matches no methods or services? All elements are removed; the file is omitted from output.
- What happens when both the old flat `annotations` key and the new structured `annotations.exclude` or `annotations.include` key are present? The system rejects the config with a clear error message and non-zero exit code, same as the include+exclude conflict. Users must migrate to the new format or remove the structured key.
- What happens when `annotations.include` is set but the proto file has no annotations at all? All methods/services are removed since none match the include list.
- What happens with types (messages) referenced only by excluded methods in include mode? Orphaned types are removed, same as current exclude behavior.
- What happens when a service has some methods matching the include list and others not? Only matching methods are kept; non-matching methods are removed. If no methods remain, the service is removed.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST support an `annotations.include` config key that accepts a list of annotation names.
- **FR-002**: When `annotations.include` is configured, the system MUST keep only services and methods whose comments contain a matching annotation, removing all others.
- **FR-003**: System MUST support an `annotations.exclude` config key that accepts a list of annotation names, replacing the current flat `annotations` key for the exclude use case.
- **FR-004**: When `annotations.exclude` is configured, the system MUST remove services and methods whose comments contain a matching annotation (current behavior, same semantics as today's `annotations`).
- **FR-005**: System MUST continue to accept the old flat `annotations` config format and treat it as `annotations.exclude` for backward compatibility.
- **FR-006**: System MUST reject configs where both `annotations.include` and `annotations.exclude` are non-empty, exiting with a clear error message and non-zero status code.
- **FR-007**: After include or exclude filtering, the system MUST remove orphaned type definitions (messages/enums) that are no longer referenced by any remaining service method, consistent with current behavior.
- **FR-008**: Both `@Name` and `[Name]` annotation syntaxes MUST work with both include and exclude modes.
- **FR-009**: When `annotations.include` is set and a file has no matching elements remaining after filtering, the system MUST omit that file from output.
- **FR-010**: System MUST reject configs where the old flat `annotations` key coexists with the new structured `annotations.include` or `annotations.exclude` key, exiting with a clear error message and non-zero status code.

### Key Entities

- **AnnotationConfig**: Structured annotation filtering configuration with optional `include` and `exclude` lists, replacing the flat annotation list. At most one of include/exclude may be non-empty.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Include mode correctly keeps only matching annotated services/methods and removes all non-matching ones, verified by golden file tests.
- **SC-002**: Exclude mode (new config format) produces identical output to the current `annotations` key behavior, verified by golden file tests.
- **SC-003**: Old flat `annotations` config format continues to work without any changes to existing config files (backward compatibility).
- **SC-004**: Configs with both include and exclude populated produce a clear error message and non-zero exit code.
- **SC-005**: All existing tests continue to pass without modification (no regressions).
- **SC-006**: Both annotation syntaxes (`@Name` and `[Name]`) work correctly with include mode.
- **SC-007**: Orphaned type cleanup works correctly in include mode (types only referenced by removed methods are cleaned up).
