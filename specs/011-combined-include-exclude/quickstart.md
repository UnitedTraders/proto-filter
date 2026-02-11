# Quickstart: Combined Include and Exclude Annotation Filters

## What This Feature Does

You can now use `annotations.include` and `annotations.exclude` together in the same config. The system first applies inclusion (keeping only matching services/methods), then applies exclusion (removing matching services, methods, and message fields).

## Before This Feature

```yaml
# config.yaml — ERROR: mutually exclusive!
annotations:
  include:
    - "PublicApi"
  exclude:
    - "Deprecated"
```

```
proto-filter: error: annotations.include and annotations.exclude are mutually exclusive
```

## After This Feature

```yaml
# config.yaml — Now works!
annotations:
  include:
    - "PublicApi"
  exclude:
    - "Deprecated"
```

```protobuf
// Input: service.proto
// [PublicApi]
service OrderService {
  rpc ListOrders(Empty) returns (ListOrdersResponse);
  rpc GetOrder(GetOrderRequest) returns (GetOrderResponse);
}

message ListOrdersResponse {
  repeated Order orders = 1;
  // [Deprecated]
  uint64 index = 2;
}

message GetOrderRequest {
  uint64 id = 1;
}

message GetOrderResponse {
  uint64 id = 1;
  string name = 2;
}
```

```protobuf
// Output: service is included (PublicApi), deprecated field removed
service OrderService {
  rpc ListOrders(Empty) returns (ListOrdersResponse);
  rpc GetOrder(GetOrderRequest) returns (GetOrderResponse);
}

message ListOrdersResponse {
  repeated Order orders = 1;
}

message GetOrderRequest {
  uint64 id = 1;
}

message GetOrderResponse {
  uint64 id = 1;
  string name = 2;
}
```

## Field-Level Filtering (Also Works Standalone)

Field-level annotation filtering also works with exclude-only configs:

```yaml
# config.yaml
annotations:
  exclude:
    - "Deprecated"
```

```protobuf
// Input
message UserProfile {
  string name = 1;
  // @Deprecated
  string legacy_email = 2;
  string email = 3;
}
```

```protobuf
// Output: deprecated field removed
message UserProfile {
  string name = 1;
  string email = 3;
}
```

## Processing Order

1. **Include pass**: Keep only services/methods with matching annotations
2. **Exclude pass**: Remove services, methods, and fields with matching annotations
3. **Cleanup**: Remove empty services, remove orphaned type definitions
