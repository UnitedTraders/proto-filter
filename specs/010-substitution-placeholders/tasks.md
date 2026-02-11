# Tasks: Substitution Placeholders

**Input**: Design documents from `/specs/010-substitution-placeholders/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/cli-interface.md

**Tests**: Required by project constitution (Principle IV: Test-Driven Development).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

---

## Phase 1: Foundational (Regex Modification)

**Purpose**: Modify `substitutionRegex` to capture annotation arguments as groups, and update all consumers of the old group indices. This is the shared prerequisite for both user stories.

- [x] T001 Update `substitutionRegex` in `internal/filter/filter.go:15` from `@(\w[\w.]*)(?:\([^)]*\))?|\[(\w[\w.]*)(?:\([^)]*\))?\]` to `@(\w[\w.]*)(?:\(([^)]*)\))?|\[(\w[\w.]*)(?:\(([^)]*)\))?\]` (add capturing groups 2 and 4 for argument text)
- [x] T002 Update `substituteInComment` in `internal/filter/filter.go:627-629` to use new group indices: `submatch[1]` for `@`-style name (unchanged), `submatch[3]` for `[`-style name (was `submatch[2]`)
- [x] T003 Update `collectLocationsFromComment` in `internal/filter/filter.go:703-705` to use new group indices: `m[1]` for `@`-style name (unchanged), `m[3]` for `[`-style name (was `m[2]`)
- [x] T004 Run all existing tests (`go test -race ./...`) and verify they pass unchanged with the new regex and updated group indices (SC-003 backward compatibility gate)

**Checkpoint**: Regex captures arguments, all existing tests pass unchanged, no behavioral changes yet

---

## Phase 2: User Story 1 — Interpolate Annotation Arguments into Substitution Text (Priority: P1) 🎯 MVP

**Goal**: When a substitution value contains `%s` and the matched annotation has arguments, replace `%s` with the argument content

**Independent Test**: Configure a substitution mapping with `%s` placeholder, run proto-filter on a proto file containing annotations with arguments, and verify the output comments contain the replacement text with the argument value interpolated.

### Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T005 [P] [US1] Write unit test `TestSubstituteAnnotationsPlaceholder` in `internal/filter/filter_test.go`: create programmatic AST with `@Min(3)` annotation, substitute with `Min: "Minimal value is %s"`, verify output contains `Minimal value is 3` and does NOT contain `%s` or `@Min`
- [x] T006 [P] [US1] Write unit test `TestSubstituteAnnotationsPlaceholderBracketStyle` in `internal/filter/filter_test.go`: create AST with `[Tag(important)]` annotation, substitute with `Tag: "Tagged: %s"`, verify output contains `Tagged: important`
- [x] T007 [P] [US1] Write unit test `TestSubstituteAnnotationsPlaceholderComplexArgs` in `internal/filter/filter_test.go`: create AST with `@HasAnyRole({"ADMIN", "MANAGER"})`, substitute with `HasAnyRole: "Requires roles: %s"`, verify output contains `Requires roles: {"ADMIN", "MANAGER"}`
- [x] T008 [P] [US1] Write unit test `TestSubstituteAnnotationsNoPlaceholderWithArgs` in `internal/filter/filter_test.go`: create AST with `@Min(3)`, substitute with `Min: "Has minimum constraint"` (no `%s`), verify output contains `Has minimum constraint` — arguments are ignored (FR-004)
- [x] T009 [P] [US1] Write unit test `TestSubstituteAnnotationsMultiplePlaceholders` in `internal/filter/filter_test.go`: create AST with `@Range(5)`, substitute with `Range: "Between %s and %s"`, verify output contains `Between 5 and %s` — only first `%s` replaced (FR-007)
- [x] T010 [P] [US1] Write CLI integration test `TestPlaceholderSubstitutionCLI` in `main_test.go`: create temp proto file with `@Min(3)` and `@Max(100)`, config with `Min: "Minimal value is %s"` and `Max: "Maximum value is %s"`, verify output file contains `Minimal value is 3` and `Maximum value is 100`

### Implementation for User Story 1

- [x] T011 [US1] Update `substituteInComment` in `internal/filter/filter.go:624-636` to extract argument from `submatch[2]` (for `@`-style) or `submatch[4]` (for `[`-style), and when `strings.Contains(replacement, "%s")` is true, use `strings.Replace(replacement, "%s", args, 1)` to interpolate the argument into the replacement value
- [x] T012 [US1] Create test fixture `testdata/substitution/placeholder_service.proto` with annotations containing arguments: `@Min(3)`, `@Max(100)`, `@HasAnyRole({"ADMIN", "MANAGER"})`, `[Tag(important)]`, and `@Deprecated` (no args)
- [x] T013 [US1] Create golden file `testdata/substitution/expected/placeholder_replaced.proto` with expected output after placeholder substitution
- [x] T014 [P] [US1] Write golden file test `TestGoldenFilePlaceholderReplaced` in `internal/filter/filter_test.go`: parse `placeholder_service.proto`, apply substitutions with `%s` placeholders, compare output to `expected/placeholder_replaced.proto`
- [x] T015 [US1] Run all tests (`go test -race ./...`) and verify T005-T010, T014 pass; verify all existing tests pass unchanged

**Checkpoint**: User Story 1 is fully functional — `%s` placeholders are interpolated with annotation arguments

---

## Phase 3: User Story 2 — Graceful Handling of Missing Arguments (Priority: P2)

**Goal**: When a substitution value contains `%s` but the annotation has no arguments (or empty parentheses), the `%s` is replaced with an empty string

**Independent Test**: Configure a substitution mapping with `%s` placeholder, run proto-filter on a proto file where annotations appear without arguments, and verify `%s` is replaced with empty string.

### Tests for User Story 2

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T016 [P] [US2] Write unit test `TestSubstituteAnnotationsPlaceholderNoArgs` in `internal/filter/filter_test.go`: create AST with `@Min` (no parentheses), substitute with `Min: "Minimal value is %s"`, verify output contains `Minimal value is` (trailing space from `%s` → empty)
- [x] T017 [P] [US2] Write unit test `TestSubstituteAnnotationsPlaceholderEmptyParens` in `internal/filter/filter_test.go`: create AST with `@Min()` (empty parentheses), substitute with `Min: "Minimal value is %s"`, verify output contains `Minimal value is` (FR-008)
- [x] T018 [P] [US2] Write unit test `TestSubstituteAnnotationsPlaceholderBracketNoArgs` in `internal/filter/filter_test.go`: create AST with `[Tag]` (bracket style, no args), substitute with `Tag: "Tagged: %s"`, verify output contains `Tagged:` (FR-005)

### Verification for User Story 2

- [x] T019 [US2] Run all tests (`go test -race ./...`) and verify T016-T018 pass. Note: The implementation from T011 already handles missing arguments naturally — when the regex doesn't match parentheses, the argument capture group returns empty string, so `strings.Replace(replacement, "%s", "", 1)` produces the correct result. These tests should pass WITHOUT additional code changes.

**Checkpoint**: Both user stories are complete — arguments interpolated when present, empty string when absent

---

## Phase 4: Polish & Cross-Cutting Concerns

**Purpose**: Edge case coverage, backward compatibility verification, and quickstart validation

- [x] T020 [P] Write unit test `TestSubstituteAnnotationsPlaceholderSpecialChars` in `internal/filter/filter_test.go`: create AST with `@Format(100%)`, substitute with `Format: "Format is %s"`, verify output contains `Format is 100%` — special characters inserted literally (FR-006)
- [x] T021 [P] Write CLI integration test `TestPlaceholderWithStrictModeCLI` in `main_test.go`: create proto with `@Min(3)` and `@Unknown`, config with `Min: "Minimal value is %s"` and `strict_substitutions: true`, verify exit code 2 and stderr contains `Unknown` but NOT `Min` — strict mode checks names not args (FR-009)
- [x] T022 [P] Write CLI integration test `TestPlaceholderLocationReportingCLI` in `main_test.go`: create proto with `@Min(3)`, config with `strict_substitutions: true` and NO `Min` mapping, verify location line shows full token `@Min(3)` — location reporting includes arguments (FR-010)
- [x] T023 Validate quickstart.md scenarios: verify examples in `specs/010-substitution-placeholders/quickstart.md` match actual tool output format by running representative test cases
- [x] T024 Run `go test -race ./...` one final time to confirm all tests pass

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Foundational)**: No dependencies — modifies regex and group indices only
- **Phase 2 (US1)**: Depends on Phase 1 (T001-T004) — uses new regex capture groups for interpolation
- **Phase 3 (US2)**: Depends on Phase 2 — tests verify graceful degradation of the US1 implementation
- **Phase 4 (Polish)**: Depends on Phase 2 and Phase 3

### User Story Dependencies

- **User Story 1 (P1)**: Core feature. Can start after Phase 1. BLOCKS User Story 2.
- **User Story 2 (P2)**: Verification-only — confirms the US1 implementation handles missing arguments correctly. Depends on US1 completion.

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Regex change before substitution logic change
- All tests green before moving to next phase

### Parallel Opportunities

- T005, T006, T007, T008, T009, T010 can run in parallel (different test cases, no file conflicts)
- T016, T017, T018 can run in parallel (different test cases)
- T020, T021, T022 can run in parallel (different test files)

---

## Parallel Example: User Story 1

```text
# Launch all US1 tests together (should FAIL initially):
Task: T005 "Unit test for @Min(3) placeholder in internal/filter/filter_test.go"
Task: T006 "Unit test for [Tag(important)] bracket-style placeholder"
Task: T007 "Unit test for complex args placeholder"
Task: T008 "Unit test for no-placeholder with args"
Task: T009 "Unit test for multiple %s placeholders"
Task: T010 "CLI integration test for placeholder substitution in main_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Regex modification + group index updates (T001-T004)
2. Complete Phase 2: Write failing tests (T005-T010), implement interpolation (T011), create fixtures (T012-T013), golden test (T014), verify (T015)
3. **STOP and VALIDATE**: Run `go test -race ./...` — all tests green
4. US1 delivers the core value: `%s` placeholder interpolation in substitution values

### Incremental Delivery

1. T001-T004 → Regex updated → Foundation ready
2. T005-T015 → US1 complete → Placeholder interpolation working (MVP!)
3. T016-T019 → US2 verified → Missing arguments handled gracefully
4. T020-T024 → Polish → Edge cases covered, quickstart validated

---

## Notes

- [P] tasks = different files or independent test cases, no dependencies
- [Story] label maps task to specific user story for traceability
- `annotationRegex` is NOT modified (Research Decision 4) — used only for filtering
- Regex change shifts group indices: old `submatch[2]` → new `submatch[3]` for `[`-style name
- The `%s` replacement uses `strings.Replace(replacement, "%s", args, 1)` — count=1 ensures only first `%s` is replaced (FR-007)
- Test fixture reuse: existing `substitution_service.proto` has `@HasAnyRole({"ADMIN", "MANAGER"})` with args but existing tests don't use `%s`, so backward compatibility is preserved
- US2 implementation note: No additional code changes expected for US2 — the US1 implementation naturally handles missing arguments because Go regex returns empty string for unmatched optional groups
