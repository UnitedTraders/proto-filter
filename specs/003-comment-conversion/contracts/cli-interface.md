# CLI Interface: Comment Style Conversion

## Behavior Change

Comment conversion is automatic and unconditional. No new CLI flags are introduced.

### Existing Command (unchanged)

```bash
proto-filter --input ./protos --output ./out [--config filter.yaml] [--verbose]
```

### New Behavior

All C-style block comments (`/* ... */`, `/** ... */`) in processed proto files are automatically converted to single-line `//` comments in the output. This applies:

- With or without `--config`
- In pass-through mode (no config)
- In filtered mode (with include/exclude/annotations)

### Verbose Output

When `--verbose` is enabled, the existing summary output is unchanged. No additional verbose output is added for comment conversion since it is a formatting concern, not a filtering action.

### Examples

**Input** (`service.proto`):
```proto
/**
 * Returns updates of price for all symbols
 * @StartsWithSnapshot
 * @SupportWindow
 */
rpc GetPriceUpdates(PriceRequest) returns (stream PriceUpdate);
```

**Output** (`service.proto`):
```proto
// Returns updates of price for all symbols
// @StartsWithSnapshot
// @SupportWindow
rpc GetPriceUpdates(PriceRequest) returns (stream PriceUpdate);
```
