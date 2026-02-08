# Feature Specification: Proto Filter CLI

**Feature Branch**: `001-proto-filter-cli`
**Created**: 2026-02-08
**Status**: Draft
**Input**: User description: "The CLI tool should take input directory path and output dir path, parse all files in input dir by '*.proto' pattern and create corresponding files in the output path. All the files should be parsed to structure to make filtering semantic and output files should be generated from the parsed form."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Parse and Copy Proto Files (Priority: P1)

A developer runs the proto-filter tool, providing an input directory
containing `.proto` files and an output directory. The tool discovers
all `.proto` files in the input directory, parses each file into an
internal structural representation (packages, services, messages,
enums, imports), and writes corresponding `.proto` files to the output
directory. Comments attached to services, methods, messages, and fields
MUST be preserved in the parsed structure and included in the generated
output. The output files are generated from the parsed structure
rather than copied verbatim, establishing the parse-then-generate
pipeline that enables future filtering.

**Why this priority**: This is the foundational capability. Without
parsing proto files into a structured form and regenerating them, no
filtering is possible. This story alone delivers a working tool that
validates the parse-generate round-trip.

**Independent Test**: Run the tool against a directory of known `.proto`
files and verify that each output file is syntactically valid and
semantically equivalent to the original.

**Acceptance Scenarios**:

1. **Given** a directory with three `.proto` files, **When** the user
   runs `proto-filter --input ./protos --output ./out`, **Then** the
   output directory contains three `.proto` files with the same names,
   each syntactically valid and containing the same definitions as the
   originals.

2. **Given** an input directory with `.proto` files in nested
   subdirectories, **When** the user runs the tool, **Then** the
   output directory mirrors the subdirectory structure with
   corresponding files in matching paths.

3. **Given** a `.proto` file that imports another `.proto` file within
   the input directory, **When** the tool processes both files,
   **Then** import paths in the output remain correct and consistent.

4. **Given** an input directory containing both `.proto` files and
   non-proto files (e.g., `.txt`, `.go`), **When** the user runs the
   tool, **Then** only `.proto` files appear in the output directory.

5. **Given** a `.proto` file with comments on services, methods,
   messages, and fields, **When** the tool processes it, **Then** all
   comments appear in the corresponding output file attached to the
   same definitions they documented in the original.

---

### User Story 2 - Filter Proto Definitions by Name (Priority: P2)

A developer wants to produce a subset of their proto definitions. They
provide filter rules (via flags or a configuration file) specifying
which packages, services, or messages to include or exclude. The tool
applies these rules during generation, producing output files that
contain only the matching definitions and their transitive dependencies.

**Why this priority**: Filtering is the core value proposition. Once
the parse-generate pipeline exists (US1), adding semantic filtering
transforms the tool from a proto reformatter into a proto filter.

**Independent Test**: Run the tool with a filter that includes a
specific service, then verify the output contains only that service,
its request/response messages, and their dependencies.

**Acceptance Scenarios**:

1. **Given** a proto file with services `OrderService` and
   `UserService`, **When** the user runs the tool with a filter to
   include only `OrderService`, **Then** the output file contains
   `OrderService` and its method request/response types, but not
   `UserService` or its exclusive types.

2. **Given** a filter that includes a message type used by multiple
   services, **When** the tool runs, **Then** the shared message
   appears in the output along with all services that reference it
   (if those services are also included).

3. **Given** a filter that includes a message with nested message
   dependencies across files, **When** the tool runs, **Then** all
   transitively required messages and their containing files appear
   in the output.

4. **Given** conflicting include and exclude rules for the same
   definition, **When** the tool runs, **Then** it reports a clear
   error and exits with a non-zero code.

---

### User Story 3 - Validate and Report Errors (Priority: P3)

A developer runs the tool against a directory with malformed proto
files or invalid filter rules. The tool reports clear, actionable
error messages to stderr and exits with a non-zero exit code. When
the tool succeeds, it optionally reports a summary of what was
processed and filtered.

**Why this priority**: Robust error handling and diagnostics make the
tool usable in CI pipelines where silent failures are unacceptable.
This builds on US1 and US2 by hardening the user experience.

**Independent Test**: Run the tool against known-bad inputs and verify
error messages are specific and exit codes are non-zero.

**Acceptance Scenarios**:

1. **Given** an input directory that does not exist, **When** the user
   runs the tool, **Then** it prints an error to stderr indicating
   the directory was not found and exits with code 1.

