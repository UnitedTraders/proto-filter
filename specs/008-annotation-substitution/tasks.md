# Tasks: Annotation Substitution

**Input**: Design documents from `/specs/008-annotation-substitution/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/cli-interface.md

**Tests**: Included — the constitution mandates TDD (Principle IV) and the project follows test-driven golden file patterns.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

## Phase 1: Setup (Test Fixtures)

**Purpose**: Create proto test fixtures needed for substitution testing

- [x] T001 [P] Create substitution test fixture in testdata/annotations/substitution_service.proto — service with 3 RPC methods: one with `@HasAnyRole({"ADMIN", "MANAGER"})` + descriptive comment, one with `@Internal` only, one with `[Public]` inline with text (e.g., `// [Public] Lists all orders.`). Use package `annotations`, include request/response message types for each method.
- [x] T002 [P] Create golden file for substitution replacement in testdata/annotations/expected/substitution_replaced.proto — expected output after substituting `HasAnyRole: "Requires authentication"`, `Internal: "For internal use only"`, `Public: "Available to all users"`. Annotations replaced with descriptions, surrounding text preserved.
- [x] T003 [P] Create golden file for empty substitution removal in testdata/annotations/expected/substitution_removed.proto — expected output after substituting `HasAnyRole: ""`, `Internal: ""`, `Public: ""`. All annotation lines/tokens removed, only descriptive text remains, no empty comment lines.

**Checkpoint**: Test fixtures ready. No production code changes yet.

---

## Phase 2: Foundational (Config + Substitution Regex)

**Purpose**: Core production code changes — add `Substitutions` and `StrictSubstitutions` to `FilterConfig`, add substitution regex to filter.go.

**⚠️ CRITICAL**: All three user stories depend on this phase completing first.

- [x] T004 Add `Substitutions` and `StrictSubstitutions` fields to `FilterConfig` in internal/config/config.go — add `Substitutions map[string]string` with `yaml:"substitutions"` tag and `StrictSubstitutions bool` with `yaml:"strict_substitutions"` tag. Add `HasSubstitutions() bool` helper method that returns `len(c.Substitutions) > 0`. Do NOT modify `IsPassThrough()` (substitution-only configs still write all files).
- [x] T005 Add substitution regex in internal/filter/filter.go — define `var substitutionRegex = regexp.MustCompile(...)` with pattern `@(\w[\w.]*)(?:\([^)]*\))?|\[(\w[\w.]*)(?:\([^)]*\))?\]` that matches full annotation expressions including parameters. Group 1 = `@`-style name, group 2 = `[]`-style name.
- [x] T006 Verify all existing tests still pass by running `go test -race ./...` — backward compatibility check. New config fields have zero-value defaults (nil map, false bool) so existing configs must parse identically.

**Checkpoint**: Config supports substitution fields. New regex ready. All existing tests pass.

---

## Phase 3: User Story 1 — Replace Annotations with Descriptions (Priority: P1) 🎯 MVP

**Goal**: `substitutions` config key maps annotation names to description text; annotations in comments are replaced in-place.

**Independent Test**: Configure `substitutions: { HasAnyRole: "Requires authentication" }`, run on a proto file with `@HasAnyRole({"ADMIN"})`, verify output contains `Requires authentication` instead of the annotation.

### Tests for User Story 1

- [x] T007 [P] [US1] Add unit test TestSubstituteAnnotations in internal/filter/filter_test.go — parse substitution_service.proto, call `SubstituteAnnotations(def, map[string]string{"HasAnyRole": "Requires authentication", "Internal": "For internal use only", "Public": "Available to all users"})`, verify: (a) `@HasAnyRole({"ADMIN", "MANAGER"})` line replaced with `Requires authentication`, (b) `@Internal` line replaced with `For internal use only`, (c) `[Public]` token in inline text replaced with `Available to all users` preserving surrounding text, (d) non-annotation comment lines unchanged.
- [x] T008 [P] [US1] Add unit test TestSubstituteAnnotationsNoMapping in internal/filter/filter_test.go — parse substitution_service.proto, call `SubstituteAnnotations(def, map[string]string{})` (empty map), verify all comments are left unchanged (FR-011).
- [x] T009 [P] [US1] Add unit test TestSubstituteAnnotationsPartialMapping in internal/filter/filter_test.go — parse substitution_service.proto, call `SubstituteAnnotations(def, map[string]string{"HasAnyRole": "Auth required"})`, verify HasAnyRole is substituted but Internal and Public annotations are left unchanged.
- [x] T010 [US1] Add golden file test TestGoldenFileSubstitutionReplaced in internal/filter/filter_test.go — parse substitution_service.proto, call `SubstituteAnnotations` with full mapping, `ConvertBlockComments`, write output and compare byte-for-byte with expected/substitution_replaced.proto.

### Implementation for User Story 1

