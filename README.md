# xlsx2csv

Go library that converts `.xlsx` workbooks to dense `[][]string` grids without external Excel dependencies. Legacy `.xls` is not supported.

## Install

```bash
go get github.com/vetcher/xlsx2csv
```

## Convert

`Convert` reads an XLSX from an `io.Reader`, selects a worksheet (first sheet by default), and returns all populated cells as rows of strings.

```go
f, err := os.Open("book.xlsx")
if err != nil {
    return err
}
defer f.Close()

rows, err := xlsx2csv.Convert(f)
if err != nil {
    return err
}
// rows[0][0] is cell A1 on the first sheet
```

See also the runnable [`ExampleConvert`](example_test.go) in the test suite.

## Inspect

`Inspect` opens the same package format but only returns workbook metadata (sheet names, paths, visibility, detected encoding) without parsing cell grids.

```go
meta, err := xlsx2csv.Inspect(bytes.NewReader(data))
if err != nil {
    return err
}
for _, sh := range meta.Sheets {
    fmt.Println(sh.Index, sh.Name, sh.Hidden)
}
```

## Options

| Option | Purpose |
|--------|---------|
| `WithSheetIndex(i int)` | Select worksheet by zero-based index |
| `WithSheetName(name string)` | Select worksheet by name |
| `WithEncoding(enc encoding.Encoding)` | Override XML encoding (default: UTF-8 or declared `encoding` in XML) |
| `WithErrorLimit(n int)` | Max cell-level errors before stopping (`-1` = unlimited; `0` treated as `1`) |
| `WithErrorPlaceholder(s string)` | Text placed in cells that error (e.g. formulas) |

```go
rows, err := xlsx2csv.Convert(r,
    xlsx2csv.WithSheetName("Data"),
    xlsx2csv.WithErrorLimit(10),
    xlsx2csv.WithErrorPlaceholder("#ERR"),
)
```

## Errors

File-level problems (invalid zip, corrupt workbook XML) return `(nil, err)` with no rows. Cell-level issues (unsupported formulas, error cells) may still return partial `rows` together with an `*ErrorList` error.

```go
rows, err := xlsx2csv.Convert(r)
if err != nil {
    var list *xlsx2csv.ErrorList
    if errors.As(err, &list) {
        for _, e := range list.Errs {
            log.Println(e)
        }
    }
    var cell *xlsx2csv.CellError
    if errors.As(err, &cell) {
        log.Println(cell.Ref, cell.Code)
    }
}
```

Typed codes include `ErrFormulaNotSupported`, `ErrDecode`, `ErrUnsupportedType`, `ErrCorrupt`, `ErrSheetNotFound`, and `ErrInvalidPackage`.

## Benchmarks

See [docs/benchmarks.md](docs/benchmarks.md) for benchmark commands and allocation notes.

## Design

See [docs/superpowers/specs/2026-07-22-xlsx2csv-design.md](docs/superpowers/specs/2026-07-22-xlsx2csv-design.md) for goals, scope, and trade-offs.
