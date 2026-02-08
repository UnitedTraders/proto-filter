# Tasks: Proto Filter CLI

**Input**: Design documents from `/specs/001-proto-filter-cli/`
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

**Purpose**: Project initialization, Go module, dependencies, directory structure

- [x] T001 Initialize Go module with `go mod init github.com/unitedtraders/proto-filter` in go.mod
- [x] T002 Add dependencies: `github.com/emicklei/proto`, `github.com/emicklei/proto-contrib`, `gopkg.in/yaml.v3` in go.mod via `go get`
- [x] T003 [P] Create directory structure: internal/parser/, internal/filter/, internal/deps/, internal/writer/, internal/config/, testdata/simple/, testdata/nested/, testdata/imports/, testdata/comments/, testdata/filter/

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Test fixtures and CLI entry point that all user stories depend on

**CRITICAL**: No user story work can begin until this phase is complete

- [x] T004 Create test fixture proto files in testdata/simple/: a basic proto3 file with one service, two messages, and one enum (testdata/simple/service.proto)
- [x] T005 [P] Create test fixture in testdata/nested/: proto files in subdirectories (testdata/nested/a/orders.proto, testdata/nested/b/users.proto) to test recursive discovery and directory mirroring
- [x] T006 [P] Create test fixture in testdata/imports/: two proto files where one imports the other (testdata/imports/common.proto, testdata/imports/service.proto) to test import handling
- [x] T007 [P] Create test fixture in testdata/comments/: proto file with leading comments, inline comments on services, methods, messages, fields, and enums (testdata/comments/commented.proto)
- [x] T008 [P] Create test fixtures in testdata/filter/: proto files with multiple services, shared messages, and cross-file dependencies for filter testing (testdata/filter/orders.proto, testdata/filter/users.proto, testdata/filter/common.proto)
- [x] T009 Implement CLI entry point with flag parsing (--input, --output, --config, --verbose) and input validation (required flags, same-path rejection) in main.go

**Checkpoint**: Foundation ready — all test fixtures exist, CLI parses flags

---

## Phase 3: User Story 1 — Parse and Copy Proto Files (Priority: P1) MVP

**Goal**: Discover proto files in input directory, parse each into an AST, regenerate via formatter, and write to output directory preserving structure and comments.

**Independent Test**: Run tool against testdata/ fixtures and verify output files are semantically equivalent to originals.

### Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T010 [P] [US1] Unit test for proto file discovery (recursive glob for *.proto, ignoring non-proto files) in internal/parser/parser_test.go
- [x] T011 [P] [US1] Unit test for single proto file parsing: verify extracted package, imports, services, messages, enums, and comments from testdata/simple/service.proto in internal/parser/parser_test.go
- [x] T012 [P] [US1] Unit test for proto file writing: parse a proto file, format with protofmt, verify output is syntactically valid and semantically equivalent in internal/writer/writer_test.go
- [x] T013 [P] [US1] Unit test for comment preservation: parse testdata/comments/commented.proto, write output, verify all leading and inline comments appear on correct definitions in internal/writer/writer_test.go
- [x] T014 [P] [US1] Integration test for full pass-through pipeline: run against testdata/simple/, testdata/nested/, testdata/imports/ and verify output directory structure, file contents, and import paths in internal/parser/parser_test.go (TestIntegrationPassThrough)

### Implementation for User Story 1

- [x] T015 [US1] Implement DiscoverProtoFiles function: recursively walk input directory, match *.proto pattern, return list of relative paths in internal/parser/parser.go
- [x] T016 [US1] Implement ParseProtoFile function: open file, parse with emicklei/proto, return parsed *proto.Proto AST in internal/parser/parser.go
- [x] T017 [US1] Implement WriteProtoFile function: take parsed AST and output path, create parent directories, format with protofmt.NewFormatter, write to file in internal/writer/writer.go
- [x] T018 [US1] Wire pass-through pipeline in main.go: discover files → parse each → write each to output dir preserving relative paths; handle zero-file warning; support --verbose summary to stderr
- [x] T019 [US1] Verify all US1 tests pass with `go test ./...`

**Checkpoint**: Tool can discover, parse, and regenerate proto files with comments preserved. MVP functional.

---

## Phase 4: User Story 2 — Filter Proto Definitions by Name (Priority: P2)

**Goal**: Load YAML filter config, apply include/exclude glob rules, resolve transitive dependencies, and generate output containing only matching definitions.

**Independent Test**: Run tool with --config against testdata/filter/ and verify only included definitions (plus deps) appear in output.

### Tests for User Story 2

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T020 [P] [US2] Unit test for YAML config loading: parse valid config with include/exclude lists, parse empty config (pass-through), reject invalid YAML in internal/config/config_test.go
- [x] T021 [P] [US2] Unit test for glob pattern matching: verify `path.Match` behavior with FQN patterns like `my.package.*`, `*.OrderService`, exact matches in internal/filter/filter_test.go
- [x] T022 [P] [US2] Unit test for dependency graph construction: build graph from parsed proto files, verify edges for service→message, message→message references in internal/deps/deps_test.go
- [x] T023 [P] [US2] Unit test for transitive dependency resolution: given a graph and a set of included FQNs, verify all transitive deps are returned in internal/deps/deps_test.go
- [x] T024 [P] [US2] Unit test for AST filtering: given a parsed proto file and a set of FQNs to keep, verify the AST is pruned to only those definitions (services, messages, enums) in internal/filter/filter_test.go
- [x] T025 [P] [US2] Unit test for conflicting rules detection: include and exclude patterns matching the same FQN must return an error in internal/filter/filter_test.go
- [x] T026 [P] [US2] Integration test for filtered pipeline: run with testdata/filter/ and a config including only OrderService, verify output contains OrderService + its request/response types but not UserService in internal/filter/filter_test.go (TestIntegrationFilter)

