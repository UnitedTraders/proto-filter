# Research: Combined Include and Exclude Annotation Filters

## Decision 1: How to Remove Mutual Exclusivity

**Decision**: Remove the `Validate()` error check in `internal/config/config.go:73-76` that rejects configs with both include and exclude populated. Update `TestValidateMutualExclusivity` to expect success instead of error. Update `TestMutualExclusivityErrorCLI` in `main_test.go` to verify combined configs are accepted.

**Rationale**: The validation is a simple 3-line `if` block. Removing it is the minimal change. The `Validate()` method can remain (for future validation rules) with the mutual exclusivity check deleted.

**Alternatives considered**:
- Add a new flag to opt-in to combined mode: Rejected — unnecessary complexity, the behavior is deterministic and safe.
- Add a warning instead of error: Rejected — warnings for intentional behavior are confusing.

## Decision 2: Orchestration of Combined Include + Exclude in main.go

**Decision**: Change the `if/else if` block in `main.go:191-199` to sequential execution: first apply include (if configured), then apply exclude (if configured), regardless of each other. This means both branches can run in a single pass through the file.

**Rationale**: The current code uses `if cfg.HasAnnotationExclude() { ... } else if cfg.HasAnnotationInclude() { ... }` which assumes mutual exclusivity. The fix is to change to two independent `if` blocks: `if cfg.HasAnnotationInclude() { ... }` followed by `if cfg.HasAnnotationExclude() { ... }`. Include must run first to narrow the set, then exclude further refines. The order matches the spec's FR-002 requirement.

**Alternatives considered**:
- Single merged pass: Rejected — harder to reason about, breaks the clean separation of include/exclude concerns.
- New combined function: Rejected — the existing functions work correctly when called sequentially.

## Decision 3: Message Field Types to Filter

**Decision**: `FilterFieldsByAnnotation` will handle three field types from `github.com/emicklei/proto`:
1. `*proto.NormalField` — regular fields (including repeated, optional, required)
2. `*proto.MapField` — map fields
3. `*proto.OneOfField` — fields inside oneof blocks (handled via `*proto.Oneof` container iteration)

Each field type embeds `*proto.Field` which has both `Comment` (leading) and `InlineComment` (trailing) properties. Both must be checked for annotations.

**Rationale**: These are all the field types that appear in message definitions per the proto library. The `*proto.Oneof` type is a container, not a field — its elements contain `*proto.OneOfField` instances that should be individually filterable. Enum fields (`*proto.EnumField`) are excluded per the spec's Assumptions section.

**Alternatives considered**:
- Only filter `NormalField`: Rejected — map fields and oneof fields can also have annotations.
- Skip oneof field iteration: Rejected — oneof fields have comments and should be filterable like any other field.

## Decision 4: Comment Checking Strategy for Fields

**Decision**: Use the existing `ExtractAnnotations(comment *proto.Comment)` function to check both `field.Comment` and `field.InlineComment` for each field. A field is removed if either comment contains a matching exclude annotation.

**Rationale**: The `ExtractAnnotations` function already handles both `@Name` and `[Name]` syntaxes and is used consistently for service and method filtering. Reusing it for fields maintains consistency and satisfies FR-004.

**Alternatives considered**:
- New dedicated field annotation extractor: Rejected — no field-specific annotation syntax exists.

## Decision 5: Verbose Output for Field Removals

**Decision**: Add a new counter `fieldsRemoved` in `main.go` alongside `servicesRemoved` and `methodsRemoved`. Report it in verbose output as: `proto-filter: removed N services by annotation, M methods by annotation, F fields by annotation, O orphaned definitions`.

**Rationale**: FR-010 requires verbose output for field removals. Extending the existing verbose line is simpler than adding a separate line and maintains the current format.

**Alternatives considered**:
- Separate verbose line for fields: Rejected — clutters output, breaks existing verbose format.
- Only report if fields > 0: Rejected — always reporting is consistent with how services/methods are reported.
