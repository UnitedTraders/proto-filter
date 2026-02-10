# Quickstart: Annotation Substitution

**Feature**: 008-annotation-substitution

## Replace Annotations with Descriptions

Given a proto file with annotated methods:

```protobuf
service OrderService {
  // @HasAnyRole({"ADMIN", "MANAGER"})
  // Creates a new order in the system.
  rpc CreateOrder(CreateOrderRequest) returns (CreateOrderResponse);

  // @Internal
  // Deletes an existing order.
  rpc DeleteOrder(DeleteOrderRequest) returns (DeleteOrderResponse);

  // Lists all orders for the current user.
  rpc ListOrders(ListOrdersRequest) returns (ListOrdersResponse);
}
```

With config:

```yaml
substitutions:
  HasAnyRole: "Requires authentication"
  Internal: "For internal use only"
```

Run:

```bash
proto-filter --input ./protos --output ./out --config filter.yaml
```

Result:

```protobuf
service OrderService {
  // Requires authentication
  // Creates a new order in the system.
  rpc CreateOrder(CreateOrderRequest) returns (CreateOrderResponse);

  // For internal use only
  // Deletes an existing order.
  rpc DeleteOrder(DeleteOrderRequest) returns (DeleteOrderResponse);

  // Lists all orders for the current user.
  rpc ListOrders(ListOrdersRequest) returns (ListOrdersResponse);
}
```

## Remove Annotations with Empty Substitution

```yaml
substitutions:
  HasAnyRole: ""
  Internal: ""
```

Result: All annotation lines are removed, only descriptive comments remain:

```protobuf
service OrderService {
  // Creates a new order in the system.
  rpc CreateOrder(CreateOrderRequest) returns (CreateOrderResponse);

  // Deletes an existing order.
  rpc DeleteOrder(DeleteOrderRequest) returns (DeleteOrderResponse);

  // Lists all orders for the current user.
  rpc ListOrders(ListOrdersRequest) returns (ListOrdersResponse);
}
```

## Strict Mode — Enforce Complete Mappings

```yaml
substitutions:
  HasAnyRole: "Requires authentication"
strict_substitutions: true
```

If the input contains `@Internal` (which has no mapping):

```bash
proto-filter --input ./protos --output ./out --config filter.yaml
# Exit code 2
# stderr: proto-filter: error: unsubstituted annotations found: Internal
```

Add the missing mapping to fix:

```yaml
substitutions:
  HasAnyRole: "Requires authentication"
  Internal: ""
strict_substitutions: true
```

Now the tool succeeds (empty string is a valid mapping).

## Combined with Annotation Filtering

Substitution works alongside annotation include/exclude filtering:

```yaml
annotations:
  exclude:
    - "Internal"
substitutions:
  HasAnyRole: "Requires authentication"
```

Result: Methods with `@Internal` are removed entirely, while `@HasAnyRole` annotations on remaining methods are replaced with "Requires authentication".

## Bracket Syntax Works Too

Both `@Name` and `[Name]` annotations are substituted:

```protobuf
// [HasAnyRole({"ADMIN"})]
// Creates a new order.
rpc CreateOrder(CreateOrderRequest) returns (CreateOrderResponse);
```

With `substitutions: { HasAnyRole: "Auth required" }` → output:

```protobuf
// Auth required
// Creates a new order.
rpc CreateOrder(CreateOrderRequest) returns (CreateOrderResponse);
```
