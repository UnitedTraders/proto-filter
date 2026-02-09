# Research: Service-Level Annotation Filtering

**Date**: 2026-02-09
**Branch**: `004-service-annotation-filter`

## Decision 1: Extend Existing FilterMethodsByAnnotation vs. New Function

**Decision**: Create a new function `FilterServicesByAnnotation` in
`internal/filter/filter.go` that runs before the existing
`FilterMethodsByAnnotation`. The new function removes entire services
from the AST when their service-level comment contains a matching
annotation.

**Rationale**: The existing `FilterMethodsByAnnotation` iterates over
services and filters their RPC elements. Service-level filtering has
different semantics — it removes the entire service element from
`def.Elements`, not individual RPCs. Mixing both concerns in one
function would complicate the logic and make testing harder. A separate
function keeps each responsibility clear and allows independent testing.

The pipeline order in `main.go` becomes:
1. `FilterServicesByAnnotation` → removes entire services by annotation
2. `FilterMethodsByAnnotation` → removes individual methods from surviving services
3. `RemoveEmptyServices` → drops services with zero remaining methods
4. `RemoveOrphanedDefinitions` → cleans up unreferenced messages/enums

**Alternatives considered**:

| Approach | Verdict | Why rejected |
|----------|---------|--------------|
| Extend `FilterMethodsByAnnotation` | Rejected | Mixes two granularities (service-level vs. method-level) in one function; harder to test independently |
| Add service-level check inside existing annotation block in `main.go` | Rejected | Inlines filtering logic in the pipeline; violates separation of concerns |

## Decision 2: Reuse ExtractAnnotations for Service Comments

**Decision**: Reuse the existing `ExtractAnnotations(comment *proto.Comment)`
function to extract annotations from `Service.Comment`. No changes to
the extraction logic are needed.

**Rationale**: The `emicklei/proto` library stores service-level comments
in `proto.Service.Comment` as a `*proto.Comment` — the same type used
for method comments. The `ExtractAnnotations` function already handles
this type generically. FR-003 requires using the same matching logic
for services as for methods, which is automatically satisfied by reusing
the same function.

**Alternatives considered**:

| Approach | Verdict | Why rejected |
|----------|---------|--------------|
| Write a service-specific annotation extractor | Rejected | Would duplicate identical logic; same `Comment` struct is used |

## Decision 3: Return Value and Verbose Output

**Decision**: `FilterServicesByAnnotation` returns the count of removed
services (int). The verbose output in `main.go` will report "removed N
services by annotation" alongside the existing method/orphan counts.

**Rationale**: Consistent with existing `FilterMethodsByAnnotation`
which returns removed method count. The verbose output helps users
understand what was filtered and why, supporting debugging.

**Alternatives considered**:

| Approach | Verdict | Why rejected |
|----------|---------|--------------|
| Return list of removed service names | Over-engineering | Count is sufficient for verbose output; names can be logged separately if needed in future |
| No return value | Insufficient | Verbose output needs the count to report to users |

## Decision 4: Test Fixture Strategy

**Decision**: Create new test fixture files under `testdata/annotations/`
alongside existing fixtures. Add a `service_annotated.proto` file with
services that have annotations at the service level.

**Rationale**: The existing `testdata/annotations/` directory already
contains `service.proto` (method-level annotations), `shared.proto`
(shared messages), and `internal_only.proto` (all-methods-annotated).
Adding a new file for service-level annotations keeps test isolation
clean and avoids modifying existing fixtures that other tests depend on.

**Alternatives considered**:

| Approach | Verdict | Why rejected |
|----------|---------|--------------|
| Modify existing `service.proto` | Rejected | Would break existing tests that count specific methods/services |
| Create a new testdata subdirectory | Rejected | Unnecessary; existing `testdata/annotations/` is the right home for annotation-related fixtures |
