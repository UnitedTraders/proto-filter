# Feature Specification: Annotation-Based Method Filtering

**Feature Branch**: `002-annotation-filter`
**Created**: 2026-02-08
**Status**: Draft
**Input**: User description: "Proto files can have special type of strings like `@Name` or `@HasAnyRole({"ADMIN"})` that are Java-style annotations. The tool should support filtering by such annotations, for ex 'Filter all methods that has @HasAnyRole annotation'. Filtering methods should check if input or output messages are used anywhere in the sources and filter out the ones that are not used after filtering methods. If a service becomes empty after filtering methods, the service should be filtered out completely."

## User Scenarios & Testing

### User Story 1 - Filter RPC Methods by Annotation (Priority: P1)

As a developer, I want to filter proto files so that RPC methods carrying a specific annotation (e.g., `@HasAnyRole`) are removed from the output. This allows me to strip out methods marked with certain metadata — for example, removing all methods that require specific authorization roles to produce a "public-only" proto definition.

Annotations appear as Java-style tokens inside proto comments attached to RPC methods, for example:

```protobuf
// @HasAnyRole({"ADMIN", "MANAGER"})
// Creates a new order in the system.
rpc CreateOrder(CreateOrderRequest) returns (CreateOrderResponse);
```

**Why this priority**: This is the core capability. Without annotation-based method filtering, the orphan cleanup and empty service removal have nothing to act on.

**Independent Test**: Run the tool with an annotation filter config against proto files containing annotated and non-annotated methods. Verify that methods with matching annotations are removed and all other methods are kept.

**Acceptance Scenarios**:

1. **Given** a proto file with three RPC methods where two have `@HasAnyRole` in their comments, **When** the tool runs with annotation filter `@HasAnyRole`, **Then** the two annotated methods are removed and only the one non-annotated method remains in the output.
2. **Given** a proto file with an RPC method annotated with `@HasAnyRole({"ADMIN"})`, **When** the tool runs with annotation filter `@HasAnyRole`, **Then** the method is removed (annotation name matching ignores arguments).
3. **Given** a proto file where no methods have annotations matching the filter, **When** the tool runs with annotation filter `@HasAnyRole`, **Then** all methods are kept unchanged.
4. **Given** a method with an annotation in a block comment (`/* @HasAnyRole */`), **When** the tool runs with annotation filter `@HasAnyRole`, **Then** the method is removed.

---

### User Story 2 - Remove Orphaned Messages After Method Filtering (Priority: P2)

As a developer, after methods are filtered by annotation, I want any request/response message types that are no longer referenced by any remaining method or message to be automatically removed. This keeps the output clean and free of unused definitions.

**Why this priority**: Without orphan cleanup, the output contains dead message types that clutter generated code and confuse consumers.

**Independent Test**: Run the tool with annotation filtering against proto files where some request/response types are shared between annotated and non-annotated methods. Verify that only messages still referenced survive.

**Acceptance Scenarios**:

1. **Given** a method `CreateOrder` is removed because it has a matching annotation and its `CreateOrderRequest` message is not used by any remaining method or message, **When** the tool writes output, **Then** `CreateOrderRequest` is absent from the output.
2. **Given** a message `CommonPagination` is used by both a removed annotated method and a kept non-annotated method, **When** the tool writes output, **Then** `CommonPagination` is preserved.
3. **Given** a message `OrderDetail` is referenced only by another message `CreateOrderResponse` which is itself orphaned after method removal, **When** the tool writes output, **Then** both are removed (transitive orphan removal).
4. **Given** an enum `Status` is referenced by a kept message, **When** the tool writes output, **Then** the enum is preserved.

---

### User Story 3 - Remove Empty Services (Priority: P3)

As a developer, if all methods in a service are filtered out by annotation, I want the entire service definition removed from the output rather than left as an empty block.

**Why this priority**: An empty service definition is useless in generated code and would confuse consumers.

**Independent Test**: Run the tool with annotation filtering against a proto file containing a service where no methods match the annotation filter. Verify the entire service block is removed.

**Acceptance Scenarios**:

1. **Given** a service `InternalService` where all of its methods have `@HasAnyRole`, **When** the tool runs with annotation filter `@HasAnyRole`, **Then** all methods are removed and `InternalService` is removed entirely from the output.
2. **Given** a service `OrderService` where some methods have `@HasAnyRole` and others do not, **When** the tool runs with annotation filter `@HasAnyRole`, **Then** the service is preserved with only the non-annotated methods.
3. **Given** a proto file that becomes completely empty after filtering (no services, no referenced messages, no referenced enums remain), **When** the tool writes output, **Then** the file is not written to the output directory.

