# Tasks: Service-Level Annotation Filtering

**Input**: Design documents from `/specs/004-service-annotation-filter/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

**Tests**: Included per Constitution Principle IV (Test-Driven Development). Uses golden file comparison pattern.

**Organization**: Single user story (US1 - P1). Tasks grouped into setup, core implementation, pipeline wiring, and polish phases.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1)
- Include exact file paths in descriptions

## Phase 1: Setup (Test Fixtures)

**Purpose**: Create input fixtures for service-level annotation filtering tests

- [x] T001 Create test fixture testdata/annotations/service_annotated.proto with two services: `AdminService` with `@Internal` annotation on the service comment and two RPC methods, and `OrderService` without annotations and one RPC method, plus all request/response message types
- [x] T002 [P] Create test fixture testdata/annotations/mixed_annotations.proto with one service annotated `@Internal` at the service level and another service with a method annotated `@HasAnyRole` at the method level, plus all request/response message types (tests acceptance scenario 2: service-level + method-level coexistence)
- [x] T003 [P] Create golden file testdata/annotations/expected/service_annotated.proto with expected output after filtering with annotation `Internal`: only `OrderService` and its message types remain
- [x] T004 Create golden file testdata/annotations/expected/mixed_annotations.proto with expected output after filtering with annotations `Internal` and `HasAnyRole`: service-annotated service removed entirely, method-annotated method removed from remaining service, orphaned messages cleaned up

---

## Phase 2: Foundational (Core FilterServicesByAnnotation Function)

**Purpose**: Implement `FilterServicesByAnnotation` and its unit tests

- [x] T005 [US1] Write unit test `TestFilterServicesByAnnotation` in internal/filter/filter_test.go that parses testdata/annotations/service_annotated.proto, calls `FilterServicesByAnnotation` with annotation `Internal`, and verifies: 1 service removed, `AdminService` gone, `OrderService` remains with its methods
- [x] T006 [US1] Write unit test `TestFilterServicesByAnnotationNoMatch` in internal/filter/filter_test.go that parses testdata/annotations/service_annotated.proto, calls `FilterServicesByAnnotation` with annotation `NonExistent`, and verifies 0 services removed and both services remain
- [x] T007 [P] [US1] Write unit test `TestFilterServicesByAnnotationMultipleAnnotations` in internal/filter/filter_test.go that creates a proto AST with a service comment containing `@Internal` and `@Deprecated`, calls `FilterServicesByAnnotation` with only `Internal`, and verifies the service is removed (any match is sufficient)
- [x] T008 [US1] Implement `FilterServicesByAnnotation(def *proto.Proto, annotations []string) int` in internal/filter/filter.go that iterates `def.Elements`, checks `Service.Comment` using existing `ExtractAnnotations`, removes services with matching annotations, and returns count of removed services

**Checkpoint**: `FilterServicesByAnnotation` function works on AST level. Unit tests pass.

---

## Phase 3: User Story 1 - Remove Services by Annotation (Priority: P1)

**Goal**: End-to-end service-level annotation filtering with orphan cleanup, empty-file handling, and golden file validation

**Independent Test**: Run the tool with annotation filter config against proto files containing annotated and non-annotated services. Verify annotated services are removed entirely, non-annotated services remain, and orphaned messages are cleaned up.

### Tests for User Story 1

- [x] T009 [US1] Write golden file comparison test `TestGoldenFileServiceAnnotated` in internal/filter/filter_test.go that parses testdata/annotations/service_annotated.proto, calls `FilterServicesByAnnotation` then `RemoveOrphanedDefinitions`, writes output via `writer.WriteProtoFile`, and compares byte-for-byte against testdata/annotations/expected/service_annotated.proto
- [x] T010 [P] [US1] Write golden file comparison test `TestGoldenFileMixedAnnotations` in internal/filter/filter_test.go that parses testdata/annotations/mixed_annotations.proto, calls `FilterServicesByAnnotation` then `FilterMethodsByAnnotation` then `RemoveEmptyServices` then `RemoveOrphanedDefinitions`, writes output, and compares against testdata/annotations/expected/mixed_annotations.proto
- [x] T011 [P] [US1] Write unit test `TestFilterServicesByAnnotationAllServicesRemoved` in internal/filter/filter_test.go that parses testdata/annotations/service_annotated.proto, annotates both services (programmatically add `@Internal` to OrderService comment), calls `FilterServicesByAnnotation`, then verifies `HasRemainingDefinitions` returns false after orphan removal (tests empty-file edge case)
- [x] T012 [US1] Write unit test `TestFilterServicesByAnnotationEmptyAnnotationList` in internal/filter/filter_test.go that calls `FilterServicesByAnnotation` with empty annotations slice and verifies 0 services removed (backward compatibility per FR-007)

**Checkpoint**: All unit and golden file tests pass. Service-level annotation filtering is correct in isolation.

---

## Phase 4: Pipeline Wiring & CLI Integration

**Purpose**: Wire `FilterServicesByAnnotation` into main.go pipeline and verify end-to-end via CLI

- [x] T013 [US1] Wire `filter.FilterServicesByAnnotation(pf.def, cfg.Annotations)` call into main.go annotation filtering block, placed before the existing `filter.FilterMethodsByAnnotation` call. Add `servicesRemoved` counter variable.
- [x] T014 [US1] Update verbose output in main.go to include services-removed-by-annotation count in the annotation summary line: `"removed %d services by annotation, %d methods by annotation, %d orphaned definitions"`
- [x] T015 [US1] Write CLI integration test `TestServiceAnnotationFilteringCLI` in main_test.go that runs proto-filter with testdata/annotations/ as input and a config with `annotations: ["Internal"]`, then verifies: service_annotated.proto output contains only `OrderService` and its message types, and `AdminService` is absent
- [x] T016 [P] [US1] Write CLI integration test `TestMixedServiceMethodAnnotationFilteringCLI` in main_test.go that runs proto-filter with testdata/annotations/ as input and a config with `annotations: ["Internal", "HasAnyRole"]`, then verifies both service-level and method-level filtering work together in a single run
- [x] T017 [US1] Write CLI integration test `TestServiceAnnotationFilteringVerbose` in main_test.go that runs proto-filter with `--verbose` and service-annotated fixtures, verifying the verbose output includes "services by annotation" count

**Checkpoint**: Full end-to-end pipeline works. `go test -race ./...` passes across all packages.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Validate all success criteria, backward compatibility, and run final checks

- [x] T018 Run `go test -race ./...` and verify all tests pass across all packages (SC-005: backward compatibility)
- [x] T019 Run quickstart.md validation: process a proto file with service-level annotations through the CLI and verify output matches expected format

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies — T001 first, then T002-T004 parallel (T003 depends on T001 content, T004 depends on T002 content)
- **Phase 2 (Foundational)**: Depends on Phase 1; T005-T007 (tests) before T008 (implementation)
- **Phase 3 (US1 Golden Files)**: Depends on Phase 2; T009-T012 after T008
- **Phase 4 (Pipeline)**: Depends on Phases 2-3; T013-T014 first (wiring), then T015-T017 (CLI tests)
- **Phase 5 (Polish)**: Depends on all previous phases

### User Story Dependencies

- **US1 (Remove Services by Annotation)**: Single user story — no cross-story dependencies

### Parallel Opportunities

- T002, T003 can run in parallel (different files)
- T007 can run in parallel with T005, T006 (different test, programmatic AST)
- T010, T011 can run in parallel with T009 (different test cases)
- T016 can run in parallel with T015 (different CLI tests)

---

## Parallel Example: Phase 1

```bash
# After T001 is created:
Task: "Create mixed_annotations.proto fixture in testdata/annotations/mixed_annotations.proto"
Task: "Create golden file testdata/annotations/expected/service_annotated.proto"
```

---

## Implementation Strategy

### MVP First (Phases 1-3)

1. Complete Phase 1: Create fixtures and golden files
2. Complete Phase 2: Implement `FilterServicesByAnnotation` with unit tests
3. Complete Phase 3: Golden file comparison tests
4. **STOP and VALIDATE**: All golden file tests pass

### Full Delivery

1. Phases 1-3 → Unit + golden file tests pass (MVP)
2. Phase 4 → CLI integration tests pass
3. Phase 5 → Full validation complete

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story
- Golden files are the primary validation mechanism for output correctness
- Constitution IV requires tests before/alongside implementation
- Commit after each phase completion
- This feature reuses existing `ExtractAnnotations`, `RemoveOrphanedDefinitions`, `RemoveEmptyServices`, and `HasRemainingDefinitions` — no modifications to those functions
