# Tasks: Combined Include and Exclude Annotation Filters

**Input**: Design documents from `/specs/011-combined-include-exclude/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/cli-interface.md

**Tests**: Required by project constitution (Principle IV: Test-Driven Development).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

---

## Phase 1: Foundational (Remove Mutual Exclusivity)

**Purpose**: Remove the config validation that rejects combined include+exclude configs. This is the shared prerequisite for all user stories.

- [x] T001 Remove the mutual exclusivity check from `Validate()` in `internal/config/config.go:73-76` — delete the `if` block that returns the "mutually exclusive" error, leaving `Validate()` with just `return nil`
- [x] T002 Update `TestValidateMutualExclusivity` in `internal/config/config_test.go:217-231` — rename to `TestValidateCombinedIncludeExcludePass` and change to expect no error when both include and exclude are populated
- [x] T003 Update `TestMutualExclusivityErrorCLI` in `main_test.go:1405-1424` — rename to `TestCombinedIncludeExcludeCLI` and change to expect exit code 0 (success) instead of exit code 2 (error)
- [x] T004 Run all existing tests (`go test -race ./...`) and verify they pass unchanged with the validation removed (SC-002 backward compatibility gate)

**Checkpoint**: Combined configs are accepted, all existing tests pass unchanged

---

## Phase 2: User Story 2 — Message Field Annotation Filtering (Priority: P2)

**Goal**: Add `FilterFieldsByAnnotation` function that removes individual message fields whose comments contain a matching exclude annotation

**Independent Test**: Create a proto with annotated fields, configure exclude, verify annotated fields are removed while others remain

**Note**: US2 is implemented before US1 because the combined filtering story (US1) depends on field filtering being available. US2 is independently testable with exclude-only configs.

### Tests for User Story 2

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T005 [P] [US2] Write unit test `TestFilterFieldsByAnnotationNormalField` in `internal/filter/filter_test.go`: create programmatic AST with a message containing three `NormalField` fields, one with `// [Deprecated]` comment, apply `FilterFieldsByAnnotation(def, ["Deprecated"])`, verify the annotated field is removed and two remain, return count = 1
- [x] T006 [P] [US2] Write unit test `TestFilterFieldsByAnnotationAtSignSyntax` in `internal/filter/filter_test.go`: create AST with a message field annotated `// @Deprecated`, apply `FilterFieldsByAnnotation(def, ["Deprecated"])`, verify the field is removed (FR-004: both syntaxes work)
- [x] T007 [P] [US2] Write unit test `TestFilterFieldsByAnnotationInlineComment` in `internal/filter/filter_test.go`: create AST with a `NormalField` that has an `InlineComment` containing `[Deprecated]` (not a leading `Comment`), verify the field is removed based on inline comment annotation
- [x] T008 [P] [US2] Write unit test `TestFilterFieldsByAnnotationMapField` in `internal/filter/filter_test.go`: create AST with a `MapField` annotated `// [Deprecated]`, verify it is removed
- [x] T009 [P] [US2] Write unit test `TestFilterFieldsByAnnotationOneofField` in `internal/filter/filter_test.go`: create AST with a `Oneof` container containing two `OneOfField` entries, one annotated `// [Deprecated]`, verify the annotated field is removed from the oneof and the other remains
- [x] T010 [P] [US2] Write unit test `TestFilterFieldsByAnnotationNoMatch` in `internal/filter/filter_test.go`: create AST with fields that have no annotations, apply `FilterFieldsByAnnotation`, verify all fields remain and return count = 0
- [x] T011 [P] [US2] Write unit test `TestFilterFieldsByAnnotationAllFieldsRemoved` in `internal/filter/filter_test.go`: create AST where all fields in a message are annotated, verify all are removed and message remains with empty elements, return count = total fields
- [x] T012 [P] [US2] Write CLI integration test `TestFieldFilteringCLI` in `main_test.go`: create temp proto file with a message containing a `// [Deprecated]` field, config with `annotations.exclude: [Deprecated]`, verify output file has the field removed

### Implementation for User Story 2

