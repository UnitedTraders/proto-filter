# Tasks: C#-Style Annotation Syntax Support

**Input**: Design documents from `/specs/006-csharp-annotation-syntax/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/cli-interface.md

**Tests**: Included — the spec requires test-driven development (SC-003, SC-004) and the constitution mandates TDD (Principle IV).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

## Phase 1: Setup (Test Fixtures)

**Purpose**: Create proto test fixtures needed by both user stories

- [x] T001 [P] Create bracket-annotated service fixture in testdata/annotations/bracket_service.proto — service with `[HasAnyRole]` and `[Internal]` bracket-style annotations on methods, plus one unannotated method. Use same package `annotations` and same message patterns as existing service.proto.
- [x] T002 [P] Create golden file for bracket service filtering in testdata/annotations/expected/bracket_service.proto — expected output after filtering `[HasAnyRole]` annotated methods (annotated methods removed, unannotated method and its types remain).
- [x] T003 [P] Create mixed-style annotation fixture in testdata/annotations/mixed_styles.proto — service where some methods use `@HasAnyRole` and others use `[HasAnyRole]`, plus one unannotated method. Same package `annotations`.
- [x] T004 [P] Create golden file for mixed-style filtering in testdata/annotations/expected/mixed_styles.proto — expected output after filtering both syntax styles (only unannotated method remains).

**Checkpoint**: All test fixtures ready. No production code changes yet.

---

## Phase 2: Foundational (Regex + Extraction Logic)

**Purpose**: Core production code change — extend annotationRegex and ExtractAnnotations

**⚠️ CRITICAL**: Both user stories depend on this phase completing first.

- [x] T005 Update annotationRegex in internal/filter/filter.go — change from `@(\w[\w.]*)` to `@(\w[\w.]*)|\[(\w[\w.]*)(?:\([^)]*\))?\]` to match both `@Name` and `[Name]`/`[Name(value)]` syntax.
- [x] T006 Update ExtractAnnotations in internal/filter/filter.go — modify the match extraction loop to check both capture groups: group 1 for `@Name`, group 2 for `[Name]`. If `m[1] != ""` append `m[1]`, else if `m[2] != ""` append `m[2]`.
- [x] T007 Verify all existing tests still pass by running `go test -race ./...` — backward compatibility check (SC-003). No test should fail from the regex change.

**Checkpoint**: Production code change complete. Both `@Name` and `[Name]` syntax now recognized. All existing tests pass.

---

## Phase 3: User Story 1 — Recognize C#-Style Annotations (Priority: P1) 🎯 MVP

**Goal**: Bracket-style annotations `[Name]` and `[Name(value)]` are correctly extracted and used for filtering services and methods.

**Independent Test**: Annotate a proto service/method with `[Name]` syntax, run filtering, verify annotated element is removed.

### Tests for User Story 1

- [x] T008 [P] [US1] Add bracket annotation test cases to TestExtractAnnotations table in internal/filter/filter_test.go — add cases: `[HasAnyRole]` → `HasAnyRole`, `[HasAnyRole("ADMIN")]` → `HasAnyRole`, `[com.example.Secure]` → `com.example.Secure`, `[]` → nil, `[ Name ]` → nil, `[RFC 7231]` → nil, `[error code]` → nil. Covers FR-001, FR-002, FR-003, FR-007.
- [x] T009 [P] [US1] Add unit test TestFilterMethodsByBracketAnnotation in internal/filter/filter_test.go — parse bracket_service.proto, filter by `[HasAnyRole]`, verify annotated methods removed and unannotated method remains. Covers acceptance scenarios 1 and 2.
- [x] T010 [P] [US1] Add unit test TestFilterServicesByBracketAnnotation in internal/filter/filter_test.go — create a proto AST with a service commented `[Internal]`, filter by annotation `Internal`, verify service is removed. Covers acceptance scenario 3.
- [x] T011 [US1] Add golden file test TestGoldenFileBracketService in internal/filter/filter_test.go — parse bracket_service.proto, filter by `HasAnyRole`, run RemoveOrphanedDefinitions, ConvertBlockComments, write output and compare byte-for-byte with expected/bracket_service.proto.
- [x] T012 [US1] Add CLI integration test TestBracketAnnotationFilteringCLI in main_test.go — copy bracket_service.proto to temp dir, run CLI with annotation config `HasAnyRole`, verify output contains unannotated method and does not contain annotated methods.

**Checkpoint**: User Story 1 complete. Bracket-style annotations are recognized and filtered correctly.

---

## Phase 4: User Story 2 — Mixed Annotation Styles (Priority: P2)

**Goal**: Both `@Name` and `[Name]` styles work simultaneously in the same file and across files.

**Independent Test**: Create a proto file mixing both annotation styles, run filter, verify both styles matched.

### Tests for User Story 2

- [x] T013 [P] [US2] Add mixed-style extraction test case to TestExtractAnnotations table in internal/filter/filter_test.go — add case with comment lines containing both `@HasAnyRole` and `[Deprecated]` → expect `["HasAnyRole", "Deprecated"]`. Covers FR-004.
- [x] T014 [US2] Add unit test TestFilterMethodsMixedAnnotationStyles in internal/filter/filter_test.go — parse mixed_styles.proto, filter by `HasAnyRole`, verify both `@HasAnyRole` and `[HasAnyRole]` methods removed, unannotated method remains. Covers acceptance scenario 1.
- [x] T015 [US2] Add golden file test TestGoldenFileMixedStyles in internal/filter/filter_test.go — parse mixed_styles.proto, filter, write output, compare byte-for-byte with expected/mixed_styles.proto. Covers SC-002.
- [x] T016 [US2] Add CLI integration test TestMixedStyleAnnotationFilteringCLI in main_test.go — copy mixed_styles.proto to temp dir, run CLI with annotation config, verify both styles filtered correctly.

**Checkpoint**: User Story 2 complete. Mixed annotation styles work seamlessly in the same file.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Final validation across all stories

- [x] T017 Run full test suite with `go test -race ./...` to verify all packages pass (SC-003, SC-004)
- [x] T018 Validate quickstart.md scenarios match actual behavior — run the examples from quickstart.md mentally against the implementation and verify consistency

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — create test fixtures immediately
- **Foundational (Phase 2)**: No dependency on Phase 1 for T005/T006, but T007 can verify backward compat independently
- **User Story 1 (Phase 3)**: Depends on Phase 1 (fixtures) + Phase 2 (regex change)
- **User Story 2 (Phase 4)**: Depends on Phase 1 (fixtures) + Phase 2 (regex change). Independent of US1.
- **Polish (Phase 5)**: Depends on all phases complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Phase 2 — No dependencies on User Story 2
- **User Story 2 (P2)**: Can start after Phase 2 — No dependencies on User Story 1

### Parallel Opportunities

- T001, T002, T003, T004 can all run in parallel (different files)
- T005, T006 are sequential (same file, T006 depends on T005)
- T008, T009, T010 can run in parallel (different test functions, same file but independent additions)
- T013 can start in parallel with US1 tasks (if Phase 2 complete)
- US1 and US2 can be implemented in parallel after Phase 2

---

## Parallel Example: Phase 1

```text
# All fixture files created in parallel:
Task: T001 "Create bracket_service.proto"
Task: T002 "Create expected/bracket_service.proto"
Task: T003 "Create mixed_styles.proto"
Task: T004 "Create expected/mixed_styles.proto"
```

## Parallel Example: User Story 1 Tests

```text
# Unit tests can be written in parallel (different test functions):
Task: T008 "Add bracket extraction test cases"
Task: T009 "Add TestFilterMethodsByBracketAnnotation"
Task: T010 "Add TestFilterServicesByBracketAnnotation"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Create bracket_service.proto and golden file (T001, T002)
2. Complete Phase 2: Update regex and ExtractAnnotations (T005, T006, T007)
3. Complete Phase 3: US1 tests (T008–T012)
4. **STOP and VALIDATE**: All bracket annotation tests pass, all existing tests pass
5. This delivers the core capability — bracket annotations work

### Incremental Delivery

1. Phase 1 + Phase 2 → Foundation ready (regex recognizes both styles)
2. Add User Story 1 → Bracket annotations work independently → MVP!
3. Add User Story 2 → Mixed styles verified → Feature complete
4. Phase 5 → Full validation → Ready for merge

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Production code changes are minimal: ~10 lines in one file (internal/filter/filter.go)
- All test changes are additive — no existing test modifications
- The regex change is the only blocking prerequisite; all tests depend on it
