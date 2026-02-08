# Data Model: Proto Filter CLI

**Date**: 2026-02-08
**Branch**: `001-proto-filter-cli`

## Entities

### ProtoFile

Represents a single parsed `.proto` file.

| Field | Type | Description |
|-------|------|-------------|
| RelativePath | string | Path relative to input directory (e.g., `a/b/foo.proto`) |
| Syntax | string | `"proto2"` or `"proto3"` |
| Package | string | Package declaration (e.g., `my.package`) |
| Imports | []string | List of import paths |
| Options | []Option | File-level options |
| Definitions | []Definition | Top-level services, messages, enums, extensions |
| RawAST | *proto.Proto | Underlying parsed AST from `emicklei/proto` |

### Definition

A named top-level proto construct.

| Field | Type | Description |
|-------|------|-------------|
| Kind | enum | `SERVICE`, `MESSAGE`, `ENUM`, `EXTENSION` |
| Name | string | Simple name (e.g., `OrderService`) |
| FullyQualifiedName | string | Package + name (e.g., `my.package.OrderService`) |
| Comment | string | Leading comment text (preserved from source) |
| InlineComment | string | Trailing inline comment |
| References | []string | FQNs of types this definition depends on |

**References extraction rules**:
- **Service**: References all request/response message types of its RPCs
- **Message**: References all field types that are message/enum types
  (including nested messages, map key/value types, oneof field types)
- **Enum**: No outgoing references (enums are leaves)
- **Extension**: References the extended message type

### FilterConfig

Parsed from the YAML configuration file.

| Field | Type | Description |
|-------|------|-------------|
| Include | []string | Glob patterns for definitions to include |
| Exclude | []string | Glob patterns for definitions to exclude |

**YAML schema**:

```yaml
include:
  - "my.package.OrderService"
  - "my.package.common.*"
exclude:
  - "my.package.internal.*"
```

**Semantics**:
- If `include` is non-empty, only matching definitions are kept
  (allowlist mode)
- If `exclude` is non-empty, matching definitions are removed
  (denylist mode)
- Both can be specified: include is applied first, then exclude
  removes from the included set
- A definition matching both include and exclude is an error
  (conflicting rules)

### DependencyGraph

Directed graph of definition-to-definition references.

| Field | Type | Description |
|-------|------|-------------|
| Nodes | map[string]Definition | FQN → Definition lookup |
| Edges | map[string][]string | FQN → list of FQNs it depends on |
| FileMap | map[string]string | FQN → RelativePath of containing file |

**Operations**:
- `TransitiveDeps(fqn string) []string` — returns all FQNs
  transitively required by the given definition
- `RequiredFiles(fqns []string) []string` — returns all file paths
  that must appear in output to satisfy the given set of definitions

## State Transitions

This is a stateless CLI tool. No persistent state or lifecycle
transitions exist. The processing pipeline is:

```
Input Directory
    → Discover .proto files
    → Parse each file into ProtoFile
    → Build DependencyGraph from all ProtoFiles
    → Load FilterConfig (if --config provided)
    → Apply filter rules + resolve transitive deps
    → Determine output file set
    → Generate .proto files via formatter
    → Write to output directory
```
