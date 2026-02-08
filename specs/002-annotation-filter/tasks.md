# Tasks: Annotation-Based Method Filtering

**Input**: Design documents from `/specs/002-annotation-filter/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Required by constitution principle IV (Test-Driven Development).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Single project**: `main.go` at root, `internal/` packages, `testdata/` fixtures

---

## Phase 1: Setup

**Purpose**: Test fixtures for annotation-based filtering scenarios

- [x] T001 Create test fixture testdata/annotations/service.proto: a service with 3 RPC methods — two annotated with `@HasAnyRole` (one with args, one without) and one non-annotated; include request/response messages for each method
- [x] T002 [P] Create test fixture testdata/annotations/shared.proto: a service with methods where some request/response messages are shared between annotated and non-annotated methods, plus a shared enum referenced by a kept message
- [x] T003 [P] Create test fixture testdata/annotations/internal_only.proto: a service where ALL methods have `@HasAnyRole` annotation, plus their request/response messages (all should become orphaned)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Extend config to support annotations — MUST be complete before any user story

**CRITICAL**: No user story work can begin until this phase is complete

- [x] T004 Add `Annotations []string` field with `yaml:"annotations"` tag to FilterConfig struct in internal/config/config.go; update IsPassThrough() to also check len(Annotations) == 0
- [x] T005 [P] Add unit tests for annotation config loading in internal/config/config_test.go: test loading config with annotations list, test empty annotations (pass-through), test config with only annotations (no include/exclude)

**Checkpoint**: Config can load and represent annotation filter rules. All existing tests still pass.

---

## Phase 3: User Story 1 — Filter RPC Methods by Annotation (Priority: P1) MVP

**Goal**: Parse annotations from RPC method comments and remove methods carrying specified annotations from output.

**Independent Test**: Run tool with `annotations: ["HasAnyRole"]` against testdata/annotations/service.proto. Verify annotated methods are removed, non-annotated methods remain.

### Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T006 [P] [US1] Unit test for ExtractAnnotations function in internal/filter/filter_test.go: test extracting annotation names from comment lines — `@HasAnyRole`, `@HasAnyRole({"ADMIN"})`, `@com.example.Secure`, no annotations, nil comment, block comment with annotation
- [x] T007 [P] [US1] Unit test for FilterMethodsByAnnotation function in internal/filter/filter_test.go: test against a parsed service AST — verify annotated methods are removed, non-annotated are kept, service with no matching annotations is unchanged
- [x] T008 [P] [US1] Integration test for annotation method filtering pipeline in internal/filter/filter_test.go (TestIntegrationAnnotationFilter): parse testdata/annotations/service.proto, apply annotation filter for "HasAnyRole", verify output has only the non-annotated method with its messages

### Implementation for User Story 1

- [x] T009 [US1] Implement ExtractAnnotations function in internal/filter/filter.go: take a *proto.Comment, return []string of annotation names using regex `@(\w[\w.]*)` against Comment.Lines
- [x] T010 [US1] Implement FilterMethodsByAnnotation function in internal/filter/filter.go: take a *proto.Proto, package name, and annotation names list; iterate service elements, remove RPC methods whose comments contain any matching annotation; return count of removed methods
- [x] T011 [US1] Verify all US1 tests pass with `go test ./internal/filter/...`

**Checkpoint**: Tool can parse annotations from comments and remove matching RPC methods from services.

---

## Phase 4: User Story 2 — Remove Orphaned Messages (Priority: P2)

**Goal**: After method filtering, detect and remove messages/enums no longer referenced by any remaining method or message.

**Independent Test**: Run tool against testdata/annotations/shared.proto with annotation filter. Verify orphaned messages are removed, shared messages are kept, transitive orphans are removed.

### Tests for User Story 2

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T012 [P] [US2] Unit test for CollectReferencedTypes function in internal/filter/filter_test.go: test collecting all FQNs referenced by remaining RPC methods and message fields from a parsed AST
- [x] T013 [P] [US2] Unit test for RemoveOrphanedDefinitions function in internal/filter/filter_test.go: test that unreferenced messages/enums are removed; shared messages are preserved; transitive orphans are removed; external imports are never orphaned
- [x] T014 [P] [US2] Integration test for orphan removal pipeline in internal/filter/filter_test.go (TestIntegrationOrphanRemoval): parse testdata/annotations/shared.proto, filter by annotation, remove orphans, verify only referenced definitions remain

### Implementation for User Story 2

- [x] T015 [US2] Implement CollectReferencedTypes function in internal/filter/filter.go: walk a *proto.Proto AST, collect all message/enum FQNs referenced by remaining RPC request/response types and message field types; return map[string]bool of referenced FQNs
- [x] T016 [US2] Implement RemoveOrphanedDefinitions function in internal/filter/filter.go: take a *proto.Proto, package name, and set of all known FQNs; iteratively collect referenced types and remove unreferenced messages/enums until stable (transitive orphan detection); return count of removed definitions
- [x] T017 [US2] Verify all US2 tests pass with `go test ./internal/filter/...`

**Checkpoint**: Orphaned messages and enums are automatically cleaned up after method filtering.

---

## Phase 5: User Story 3 — Remove Empty Services (Priority: P3)

**Goal**: Remove service definitions that have zero remaining methods after annotation filtering. Skip writing files with no remaining definitions.

**Independent Test**: Run tool against testdata/annotations/internal_only.proto with annotation filter. Verify the entire service is removed and the file is not written.

### Tests for User Story 3

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T018 [P] [US3] Unit test for RemoveEmptyServices function in internal/filter/filter_test.go: test that a service with zero methods after filtering is removed from elements; a service with remaining methods is kept; non-service elements are unaffected
- [x] T019 [P] [US3] Unit test for HasRemainingDefinitions function in internal/filter/filter_test.go: test that a *proto.Proto with no services/messages/enums returns false; one with any definition returns true
- [x] T020 [P] [US3] Integration test for empty service removal in internal/filter/filter_test.go (TestIntegrationEmptyServiceRemoval): parse testdata/annotations/internal_only.proto, filter by annotation, verify service is removed, all messages are orphaned and removed

### Implementation for User Story 3

- [x] T021 [US3] Implement RemoveEmptyServices function in internal/filter/filter.go: iterate *proto.Proto elements, remove services with zero RPC method children; return count of removed services
- [x] T022 [US3] Implement HasRemainingDefinitions function in internal/filter/filter.go: check if a *proto.Proto has any remaining service, message, or enum elements (ignoring syntax/package/imports/options)
- [x] T023 [US3] Verify all US3 tests pass with `go test ./internal/filter/...`

**Checkpoint**: Empty services are removed and empty files are skipped.

---

## Phase 6: Pipeline Integration & Polish

**Purpose**: Wire annotation filtering into main.go pipeline, add verbose output, run full validation

- [x] T024 Wire annotation filtering pipeline in main.go: after existing PruneAST, if config has annotations: call FilterMethodsByAnnotation on each parsed file → call RemoveEmptyServices → call RemoveOrphanedDefinitions → check HasRemainingDefinitions before writing; track counts for verbose output
- [x] T025 Update verbose output in main.go: when --verbose is set and annotations were configured, print "proto-filter: removed N methods by annotation, M orphaned definitions" to stderr per contracts/cli-interface.md
- [x] T026 [P] Add integration test in main_test.go: build binary, run with testdata/annotations/ and a config file containing `annotations: ["HasAnyRole"]`, verify correct exit code 0, verify verbose output contains method/orphan counts
- [x] T027 [P] Add backward compatibility test in main_test.go: run with existing testdata/filter/ and a config with no annotations key, verify output is identical to before (SC-004)
- [x] T028 Run full test suite with race detection: `go test -race ./...`
- [x] T029 Run quickstart.md validation: create temp proto files per quickstart example, run tool, verify output matches expected

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion — BLOCKS all user stories
- **User Story 1 (Phase 3)**: Depends on Foundational (Phase 2) — core annotation parsing
- **User Story 2 (Phase 4)**: Depends on User Story 1 (Phase 3) — needs method filtering to produce orphans
- **User Story 3 (Phase 5)**: Depends on User Story 1 (Phase 3) — needs method filtering to produce empty services; can run in parallel with US2
- **Polish (Phase 6)**: Depends on all user stories being complete

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Core functions before integration
- Story complete before moving to next priority

### Parallel Opportunities

- Phase 1: T002, T003 can run in parallel (independent fixture files)
- Phase 2: T005 can run in parallel with T004 (test file vs source file)
- Phase 3 tests: T006, T007, T008 can all run in parallel
- Phase 4 tests: T012, T013, T014 can all run in parallel
- Phase 5 tests: T018, T019, T020 can all run in parallel
- Phase 5 can run in parallel with Phase 4 (both depend only on Phase 3)
- Phase 6: T026, T027 can run in parallel

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001–T003)
2. Complete Phase 2: Foundational (T004–T005)
3. Complete Phase 3: User Story 1 (T006–T011)
4. **STOP and VALIDATE**: `go test ./...` passes, annotated methods are removed from output

### Incremental Delivery

1. Setup + Foundational → Config ready
2. User Story 1 → Annotation method filtering works (MVP!)
3. User Story 2 → Orphaned messages cleaned up
4. User Story 3 → Empty services removed
5. Polish → Pipeline wired, verbose output, full validation

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story
- Constitution principle IV requires TDD: tests before implementation
- All file paths are relative to repository root
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Existing tests MUST continue to pass at every checkpoint (backward compatibility)
