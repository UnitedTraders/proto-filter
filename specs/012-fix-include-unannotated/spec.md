# Feature Specification: Fix Include Filter Keeping Unannotated Services

**Feature Branch**: `012-fix-include-unannotated`
**Created**: 2026-02-11
**Status**: Draft
**Input**: User description: "Combining 'include' and 'exclude' gives wrong results when service has no annotations. Services without annotations should be excluded when include annotations are configured, because the required include annotation is missing."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Unannotated Services Removed When Include Is Configured (Priority: P1)

As a user, when I configure `annotations.include: ["PublishedApi"]`, I expect that only services explicitly marked with `[PublishedApi]` appear in the output. A service with no annotations at all should be removed, because it does not carry the required include marker.

**Why this priority**: This is the core bug. The current behavior contradicts the documented spec (feature 011, assumption: "Services without annotations and without include matches are excluded") and produces incorrect output.

**Independent Test**: Create a proto file with one unannotated service and configure `annotations.include: ["PublishedApi"]`. Run proto-filter and verify the output is empty (service removed, its message dependencies removed as orphans).

**Acceptance Scenarios**:

1. **Given** a proto file with a single service that has no annotation comments, **When** the config has `annotations.include: ["PublishedApi"]`, **Then** the service is removed from the output and all its exclusively-referenced messages are removed as orphans.
2. **Given** a proto file with two services — one annotated `[PublishedApi]` and one with no annotations, **When** the config has `annotations.include: ["PublishedApi"]`, **Then** only the annotated service and its dependencies remain in the output.
3. **Given** a proto file with two services — one annotated `[PublishedApi]` and one with no annotations — that share a common message type, **When** the config has `annotations.include: ["PublishedApi"]`, **Then** the unannotated service is removed, the annotated service remains, and the shared message remains (still referenced by the kept service).

---

### User Story 2 - Combined Include + Exclude With Unannotated Services (Priority: P1)

As a user, when I configure both `annotations.include: ["PublishedApi"]` and `annotations.exclude: ["Deprecated"]`, unannotated services must still be removed by the include gate before exclude filtering runs.

**Why this priority**: This is the exact scenario described in the bug report. Combined mode must correctly remove unannotated services.

**Independent Test**: Create a proto file with an unannotated service whose messages contain deprecated fields. Configure both include and exclude. Verify the service and its messages are completely removed — the deprecated field filtering never even applies because the service was already excluded by the include gate.

**Acceptance Scenarios**:

1. **Given** a proto file with an unannotated service `OrderService` and a message with a `[Deprecated]` field, **When** the config has `annotations.include: ["PublishedApi"]` and `annotations.exclude: ["Deprecated"]`, **Then** the output is empty because `OrderService` lacks the required `[PublishedApi]` annotation.
2. **Given** a proto file with an annotated service `[PublishedApi] OrderService` and an unannotated service `InternalService`, **When** the config has `annotations.include: ["PublishedApi"]` and `annotations.exclude: ["Deprecated"]`, **Then** only `OrderService` and its dependencies remain.

---

### User Story 3 - Include-Only Mode Consistency (Priority: P2)

As a user, when I configure only `annotations.include` (no exclude), the include gate must behave identically: unannotated services are removed, and within annotated services, only methods with matching include annotations are kept.

**Why this priority**: The fix must apply consistently in include-only mode for the same reason. An unannotated service not matching the include list should never pass through.

**Independent Test**: Run include-only test cases with the fix applied. Verify unannotated services are removed while annotated services and their matched methods are kept.

**Acceptance Scenarios**:

1. **Given** a proto file with services where some have `[Public]` annotations and some have no annotations, **When** the config has only `annotations.include: ["Public"]`, **Then** unannotated services are removed and only `[Public]`-annotated services remain.
2. **Given** a proto file where all services have annotations matching the include list, **When** the config has only `annotations.include`, **Then** all services are kept (no change from current behavior for annotated services).

---

### Edge Cases

- What happens when a service has annotations but none match the include list? The service is removed (this already works correctly).
- What happens when all services are unannotated and include is configured? All services are removed, resulting in an empty output file (file may be omitted from output).
- What happens when exclude-only mode is configured (no include)? Unannotated services are kept as before — this fix only affects behavior when include annotations are configured.
- What happens when a file has no services at all (only messages)? Messages without any referencing service are removed as orphans, consistent with existing orphan removal logic.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: When `annotations.include` is configured, services with no annotation comments MUST be removed from the output (they do not satisfy the include requirement).
- **FR-002**: This behavior MUST apply in both include-only mode and combined include+exclude mode.
- **FR-003**: Services with annotations that do not match the include list MUST continue to be removed (existing behavior, no change).
- **FR-004**: Services with annotations matching the include list MUST continue to be kept (existing behavior, no change).
- **FR-005**: Exclude-only mode MUST NOT be affected by this change — unannotated services remain kept when only exclude is configured.
- **FR-006**: After removing unannotated services, orphaned type definitions MUST be cleaned up as usual.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: An unannotated service produces empty output when `annotations.include` is configured, verified by automated tests.
- **SC-002**: Combined include+exclude correctly removes unannotated services before applying exclude filtering, verified by integration tests matching the bug report scenario.
- **SC-003**: All existing tests for exclude-only and annotated-service include scenarios continue to pass without modification.
- **SC-004**: The fix aligns with the documented specification in feature 011 (assumption: "Services without annotations and without include matches are excluded").

### Assumptions

- The fix targets the service-level include filtering logic, which currently has an explicit code path that keeps services with no annotations. This code path should instead remove them (treat "no annotations" as "does not match the include list").
- Existing golden files and tests that depend on the old behavior (keeping unannotated services) will need to be updated to reflect the corrected behavior.
- The method-level include filtering already correctly removes unannotated methods — only the service-level function has the bug.
