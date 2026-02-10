# Quickstart: Cross-File Orphan Detection Tests

## What Changed

No production code changes. This feature adds test coverage for cross-file
orphan detection behavior when annotation filtering is applied.

## What Is Tested

When proto files reference message types from other "common" files, the
filtering pipeline must correctly:

1. **Preserve** shared message types that are still referenced by surviving
   methods after annotation filtering
2. **Remove** shared message types that are no longer referenced by any
   surviving method after annotation filtering

## Example Scenario

**Input files:**

`common.proto` — shared types, no services:
```protobuf
syntax = "proto3";
package crossfile;

message Pagination { int32 page = 1; int32 per_page = 2; }
message Money { string currency = 1; int64 amount = 2; }
message ErrorDetail { string code = 1; string message = 2; }
```

`orders.proto` — service referencing common types:
```protobuf
syntax = "proto3";
package crossfile;
import "common.proto";

service OrderService {
  rpc ListOrders(ListOrdersRequest) returns (ListOrdersResponse);
}

message ListOrdersRequest { Pagination pagination = 1; }
message ListOrdersResponse { repeated Money totals = 1; }
```

`payments.proto` — service with `@Internal` annotation:
```protobuf
syntax = "proto3";
package crossfile;
import "common.proto";

// @Internal
service PaymentService {
  rpc ProcessPayment(ProcessPaymentRequest) returns (ProcessPaymentResponse);
}

message ProcessPaymentRequest { Money amount = 1; }
message ProcessPaymentResponse { ErrorDetail error = 1; }
```

**Config** `filter.yaml`:
```yaml
annotations:
  - "Internal"
```

**Expected behavior:**
- `PaymentService` is removed (has `@Internal`)
- `Money` survives in `common.proto` (still referenced by `OrderService`)
- `Pagination` survives (still referenced by `OrderService`)
- `ErrorDetail` survives in `common.proto` (common files with no services are
  not subject to per-file orphan removal — their types were included by the
  dependency graph because other files reference them)

## Running Tests

```bash
go test -race ./...
```
