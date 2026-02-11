# Tasks: Annotation Error Locations

**Input**: Design documents from `/specs/009-annotation-error-locations/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/cli-interface.md

**Tests**: Required by project constitution (Principle IV: Test-Driven Development).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

---

## Phase 1: Foundational (Shared Data Type)

**Purpose**: Define the `AnnotationLocation` struct that both user stories depend on

- [x] T001 [US1] Add `AnnotationLocation` struct to `internal/filter/filter.go` with fields: `File string`, `Line int`, `Name string`, `Token string`

**Checkpoint**: New struct compiles, existing tests pass unchanged (`go test ./...`)

---

## Phase 2: User Story 1 — See Source Locations for Unsubstituted Annotations (Priority: P1) 🎯 MVP

**Goal**: When strict substitution check fails, each unsubstituted annotation is reported with its file path and 1-based line number

**Independent Test**: Run proto-filter with `strict_substitutions: true` and incomplete mapping against `testdata/substitution/substitution_service.proto`. Verify stderr includes location lines like `  substitution_service.proto:6: @HasAnyRole({"ADMIN", "MANAGER"})` with correct line numbers and exit code 2.

### Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T002 [P] [US1] Write unit tests for `CollectAnnotationLocations` in `internal/filter/filter_test.go`: test that parsing `testdata/substitution/substitution_service.proto` returns correct locations (file path, line numbers 6, 10, 13; annotation names `HasAnyRole`, `Internal`, `Public`; full tokens `@HasAnyRole({"ADMIN", "MANAGER"})`, `@Internal`, `[Public]`)
- [x] T003 [P] [US1] Write unit test for `CollectAnnotationLocations` with nil/empty comments in `internal/filter/filter_test.go`: verify empty proto returns empty slice
- [x] T004 [P] [US1] Write CLI integration test `TestStrictSubstitutionErrorWithLocations` in `main_test.go`: run with incomplete mapping, assert stderr contains summary line AND location lines with correct file:line format, exit code 2

### Implementation for User Story 1

- [x] T005 [US1] Implement `CollectAnnotationLocations(def *proto.Proto, relPath string) []AnnotationLocation` in `internal/filter/filter.go`: walk services, RPCs, messages, enums, fields; for each comment line use `substitutionRegex.FindAllStringSubmatch` to extract tokens and names; compute line number as `comment.Position.Line + lineIndex`
- [x] T006 [US1] Update `main.go` strict check: replace `allAnnotations map[string]bool` with `allLocations []filter.AnnotationLocation`; replace `filter.CollectAllAnnotations(pf.def)` call with `filter.CollectAnnotationLocations(pf.def, pf.rel)`; after collecting missing names, sort locations by File then Line, print each unsubstituted location as `  %s:%d: %s\n`
- [x] T007 [US1] Run all tests (`go test -race ./...`) and verify T002, T003, T004 pass; verify all existing tests pass unchanged

**Checkpoint**: User Story 1 is fully functional — strict mode errors include location lines

---

## Phase 3: User Story 2 — Preserve Summary Line for Backward Compatibility (Priority: P2)

**Goal**: The existing summary line (`proto-filter: error: unsubstituted annotations found: Name1, Name2`) remains as the first line of error output, with location lines appended after it

**Independent Test**: Run strict mode with missing annotations. Verify stderr first line matches existing format exactly, followed by location detail lines.

### Tests for User Story 2

- [x] T008 [US2] Write CLI integration test `TestStrictErrorSummaryLinePreserved` in `main_test.go`: verify the existing `TestStrictSubstitutionErrorCLI` test (from feature 008) still passes without modification — the summary line format is unchanged; additionally verify that a single missing annotation produces exactly one summary line and one location line

### Verification for User Story 2

- [x] T009 [US2] Run full test suite (`go test -race ./...`) and confirm all existing strict mode tests from feature 008 pass without modification (SC-003, SC-004)

**Checkpoint**: Both user stories are complete — summary line preserved, location lines added

---

## Phase 4: Polish & Cross-Cutting Concerns

**Purpose**: Edge case coverage and validation

- [x] T010 [P] Write CLI integration test for multiple files with annotations in `main_test.go`: create two proto files in temp input, run with strict mode and incomplete mapping, verify location lines are ordered by file path then line number (FR-005)
- [x] T011 Validate quickstart.md scenarios manually or via test: verify examples in `specs/009-annotation-error-locations/quickstart.md` match actual tool output format
- [x] T012 Run `go test -race ./...` one final time to confirm all tests pass

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Foundational)**: No dependencies — adds struct only
- **Phase 2 (US1)**: Depends on Phase 1 (T001) — uses AnnotationLocation struct
- **Phase 3 (US2)**: Depends on Phase 2 — verifies backward compatibility of the implementation done in US1
- **Phase 4 (Polish)**: Depends on Phase 2 and Phase 3

### User Story Dependencies

- **User Story 1 (P1)**: Core feature. Can start after T001. BLOCKS User Story 2.
- **User Story 2 (P2)**: Verification-only — confirms backward compatibility of US1 implementation. Depends on US1 completion.

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Struct definition before function implementation
- Function implementation before main.go integration
- All tests green before moving to next phase

### Parallel Opportunities

- T002, T003, T004 can run in parallel (different test files / different test cases)
- T010 can run in parallel with T011

---

## Parallel Example: User Story 1

```text
# Launch all US1 tests together (should FAIL initially):
Task: T002 "Unit tests for CollectAnnotationLocations in internal/filter/filter_test.go"
Task: T003 "Unit test for empty proto in internal/filter/filter_test.go"
Task: T004 "CLI integration test for location output in main_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Add AnnotationLocation struct (T001)
2. Complete Phase 2: Write failing tests (T002–T004), implement (T005–T006), verify (T007)
3. **STOP and VALIDATE**: Run `go test -race ./...` — all tests green
4. US1 delivers the core value: location lines in strict mode errors

### Incremental Delivery

1. T001 → Struct defined → Foundation ready
2. T002–T007 → US1 complete → Location lines working (MVP!)
3. T008–T009 → US2 verified → Backward compatibility confirmed
4. T010–T012 → Polish → Edge cases covered, quickstart validated

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Existing `CollectAllAnnotations` function is kept as-is (Research Decision 4)
- Test fixture reuse: `testdata/substitution/substitution_service.proto` has annotations at known lines (6, 10, 13)
- Line number computation: `comment.Position.Line + lineIndex` (Research Decision 1)
- Token extraction: `substitutionRegex.FindAllStringSubmatch` (Research Decision 2)
- Sorting done in `main.go`, not in collection function (Research Decision 3)
