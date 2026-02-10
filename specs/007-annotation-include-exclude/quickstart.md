# Quickstart: Annotation Include/Exclude Filtering Modes

**Feature**: 007-annotation-include-exclude

## Include Mode — Keep Only Annotated Services/Methods

Given a proto file with annotated and unannotated methods:

```protobuf
service OrderService {
  // @Public
  rpc ListOrders(ListOrdersRequest) returns (ListOrdersResponse);

  // @Internal
  rpc DeleteOrder(DeleteOrderRequest) returns (DeleteOrderResponse);

  rpc GetStatus(StatusRequest) returns (StatusResponse);
}
```

With config:

```yaml
annotations:
  include:
    - "Public"
```

Run:

```bash
proto-filter --input ./protos --output ./out --config filter.yaml
```

Result: Only `ListOrders` remains. `DeleteOrder` and `GetStatus` are removed (they don't have `@Public`). Orphaned types (`DeleteOrderRequest`, `DeleteOrderResponse`, `StatusRequest`, `StatusResponse`) are cleaned up.

## Exclude Mode — New Config Format

Same as current behavior but with the new structured config key:

```yaml
annotations:
  exclude:
    - "Internal"
```

Result: `DeleteOrder` is removed (has `@Internal`). `ListOrders` and `GetStatus` remain.

## Backward Compatibility — Old Flat Format

The old format still works identically:

```yaml
annotations:
  - "Internal"
```

This is treated as `annotations.exclude: [Internal]` — same behavior as before.

## Error: Both Include and Exclude

```yaml
annotations:
  include:
    - "Public"
  exclude:
    - "Internal"
```

```bash
proto-filter --input ./protos --output ./out --config filter.yaml
# Exit code 2
# stderr: proto-filter: error: annotations.include and annotations.exclude are mutually exclusive
```

## Bracket Syntax Works Too

Both `@Name` and `[Name]` syntaxes work with include mode:

```protobuf
service OrderService {
  // [Public]
  rpc ListOrders(ListOrdersRequest) returns (ListOrdersResponse);

  rpc DeleteOrder(DeleteOrderRequest) returns (DeleteOrderResponse);
}
```

With `annotations.include: [Public]` → `ListOrders` is kept, `DeleteOrder` is removed.
