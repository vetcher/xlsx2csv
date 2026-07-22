# xlsx2csv Design

Go library: convert `.xlsx` workbooks to `[][]string` with typed cell errors and encoding auto-detection (user-overridable).

## Goals

- Read `.xlsx` (OOXML) from `io.Reader` into `[][]string`
- Report errors with sheet + cell coordinates
- Auto-detect character encoding; allow override via `encoding.Encoding`
- Preserve file text literals for cell values (no needless re-format)
- Reject formulas (raw data only)
- Inspect workbook metadata (sheet list/names and related bits)
- Keep allocations low on the hot path; ship benchmarks + docs

## Non-goals (v1)

- `.xls` / BIFF
- Formula evaluation or accepting formula cells as data
- Merged cells (`mergeCells`) — rejected with typed error
- Number-format / locale display rendering (dates stay numeric literals)
- Writing workbooks or emitting `.csv` files
- True streaming of multi-GB sheets without buffering the zip entry (full entry buffer is OK)
- Trimming accidental far-away whitespace cells from the used range (they widen the dense grid)

## Approach

First-party thin OOXML reader. No excelize, no `.xls` dependency.

Pipeline:

1. Buffer reader into bytes (zip needs `ReaderAt`)
2. Open zip; locate workbook, shared strings, target sheet parts
3. Decode part bytes through charset layer (override or detect) to UTF-8
4. Parse XML; build rows with literal cell text
5. Collect typed errors up to configured limit; return partial rows + error list

Dependencies: stdlib (`archive/zip`, `encoding/xml`, …) and `golang.org/x/text` for encodings/detection.

## Public API

Package name: `xlsx2csv`.

```go
func Convert(r io.Reader, opts ...Option) (rows [][]string, err error)
func Inspect(r io.Reader, opts ...Option) (Meta, error)
```

### Options

| Option | Default | Meaning |
|--------|---------|---------|
| Sheet by index or name | first sheet | Which sheet `Convert` reads |
| `Encoding` | `nil` (auto-detect) | `encoding.Encoding` from `golang.org/x/text/encoding` |
| `ErrorLimit` | `1` (fail-fast) | Stop after N errors; `-1` = collect all; `0` treated as `1` |
| `ErrorPlaceholder` | `""` | Value written into matrix for failed cells |

Exact option constructors (`WithSheetIndex`, `WithSheetName`, `WithEncoding`, …) follow idiomatic functional options.

### Return behavior

- Always return whatever rows were parsed when sheet reading started
- `err == nil` only when no errors
- `err != nil` when any error occurred; typically `*ErrorList`
- File-level failures before any grid exists (invalid zip, sheet not found, charset hard-fail on workbook) may return `rows == nil`

### Meta (`Inspect`)

```go
type Meta struct {
    Sheets       []SheetMeta
    EncodingName string // detected or override label when known
}

type SheetMeta struct {
    Name   string
    Index  int    // 0-based
    Path   string // zip part path
    Hidden bool
}
```

`Inspect` opens the package, runs charset on workbook XML, parses sheet list; does not fully load sheet grids.

## Cell conversion

Supported raw types:

| Kind | Result |
|------|--------|
| Shared / inline string | Exact text from file |
| Number | XML literal as written (no float→string round-trip) |
| Bool | `TRUE` / `FALSE` |
| Blank / missing | `""` |
| Date-serialized number | Same as number (literal); format honor out of scope |

Rejected (`CellError`):

- Formula cells (`f` present / formula type)
- Error-type cells (`t="e"`)
- Unknown / unsupported types

Grid shape:

- Sparse cells densified per row; gaps are `""`
- Row width = max column seen on that row (no pad to worksheet dimension)

## Error model

```go
type CellRef struct {
    Sheet string
    Row   int // 1-based
    Col   int // 1-based; String() yields A1-style
}

type ErrCode int // FormulaNotSupported, Decode, UnsupportedType, Corrupt, …

type CellError struct {
    Ref   CellRef
    Code  ErrCode
    Msg   string
    Cause error
}

type ErrorList struct {
    Errs []error // *CellError and file-level errors
}
```

- `Error()` / `Unwrap()` implemented for `errors.Is` / `errors.As`
- Cell errors: write `ErrorPlaceholder`, continue until `ErrorLimit`
- `ErrorLimit == 1` ≈ fail-fast; `-1` ≈ collect all

## Encoding

Applies to OOXML XML parts (old Excel / local tools may emit non-UTF-8, e.g. windows-1251).

Order:

1. If `Encoding` option set → use it
2. Else: BOM → XML `encoding=` declaration → `x/text` detection on sample → fallback UTF-8
3. Decode bytes to UTF-8, then XML-parse
4. Detection/decode failure → file-level error

Shared by `Convert` and `Inspect`.

## Package layout

```
xlsx2csv/
  convert.go           # Convert, Option
  inspect.go           # Inspect, Meta
  errors.go            # CellRef, CellError, ErrorList, ErrCode
  charset.go           # detect + decode
  internal/ooxml/      # zip, workbook, sharedStrings, sheet parse
  convert_test.go
  convert_bench_test.go
  testdata/            # fixtures: utf-8, windows-1251, formula, multi-sheet
docs/
  benchmarks.md        # how to run, numbers, alloc notes
  superpowers/specs/   # this design
```

## Performance

Goal: minimize allocations on hot path (sheet walk, charset decode, string materialization).

Practices:

- Stream sheet XML with a token decoder loop
- Avoid unnecessary `[]byte` ↔ `string` copies; reuse row buffers
- Build shared-string table once; index lookups only
- Prefer slicing decoded buffers where lifetimes allow; `strings.Clone` when a row escapes the parse buffer
- Use `testing.B` and `testing.AllocsPerRun` with budgets on representative fixtures
- Profile with `go test -bench -cpuprofile -memprofile` before treating hot path as done; fix alloc hotspots that profiles show

Artifacts:

- Separate `convert_bench_test.go` (not mixed into correctness-only tests)
- `docs/benchmarks.md`: run commands, current numbers, alloc notes, regression watch list

Correctness fixtures green before micro-optimization.

## Testing

Unit / integration:

- Happy path → expected `[][]string`
- Formula cell → `CellError` + placeholder + limit (`1`, `N`, `-1`)
- Sheet select by name/index; default first
- Charset override + auto-detect fixture
- `Inspect` names / hidden flags
- Bad zip / missing sheet → file-level error, `rows == nil`

Benchmarks: separate file; document in `docs/benchmarks.md`.

## Implementation notes

- Functional options pattern for `Convert` / `Inspect`
- OOXML details live under `internal/ooxml` (callers cannot import)
- Module path follows repo remote at `go mod init` time
- v1 focuses on read path only
- Implementation plan must include a profile → alloc-fix → re-bench loop before release checklist
