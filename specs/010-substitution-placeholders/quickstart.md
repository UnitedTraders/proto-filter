# Quickstart: Substitution Placeholders

## What This Feature Does

When you define a substitution mapping with a `%s` placeholder, proto-filter now interpolates the annotation's argument into the replacement text.

## Before This Feature

```yaml
# config.yaml
substitutions:
  Min: "Minimal value constraint"  # Can't include the actual min value
```

```protobuf
// Input: service.proto
service OrderService {
  // @Min(3)
  // Minimum quantity for an order.
  rpc CreateOrder(CreateOrderRequest) returns (CreateOrderResponse);
}
```

```protobuf
// Output: The "3" from @Min(3) is lost
service OrderService {
  // Minimal value constraint
  // Minimum quantity for an order.
  rpc CreateOrder(CreateOrderRequest) returns (CreateOrderResponse);
}
```

## After This Feature

```yaml
# config.yaml
substitutions:
  Min: "Minimal value is %s"
  Max: "Maximum value is %s"
  HasAnyRole: "Requires roles: %s"
  Deprecated: "This method is deprecated"
```

```protobuf
// Input: service.proto
service OrderService {
  // @Min(3)
  // Minimum quantity for an order.
  rpc CreateOrder(CreateOrderRequest) returns (CreateOrderResponse);

  // @Max(100)
  rpc UpdateOrder(UpdateOrderRequest) returns (UpdateOrderResponse);

  // @HasAnyRole({"ADMIN", "MANAGER"})
  rpc DeleteOrder(DeleteOrderRequest) returns (DeleteOrderResponse);

  // @Deprecated
  rpc OldMethod(OldReq) returns (OldResp);
}
```

```protobuf
// Output: Arguments are interpolated into the replacement text
service OrderService {
  // Minimal value is 3
  // Minimum quantity for an order.
  rpc CreateOrder(CreateOrderRequest) returns (CreateOrderResponse);

  // Maximum value is 100
  rpc UpdateOrder(UpdateOrderRequest) returns (UpdateOrderResponse);

  // Requires roles: {"ADMIN", "MANAGER"}
  rpc DeleteOrder(DeleteOrderRequest) returns (DeleteOrderResponse);

  // This method is deprecated
  rpc OldMethod(OldReq) returns (OldResp);
}
```

## Bracket-Style Annotations

Works identically with `[Name(args)]` style:

```protobuf
// Input:  [Min(3)] Minimum quantity.
// Output: Minimal value is 3 Minimum quantity.
```

## Edge Cases

- **No arguments**: `@Min` with `"Minimal value is %s"` → `Minimal value is`
- **Empty arguments**: `@Min()` with `"Minimal value is %s"` → `Minimal value is`
- **Multiple `%s`**: Only the first `%s` is replaced; subsequent ones remain as literal text.
- **No `%s` in value**: Arguments are ignored — existing behavior unchanged.
