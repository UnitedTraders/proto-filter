# Quickstart: Annotation-Based Method Filtering

## Prerequisites

- Built `proto-filter` binary (see main README)
- Proto files with Java-style annotations in RPC method comments

## 1. Prepare Proto Files

Example `service.proto`:

```protobuf
syntax = "proto3";
package myapp;

service OrderService {
  // @HasAnyRole({"ADMIN"})
  // Admin-only: creates orders directly.
  rpc CreateOrder(CreateOrderRequest) returns (CreateOrderResponse);

  // Public: lists orders for the current user.
  rpc ListOrders(ListOrdersRequest) returns (ListOrdersResponse);
}

message CreateOrderRequest { string name = 1; }
message CreateOrderResponse { string id = 1; }
message ListOrdersRequest { int32 page = 1; }
message ListOrdersResponse { repeated string ids = 1; }
```

## 2. Create Filter Config

Create `filter.yaml`:

```yaml
annotations:
  - "HasAnyRole"
```

## 3. Run the Tool

```bash
proto-filter --input ./protos --output ./out --config filter.yaml --verbose
```

Expected verbose output:

```
proto-filter: processed 1 files, 5 definitions
proto-filter: removed 1 methods by annotation, 2 orphaned definitions
proto-filter: wrote 1 files to ./out
```

## 4. Verify Output

The output `service.proto` should contain:

```protobuf
syntax = "proto3";
package myapp;

service OrderService {
  // Public: lists orders for the current user.
  rpc ListOrders(ListOrdersRequest) returns (ListOrdersResponse);
}

message ListOrdersRequest { int32 page = 1; }
message ListOrdersResponse { repeated string ids = 1; }
```

Note:
- `CreateOrder` method is removed (had `@HasAnyRole`)
- `CreateOrderRequest` and `CreateOrderResponse` are removed (orphaned)
- `OrderService` is kept (still has `ListOrders`)

## 5. Combine with Name-Based Filtering

Annotation filtering works alongside existing include/exclude rules:

```yaml
include:
  - "myapp.*"
exclude:
  - "myapp.internal.*"
annotations:
  - "HasAnyRole"
  - "Internal"
```

This first applies name-based filtering, then removes annotated methods from the result.
