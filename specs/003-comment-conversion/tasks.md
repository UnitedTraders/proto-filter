# Tasks: Comment Style Conversion

**Input**: Design documents from `/specs/003-comment-conversion/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

**Tests**: Included per Constitution Principle IV (Test-Driven Development). Uses golden file comparison pattern.

**Organization**: Tasks grouped by user story. Both US1 and US2 are P1 but US2 (content preservation) is tested as part of golden file comparison, so they share a single implementation phase.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

## Phase 1: Setup (Test Fixtures & Golden Files)

**Purpose**: Create input fixtures and expected output golden files for comparison testing

- [x] T001 [P] Create block comment input fixture in testdata/comments/block_comments.proto with Javadoc-style `/** */` comments on services, messages, enums, RPCs, and fields including annotations like `@StartsWithSnapshot`, inline block comments, and empty block comments
- [x] T002 [P] Create golden file testdata/comments/expected/commented.proto by copying testdata/comments/commented.proto (already all `//` style, should be unchanged after conversion)
- [x] T003 [P] Create golden file testdata/comments/expected/multiline.proto with all block comments from testdata/comments/multiline.proto converted to `//` style, preserving text content
- [x] T004 Create golden file testdata/comments/expected/block_comments.proto with all block comments from T001 input converted to `//` style, preserving text content and annotations

---

## Phase 2: Foundational (Core Conversion Function)

**Purpose**: Implement `ConvertBlockComments` and the golden file test helper

- [x] T005 [US1] Write unit test `TestConvertBlockComments` in internal/filter/filter_test.go that parses testdata/comments/multiline.proto, calls `ConvertBlockComments`, and verifies no `Cstyle=true` comments remain in the AST
- [x] T006 [US1] Write unit test `TestConvertBlockCommentsPreservesExisting` in internal/filter/filter_test.go that parses testdata/comments/commented.proto, calls `ConvertBlockComments`, and verifies all comments are still present and unchanged
- [x] T007 [US1] Implement `ConvertBlockComments(def *proto.Proto)` in internal/filter/filter.go that walks the AST using `proto.Walk`, sets `Cstyle = false` on all block comments, and strips leading `* ` prefixes from `Lines` entries
- [x] T008 Implement helper function `convertComment(comment *proto.Comment)` in internal/filter/filter.go that converts a single Comment: skips nil/non-Cstyle, sets `Cstyle = false`, and cleans `Lines`

**Checkpoint**: `ConvertBlockComments` function works on AST level. Unit tests pass.

---

## Phase 3: User Story 1 - Block-to-Single-Line Conversion (Priority: P1)

**Goal**: Verify complete block comment conversion via golden file comparison

**Independent Test**: Process each input fixture, write output, compare byte-for-byte against golden file

### Tests for User Story 1

- [x] T009 [US1] Write golden file comparison helper `testGoldenFile(t, inputName)` in internal/filter/filter_test.go that parses input from `testdata/comments/<name>.proto`, calls `ConvertBlockComments`, writes to temp file via `writer.WriteProtoFile`, reads golden from `testdata/comments/expected/<name>.proto`, and compares content
- [x] T010 [US1] Write test `TestGoldenFileMultiline` in internal/filter/filter_test.go that calls `testGoldenFile(t, "multiline")` to verify block comments are converted correctly
- [x] T011 [P] [US1] Write test `TestGoldenFileCommented` in internal/filter/filter_test.go that calls `testGoldenFile(t, "commented")` to verify single-line comments pass through unchanged
- [x] T012 [P] [US1] Write test `TestGoldenFileBlockComments` in internal/filter/filter_test.go that calls `testGoldenFile(t, "block_comments")` to verify the new fixture with varied patterns converts correctly

**Checkpoint**: All golden file comparison tests pass. Block comments are converted, single-line comments are unchanged.

---

## Phase 4: User Story 2 - Content Preservation (Priority: P1)

**Goal**: Verify annotations and text content are preserved exactly during conversion