- [x] T013 [US2] Implement `FilterFieldsByAnnotation(def *proto.Proto, annotations []string) int` in `internal/filter/filter.go`: iterate all `*proto.Message` elements in `def.Elements`, for each message iterate `m.Elements` handling `*proto.NormalField`, `*proto.MapField`, and `*proto.Oneof` (with inner `*proto.OneOfField`), check both `.Comment` and `.InlineComment` via `ExtractAnnotations`, remove matching fields, return total count removed
- [x] T014 [US2] Update annotation filtering block in `main.go:191-210` to call `filter.FilterFieldsByAnnotation(pf.def, cfg.Annotations.Exclude)` when `cfg.HasAnnotationExclude()` is true, add `fieldsRemoved` counter alongside `servicesRemoved` and `methodsRemoved`, include `fieldsRemoved` in orphan removal condition
- [x] T015 [US2] Update verbose output in `main.go:281` to include field count: change format to `"proto-filter: removed %d services by annotation, %d methods by annotation, %d fields by annotation, %d orphaned definitions"` (FR-010)
- [x] T016 [US2] Run all tests (`go test -race ./...`) and verify T005-T012 pass; verify all existing tests pass unchanged

**Checkpoint**: Field-level annotation filtering works with exclude-only configs

---

## Phase 3: User Story 1 — Combined Include + Exclude Filtering (Priority: P1)

**Goal**: Enable combined include+exclude annotation filtering by changing the orchestration from `if/else if` to sequential `if/if`

**Independent Test**: Configure both `annotations.include: [PublicApi]` and `annotations.exclude: [Deprecated]`, run proto-filter, verify include narrows first, then exclude refines

### Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T017 [P] [US1] Write unit test `TestCombinedIncludeExcludeFiltering` in `internal/filter/filter_test.go`: create AST with a service annotated `[PublicApi]` containing two RPCs, one RPC annotated `[Deprecated]`, call include then exclude sequentially on the AST, verify the `PublicApi` service remains but the `Deprecated` method is removed
- [x] T018 [P] [US1] Write unit test `TestCombinedIncludeExcludeWithFieldFiltering` in `internal/filter/filter_test.go`: create AST with a service annotated `[PublicApi]` and a message with a field annotated `[Deprecated]`, apply include (service level), then exclude (field level), verify service included and deprecated field removed
- [x] T019 [P] [US1] Write unit test `TestCombinedSameAnnotationInBothLists` in `internal/filter/filter_test.go`: create AST with a service annotated `[PublicApi]`, config include=[PublicApi] and exclude=[PublicApi], verify the service is included then immediately excluded — net result: removed
- [x] T020 [P] [US1] Write CLI integration test `TestCombinedIncludeExcludeCLI` in `main_test.go`: create temp proto matching the quickstart example (service with `[PublicApi]`, message field with `[Deprecated]`), config with include=[PublicApi] and exclude=[Deprecated], verify output matches expected (service included, field removed)
- [x] T021 [P] [US1] Write CLI integration test `TestCombinedIncludeExcludeMethodRemoval` in `main_test.go`: create proto with `[PublicApi]` service containing a `[Deprecated]` method and a non-annotated method, combined config, verify deprecated method removed but non-annotated method remains
- [x] T022 [P] [US1] Write CLI integration test `TestCombinedIncludeExcludeVerbose` in `main_test.go`: run combined config with `--verbose`, verify stderr contains "fields by annotation" in the output line

### Implementation for User Story 1

- [x] T023 [US1] Update annotation filtering orchestration in `main.go:191-199`: change `if cfg.HasAnnotationExclude() { ... } else if cfg.HasAnnotationInclude() { ... }` to sequential: first `if cfg.HasAnnotationInclude()` (include services/methods), then `if cfg.HasAnnotationExclude()` (exclude services/methods/fields). Include must run before exclude to narrow the set first (FR-002)
- [x] T024 [US1] Create test fixture `testdata/combined/combined_service.proto` with: service annotated `// [PublicApi]` with two RPCs, a second unannotated service, a message with a `// [Deprecated]` field, and a message with a `// @Deprecated` method
- [x] T025 [US1] Create golden file `testdata/combined/expected/combined_filtered.proto` with expected output after combined include+exclude filtering
- [x] T026 [P] [US1] Write golden file test `TestGoldenFileCombinedFiltering` in `internal/filter/filter_test.go`: parse `combined_service.proto`, apply combined include+exclude filtering, compare output to `expected/combined_filtered.proto`
- [x] T027 [US1] Run all tests (`go test -race ./...`) and verify T017-T022, T026 pass; verify all existing tests pass unchanged

**Checkpoint**: Combined include+exclude filtering is fully functional

---

## Phase 4: User Story 3 — Backward Compatibility Verification (Priority: P3)

**Goal**: Verify that include-only and exclude-only configs continue to work identically to before

**Independent Test**: Run all existing test suites — they must pass without modification

