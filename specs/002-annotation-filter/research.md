# Research: Annotation-Based Method Filtering

**Date**: 2026-02-08
**Branch**: `002-annotation-filter`

## Decision 1: Annotation Parsing Strategy

**Decision**: Parse annotations from `proto.Comment.Lines` using a
regex pattern `@(\w[\w.]*)` to extract annotation names from RPC
method comments.

**Rationale**: Annotations are Java-style tokens embedded in proto
comments (e.g., `// @HasAnyRole({"ADMIN"})`). The `emicklei/proto`
library preserves comments as `Comment.Lines []string` on every AST
element including `*proto.RPC`. Parsing is straightforward string
matching — no need for a full annotation parser since we match by
name only and ignore arguments (per FR-003).

The regex `@(\w[\w.]*)` captures:
- `@HasAnyRole` → name: `HasAnyRole`
- `@HasAnyRole({"ADMIN"})` → name: `HasAnyRole` (parenthesized args ignored)
- `@com.example.Secure` → name: `com.example.Secure` (dots allowed)
- `@Deprecated` → name: `Deprecated`

Both `//` and `/* */` comment styles are handled identically since
`emicklei/proto` normalizes both into `Comment.Lines`.

**Alternatives considered**:

| Approach | Verdict | Why rejected |
|----------|---------|--------------|
| Full annotation parser (with arg parsing) | Unnecessary | FR-003 explicitly says ignore arguments; adds complexity for no value |
| Simple `strings.Contains` with `@` prefix | Fragile | Would match `@` in email addresses or doc text; regex is more precise |
| External annotation library | Overkill | No Go library exists for this niche; a single regex suffices |

## Decision 2: Method-Level Filtering Architecture

**Decision**: Implement annotation filtering as a separate pass that
operates on service elements (RPC methods) after the existing
definition-level filtering, rather than extending the existing
`ApplyFilter`/`PruneAST` functions.

**Rationale**: The existing filter pipeline operates on top-level
definitions (services, messages, enums) using FQN glob matching.
Annotation filtering operates at a different granularity (individual
RPC methods within services) and uses a fundamentally different
matching mechanism (comment content, not FQN patterns). Separating
these concerns keeps the existing filter logic unchanged (backward
compatibility per SC-004) and makes the new logic independently
testable.

The new pipeline step is:
1. Existing: `ApplyFilter` → determines which services/messages/enums to keep
2. **New**: `FilterMethodsByAnnotation` → removes annotated RPC methods from kept services
3. **New**: `RemoveEmptyServices` → drops services with zero remaining methods
4. **New**: `FindOrphanedDefinitions` → removes messages/enums no longer referenced

**Alternatives considered**:

| Approach | Verdict | Why rejected |
|----------|---------|--------------|
| Extend `ApplyFilter` with annotation logic | Rejected | Mixes two different matching paradigms (FQN glob vs. comment content); harder to test and maintain |
| Track methods as graph nodes | Rejected | Over-engineering; methods don't need full graph traversal — orphan detection can use a simpler reference-counting approach on the already-pruned AST |

## Decision 3: Orphan Detection Strategy

**Decision**: Use reverse-reference counting on the post-method-filtered
AST rather than extending the dependency graph with method-level nodes.

**Rationale**: After method filtering, we need to find messages/enums
that are no longer referenced. The existing dependency graph tracks
definition-level edges which are too coarse (a service's references
include ALL its methods' types, not just surviving ones). Rather than
rebuilding the graph, we can:
1. Collect all type references from surviving RPC methods and surviving messages
2. Mark any message/enum not in this reference set as orphaned
3. Repeat until no new orphans are found (transitive orphan detection)

This is simpler than extending the graph and aligns with the feature's
scope — orphan detection is only needed when annotation filtering
removes methods.

**Alternatives considered**:

| Approach | Verdict | Why rejected |
|----------|---------|--------------|
| Extend `deps.Graph` with method nodes | Over-engineering | Adds complexity to a working system; method-level graph not needed elsewhere |
| Single-pass reference scan | Insufficient | Misses transitive orphans (A → B → removed method) |

## Decision 4: New Package or Existing Package

**Decision**: Add annotation filtering logic to the existing
`internal/filter/` package as new functions, keeping the existing
functions unchanged.

**Rationale**: Constitution principle V (Simplicity) caps at "one or
two internal packages." We already have five (`parser`, `filter`,
`deps`, `writer`, `config`). Adding a sixth package for annotation
filtering would violate the spirit of this constraint. The annotation
filter functions naturally belong alongside the existing filter logic.
The config extension (`Annotations` field) belongs in the existing
config package.

**Alternatives considered**:

| Approach | Verdict | Why rejected |
|----------|---------|--------------|
| New `internal/annotations/` package | Rejected | Would add a 6th internal package; violates constitution principle V spirit |
| Inline everything in `main.go` | Rejected | Would bloat `main.go` and make logic untestable in isolation |

## Decision 5: Config Format for Annotations

**Decision**: Add an `annotations` key to the existing YAML config
as a simple string list. Names are specified without the `@` prefix.

**Rationale**: Clarified during `/speckit.clarify` session. The
simplest format that meets the requirement — methods with any of the
listed annotations are removed. No include/exclude sub-structure
needed since the entire feature's semantic is removal of annotated
methods.

```yaml
include:
  - "my.package.OrderService"
exclude:
  - "my.package.internal.*"
annotations:
  - "HasAnyRole"
  - "Internal"
```

**Alternatives considered**:

| Approach | Verdict | Why rejected |
|----------|---------|--------------|
| Include/exclude sub-keys for annotations | Rejected | Over-engineering; user requested simple removal semantics |
| Prefixed format (`@HasAnyRole`) | Rejected | Redundant prefix; annotation names are inherently `@`-prefixed in source, config should be clean |