**Independent Test**: Parse converted output and verify annotation text and descriptive content matches original

### Tests for User Story 2

- [x] T013 [US2] Write test `TestConvertBlockCommentsPreservesAnnotations` in internal/filter/filter_test.go that parses testdata/comments/block_comments.proto, calls `ConvertBlockComments`, and verifies annotations like `@StartsWithSnapshot` appear in the resulting `Lines` unchanged
- [x] T014 [US2] Write test `TestConvertBlockCommentsEmptyComment` in internal/filter/filter_test.go that creates a proto AST with an empty block comment (`/* */`), calls `ConvertBlockComments`, and verifies the comment is converted to `//` style with empty or no lines

**Checkpoint**: Content preservation verified. Annotations and text survive conversion.

---

## Phase 5: Pipeline Wiring & CLI Integration

**Purpose**: Wire `ConvertBlockComments` into main.go pipeline and verify end-to-end

- [x] T015 Wire `filter.ConvertBlockComments(pf.def)` call into main.go processing loop, placed after all filtering passes but before `writer.WriteProtoFile`, applied unconditionally to every processed file
- [x] T016 Write CLI integration test `TestCommentConversionCLI` in main_test.go that runs proto-filter with testdata/comments/ as input (no config), reads output files, and compares against golden files in testdata/comments/expected/
- [x] T017 Write CLI integration test `TestCommentConversionWithFiltering` in main_test.go that runs proto-filter with both a filter config and comment conversion active, verifying comments are converted in filtered output

**Checkpoint**: Full end-to-end pipeline works. `go test -race ./...` passes across all packages.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Validate all success criteria and run final checks

- [x] T018 Run `go test -race ./...` and verify all tests pass across all packages
- [x] T019 Run quickstart.md validation: process a proto file with block comments through the CLI and verify output matches expected format

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies - T001, T002, T003 are parallel; T004 depends on T001
- **Phase 2 (Foundational)**: Depends on Phase 1; T005-T006 (tests) before T007-T008 (implementation)
- **Phase 3 (US1 Golden Files)**: Depends on Phase 2; T009 first, then T010-T012 parallel
- **Phase 4 (US2 Content)**: Depends on Phase 2; can run in parallel with Phase 3
- **Phase 5 (Pipeline)**: Depends on Phases 2-4; T015 first, then T016-T017 parallel
- **Phase 6 (Polish)**: Depends on all previous phases

### User Story Dependencies

- **US1 (Block-to-Single-Line)**: Can start after Phase 2 - no US2 dependency
- **US2 (Content Preservation)**: Can start after Phase 2 - no US1 dependency

### Parallel Opportunities

- T001, T002, T003 can run in parallel (different files)
- T005, T006 can run in parallel (different test cases)
- T010, T011, T012 can run in parallel after T009 (different golden file tests)
- T013, T014 can run in parallel (different test cases)
- T016, T017 can run in parallel after T015 (different CLI tests)

---

## Parallel Example: Phase 1

```bash
# Launch all fixture creation tasks together:
Task: "Create block_comments.proto fixture in testdata/comments/block_comments.proto"
Task: "Create golden file testdata/comments/expected/commented.proto"
Task: "Create golden file testdata/comments/expected/multiline.proto"
```

---

## Implementation Strategy

### MVP First (Phases 1-3)

1. Complete Phase 1: Create fixtures and golden files
2. Complete Phase 2: Implement `ConvertBlockComments` with unit tests
3. Complete Phase 3: Golden file comparison tests
4. **STOP and VALIDATE**: All golden file tests pass

### Full Delivery

1. Phases 1-3 → Golden file tests pass (MVP)
2. Phase 4 → Content preservation tests pass
3. Phase 5 → CLI integration tests pass
4. Phase 6 → Full validation complete

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story
- Golden files are the primary validation mechanism
- Constitution IV requires tests before/alongside implementation
- Commit after each phase completion