- [x] T028 [US3] Run `go test -race ./...` and verify all existing tests pass unchanged — this confirms SC-002 (backward compatibility) for include-only, exclude-only, and flat `annotations:` format configs
- [x] T029 [P] [US3] Write CLI integration test `TestExcludeOnlyWithFieldFilteringCLI` in `main_test.go`: create proto with a message field annotated `// @Deprecated`, config with exclude-only `annotations.exclude: [Deprecated]`, verify field is removed (confirms field filtering works in exclude-only mode per FR-006)

**Checkpoint**: All backward compatibility verified, field filtering works in all exclude configurations

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Edge case coverage, verbose output verification, and quickstart validation

- [x] T030 [P] Write unit test `TestFilterFieldsByAnnotationNestedMessage` in `internal/filter/filter_test.go`: create AST with a nested message (message inside message) containing an annotated field, verify the nested field is also filtered
- [x] T031 [P] Write CLI integration test `TestCombinedWithSubstitutionCLI` in `main_test.go`: create proto with `[PublicApi]` service containing `@HasAnyRole(ADMIN)` annotation, config with include=[PublicApi], exclude=[Deprecated], substitutions with `%s` placeholder for HasAnyRole, verify substitution still works correctly alongside combined filtering
- [x] T032 [P] Write CLI integration test `TestCombinedWithStrictSubstitutionCLI` in `main_test.go`: create proto with `[PublicApi]` service containing `@Unknown` annotation, config with include=[PublicApi], exclude=[Deprecated], strict_substitutions=true and no Unknown mapping, verify exit code 2 and stderr contains Unknown
- [x] T033 Validate quickstart.md scenarios: verify examples in `specs/011-combined-include-exclude/quickstart.md` match actual tool output by running representative test cases
- [x] T034 Run `go test -race ./...` one final time to confirm all tests pass

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Foundational)**: No dependencies — removes validation only
- **Phase 2 (US2 - Field Filtering)**: Depends on Phase 1 — uses exclude annotations that were previously blocked in combined mode
- **Phase 3 (US1 - Combined Filtering)**: Depends on Phase 2 — combined mode needs field filtering to be available
- **Phase 4 (US3 - Backward Compat)**: Depends on Phase 2 and Phase 3 — verifies nothing broke
- **Phase 5 (Polish)**: Depends on all prior phases

### User Story Dependencies

- **User Story 2 (P2 — Field Filtering)**: Implemented first despite being P2 priority, because US1 depends on it. Independently testable with exclude-only configs.
- **User Story 1 (P1 — Combined Filtering)**: Core feature. Depends on US2 (field filtering). BLOCKS US3 verification.
- **User Story 3 (P3 — Backward Compat)**: Verification-only — confirms no regressions. Depends on US1 and US2 completion.

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Config/validation changes before filter logic changes
- All tests green before moving to next phase

### Parallel Opportunities

- T005, T006, T007, T008, T009, T010, T011, T012 can run in parallel (different test cases, no file conflicts)
- T017, T018, T019, T020, T021, T022 can run in parallel (different test cases)
- T030, T031, T032 can run in parallel (different test files)

---

## Implementation Strategy

### MVP First (User Story 2 + User Story 1)

1. Complete Phase 1: Remove mutual exclusivity (T001-T004)
2. Complete Phase 2: Field filtering implementation (T005-T016)
3. Complete Phase 3: Combined orchestration (T017-T027)
4. **STOP and VALIDATE**: Run `go test -race ./...` — all tests green
5. US1 + US2 together deliver the core value: combined include+exclude with field-level filtering

### Incremental Delivery

1. T001-T004 → Validation removed → Foundation ready
2. T005-T016 → US2 complete → Field filtering working (independently testable)
3. T017-T027 → US1 complete → Combined filtering working (MVP!)
4. T028-T029 → US3 verified → Backward compatibility confirmed
5. T030-T034 → Polish → Edge cases covered, quickstart validated

---

## Notes

- [P] tasks = different files or independent test cases, no dependencies
- [Story] label maps task to specific user story for traceability
- US2 is implemented before US1 despite lower priority because US1 depends on field filtering being available
- The `FilterFieldsByAnnotation` function follows the exact pattern of `FilterMethodsByAnnotation` and `FilterServicesByAnnotation` — same signature, same iteration approach, same use of `ExtractAnnotations`
- The main.go orchestration change is minimal: replace `if/else if` with two independent `if` blocks, reorder so include runs before exclude
- Message field types to handle: `*proto.NormalField`, `*proto.MapField`, `*proto.OneOfField` (via `*proto.Oneof` container)
- Both `Comment` (leading) and `InlineComment` (trailing) must be checked on each field
- Existing `TestMutualExclusivityErrorCLI` and `TestValidateMutualExclusivity` are updated in-place (renamed and assertions inverted), not deleted
