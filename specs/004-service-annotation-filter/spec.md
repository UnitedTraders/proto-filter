# Feature Specification: Service-Level Annotation Filtering

**Feature Branch**: `004-service-annotation-filter`
**Created**: 2026-02-09
**Status**: Draft
**Input**: User description: "Annotations can be set on services. If annotation is set to service and annotation is in config - whole service with all the methods should be removed from the output"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Remove Services by Annotation (Priority: P1)

As a developer, I want the annotation filter to also inspect comments on service declarations. If a service has an annotation in its comment that matches any annotation in the filter config, the entire service (including all its RPC methods) should be removed from the output. This allows me to strip out entire services that are marked as internal, restricted, or otherwise unsuitable for the target audience.

For example, given:
```protobuf
// @Internal
// Administrative operations for system management.
service AdminService {
  rpc ResetCache(ResetCacheRequest) returns (ResetCacheResponse);
  rpc GetMetrics(MetricsRequest) returns (MetricsResponse);
}
```

If the config includes annotation `Internal`, the entire `AdminService` and all its methods should be removed.

**Why this priority**: This is the core and only capability of this feature. It extends the existing method-level annotation filtering to the service level.

**Independent Test**: Run the tool with an annotation filter config against proto files containing annotated and non-annotated services. Verify that services with matching annotations are removed entirely with all their methods, while non-annotated services remain unchanged.

**Acceptance Scenarios**:

1. **Given** a proto file with two services where one has `@Internal` in its comment, **When** the tool runs with annotation filter `Internal`, **Then** the annotated service and all its methods are removed, and the other service remains unchanged.
2. **Given** a proto file with a service annotated at the service level and individual methods annotated within a different service, **When** the tool runs with the annotation filter, **Then** the service-level annotated service is removed entirely, and only the annotated methods within the other service are removed (existing behavior preserved).
3. **Given** a proto file where a service has an annotation that does NOT match the filter config, **When** the tool runs, **Then** the service is kept unchanged.
4. **Given** a proto file where a service is removed by annotation, **When** its request/response message types are not used by any remaining service or method, **Then** those orphaned messages are removed from the output (existing orphan cleanup applies).

---

### Edge Cases

- What happens when a service has multiple annotations and only one matches the filter? The service should be removed (any match is sufficient).
- What happens when a service has annotations on both the service comment and on individual methods? The service-level annotation takes precedence — if the service matches, the entire service is removed regardless of method-level annotations.
- What happens when all services in a file are removed by annotation? The file should be omitted from the output (existing empty-file behavior applies).
- What happens when no annotations config is provided? Service-level annotation filtering is skipped entirely (existing behavior preserved, backward compatible).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The tool MUST check annotations on service-level comments in addition to method-level comments when annotation filtering is configured.
- **FR-002**: If a service comment contains an annotation matching any annotation in the filter config, the tool MUST remove the entire service including all its RPC methods.
- **FR-003**: The tool MUST use the same annotation matching logic for services as it already uses for methods (name-only matching, ignoring parenthesized arguments).
- **FR-004**: The tool MUST continue to apply method-level annotation filtering on services that are NOT removed at the service level.
- **FR-005**: After service removal, the existing orphan detection MUST apply to clean up message types that are no longer referenced.
- **FR-006**: After service removal, the existing empty-file detection MUST apply to skip files with no remaining definitions.
- **FR-007**: The tool MUST NOT change behavior when no annotations config is provided (backward compatible).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Services with matching annotations in their comments are removed entirely from the output, including all their methods.
- **SC-002**: Non-annotated services and services with non-matching annotations remain unchanged in the output.
- **SC-003**: Orphaned messages from removed services are cleaned up automatically.
- **SC-004**: Existing method-level annotation filtering continues to work unchanged for services that are not removed at the service level.
- **SC-005**: All existing tests continue to pass (backward compatibility preserved).