### Implementation for User Story 2

- [x] T027 [US2] Implement FilterConfig struct and LoadConfig function: parse YAML file into FilterConfig with Include/Exclude string slices in internal/config/config.go
- [x] T028 [US2] Implement FQN extraction: walk parsed AST to extract fully qualified names for all top-level definitions (package + name) in internal/parser/parser.go (ExtractDefinitions function)
- [x] T029 [US2] Implement reference extraction: for each definition, extract FQNs of referenced types (service RPC request/response types, message field types) in internal/deps/deps.go (ExtractReferences function)
- [x] T030 [US2] Implement DependencyGraph: BuildGraph from list of definitions with references, TransitiveDeps function using BFS/DFS, RequiredFiles function in internal/deps/deps.go
- [x] T031 [US2] Implement ApplyFilter function: take FilterConfig + list of all FQNs, apply include glob patterns, apply exclude glob patterns, detect conflicts, return filtered FQN set in internal/filter/filter.go
- [x] T032 [US2] Implement PruneAST function: given a parsed *proto.Proto and a set of FQNs to keep, remove top-level elements not in the set while preserving syntax, package, imports, and options in internal/filter/filter.go
- [x] T033 [US2] Wire filter pipeline in main.go: if --config provided, load config → extract definitions → build dep graph → apply filter → resolve transitive deps → prune ASTs → write filtered files; update --verbose output with include/exclude counts
- [x] T034 [US2] Verify all US2 tests pass with `go test ./...`

**Checkpoint**: Tool filters proto definitions by glob pattern with transitive dependency resolution.

---

## Phase 5: User Story 3 — Validate and Report Errors (Priority: P3)

**Goal**: Structured error messages to stderr with file/line info, distinct exit codes (1 for runtime, 2 for config), verbose summary.

**Independent Test**: Run tool against invalid inputs and verify specific error messages and exit codes.

### Tests for User Story 3

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T035 [P] [US3] Integration test for missing input directory: run tool with nonexistent --input, verify stderr contains "not found" and exit code 1 in main_test.go
- [x] T036 [P] [US3] Integration test for same input/output path: verify stderr contains error about same path and exit code 1 in main_test.go
- [x] T037 [P] [US3] Integration test for invalid YAML config: run with malformed --config file, verify stderr contains parse error and exit code 2 in main_test.go
- [x] T038 [P] [US3] Integration test for verbose output: run with --verbose against testdata/simple/, verify stderr contains "processed N files" summary in main_test.go

### Implementation for User Story 3

- [x] T039 [US3] Implement structured error formatting: "proto-filter: error: <message>" and "proto-filter: warning: <message>" pattern for all error/warning output to stderr in main.go
- [x] T040 [US3] Implement distinct exit codes: exit 1 for runtime errors (missing dir, parse failure, I/O), exit 2 for config errors (invalid YAML, conflicting rules) in main.go
- [x] T041 [US3] Implement verbose summary: when --verbose is set, print to stderr the count of files processed, definitions found, definitions included/excluded, and files written in main.go
- [x] T042 [US3] Add edge case handling: zero proto files warning, permission errors, proto parse errors with file name in error message in main.go
- [x] T043 [US3] Verify all US3 tests pass with `go test ./...`

**Checkpoint**: All user stories functional with robust error handling.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final validation, cleanup, and documentation

- [x] T044 Run full test suite with race detection: `go test -race ./...`
- [x] T045 Run quickstart.md validation: follow quickstart steps against testdata/ to verify documented usage works
- [x] T046 [P] Verify `CGO_ENABLED=0 go build -o proto-filter .` produces a working static binary
- [x] T047 Code cleanup: remove any TODO comments, ensure consistent error wrapping with `fmt.Errorf` and `%w` across all packages

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion — BLOCKS all user stories
- **User Story 1 (Phase 3)**: Depends on Foundational (Phase 2)
- **User Story 2 (Phase 4)**: Depends on User Story 1 (Phase 3) — uses parser and writer from US1
- **User Story 3 (Phase 5)**: Depends on User Story 2 (Phase 4) — hardens error handling across full pipeline
- **Polish (Phase 6)**: Depends on all user stories being complete

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Parser/model functions before service logic
- Service logic before CLI wiring
- Story complete before moving to next priority

### Parallel Opportunities

- Phase 2: T005, T006, T007, T008 can all run in parallel (independent fixture files)
- Phase 3 tests: T010–T014 can all run in parallel (independent test files/functions)
- Phase 4 tests: T020–T026 can all run in parallel
- Phase 5 tests: T035–T038 can all run in parallel
- Phase 6: T044 and T046 can run in parallel

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001–T003)
2. Complete Phase 2: Foundational (T004–T009)
3. Complete Phase 3: User Story 1 (T010–T019)
4. **STOP and VALIDATE**: `go test ./...` passes, tool round-trips proto files

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. User Story 1 → Parse-generate pipeline works (MVP!)
3. User Story 2 → Semantic filtering with dependency resolution
4. User Story 3 → Robust error handling and diagnostics
5. Polish → CI-ready, documented, static binary

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story
- Constitution principle IV requires TDD: tests before implementation
- All file paths are relative to repository root
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
