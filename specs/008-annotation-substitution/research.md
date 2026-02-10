# Research: Annotation Substitution

**Feature**: 008-annotation-substitution

## Decision 1: Regex for full annotation expression matching

**Decision**: Create a new regex that matches the full annotation expression including parameters, suitable for replacement. The existing `annotationRegex` extracts only the annotation name (for filtering decisions). Substitution needs to match the entire `@Name(...)` or `[Name(...)]` token so it can be replaced in-place.

**Rationale**: The existing regex `@(\w[\w.]*)|\[(\w[\w.]*)(?:\([^)]*\))?\]` is designed for extraction — it captures names but doesn't fully match `@Name(...)` with parameters. For `@HasAnyRole({"ADMIN", "MANAGER"})`, the current regex only matches `@HasAnyRole` (the `(...)` part is not consumed for `@` style). Substitution requires matching the full token so the replacement is clean.

**New regex approach**: A substitution-specific regex that matches:
- `@Name` (no params) — e.g., `@Internal`
- `@Name(...)` (with params, balanced parens) — e.g., `@HasAnyRole({"ADMIN", "MANAGER"})`
- `[Name]` (no params) — e.g., `[Public]`
- `[Name(...)]` (with params) — e.g., `[HasAnyRole({"ADMIN"})]`

Pattern: `@(\w[\w.]*)(?:\([^)]*\))?|\[(\w[\w.]*)(?:\([^)]*\))?\]`

This captures the full match (group 0) for replacement and the name (group 1 or 2) for lookup in the substitution map.

**Alternatives considered**:
- Reuse existing regex: Rejected — it doesn't match `@Name(...)` parameters for `@`-style annotations.
- String manipulation instead of regex: Rejected — error-prone with edge cases like nested brackets.

## Decision 2: Config structure for substitutions

**Decision**: Add `Substitutions map[string]string` and `StrictSubstitutions bool` as top-level fields in `FilterConfig`.

**Rationale**: Substitutions are independent of the annotation include/exclude filtering. They operate on comment text, not on which services/methods to keep. Placing them at the top level of the config keeps the YAML structure flat and simple:

```yaml
substitutions:
  HasAnyRole: "Requires authentication"
  Internal: ""
strict_substitutions: true
```

**Alternatives considered**:
- Nest under `annotations:` key: Rejected — substitutions are a comment rewriting feature, not a filtering feature. Nesting would create confusion with the existing `annotations.include`/`annotations.exclude` semantic.
- Separate config file: Rejected — adds unnecessary complexity for a simple key-value mapping.

## Decision 3: Processing order in the pipeline

**Decision**: Apply substitutions after annotation-based include/exclude filtering and after block comment conversion, but before writing output. Strict mode validation (collecting all annotations) must happen as a separate pass before any substitution or writing.

**Rationale**:
1. Include/exclude filtering may remove entire services/methods. There's no point substituting annotations in elements that will be removed.
2. Block comment conversion (`ConvertBlockComments`) normalizes comment format first, making substitution logic simpler (only deal with single-line `//` style comments).
3. Strict mode must scan all files first to collect all unsubstituted annotations before any files are written (FR-010 requires no output on strict failure).

**Pipeline order**:
1. Parse files
2. FQN filtering (include/exclude glob patterns)
3. Annotation-based filtering (include/exclude by annotation)
4. Block comment conversion
5. **Strict mode check** (if enabled): scan all remaining comments, collect annotation names, compare against substitution map. If any missing → error, exit, no output.
6. **Apply substitutions**: replace annotation tokens in comments
7. Write output files

**Alternatives considered**:
- Apply substitutions inline during annotation filtering: Rejected — conflates two independent features, makes testing harder.
- Apply substitutions before block comment conversion: Rejected — would need to handle both `/* */` and `//` comment styles.

## Decision 4: Empty line cleanup strategy

**Decision**: After substituting annotations in a comment, remove any lines that are empty or whitespace-only. If all lines are removed, set the comment pointer to nil on the AST element.

**Rationale**: When an annotation like `@HasAnyRole({"ADMIN"})` occupies an entire comment line and is mapped to `""`, the line becomes empty after substitution. Leaving empty comment lines (`//`) in the output is visually noisy. Setting the comment to nil when all lines are removed avoids orphaned empty comment markers in the output.

**Alternatives considered**:
- Leave empty lines as-is: Rejected — produces messy output like `//\n// Creates an order.`
- Remove only the annotation but keep surrounding whitespace: Rejected — same messy output problem.

## Decision 5: Strict mode annotation collection scope

**Decision**: Strict mode collects annotations from all comments on surviving elements (after include/exclude filtering). Annotations that were on removed services/methods are NOT counted as unsubstituted.

**Rationale**: If a service is excluded by `annotations.exclude`, its annotations are irrelevant to the output. Only annotations that will appear in the output need substitution mappings. This aligns with FR-013 (substitution operates after filtering).

**Alternatives considered**:
- Collect from all files before filtering: Rejected — would require substitution mappings for annotations on elements that are being excluded, which is counterintuitive.