2. **Given** a `.proto` file with syntax errors, **When** the tool
   processes it, **Then** it reports the file name, line number (if
   available), and nature of the error to stderr.

3. **Given** a valid run, **When** the user passes a `--verbose` flag,
   **Then** the tool prints a summary to stderr listing how many files
   were processed, how many definitions were included/excluded, and
   how many output files were written.

---

### Edge Cases

- What happens when the input directory contains zero `.proto` files?
  The tool MUST exit successfully with a warning to stderr.
- What happens when the output directory already exists and contains
  files? The tool MUST overwrite existing files with the same name
  and leave unrelated files untouched.
- What happens when a `.proto` file has `import` statements pointing
  outside the input directory (e.g., well-known types like
  `google/protobuf/timestamp.proto`)? The tool MUST pass through
  the `import` line as-is without attempting to resolve or copy the
  external file. No warning is needed for standard well-known imports.
- What happens when the input and output directories are the same?
  The tool MUST reject this with an error to prevent data corruption.
- What happens when the user lacks write permissions on the output
  directory? The tool MUST report a permission error and exit
  with a non-zero code.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The tool MUST accept an input directory path and an
  output directory path as required arguments.
- **FR-002**: The tool MUST recursively discover all files matching
  the `*.proto` glob pattern within the input directory.
- **FR-003**: The tool MUST parse each discovered `.proto` file into
  a structured representation capturing packages, imports, options,
  services, methods, messages, enums, and extensions.
- **FR-004**: The tool MUST generate output `.proto` files from the
  parsed structural representation, not by copying source text.
- **FR-005**: The tool MUST preserve the relative directory structure
  from input to output (e.g., `input/a/b/foo.proto` produces
  `output/a/b/foo.proto`).
- **FR-006**: The tool MUST create the output directory and any
  necessary subdirectories if they do not exist.
- **FR-007**: The tool MUST produce output `.proto` files that are
  syntactically valid and compilable by `protoc`.
- **FR-008**: The tool MUST support filtering definitions by fully
  qualified name using include and/or exclude rules provided via
  a configuration file specified with a `--config` flag.
- **FR-009**: When filtering is applied, the tool MUST automatically
  include all transitive dependencies (message types, enums) required
  by the included definitions.
- **FR-010**: The tool MUST report errors to stderr and return a
  non-zero exit code on failure.
- **FR-011**: The tool MUST reject the case where input and output
  directories resolve to the same path.
- **FR-012**: The tool MUST support a `--verbose` flag that prints
  processing summary information to stderr.

### Key Entities

- **ProtoFile**: Represents a single `.proto` file; contains its
  relative path, package declaration, imports, and a list of
  top-level definitions.
- **Definition**: A named proto construct — service, message, enum,
  or extension. Each has a fully qualified name derived from its
  package and nesting hierarchy.
- **FilterRule**: An include or exclude directive targeting definitions
  by fully qualified name using glob/wildcard patterns (e.g.,
  `my.package.*`, `*.OrderService`).
- **DependencyGraph**: Tracks which definitions reference other
  definitions, enabling transitive dependency resolution during
  filtering.

### Assumptions

- The tool processes proto2 and proto3 syntax files.
- Filtering granularity is at the top-level definition level
  (services, messages, enums) — not individual fields or methods.
- Import paths in proto files follow the standard `protoc` convention
  (relative to a proto root, not filesystem-absolute).
- The tool does not invoke `protoc` or any external compiler;
  validation means the output is structurally sound, not that it
  links against all external dependencies.
- Filter rules are specified via a YAML configuration file passed
  with a `--config` flag. This suits complex rule sets and keeps
  the CLI interface clean.

## Clarifications

### Session 2026-02-08

- Q: What format should the filter configuration file use? → A: YAML
- Q: What pattern matching style for filter rules? → A: Glob/wildcard patterns
- Q: How to handle well-known proto imports outside input dir? → A: Pass through import lines as-is

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: The tool processes a directory of 100 `.proto` files
  and produces valid output in under 5 seconds on standard hardware.
- **SC-002**: Output `.proto` files pass `protoc` compilation with
  zero errors when all dependencies are present.
- **SC-003**: Round-trip fidelity — processing proto files without
  any filter rules produces output that is semantically equivalent
  to the input (same definitions, same structure).
- **SC-004**: Filtered output contains no orphaned references — every
  type referenced in the output is defined in the output.
- **SC-005**: The tool returns a non-zero exit code for every
  documented error condition (invalid input path, parse errors,
  conflicting filter rules).
