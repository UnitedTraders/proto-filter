# Data Model: Cross-File Orphan Detection Tests

**Date**: 2026-02-09
**Branch**: `005-cross-file-orphan-tests`

## No New Entities

This feature adds tests only — no new data types, configuration fields,
or production code changes.

## Test Fixture Structure

### Cross-File Test Directory

```
testdata/crossfile/
├── common.proto          # Shared message types (no services)
├── orders.proto          # Service file referencing common types
└── payments.proto        # Second service file referencing common types
```

### common.proto

Contains only shared message and enum types, no services:
- `Pagination` — shared pagination type used by multiple services
- `Money` — shared monetary type
- `ErrorDetail` — shared error type referenced only by one service
  (used to test orphan removal when that service is removed)

### orders.proto

Contains `OrderService` with methods referencing types from both
`common.proto` and locally-defined messages:
- Methods with and without annotations
- Request/response types defined locally
- Some request/response types using `Pagination` or `Money` from common

### payments.proto

Contains `PaymentService` with an `@Internal` annotation at the service
level, referencing types from `common.proto`:
- All methods reference common types like `Money` and `ErrorDetail`
- When service-level filtering removes `PaymentService`, `ErrorDetail`
  (referenced only by this service) should become orphaned, while `Money`
  (also referenced by OrderService) should survive

## Processing Pipeline (unchanged)

No changes to the processing pipeline. Tests verify existing behavior.
