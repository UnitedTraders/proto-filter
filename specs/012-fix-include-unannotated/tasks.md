# Tasks: Fix Include Filter Keeping Unannotated Services

**Input**: Design documents from `/specs/012-fix-include-unannotated/`
**Prerequisites**: plan.md (required), spec.md (required), research.md

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Core Fix

**Purpose**: Fix the root cause — `IncludeServicesByAnnotation` incorrectly keeping unannotated services

- [x] T001 Remove the `if len(annots) == 0` early-return block in `IncludeServicesByAnnotation` in internal/filter/filter.go (~line 280-285). Delete the entire block so unannotated services fall through to the `hasMatch` check and are correctly excluded.

**Checkpoint**: Core bug is fixed. All downstream tests will need updating to match corrected behavior.

---

## Phase 2: User Story 1 — Unannotated Services Removed When Include Is Configured (Priority: P1)

**Goal**: Verify that unannotated services are removed when `annotations.include` is configured in include-only mode.

**Independent Test**: `go test -race -run TestIncludeServicesByAnnotation ./internal/filter/`

### Implementation for User Story 1

- [x] T002 [US1] Update `TestIncludeServicesByAnnotation` in internal/filter/filter_test.go (~line 1544-1591): change expected removed count from 1 to 2, replace the assertion that `UnannotatedService` should remain with an assertion that it should be removed, and update comments to reflect corrected behavior.

**Checkpoint**: Unit test for `IncludeServicesByAnnotation` passes with corrected expectations.

---

## Phase 3: User Story 2 — Combined Include + Exclude With Unannotated Services (Priority: P1)

**Goal**: Verify that combined include+exclude mode correctly removes unannotated services before applying exclude filtering.

**Independent Test**: `go test -race -run TestCombinedIncludeExclude ./internal/filter/ && go test -race -run TestCombinedIncludeExclude ./`

### Implementation for User Story 2

- [x] T003 [US2] Add a new CLI integration test `TestCombinedIncludeExcludeUnannotatedServiceCLI` in main_test.go that creates a proto file with an unannotated service (matching the bug report scenario: `OrderService` with no annotations, messages with `[Deprecated]` fields), configures both `annotations.include: ["PublishedApi"]` and `annotations.exclude: ["Deprecated"]`, and asserts the output file is either empty or not written.

**Checkpoint**: Combined mode integration test passes, confirming the bug report scenario is fixed.

---

## Phase 4: User Story 3 — Include-Only Mode Consistency (Priority: P2)

**Goal**: Ensure existing include-only golden file tests and CLI integration tests pass with corrected behavior.

**Independent Test**: `go test -race -run TestGoldenFileIncludeService ./internal/filter/ && go test -race -run TestIncludeAnnotationFilteringCLI ./`

### Implementation for User Story 3

- [x] T004 [P] [US3] Add `// [Public]` service-level annotation comment above `service OrderService {` in testdata/annotations/include_service.proto (input fixture) so the service passes the include gate. The golden file testdata/annotations/expected/include_service.proto should already be correct.
- [x] T005 [P] [US3] Verify `TestGoldenFileIncludeService` in internal/filter/filter_test.go (~line 1616) still passes after the input fixture change. If needed, update the test to call `IncludeServicesByAnnotation` before `IncludeMethodsByAnnotation` to match the corrected pipeline.

**Checkpoint**: Include-only mode golden file test and CLI integration test both pass.

---

## Phase 5: Verification

**Purpose**: Full test suite validation

- [x] T006 Run full test suite with `go test -race ./...` and verify all tests pass, including exclude-only tests (confirming FR-005: exclude mode unaffected).

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Core Fix)**: No dependencies — start immediately
- **Phase 2 (US1)**: Depends on Phase 1 (T001)
- **Phase 3 (US2)**: Depends on Phase 1 (T001), independent of Phase 2
- **Phase 4 (US3)**: Depends on Phase 1 (T001), independent of Phases 2-3
- **Phase 5 (Verification)**: Depends on all previous phases

### User Story Dependencies

- **US1 (P1)**: Depends only on T001
- **US2 (P1)**: Depends only on T001 — can run in parallel with US1
- **US3 (P2)**: Depends only on T001 — can run in parallel with US1/US2

### Parallel Opportunities

- T002, T003, T004, T005 can all run in parallel after T001 completes (different files or independent test functions)
- T004 and T005 within US3 can run in parallel (different files)

---

## Implementation Strategy

### MVP First (T001 + T002 + T006)

1. Apply the one-line fix (T001)
2. Update the unit test (T002)
3. Run full suite (T006) to identify any remaining failures
4. Fix remaining test failures as discovered

### Full Delivery

1. T001 → Core fix
2. T002 + T003 + T004 + T005 in parallel → All test updates
3. T006 → Full verification

---

## Notes

- This is a bug fix with minimal scope: 1 code change + test updates
- The core fix (T001) is a deletion of 5 lines, not addition of new code
- All user stories share the same root fix; their tasks differ only in test coverage
- No new dependencies, no new files, no API changes
