# Feature Specification: Cross-File Orphan Detection Tests

**Feature Branch**: `005-cross-file-orphan-tests`
**Created**: 2026-02-09
**Status**: Draft
**Input**: User description: "Messages for service methods can be included from other 'common' files. Please create tests that check situation when messages used from external files do not become orphaned."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Verify Cross-File Message Types Survive Annotation Filtering (Priority: P1)

As a developer using proto-filter with annotation-based filtering, I want to be confident that message types defined in a separate "common" or "shared" proto file are NOT incorrectly removed as orphans when they are still referenced by surviving methods in other files. This is a test-only feature that adds regression tests for existing cross-file orphan detection behavior.

For example, given two files:
- `common.proto` containing shared message types (`Timestamp`, `Money`, `Pagination`)
- `service.proto` containing services whose RPC methods reference those shared types

If annotation filtering removes some methods from `service.proto` but other surviving methods still reference types from `common.proto`, those shared types must remain in the output.

**Why this priority**: This is the only capability of this feature — adding test coverage for a critical correctness guarantee. Cross-file message references are a common proto pattern and incorrect orphan removal would produce invalid output.

**Independent Test**: Create test fixtures with cross-file message references, run annotation filtering, and verify that referenced messages from common files are preserved while truly unreferenced messages are removed.

**Acceptance Scenarios**:

1. **Given** a service file with methods referencing message types from a common file, and annotation filtering removes some methods, **When** the tool runs, **Then** message types from the common file that are still referenced by surviving methods remain in the output.
2. **Given** a service file where ALL methods referencing a particular common message type are removed by annotation, **When** the tool runs, **Then** that common message type is removed from the output (it is truly orphaned).
3. **Given** a service file with methods referencing message types from a common file, and service-level annotation filtering removes an entire service, **When** the tool runs, **Then** common message types referenced only by the removed service are removed, while types referenced by surviving services remain.
4. **Given** multiple service files referencing the same common message type, and annotation filtering removes all methods in one file but leaves methods in another, **When** the tool runs, **Then** the common message type remains in the output because it is still referenced.

---

### Edge Cases

- What happens when a common file contains only message/enum types and no services? The file should be preserved if any of its types are referenced by surviving methods in other files.
- What happens when a common message type is transitively referenced (e.g., a surviving method's response type contains a field of the common type)? The common type should be preserved via transitive dependency resolution.
- What happens when annotation filtering is combined with name-based include/exclude filtering, and the common file's types are needed by included services? The common types must be preserved.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Tests MUST verify that message types defined in a separate "common" proto file are preserved in output when they are still referenced by surviving RPC methods after annotation-based filtering.
- **FR-002**: Tests MUST verify that message types from a common file ARE removed when no surviving method or message references them after annotation filtering.
- **FR-003**: Tests MUST cover both method-level and service-level annotation filtering in combination with cross-file message references.
- **FR-004**: Tests MUST verify the end-to-end CLI pipeline (not just unit-level filter functions) to ensure cross-file references are handled correctly in the full processing flow.
- **FR-005**: Tests MUST NOT require any changes to existing production code — this feature adds test coverage only.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: At least one test verifies that common-file message types referenced by surviving methods are preserved after annotation filtering.
- **SC-002**: At least one test verifies that common-file message types NOT referenced by any surviving method are correctly removed.
- **SC-003**: At least one test exercises the full CLI pipeline with cross-file annotation filtering and verifies correct output.
- **SC-004**: All new tests pass with `go test -race ./...`.
- **SC-005**: All existing tests continue to pass (no regressions).