---

### Edge Cases

- What happens when an annotation appears in a comment on a message or field (not an RPC method)? Only annotations on RPC method comments are considered for method filtering.
- What happens when a method has multiple annotations (e.g., `@HasAnyRole` and `@Deprecated`)? The method is removed if any of its annotations match the filter.
- What happens when annotation filter is combined with existing include/exclude name filters? Annotation filtering is applied after name-based filtering. Both filters must pass for a definition to be kept.
- What happens when an annotation name contains special characters like dots (e.g., `@com.example.Secure`)? The annotation name is matched literally after the `@` prefix.
- What happens when a file has no services (only messages/enums)? Those definitions follow existing name-based filtering rules; annotation filtering only targets RPC methods.

## Requirements

### Functional Requirements

- **FR-001**: System MUST recognize Java-style annotations in proto comments attached to RPC methods. Annotations follow the pattern `@AnnotationName` optionally followed by arguments in parentheses (e.g., `@HasAnyRole({"ADMIN"})`).
- **FR-002**: System MUST support an `annotations` key in the YAML filter configuration as a simple list of annotation names (without `@` prefix) whose matching methods are removed. Example: `annotations: ["HasAnyRole", "Internal"]`.
- **FR-003**: System MUST match annotations by name only, ignoring any arguments in parentheses. For example, filtering by `@HasAnyRole` matches both `@HasAnyRole` and `@HasAnyRole({"ADMIN"})`.
- **FR-004**: System MUST detect annotations in both single-line (`//`) and block (`/* */`) comments attached to RPC methods.
- **FR-005**: System MUST remove RPC methods that carry any of the specified filter annotations.
- **FR-006**: System MUST remove a service definition entirely if all of its RPC methods are filtered out.
- **FR-007**: System MUST remove message and enum definitions that are no longer referenced by any remaining RPC method or any other remaining message after method filtering.
- **FR-008**: System MUST preserve message and enum definitions that are still referenced by at least one remaining method or message, even if they were also referenced by filtered-out methods.
- **FR-009**: System MUST perform transitive orphan detection — if message A is only referenced by message B, and message B is orphaned, then message A is also removed.
- **FR-010**: System MUST not write a proto file to the output directory if it contains no remaining definitions after filtering.
- **FR-011**: System MUST support annotation filtering in combination with existing name-based include/exclude filtering. Annotation filtering is applied after name-based filtering.
- **FR-012**: System MUST preserve comments on methods, messages, enums, and fields that survive filtering.
- **FR-013**: When `--verbose` is enabled, system MUST report the number of methods filtered by annotation and the number of orphaned definitions removed.

### Key Entities

- **Annotation**: A token in an RPC method's comment matching the pattern `@Name` or `@Name(...)`. Identified by its name (the part after `@` before any parentheses or whitespace).
- **Orphaned Definition**: A message or enum that is not referenced by any remaining RPC method's input/output types or by any other remaining message's field types after method filtering.

## Success Criteria

### Measurable Outcomes

- **SC-001**: Users can filter proto files by annotation name and receive output with all annotated methods removed, with zero false positives or false negatives.
- **SC-002**: Output proto files contain no orphaned message or enum definitions — every remaining definition is reachable from at least one remaining RPC method.
- **SC-003**: No empty service blocks appear in any output file.
- **SC-004**: Existing name-based filtering continues to work identically when no annotation filters are configured (full backward compatibility).
- **SC-005**: The filtering pipeline completes without errors on proto files containing mixed annotated and non-annotated methods across multiple files.

## Assumptions

- Annotations always appear in the comment block immediately preceding an RPC method definition.
- Annotation names are matched case-sensitively (e.g., `@hasAnyRole` does not match a filter for `@HasAnyRole`).
- The annotation argument syntax (parenthesized content) is opaque — the tool does not parse or validate annotation arguments.
- Cross-file dependency resolution for orphan detection reuses the existing dependency graph infrastructure.

## Clarifications

### Session 2026-02-08

- Q: How should annotation filters be structured in the YAML config? → A: Simple list under `annotations` key with bare names (no `@` prefix). Example: `annotations: ["HasAnyRole", "Internal"]`. All listed annotations trigger method removal.
