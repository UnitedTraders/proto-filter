# Quickstart: Annotation Error Locations

**Feature**: 009-annotation-error-locations

## Strict Mode Error with Locations

Given proto files with annotated methods:

```protobuf
// orders.proto
service OrderService {
  // @Deprecated
  // Use CreateOrderV2 instead.
  rpc CreateOrder(CreateOrderRequest) returns (CreateOrderResponse);

  // @SupportWindow({duration: "6M"})
  // Returns order details.
  rpc GetOrder(GetOrderRequest) returns (GetOrderResponse);

  // Lists all orders.
  rpc ListOrders(ListOrdersRequest) returns (ListOrdersResponse);
}
```

With config:

```yaml
substitutions:
  SupportWindow: "Supported for 6 months"
strict_substitutions: true
```

Run:

```bash
proto-filter --input ./protos --output ./out --config filter.yaml
# Exit code 2
```

Stderr output:

```
proto-filter: error: unsubstituted annotations found: Deprecated
  orders.proto:4: @Deprecated
```

The summary line lists the unique missing annotation names. Below it, each occurrence is shown with its file path and line number, so you can navigate directly to the source.

## Multiple Files, Multiple Annotations

Given two proto files with various annotations, and a config mapping only some of them:

```yaml
substitutions:
  HasAnyRole: "Requires authentication"
strict_substitutions: true
```

Stderr output:

```
proto-filter: error: unsubstituted annotations found: Deprecated, Internal
  orders.proto:4: @Deprecated
  payments.proto:5: @Internal
  payments.proto:11: @Deprecated
```

All occurrences across all files are listed, ordered by file path then line number.

## Complete Mapping — No Error

When all annotations have a substitution mapping, strict mode succeeds silently:

```yaml
substitutions:
  HasAnyRole: "Requires authentication"
  Deprecated: ""
  Internal: "For internal use only"
strict_substitutions: true
```

```bash
proto-filter --input ./protos --output ./out --config filter.yaml
# Exit code 0 — no error output
```