- [x] T011 [US1] Implement `SubstituteAnnotations` function in internal/filter/filter.go — walk all elements in the proto AST (services, RPCs, messages, enums, fields). For each comment, iterate lines and use `substitutionRegex.ReplaceAllStringFunc` to replace matched annotation tokens. Look up annotation name (group 1 or 2) in the substitutions map; if found, replace with description text; if not found, leave unchanged. After replacement, trim whitespace from each line. Return count of substitutions made.
- [x] T012 [US1] Generate/verify golden file for substitution_replaced.proto — run the actual substitution pipeline on the fixture to produce testdata/annotations/expected/substitution_replaced.proto. Verify golden file test (T010) passes.
- [x] T013 [US1] Add CLI integration test TestSubstitutionReplacementCLI in main_test.go — copy substitution_service.proto to temp dir, write config with `substitutions: { HasAnyRole: "Requires authentication", Internal: "For internal use only", Public: "Available to all users" }`, run CLI, verify output contains description text and does not contain annotation markers.
- [x] T014 [US1] Wire SubstituteAnnotations into main.go pipeline — in the per-file processing loop, after `ConvertBlockComments` and before writing, call `filter.SubstituteAnnotations(pf.def, cfg.Substitutions)` when `cfg.HasSubstitutions()` is true. Add substitution count to verbose output.

**Checkpoint**: User Story 1 complete. Annotations replaced with descriptions in output.

---

## Phase 4: User Story 2 — Remove Annotations via Empty Substitution (Priority: P2)

**Goal**: Empty string substitution values cleanly remove annotation lines/tokens with no leftover empty lines or orphaned comments.

**Independent Test**: Configure `substitutions: { HasAnyRole: "" }`, run on a proto file with `// @HasAnyRole\n// Creates an order.`, verify output is `// Creates an order.` only.

### Tests for User Story 2

- [x] T015 [P] [US2] Add unit test TestSubstituteAnnotationsEmptyRemoval in internal/filter/filter_test.go — parse substitution_service.proto, call `SubstituteAnnotations(def, map[string]string{"HasAnyRole": "", "Internal": "", "Public": ""})`, verify: (a) annotation-only lines are removed, (b) no empty comment lines remain, (c) descriptive text lines preserved, (d) inline annotation removed but surrounding text preserved.
- [x] T016 [P] [US2] Add unit test TestSubstituteAnnotationsFullCommentRemoval in internal/filter/filter_test.go — create a proto AST with a method whose comment consists of only `@Internal` (no descriptive text). Call `SubstituteAnnotations(def, map[string]string{"Internal": ""})`, verify the comment is set to nil (removed entirely from the element), not left as an empty comment.
- [x] T017 [US2] Add golden file test TestGoldenFileSubstitutionRemoved in internal/filter/filter_test.go — parse substitution_service.proto, call `SubstituteAnnotations` with all-empty mapping, `ConvertBlockComments`, write output and compare byte-for-byte with expected/substitution_removed.proto.

### Implementation for User Story 2

- [x] T018 [US2] Add empty line cleanup to `SubstituteAnnotations` in internal/filter/filter.go — after replacing annotations in a comment's lines: (a) remove lines that are empty/whitespace-only after substitution, (b) if all lines are removed, set the comment pointer to nil on the AST element. This requires the function to accept or return the comment pointer so it can be nullified.
- [x] T019 [US2] Generate/verify golden file for substitution_removed.proto — run the empty substitution pipeline on the fixture to produce testdata/annotations/expected/substitution_removed.proto. Verify golden file test (T017) passes.
- [x] T020 [US2] Add CLI integration test TestSubstitutionEmptyRemovalCLI in main_test.go — copy substitution_service.proto to temp dir, write config with `substitutions: { HasAnyRole: "", Internal: "", Public: "" }`, run CLI, verify output contains no annotation markers and no empty comment lines.

**Checkpoint**: User Story 2 complete. Empty substitutions cleanly remove annotations.

---

## Phase 5: User Story 3 — Strict Mode: Detect Unsubstituted Annotations (Priority: P3)

**Goal**: `strict_substitutions: true` scans all annotations across all processed files and fails if any lack a substitution mapping.

**Independent Test**: Enable strict mode with incomplete mapping, run on a file with an unmapped annotation, verify exit code 2 and error listing the missing annotation name.

### Tests for User Story 3

- [x] T021 [P] [US3] Add unit test TestCollectAllAnnotations in internal/filter/filter_test.go — parse substitution_service.proto, call `CollectAllAnnotations(def)`, verify it returns a set containing `HasAnyRole`, `Internal`, `Public` (the unique annotation names found in all comments).
- [x] T022 [P] [US3] Add unit test TestCollectAllAnnotationsEmpty in internal/filter/filter_test.go — create a proto AST with no annotations in comments, call `CollectAllAnnotations(def)`, verify it returns an empty set.
- [x] T023 [P] [US3] Add config test TestLoadConfigWithSubstitutions in internal/config/config_test.go — write YAML with `substitutions: { HasAnyRole: "Auth", Internal: "" }` and `strict_substitutions: true`, load config, verify `cfg.Substitutions` contains 2 entries, `cfg.StrictSubstitutions` is true, `cfg.HasSubstitutions()` returns true.
- [x] T024 [P] [US3] Add config test TestLoadConfigNoSubstitutions in internal/config/config_test.go — write YAML with only `include` patterns (no substitutions keys), load config, verify `cfg.Substitutions` is nil/empty, `cfg.StrictSubstitutions` is false, `cfg.HasSubstitutions()` returns false.

