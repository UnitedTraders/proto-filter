# Research: Fix Include Filter Keeping Unannotated Services

**Date**: 2026-02-11

## No NEEDS CLARIFICATION Items

All technical context is fully resolved. This is a focused bug fix with no unknowns.

## Root Cause Analysis

- **Decision**: The bug is in `IncludeServicesByAnnotation` at `internal/filter/filter.go:280-285`
- **Rationale**: The function explicitly keeps services with `len(annots) == 0`, but the spec requires that include annotations act as a gate — services without any annotations don't pass the gate
- **Alternatives considered**: None — the spec is unambiguous. Feature 011's spec explicitly states: "Services without annotations and without include matches are excluded."

## Impact Analysis

- **Decision**: 2 tests will break and need updating; 1 test fixture needs a service-level annotation added
- **Rationale**: `TestIncludeServicesByAnnotation` asserts unannotated services are kept; `TestIncludeAnnotationFilteringCLI` uses an input fixture with an unannotated service
- **Alternatives considered**: Could add a separate flag/parameter to control behavior, but this adds unnecessary complexity for what is clearly a bug

## Fix Approach

- **Decision**: Delete the `if len(annots) == 0` early-return block entirely
- **Rationale**: When `annots` is empty, the existing `hasMatch` loop produces `false`, and the service is correctly excluded. No new code needed — just removing the incorrect special case
- **Alternatives considered**: Could add a boolean parameter to control the behavior, but the spec is clear that unannotated services should never pass the include gate
