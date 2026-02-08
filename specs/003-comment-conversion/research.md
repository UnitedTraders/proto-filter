# Research: Comment Style Conversion

## Decision 1: Comment Conversion Mechanism

**Decision**: Set `Comment.Cstyle = false` on all block comments in the AST before writing.

**Rationale**: The `emicklei/proto` parser stores comment text in `Comment.Lines` with syntax prefixes already stripped. The `protofmt.Formatter` decides output format based solely on the `Cstyle` boolean flag. Setting it to `false` causes the formatter to emit `// ...` lines instead of `/* ... */` blocks. No text manipulation of `Lines` content is needed since the parser already handles asterisk stripping.

**Alternatives considered**:
- Post-processing output files with regex: Rejected because it would bypass the AST and risk corrupting proto syntax.
- Custom formatter: Rejected because the existing `protofmt.Formatter` already supports `//` output when `Cstyle = false`.

## Decision 2: Pipeline Placement

**Decision**: Apply comment conversion after all filtering passes but before writing, as an unconditional step in the main processing loop.

**Rationale**: Comment conversion is independent of filtering logic. Placing it just before write ensures it applies to the final AST regardless of filter configuration. This matches FR-006 (must apply in pass-through mode too).

**Alternatives considered**:
- During parsing: Rejected because it would modify the AST before filtering, which could interfere with annotation extraction that reads comment content.
- In the writer: Rejected because the writer should remain a thin wrapper around `protofmt.Formatter`.

## Decision 3: AST Walking Strategy

**Decision**: Use `proto.Walk` with handlers for all element types that carry comments (Service, Message, Enum, RPC, NormalField, MapField, OneOfField, EnumField), converting both `.Comment` and `.InlineComment` fields.

**Rationale**: The `proto.Walk` function is the established pattern in this codebase for AST traversal. It covers all element types that can have attached comments. A single walk pass is sufficient.

**Alternatives considered**:
- Manual recursive traversal: Rejected because `proto.Walk` already handles the recursion and is used consistently throughout the codebase.

## Decision 4: Line Content Cleaning

**Decision**: Strip leading ` * ` and ` *` prefixes from `Lines` entries as a safety measure, even though the parser typically handles this.

**Rationale**: While the `emicklei/proto` parser strips asterisk prefixes in most cases, edge cases exist where leading asterisks or spaces may be preserved differently. A defensive trim of leading `* ` or `*` patterns ensures clean output regardless of parser behavior.

**Alternatives considered**:
- Trust parser completely: Could work for most cases but risks edge case corruption if the parser preserves some asterisk patterns.

## Decision 5: Function Placement

**Decision**: Add `ConvertBlockComments(def *proto.Proto)` to `internal/filter/filter.go`.

**Rationale**: The `filter` package already contains AST transformation functions (`FilterMethodsByAnnotation`, `RemoveEmptyServices`, `RemoveOrphanedDefinitions`). Comment conversion is another AST transformation that fits naturally here. No new package needed, consistent with Constitution Principle V (simplicity).

**Alternatives considered**:
- New `internal/comments` package: Rejected as over-engineering for a single function (Constitution V).
- In `writer` package: Rejected because the writer is responsible for serialization, not transformation.
