# Quickstart: Comment Style Conversion

## What Changed

Proto-filter now automatically converts all block comments (`/* ... */`) to single-line comments (`// ...`) in output files. This is applied unconditionally to all processed files.

## Usage

No changes to command usage. Run proto-filter as before:

```bash
# Pass-through mode - comments are still converted
proto-filter --input ./protos --output ./out

# With filtering - comments are converted in addition to filtering
proto-filter --input ./protos --output ./out --config filter.yaml
```

## Before / After

**Before** (input file):
```proto
/**
 * Returns updates of price for all symbols from symbolSpec
 * @StartsWithSnapshot
 * @SupportWindow
 */
rpc GetPriceUpdates(PriceRequest) returns (stream PriceUpdate);

/* Order status enum */
enum OrderStatus {
  PENDING = 0;
  FILLED = 1;
}
```

**After** (output file):
```proto
// Returns updates of price for all symbols from symbolSpec
// @StartsWithSnapshot
// @SupportWindow
rpc GetPriceUpdates(PriceRequest) returns (stream PriceUpdate);

// Order status enum
enum OrderStatus {
  PENDING = 0;
  FILLED = 1;
}
```

## Validation

Run the test suite to verify:

```bash
go test -race ./...
```
