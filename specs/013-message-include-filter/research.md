# Research: Include Annotation Filtering for Top-Level Messages

**Date**: 2026-02-11

## No NEEDS CLARIFICATION Items

All technical context is fully resolved.

## Implementation Approach

- **Decision**: Add `IncludeMessagesByAnnotation` function following the `IncludeServicesByAnnotation` pattern, then rely on existing orphan removal for dependency preservation
- **Rationale**: `CollectReferencedTypes` already collects references from message fields (NormalField, MapField, OneOfField). When an annotated message is kept, its field references are collected, and orphan removal preserves those referenced types. This avoids building a separate dependency resolution mechanism.
- **Alternatives considered**: (1) Modify `RemoveOrphanedDefinitions` to handle include filtering directly — rejected because it conflates two concerns and the existing function is well-tested. (2) Build a separate dependency graph for messages — rejected because `CollectReferencedTypes` + iterative orphan removal already provides this.

## Orphan Removal Gate

- **Decision**: Add message removal count (`msgr`) to the orphan removal gate condition in `main.go`
- **Rationale**: The current gate `if sr > 0 || mr > 0 || fr > 0` skips orphan removal when no services/methods/fields are removed. For messages-only files, `sr == 0` and no gate triggers, so orphan removal never runs. Adding `msgr > 0` ensures orphan removal runs after message-level include filtering.
- **Alternatives considered**: Always run orphan removal when include is configured — rejected because it adds unnecessary work for files where include doesn't remove anything.

## Enum Handling

- **Decision**: Handle enums identically to messages in `IncludeMessagesByAnnotation`
- **Rationale**: Enums are top-level definitions that should be subject to the same include gating. The function checks both `*proto.Message` and `*proto.Enum` in the same pass.
- **Alternatives considered**: Separate `IncludeEnumsByAnnotation` function — rejected because enums and messages have the same comment structure and the combined function is simpler.

## Annotation Stripping

- **Decision**: No changes needed for annotation stripping on messages
- **Rationale**: `StripAnnotations` calls `SubstituteAnnotations`, which already walks messages and their comments (lines 665-679 in filter.go). Include annotations on kept messages will be automatically stripped.
