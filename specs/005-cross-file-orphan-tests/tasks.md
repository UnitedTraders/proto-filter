# Tasks: Cross-File Orphan Detection Tests

**Input**: Design documents from `/specs/005-cross-file-orphan-tests/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

**Tests**: This feature IS tests — all tasks produce test code and test fixtures. Constitution Principle IV satisfied by nature of the feature.

**Organization**: Single user story (US1 - P1). Tasks grouped into setup, unit tests, CLI integration tests, and polish phases.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1)
- Include exact file paths in descriptions

## Phase 1: Setup (Test Fixtures)

**Purpose**: Create the `testdata/crossfile/` directory and all proto fixture files

- [x] T001 Create test fixture testdata/crossfile/common.proto with shared message types only (no services): `Pagination` (fields: page, per_page), `Money` (fields: currency, amount), `ErrorDetail` (fields: code, message). Package: `crossfile`. Syntax: `proto3`.
- [x] T002 Create test fixture testdata/crossfile/orders.proto with `OrderService` containing two methods: `ListOrders(ListOrdersRequest) returns (ListOrdersResponse)` (no annotation) and `GetOrderDetails(GetOrderDetailsRequest) returns (GetOrderDetailsResponse)` with `@HasAnyRole` annotation. Request/response types defined locally: `ListOrdersRequest` uses `Pagination` from common, `ListOrdersResponse` uses `Money` from common, `GetOrderDetailsRequest` has an `order_id` field, `GetOrderDetailsResponse` has a `status` field. Package: `crossfile`. Import: `common.proto`.
- [x] T003 [P] Create test fixture testdata/crossfile/payments.proto with `PaymentService` annotated `@Internal` at the service level, containing one method: `ProcessPayment(ProcessPaymentRequest) returns (ProcessPaymentResponse)`. `ProcessPaymentRequest` uses `Money` from common, `ProcessPaymentResponse` uses `ErrorDetail` from common. Package: `crossfile`. Import: `common.proto`.
- [x] T004 Create golden file testdata/crossfile/expected/common.proto with expected output after filtering with annotation `Internal`: `Pagination` and `Money` remain (referenced by surviving `OrderService`), `ErrorDetail` is removed (only referenced by removed `PaymentService`). Note: also filter `HasAnyRole` — `GetOrderDetails` method is removed, but `ListOrders` survives referencing `Pagination` and `Money`.
- [x] T005 [P] Create golden file testdata/crossfile/expected/orders.proto with expected output after filtering with annotations `Internal` and `HasAnyRole`: `OrderService` remains with only `ListOrders` method (the `GetOrderDetails` method is removed by `@HasAnyRole`), `ListOrdersRequest` and `ListOrdersResponse` remain, `GetOrderDetailsRequest` and `GetOrderDetailsResponse` are removed as orphans.
- [x] T006 [P] N/A — payments.proto is excluded from output (no remaining definitions after PaymentService removal) with expected output after filtering: `PaymentService` is removed by `@Internal`, `ProcessPaymentRequest` and `ProcessPaymentResponse` are removed as orphans. File should contain only syntax, package, and import (or be excluded from output if `HasRemainingDefinitions` returns false).

**Checkpoint**: All fixture and golden files are in place. No code changes yet.

---

## Phase 2: User Story 1 - Verify Cross-File Message Types Survive Annotation Filtering (Priority: P1)

**Goal**: Verify that shared message types from common proto files are preserved when still referenced by surviving services, and removed when orphaned by annotation filtering.

**Independent Test**: Run annotation filtering against `testdata/crossfile/` fixtures and verify output matches golden files.

### Unit Tests (internal/filter/filter_test.go)

- [x] T007 [US1] Write unit test `TestCrossFileAnnotationFilterPreservesReferencedTypes` in internal/filter/filter_test.go that: (1) parses all files in testdata/crossfile/, (2) builds a dependency graph, (3) applies the full filtering pipeline (ApplyFilter → FilterServicesByAnnotation → FilterMethodsByAnnotation → RemoveEmptyServices → RemoveOrphanedDefinitions) with annotations `["Internal", "HasAnyRole"]`, and (4) verifies `Pagination` and `Money` survive in common.proto output while `ErrorDetail` is removed (FR-001, FR-002, SC-001, SC-002)
- [x] T008 [P] [US1] Write unit test `TestCrossFileServiceRemovalOrphansCommonTypes` in internal/filter/filter_test.go that: (1) parses all files in testdata/crossfile/, (2) applies filtering with annotation `["Internal"]` only (service-level filtering), and (3) verifies `ErrorDetail` is removed from common.proto (only referenced by PaymentService) while `Money` and `Pagination` survive (referenced by OrderService) (FR-003, acceptance scenario 3)
- [x] T009 [P] [US1] Write unit test `TestCrossFileCommonFileNoServicesPreserved` in internal/filter/filter_test.go that: (1) parses testdata/crossfile/common.proto independently, (2) verifies it has no services, (3) applies annotation filtering and verifies no types are removed from common.proto itself (since it has no services, annotation filtering has no effect on it — per research Decision 1, edge case 1)

### CLI Integration Tests (main_test.go)

- [x] T010 [US1] Write CLI integration test `TestCrossFileAnnotationFilteringCLI` in main_test.go that: (1) runs proto-filter with `--input testdata/crossfile/ --output <tmpdir> --config <config>` where config has `annotations: ["Internal", "HasAnyRole"]`, (2) reads output files from tmpdir, (3) verifies common.proto output contains `Pagination` and `Money` but NOT `ErrorDetail`, (4) verifies orders.proto output contains `OrderService` with `ListOrders` only, (5) verifies payments.proto is excluded from output or empty (FR-004, SC-003)
- [x] T011 [P] [US1] Write CLI integration test `TestCrossFileSharedTypesSurvivePartialServiceRemoval` in main_test.go that: (1) runs proto-filter with only `annotations: ["Internal"]` (no method filtering), (2) verifies `Money` and `Pagination` survive in common.proto (referenced by surviving OrderService), (3) verifies `ErrorDetail` is removed (only referenced by removed PaymentService), (4) verifies orders.proto is fully intact with all methods (acceptance scenario 4)
- [x] T012 [US1] Write golden file comparison CLI test `TestCrossFileGoldenFileComparison` in main_test.go that: (1) runs proto-filter end-to-end with testdata/crossfile/ input and annotations `["Internal", "HasAnyRole"]`, (2) compares each output file byte-for-byte against corresponding testdata/crossfile/expected/ golden files (SC-003)

**Checkpoint**: All unit and CLI integration tests pass. Cross-file orphan detection is verified at both levels.

---

## Phase 3: Polish & Cross-Cutting Concerns

**Purpose**: Validate all success criteria and ensure no regressions

- [x] T013 Run `go test -race ./...` and verify ALL tests pass across all packages — both new cross-file tests and all existing tests (SC-004, SC-005)
- [x] T014 Run quickstart.md validation: manually verify the example scenario from quickstart.md produces the documented expected behavior by running the CLI against testdata/crossfile/ fixtures

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies — T001 first, then T002-T006 (T002/T003 depend on T001 for common types, T004-T006 depend on T001-T003 for golden file content)
- **Phase 2 (US1)**: Depends on Phase 1; unit tests T007-T009 can start after fixtures are ready; CLI tests T010-T012 after unit tests pass
- **Phase 3 (Polish)**: Depends on Phase 2 completion

### User Story Dependencies

- **US1 (Verify Cross-File Message Types Survive Annotation Filtering)**: Single user story — no cross-story dependencies

### Parallel Opportunities

- T003 can run in parallel with T002 (different files)
- T005, T006 can run in parallel with T004 (different golden files)
- T008, T009 can run in parallel with T007 (different test cases in same file)
- T011 can run in parallel with T010 (different CLI test cases)

---

## Parallel Example: Phase 1

```bash
# After T001 (common.proto) is created:
Task: "Create orders.proto fixture in testdata/crossfile/orders.proto"
Task: "Create payments.proto fixture in testdata/crossfile/payments.proto"

# After T001-T003 are created:
Task: "Create golden file testdata/crossfile/expected/common.proto"
Task: "Create golden file testdata/crossfile/expected/orders.proto"
Task: "Create golden file testdata/crossfile/expected/payments.proto"
```

---

## Implementation Strategy

### MVP First (Phases 1-2)

1. Complete Phase 1: Create all fixtures and golden files
2. Complete Phase 2: Unit tests + CLI integration tests
3. **STOP and VALIDATE**: All tests pass with `go test -race ./...`

### Full Delivery

1. Phase 1 → Fixtures ready
2. Phase 2 → All tests pass (unit + CLI)
3. Phase 3 → Full validation complete, no regressions

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story
- Golden files are the primary validation mechanism for output correctness
- This feature adds tests only — NO production code changes (FR-005)
- The dependency graph and `RemoveOrphanedDefinitions` work together to handle cross-file orphans correctly (see research.md Decision 1)
- Common files with only messages are unaffected by annotation filtering since they have no services
- Commit after each phase completion
