# Quickstart: C#-Style Annotation Syntax Support

## What Changed

proto-filter now recognizes C#-style bracket annotation syntax
(`[Name]` and `[Name(value)]`) in proto comments, in addition to the
existing Java-style `@Name` syntax. Both styles can be used
interchangeably and are matched by the same config entries.

## New Syntax Examples

You can now annotate proto services and methods using brackets:

```protobuf
// [Internal]
// Administrative operations.
service AdminService {
  rpc ManageUsers(ManageUsersRequest) returns (ManageUsersResponse);
}

// [HasAnyRole("ADMIN")]
// Creates a new order.
rpc CreateOrder(CreateOrderRequest) returns (CreateOrderResponse);
```

## Config Is Unchanged

The `annotations` key in your YAML config works the same way — specify
annotation names without any syntax prefix:

```yaml
annotations:
  - "Internal"
  - "HasAnyRole"
```

This config matches both `@Internal` and `[Internal]`, and both
`@HasAnyRole("ADMIN")` and `[HasAnyRole("ADMIN")]`.

## Mixed Styles Work

You can mix `@Name` and `[Name]` styles in the same file or across
files. For example:

```protobuf
service OrderService {
  // @HasAnyRole({"ADMIN"})
  rpc CreateOrder(CreateOrderRequest) returns (CreateOrderResponse);

  // [HasAnyRole]
  rpc DeleteOrder(DeleteOrderRequest) returns (DeleteOrderResponse);

  // Public method, no annotation.
  rpc ListOrders(ListOrdersRequest) returns (ListOrdersResponse);
}
```

With `annotations: ["HasAnyRole"]`, both `CreateOrder` and `DeleteOrder`
are removed regardless of their annotation syntax style.

## What Is Not Matched

Plain English text in brackets is NOT treated as an annotation:
- `// See [RFC 7231]` — contains a space, not matched
- `// Returns [error code]` — contains a space, not matched
- `// []` — empty brackets, not matched
- `// [ Name ]` — leading space inside bracket, not matched

## Running

```bash
proto-filter -input ./protos -output ./filtered -config filter.yaml
```

No command-line changes. The tool automatically recognizes both syntax
styles.