### Implementation for User Story 3

- [x] T025 [US3] Implement `CollectAllAnnotations` function in internal/filter/filter.go — walk all elements in the proto AST (services, RPCs, messages, enums, fields) and collect all unique annotation names from comments using the existing `ExtractAnnotations` helper. Return `map[string]bool` of unique annotation names.
- [x] T026 [US3] Implement strict mode check in main.go — refactor the per-file loop into two passes: (1) first pass: filter + convert + collect annotations per file into a global set, (2) after first pass: if `cfg.StrictSubstitutions` is true, compare collected annotations against `cfg.Substitutions` keys; if any unsubstituted, print error listing all missing names (sorted alphabetically) to stderr and return exit code 2 without writing any output, (3) second pass: substitute + write. This satisfies FR-010 (no output on strict failure).
- [x] T027 [US3] Add CLI integration test TestStrictSubstitutionErrorCLI in main_test.go — copy substitution_service.proto to temp dir, write config with `substitutions: { HasAnyRole: "Auth required" }` and `strict_substitutions: true`, run CLI, verify exit code 2, stderr contains "unsubstituted annotations", stderr lists `Internal` and `Public`, no output files written.
- [x] T028 [US3] Add CLI integration test TestStrictSubstitutionSuccessCLI in main_test.go — copy substitution_service.proto to temp dir, write config with complete mappings for all 3 annotations and `strict_substitutions: true`, run CLI, verify exit code 0 and output files are written correctly.
- [x] T029 [US3] Add CLI integration test TestStrictSubstitutionNoAnnotationsCLI in main_test.go — create a proto file with no annotations, write config with `strict_substitutions: true` and no `substitutions`, run CLI, verify exit code 0 (no annotations to flag).

**Checkpoint**: User Story 3 complete. Strict mode detects all unsubstituted annotations.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final validation across all stories

- [x] T030 Run full test suite with `go test -race ./...` to verify all packages pass
- [x] T031 Validate quickstart.md scenarios match actual behavior — run the examples from quickstart.md against the implementation and verify consistency
- [x] T032 Add CLI integration test TestSubstitutionWithAnnotationFilterCLI in main_test.go — verify substitution works correctly alongside annotation exclude filtering (SC-004). Config with both `annotations.exclude: ["Internal"]` and `substitutions: { HasAnyRole: "Auth required" }`. Verify Internal methods removed, HasAnyRole substituted on remaining methods.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — create test fixtures immediately
- **Foundational (Phase 2)**: No dependency on Phase 1 for T004-T005, but T006 verifies backward compat
- **User Story 1 (Phase 3)**: Depends on Phase 1 (fixtures) + Phase 2 (config fields + regex)
- **User Story 2 (Phase 4)**: Depends on Phase 3 (SubstituteAnnotations function must exist first)
- **User Story 3 (Phase 5)**: Depends on Phase 2 (config fields). CollectAllAnnotations is independent of US1/US2. Pipeline refactoring (T026) depends on US1 (T014 wires substitution into main.go)
- **Polish (Phase 6)**: Depends on all phases complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Phase 2 — Creates the core SubstituteAnnotations function
- **User Story 2 (P2)**: Depends on US1 — Extends SubstituteAnnotations with empty line cleanup
- **User Story 3 (P3)**: Partially independent of US1/US2 (CollectAllAnnotations is new), but pipeline refactoring needs US1's wiring in main.go

### Parallel Opportunities

- T001, T002, T003 can run in parallel (different files)
- T004, T005 can run in parallel (different files)
- T007, T008, T009 can run in parallel (different test functions)
- T015, T016 can run in parallel (different test functions)
- T021, T022, T023, T024 can run in parallel (different test functions/files)

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Create substitution test fixtures (T001, T002, T003)
2. Complete Phase 2: Config fields + regex (T004–T006)
3. Complete Phase 3: US1 substitution implementation and tests (T007–T014)
4. **STOP and VALIDATE**: Annotations replaced with descriptions, all existing tests pass
5. This delivers the core capability — annotation substitution

### Incremental Delivery

1. Phase 1 + Phase 2 → Foundation ready (config supports substitution)
2. Add User Story 1 → Substitution works → MVP!
3. Add User Story 2 → Empty substitutions clean up lines → Feature polished
4. Add User Story 3 → Strict mode enforces completeness → Full safety guards
5. Phase 6 → Full validation → Ready for merge

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Production code changes in 3 files: `internal/config/config.go`, `internal/filter/filter.go`, `main.go`
- US2 extends US1's `SubstituteAnnotations` function (adds empty line cleanup to existing function)
- US3 introduces a new function (`CollectAllAnnotations`) and refactors main.go's write loop into a two-pass architecture
- The substitution regex differs from the existing annotation regex — it matches `@Name(...)` parameters for `@`-style too
- Substitution runs after `ConvertBlockComments`, so only single-line comment style needs handling
