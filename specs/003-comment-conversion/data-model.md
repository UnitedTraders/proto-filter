# Data Model: Comment Style Conversion

## Entities

### Comment (existing, from emicklei/proto)

Represents a comment attached to a proto element.

| Field      | Type     | Description                                      |
|------------|----------|--------------------------------------------------|
| Lines      | []string | Comment text lines, stripped of syntax prefixes   |
| Cstyle     | bool     | `true` for `/* */` style, `false` for `//` style |
| ExtraSlash | bool     | `true` if comment starts with `///`               |

**Conversion operation**: Set `Cstyle = false` to switch from block to single-line output. Clean leading `*` prefixes from `Lines` entries.

### Comment Attachment Points (existing)

Comments are attached to AST elements at two positions:

| Element    | Leading Comment | Inline Comment |
|------------|----------------|----------------|
| Service    | `.Comment`     | -              |
| Message    | `.Comment`     | -              |
| Enum       | `.Comment`     | -              |
| RPC        | `.Comment`     | `.InlineComment` |
| NormalField| `.Comment`     | `.InlineComment` |
| MapField   | `.Comment`     | `.InlineComment` |
| OneOfField | `.Comment`     | `.InlineComment` |
| EnumField  | `.Comment`     | `.InlineComment` |

## Pipeline Flow

```
Parse → Filter → Annotate → ConvertBlockComments → Write
                                    ↑
                          Set Cstyle=false on
                          all Comment structs
                          with Cstyle=true
```

No new entities are introduced. The feature modifies existing `Comment` structs in-place.
