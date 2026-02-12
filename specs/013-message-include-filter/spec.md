# Feature Specification: Include Annotation Filtering for Top-Level Messages

**Feature Branch**: `013-message-include-filter`
**Created**: 2026-02-11
**Status**: Draft
**Input**: User description: "When annotations.include is configured, top-level messages without a matching include annotation should be removed (along with their unreferenced dependencies). Currently, include annotations only gate services — messages pass through unfiltered."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Unannotated Messages Removed When Include Is Configured (Priority: P1)

As a user, when I configure `annotations.include: ["PublishedApi"]` and my proto file contains only messages (no services), I expect that only messages explicitly annotated with `[PublishedApi]` appear in the output. Messages without the annotation should be removed. If no messages match, the output should be empty.

**Why this priority**: This is the core bug. Currently, all messages pass through regardless of include annotations because include filtering only operates on services. This produces incorrect output for proto files that define message types without services.

**Independent Test**: Create a proto file with only messages (no services), none annotated with `[PublishedApi]`. Configure `annotations.include: ["PublishedApi"]`. Verify the output is empty.

**Acceptance Scenarios**:

1. **Given** a proto file with three unannotated messages and no services, **When** the config has `annotations.include: ["NotUsed"]`, **Then** the output is empty because no messages have the required annotation.
2. **Given** a proto file with one message annotated `[PublishedApi]` that references two other messages, **When** the config has `annotations.include: ["PublishedApi"]`, **Then** the annotated message and its referenced dependencies are kept, with the include annotation stripped from the output.
3. **Given** a proto file with one annotated message `[PublishedApi] Parameters` referencing `Params1` and `Params2`, and one unreferenced unannotated message `Unrelated`, **When** the config has `annotations.include: ["PublishedApi"]`, **Then** `Parameters`, `Params1`, and `Params2` are kept, and `Unrelated` is removed.

---

### User Story 2 - Mixed Services and Messages With Include Annotations (Priority: P2)

As a user, when my proto file contains both services and messages, include annotations should gate both: services without the annotation are removed (existing behavior), and top-level messages without the annotation are also removed — unless they are referenced by a kept service or kept message.

**Why this priority**: This extends the core fix to mixed files, ensuring consistent behavior across element types.

**Independent Test**: Create a proto file with one annotated service, one unannotated top-level message (not referenced by the service), and the service's dependency messages. Verify the unreferenced unannotated message is removed.

**Acceptance Scenarios**:

1. **Given** a proto file with an annotated service `[PublicApi] OrderService` and an unannotated standalone message `AuditLog`, **When** the config has `annotations.include: ["PublicApi"]`, **Then** `OrderService` and its dependency messages are kept, but `AuditLog` is removed.
2. **Given** a proto file with an annotated service and an annotated message `[PublicApi] SharedConfig`, **When** the config has `annotations.include: ["PublicApi"]`, **Then** both the service and `SharedConfig` (plus its dependencies) are kept.

---

### User Story 3 - Combined Include + Exclude With Messages (Priority: P2)

As a user, when I configure both `annotations.include` and `annotations.exclude`, include filtering should gate messages first, and then exclude filtering should remove annotated fields from the included messages.

**Why this priority**: Ensures combined mode works correctly with messages, not just services.

**Independent Test**: Create a proto file with an annotated message containing a deprecated field. Configure both include and exclude. Verify the message is kept but the deprecated field is removed.

**Acceptance Scenarios**:

1. **Given** a proto file with `[PublishedApi] Parameters` containing a `[Deprecated]` field, **When** the config has `annotations.include: ["PublishedApi"]` and `annotations.exclude: ["Deprecated"]`, **Then** `Parameters` is kept with the deprecated field removed.

---

### Edge Cases

- What happens when a message references another message that also has a matching include annotation? Both are kept — the annotation on the dependency is redundant but not harmful.
- What happens when an included message references a message that itself references further messages? All transitive dependencies are kept (consistent with existing orphan removal behavior).
- What happens when only exclude is configured (no include)? Messages are NOT subject to include gating — exclude-only mode behavior is unchanged.
- What happens with enum types? Enums follow the same logic as messages: if include is configured, only annotated enums (and enums referenced by included elements) are kept.
- What happens with the `oneof` construct inside a message? Oneof fields reference other message types; these references are followed during dependency resolution as usual.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: When `annotations.include` is configured, top-level messages without a matching include annotation MUST be removed from the output, unless they are referenced (directly or transitively) by a kept service or kept message.
- **FR-002**: When `annotations.include` is configured, top-level messages WITH a matching include annotation MUST be kept, along with all their transitive message/enum dependencies.
- **FR-003**: Include annotation stripping MUST apply to kept messages (consistent with existing service annotation stripping).
- **FR-004**: Exclude-only mode MUST NOT be affected — messages are not subject to include gating when only `annotations.exclude` is configured.
- **FR-005**: Orphan removal MUST run after message-level include filtering to clean up unreferenced dependencies.
- **FR-006**: This behavior MUST apply consistently in include-only mode and combined include+exclude mode.
- **FR-007**: Both `@Name` and `[Name]` annotation syntaxes MUST be recognized on messages (consistent with existing annotation parsing).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A proto file containing only unannotated messages produces empty output when `annotations.include` is configured, verified by automated tests.
- **SC-002**: A proto file with one annotated message and its dependencies produces correct output containing only the annotated message and its referenced types, verified by golden file tests.
- **SC-003**: All existing tests for service-level include filtering, exclude-only mode, and combined mode continue to pass.
- **SC-004**: The user's exact example (three messages, one annotated `[PublishedApi]`) produces the expected output, verified by a dedicated test.

### Assumptions

- Include filtering for messages follows the same annotation extraction logic used for services (`ExtractAnnotations` on the message's comment).
- Messages kept by include annotations act as "roots" for dependency resolution — their referenced types are kept even if those types lack an include annotation.
- The existing orphan removal mechanism (`RemoveOrphanedDefinitions`) handles transitive dependency resolution and can be reused after message-level include filtering.
- Enum types are treated identically to messages for include gating purposes.
