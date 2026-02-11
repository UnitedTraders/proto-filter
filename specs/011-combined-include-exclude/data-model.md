# Data Model: Combined Include and Exclude Annotation Filters

## Entity Changes

### AnnotationConfig (modified)

**Location**: `internal/config/config.go:13-16`

```
AnnotationConfig
├── Include []string   — annotation names for include mode (service/method level)
└── Exclude []string   — annotation names for exclude mode (service/method/field level)
```

**Change**: No structural change. The only modification is removing the validation constraint that prevents both `Include` and `Exclude` from being non-empty simultaneously.

**Valid states (before)**:
- Include populated, Exclude empty
- Include empty, Exclude populated
- Both empty

**Valid states (after)**:
- Include populated, Exclude empty
- Include empty, Exclude populated
- Both populated ← NEW
- Both empty

### FilterConfig.Validate() (modified)

**Location**: `internal/config/config.go:73-78`

**Before**: Returns error when both `Annotations.Include` and `Annotations.Exclude` are non-empty.
**After**: No annotation-specific validation. Method body may become empty or contain only future validation rules.

## New Function

### FilterFieldsByAnnotation

**Signature**: `func FilterFieldsByAnnotation(def *proto.Proto, annotations []string) int`

**Purpose**: Remove individual message fields whose comments contain a matching exclude annotation.

**Input**: Parsed proto AST, list of annotation names to filter.
**Output**: Count of fields removed.

**Iteration targets** (within each `*proto.Message` in `def.Elements`):

| Element Type | Comment Sources | Action |
|---|---|---|
| `*proto.NormalField` | `.Comment`, `.InlineComment` | Remove if annotation matches |
| `*proto.MapField` | `.Comment`, `.InlineComment` | Remove if annotation matches |
| `*proto.Oneof` | Iterate `.Elements` for `*proto.OneOfField` | Remove matching fields from oneof |

**Nested messages**: The function must also iterate into nested `*proto.Message` elements within a message (recursive or single-level depending on codebase patterns).

## Processing Order (main.go)

```text
1. Parse all files, build dependency graph
2. Apply name-based include/exclude filter (ApplyFilter)
3. Resolve transitive dependencies
4. For each file:
   a. PruneAST (name-based filtering)
   b. IncludeServicesByAnnotation (if annotations.include configured)
   c. IncludeMethodsByAnnotation (if annotations.include configured)
   d. FilterServicesByAnnotation (if annotations.exclude configured)
   e. FilterMethodsByAnnotation (if annotations.exclude configured)
   f. FilterFieldsByAnnotation (if annotations.exclude configured)  ← NEW
   g. RemoveEmptyServices
   h. RemoveOrphanedDefinitions (if any services/methods/fields removed)
   i. ConvertBlockComments
   j. CollectAnnotationLocations (if strict substitution)
5. Strict substitution check
6. SubstituteAnnotations + write output
```
