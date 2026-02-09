# Quickstart: Service-Level Annotation Filtering

## What Changed

Proto-filter now checks annotations on service declarations in addition
to individual RPC methods. If a service's comment contains an annotation
that matches the filter config, the entire service and all its methods
are removed from the output.

## Usage

No changes to command usage. Use the same `annotations` config key:

```bash
proto-filter --input ./protos --output ./out --config filter.yaml
```

## Example

**Input** `admin.proto`:
```protobuf
syntax = "proto3";
package myapp;

// @Internal
// Administrative operations for system management.
service AdminService {
  rpc ResetCache(ResetCacheRequest) returns (ResetCacheResponse);
  rpc GetMetrics(MetricsRequest) returns (MetricsResponse);
}

// Public-facing order operations.
service OrderService {
  rpc ListOrders(ListOrdersRequest) returns (ListOrdersResponse);
}

message ResetCacheRequest {}
message ResetCacheResponse {}
message MetricsRequest {}
message MetricsResponse {}
message ListOrdersRequest { int32 page = 1; }
message ListOrdersResponse { repeated string ids = 1; }
```

**Config** `filter.yaml`:
```yaml
annotations:
  - "Internal"
```

**Output** `admin.proto`:
```protobuf
syntax = "proto3";
package myapp;

// Public-facing order operations.
service OrderService {
  rpc ListOrders(ListOrdersRequest) returns (ListOrdersResponse);
}

message ListOrdersRequest { int32 page = 1; }
message ListOrdersResponse { repeated string ids = 1; }
```

Note:
- `AdminService` is removed entirely (had `@Internal` in its comment)
- `ResetCacheRequest`, `ResetCacheResponse`, `MetricsRequest`, `MetricsResponse` are removed (orphaned)
- `OrderService` is kept unchanged (no matching annotation)

## Combining with Method-Level Filtering

Service-level and method-level annotation filtering work together:

```yaml
annotations:
  - "Internal"
  - "HasAnyRole"
```

- Services annotated with `@Internal` are removed entirely
- Individual methods annotated with `@HasAnyRole` are removed from remaining services

## Validation

Run the test suite to verify:

```bash
go test -race ./...
```
