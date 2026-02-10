# Tasks: Annotation Include/Exclude Filtering Modes

**Input**: Design documents from `/specs/007-annotation-include-exclude/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/cli-interface.md

**Tests**: Included — the constitution mandates TDD (Principle IV) and the project follows test-driven golden file patterns.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

## Phase 1: Setup (Test Fixtures)

**Purpose**: Create proto test fixtures needed for include mode testing

- [x] T001 [P] Create include-mode service fixture in testdata/annotations/include_service.proto — service with 3 methods: one annotated `@Public`, one annotated `[Public]` (bracket syntax), one unannotated. Use package `annotations`, same message patterns as existing service.proto.
- [x] T002 [P] Create golden file for include-mode filtering in testdata/annotations/expected/include_service.proto — expected output after include-filtering by `Public` annotation (only the two annotated methods remain, unannotated method and its types removed).

**Checkpoint**: Test fixtures ready. No production code changes yet.

---

## Phase 2: Foundational (Config Restructure + Validation)

**Purpose**: Core production code changes — restructure `FilterConfig` to support `AnnotationConfig` with include/exclude sub-keys, backward compatibility for old flat format, and validation.

**⚠️ CRITICAL**: All three user stories depend on this phase completing first.

- [x] T003 Define `AnnotationConfig` struct in internal/config/config.go — add struct with `Include []string` and `Exclude []string` fields. Implement `UnmarshalYAML(value *yaml.Node) error` that detects YAML node kind: if sequence (list), populate `Exclude` from items (old flat format); if mapping, parse `include` and `exclude` sub-keys.
- [x] T004 Update `FilterConfig` in internal/config/config.go — change `Annotations []string` field to `Annotations AnnotationConfig` (keeping the `yaml:"annotations"` tag). Update `IsPassThrough()` to check `len(c.Annotations.Include) == 0 && len(c.Annotations.Exclude) == 0`. Update `HasAnnotations()` to return `len(c.Annotations.Include) > 0 || len(c.Annotations.Exclude) > 0`. Add `HasAnnotationInclude() bool` and `HasAnnotationExclude() bool` helper methods.
- [x] T005 Add `Validate() error` method to `FilterConfig` in internal/config/config.go — return error if both `Annotations.Include` and `Annotations.Exclude` are non-empty: `"annotations.include and annotations.exclude are mutually exclusive"`.
- [x] T006 Update main.go annotation filtering section — after `LoadConfig()`, call `cfg.Validate()` and exit with code 2 on error. Replace `cfg.HasAnnotations()` block: if `cfg.HasAnnotationExclude()`, call existing `FilterServicesByAnnotation`/`FilterMethodsByAnnotation` with `cfg.Annotations.Exclude`; if `cfg.HasAnnotationInclude()`, call new `IncludeServicesByAnnotation`/`IncludeMethodsByAnnotation` (to be created in Phase 3) with `cfg.Annotations.Include`.
- [x] T007 Update all existing references to `cfg.Annotations` in main.go — change `cfg.Annotations` (was `[]string`) to `cfg.Annotations.Exclude` where currently used in `FilterServicesByAnnotation`/`FilterMethodsByAnnotation` calls.
- [x] T008 Verify all existing tests still pass by running `go test -race ./...` — backward compatibility check. Tests using YAML configs with `annotations: [list]` should still work via custom unmarshaler. Tests using `config.FilterConfig{Annotations: []string{...}}` struct literals will need updating to `config.FilterConfig{Annotations: config.AnnotationConfig{Exclude: []string{...}}}`.
- [x] T009 Update existing test struct literals that reference `Annotations` field — scan internal/config/config_test.go, internal/filter/filter_test.go, and main_test.go for `Annotations: []string{` and update to `Annotations: config.AnnotationConfig{Exclude: []string{`. Run `go test -race ./...` to confirm all pass.

**Checkpoint**: Config restructured. Old flat format works via custom unmarshaler. All existing tests pass. Validation rejects ambiguous configs.

---

## Phase 3: User Story 1 — Include Mode (Priority: P1) 🎯 MVP

**Goal**: `annotations.include` list in config keeps only matching annotated services/methods, removing all others.

**Independent Test**: Configure `annotations.include: [Public]`, run filter on a proto file with `@Public` and unannotated methods, verify only annotated methods remain.

### Tests for User Story 1

- [x] T010 [P] [US1] Add unit test TestIncludeMethodsByAnnotation in internal/filter/filter_test.go — parse include_service.proto, call `IncludeMethodsByAnnotation(def, []string{"Public"})`, verify unannotated method is removed and annotated methods remain (both `@Public` and `[Public]`). Covers US1 acceptance scenarios 1 and 4.
- [x] T011 [P] [US1] Add unit test TestIncludeServicesByAnnotation in internal/filter/filter_test.go — create a proto AST with two services: one annotated `@Public`, one unannotated. Call `IncludeServicesByAnnotation(def, []string{"Public"})`, verify unannotated service is removed. Covers US1 acceptance scenario 2.
- [x] T012 [P] [US1] Add unit test TestIncludeMethodsByAnnotationNoMatch in internal/filter/filter_test.go — parse include_service.proto, call `IncludeMethodsByAnnotation(def, []string{"NonExistent"})`, verify all methods are removed. Covers US1 acceptance scenario 3.
- [x] T013 [US1] Add golden file test TestGoldenFileIncludeService in internal/filter/filter_test.go — parse include_service.proto, call `IncludeMethodsByAnnotation` with `Public`, run `RemoveOrphanedDefinitions`, `ConvertBlockComments`, write output and compare byte-for-byte with expected/include_service.proto.

### Implementation for User Story 1

- [x] T014 [US1] Implement `IncludeServicesByAnnotation` in internal/filter/filter.go — inverse of `FilterServicesByAnnotation`: remove services whose comments do NOT contain any of the specified annotations. Return count of removed services.
- [x] T015 [US1] Implement `IncludeMethodsByAnnotation` in internal/filter/filter.go — inverse of `FilterMethodsByAnnotation`: remove RPC methods whose comments do NOT contain any of the specified annotations. Return count of removed methods.
- [x] T016 [US1] Generate golden file for include_service.proto — run the actual include pipeline on the fixture to produce testdata/annotations/expected/include_service.proto with correct output. Verify golden file test passes.
- [x] T017 [US1] Add CLI integration test TestIncludeAnnotationFilteringCLI in main_test.go — copy include_service.proto to temp dir, write config with `annotations:\n  include:\n    - "Public"`, run CLI, verify output contains only annotated methods and does not contain unannotated method.

**Checkpoint**: User Story 1 complete. Include mode keeps only matching annotated elements.

---

## Phase 4: User Story 2 — Config Rename with Backward Compatibility (Priority: P2)

**Goal**: The new structured `annotations.exclude` config key works identically to the old flat `annotations` list.

**Independent Test**: Run proto-filter with `annotations:\n  exclude:\n    - "HasAnyRole"` and verify output is identical to running with `annotations:\n  - "HasAnyRole"`.

### Tests for User Story 2

- [x] T018 [P] [US2] Add config test TestLoadConfigStructuredExclude in internal/config/config_test.go — write YAML with `annotations:\n  exclude:\n    - "HasAnyRole"`, load config, verify `cfg.Annotations.Exclude` contains `["HasAnyRole"]` and `cfg.Annotations.Include` is empty.
- [x] T019 [P] [US2] Add config test TestLoadConfigStructuredInclude in internal/config/config_test.go — write YAML with `annotations:\n  include:\n    - "Public"`, load config, verify `cfg.Annotations.Include` contains `["Public"]` and `cfg.Annotations.Exclude` is empty.
- [x] T020 [P] [US2] Add config test TestLoadConfigFlatAnnotationsBackwardCompat in internal/config/config_test.go — write YAML with `annotations:\n  - "HasAnyRole"`, load config, verify `cfg.Annotations.Exclude` contains `["HasAnyRole"]` (old flat format treated as exclude).
- [x] T021 [US2] Add CLI integration test TestStructuredExcludeAnnotationFilteringCLI in main_test.go — copy service.proto from testdata/annotations to temp dir, write config with `annotations:\n  exclude:\n    - "HasAnyRole"`, run CLI, verify output is identical to existing exclude behavior (annotated methods removed, ListOrders remains).

**Checkpoint**: User Story 2 complete. New structured format works. Old flat format backward compatible.

---

## Phase 5: User Story 3 — Mutual Exclusivity Validation (Priority: P3)

**Goal**: Clear error when both include and exclude are set.

**Independent Test**: Create config with both include and exclude populated, run proto-filter, verify exit code 2 and error message.

### Tests for User Story 3

- [x] T022 [P] [US3] Add config validation test TestValidateMutualExclusivity in internal/config/config_test.go — create `FilterConfig` with both `Annotations.Include` and `Annotations.Exclude` populated, call `Validate()`, verify error contains "mutually exclusive".
- [x] T023 [P] [US3] Add config validation test TestValidateEmptyListsPass in internal/config/config_test.go — create `FilterConfig` with `Annotations.Include` set and `Annotations.Exclude` as empty slice, call `Validate()`, verify no error (empty list treated as absent).
- [x] T024 [P] [US3] Add config validation test TestValidateNoAnnotationsPass in internal/config/config_test.go — create `FilterConfig` with neither include nor exclude, call `Validate()`, verify no error.
- [x] T025 [US3] Add CLI integration test TestMutualExclusivityErrorCLI in main_test.go — write config with `annotations:\n  include:\n    - "Public"\n  exclude:\n    - "Internal"`, run CLI, verify exit code 2 and stderr contains "mutually exclusive".

**Checkpoint**: User Story 3 complete. Ambiguous configs rejected with clear error.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final validation across all stories

- [x] T026 Run full test suite with `go test -race ./...` to verify all packages pass
- [x] T027 Validate quickstart.md scenarios match actual behavior — run the examples from quickstart.md mentally against the implementation and verify consistency

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — create test fixtures immediately
- **Foundational (Phase 2)**: No dependency on Phase 1 for T003-T005, but T008/T009 verify backward compat
- **User Story 1 (Phase 3)**: Depends on Phase 1 (fixtures) + Phase 2 (config restructure + main.go dispatch)
- **User Story 2 (Phase 4)**: Depends on Phase 2 (config restructure). Independent of US1.
- **User Story 3 (Phase 5)**: Depends on Phase 2 (validation method). Independent of US1 and US2.
- **Polish (Phase 6)**: Depends on all phases complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Phase 2 — No dependencies on US2 or US3
- **User Story 2 (P2)**: Can start after Phase 2 — No dependencies on US1 or US3
- **User Story 3 (P3)**: Can start after Phase 2 — No dependencies on US1 or US2

### Parallel Opportunities

- T001, T002 can run in parallel (different files)
- T003, T004, T005 are sequential (same file, each builds on previous)
- T010, T011, T012 can run in parallel (different test functions)
- T018, T019, T020 can run in parallel (different test functions)
- T022, T023, T024 can run in parallel (different test functions)
- US1, US2, US3 can be implemented in parallel after Phase 2

---

## Parallel Example: Phase 1

```text
# All fixture files created in parallel:
Task: T001 "Create include_service.proto"
Task: T002 "Create expected/include_service.proto"
```

## Parallel Example: User Story 1 Tests

```text
# Unit tests can be written in parallel (different test functions):
Task: T010 "Add TestIncludeMethodsByAnnotation"
Task: T011 "Add TestIncludeServicesByAnnotation"
Task: T012 "Add TestIncludeMethodsByAnnotationNoMatch"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Create include_service.proto fixtures (T001, T002)
2. Complete Phase 2: Restructure config, update main.go, fix existing tests (T003–T009)
3. Complete Phase 3: US1 include mode implementation and tests (T010–T017)
4. **STOP and VALIDATE**: Include mode works, all existing tests pass
5. This delivers the core capability — include annotation filtering

### Incremental Delivery

1. Phase 1 + Phase 2 → Foundation ready (config supports both modes)
2. Add User Story 1 → Include mode works → MVP!
3. Add User Story 2 → Structured exclude config verified → Feature complete for config
4. Add User Story 3 → Mutual exclusivity validation → Full safety guards
5. Phase 6 → Full validation → Ready for merge

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Production code changes in 3 files: `internal/config/config.go`, `internal/filter/filter.go`, `main.go`
- Phase 2 is the largest phase — it restructures config and updates all existing references
- The custom `UnmarshalYAML` is the key backward-compat mechanism (research.md Decision 2)
- Test updates in Phase 2 (T008/T009) handle struct literal migration from `[]string` to `AnnotationConfig`
