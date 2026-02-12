# Tasks: Include Annotation Filtering for Top-Level Messages

**Input**: Design documents from `/specs/013-message-include-filter/`
**Prerequisites**: plan.md (required), spec.md (required), research.md

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Core Implementation

**Purpose**: Add `IncludeMessagesByAnnotation` function and integrate it into the pipeline

- [x] T001 [US1] Add `IncludeMessagesByAnnotation` function in internal/filter/filter.go following the `IncludeServicesByAnnotation` pattern. Walk `def.Elements`, check each `*proto.Message` and `*proto.Enum` for matching include annotations using `ExtractAnnotations(msg.Comment)`. Keep elements with matching annotations, remove those without. Return count of removed elements. Non-message/non-enum elements (services, syntax, package, imports, options) pass through unchanged.

- [x] T002 [US1] Update main.go annotation filtering block (~line 194): declare `msgr` variable, add `msgr += filter.IncludeMessagesByAnnotation(pf.def, cfg.Annotations.Include)` inside the `if cfg.HasAnnotationInclude()` block, update orphan removal gate from `if sr > 0 || mr > 0 || fr > 0` to `if sr > 0 || mr > 0 || fr > 0 || msgr > 0`, and add verbose output for message removals.

**Checkpoint**: Core function exists and is wired into the pipeline. No tests yet.

---

## Phase 2: User Story 1 — Unannotated Messages Removed When Include Is Configured (Priority: P1)

**Goal**: Verify that unannotated messages are removed when `annotations.include` is configured in a messages-only file.

**Independent Test**: `go test -race -run TestIncludeMessagesByAnnotation ./internal/filter/`

### Implementation for User Story 1

- [x] T003 [P] [US1] Add unit test `TestIncludeMessagesByAnnotation` in internal/filter/filter_test.go: parse a proto string with three messages (one annotated `[PublishedApi]`, one unannotated, one enum annotated `[PublishedApi]`). Call `IncludeMessagesByAnnotation(def, []string{"PublishedApi"})`. Assert removed count is 1 (the unannotated message), assert the annotated message and enum are kept.

- [x] T004 [P] [US1] Add unit test `TestIncludeMessagesByAnnotation_NoAnnotations` in internal/filter/filter_test.go: parse a proto string with two unannotated messages. Call `IncludeMessagesByAnnotation(def, []string{"PublishedApi"})`. Assert removed count is 2.

- [x] T005 [P] [US1] Add unit test `TestIncludeMessagesByAnnotation_EmptyList` in internal/filter/filter_test.go: parse a proto string with messages. Call `IncludeMessagesByAnnotation(def, []string{})`. Assert removed count is 0 (empty annotation list means no filtering).

- [x] T006 [P] [US1] Create test fixture testdata/messages/messages_annotated.proto: a proto file with `[PublishedApi] Parameters` message referencing `Params1` and `Params2` messages, plus an unreferenced `Unrelated` message. Create golden file testdata/messages/expected/messages_annotated.proto containing only `Parameters`, `Params1`, `Params2` (with `[PublishedApi]` annotation stripped).

- [x] T007 [US1] Add CLI integration test `TestIncludeMessageOnlyCLI` in main_test.go: create a proto file with only unannotated messages, configure `annotations.include: ["NotUsed"]`, run the tool, assert the output file is not written (empty output).

- [x] T008 [US1] Add CLI integration test `TestIncludeAnnotatedMessageCLI` in main_test.go: use the testdata/messages/messages_annotated.proto fixture with `annotations.include: ["PublishedApi"]`, run the tool, compare output against testdata/messages/expected/messages_annotated.proto golden file.

**Checkpoint**: Messages-only include filtering works end-to-end with unit and integration tests.

---

## Phase 3: User Story 2 — Mixed Services and Messages With Include Annotations (Priority: P2)

**Goal**: Verify that include annotations gate both services and messages in mixed files.

**Independent Test**: `go test -race -run TestIncludeMessageMixed ./`

### Implementation for User Story 2

- [x] T009 [US2] Add CLI integration test `TestIncludeMessageMixedCLI` in main_test.go: create a proto file with an annotated service `[PublicApi] OrderService`, the service's dependency messages, and an unannotated standalone message `AuditLog`. Configure `annotations.include: ["PublicApi"]`. Assert that `OrderService` and its dependencies are kept but `AuditLog` is removed from the output.

**Checkpoint**: Mixed services + messages filtering works correctly.

---

## Phase 4: User Story 3 — Combined Include + Exclude With Messages (Priority: P2)

**Goal**: Verify that combined include + exclude mode works correctly with messages.

**Independent Test**: `go test -race -run TestIncludeExcludeMessage ./`

### Implementation for User Story 3

- [x] T010 [US3] Add CLI integration test `TestIncludeExcludeMessageCLI` in main_test.go: create a proto file with `[PublishedApi] Parameters` containing a `[Deprecated]` field. Configure `annotations.include: ["PublishedApi"]` and `annotations.exclude: ["Deprecated"]`. Assert that `Parameters` is kept but the deprecated field is removed from the output.

**Checkpoint**: Combined mode works correctly with messages.

---

## Phase 5: Verification

**Purpose**: Full test suite validation

- [x] T011 Run full test suite with `go test -race ./...` and verify all tests pass, including existing service-level include tests, exclude-only tests, and combined mode tests (confirming FR-004: exclude-only mode unaffected, SC-003: existing tests still pass).

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Core)**: No dependencies — start immediately
- **Phase 2 (US1)**: Depends on Phase 1 (T001 + T002)
- **Phase 3 (US2)**: Depends on Phase 1 (T001 + T002), independent of Phase 2
- **Phase 4 (US3)**: Depends on Phase 1 (T001 + T002), independent of Phases 2-3
- **Phase 5 (Verification)**: Depends on all previous phases

### Task Dependencies

- T001 → T002 (function must exist before pipeline integration)
- T003, T004, T005 can run in parallel after T001 (unit tests for the function)
- T006 must complete before T008 (fixture needed for integration test)
- T007, T008 depend on T002 (pipeline must be wired)
- T009 depends on T002
- T010 depends on T002
- T011 depends on all previous tasks

### Parallel Opportunities

- T003, T004, T005 can all run in parallel after T001 (same file, independent test functions)
- T006 can run in parallel with T003-T005 (different files)
- T007, T009, T010 can run in parallel after T002 (independent CLI tests in main_test.go)

---

## Implementation Strategy

### MVP First (T001 + T002 + T003 + T007 + T011)

1. Add `IncludeMessagesByAnnotation` (T001)
2. Wire into pipeline (T002)
3. Add basic unit test (T003)
4. Add messages-only CLI test (T007)
5. Run full suite (T011)

### Full Delivery

1. T001 → T002 → Core implementation
2. T003 + T004 + T005 + T006 in parallel → Unit tests + fixtures
3. T007 + T008 + T009 + T010 in parallel → CLI integration tests
4. T011 → Full verification

---

## Notes

- `IncludeMessagesByAnnotation` follows the exact same pattern as `IncludeServicesByAnnotation`
- Existing `CollectReferencedTypes` already collects message field references — no changes needed
- Existing `StripAnnotations` already handles message comments — no changes needed
- The orphan removal gate update (`|| msgr > 0`) is critical for messages-only files
- Empty annotation list (`[]string{}`) should return 0 — consistent with `IncludeServicesByAnnotation` behavior
