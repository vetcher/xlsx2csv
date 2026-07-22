# Benchmarks

## Commands

```bash
go test ./... -count=1
go test -bench=BenchmarkConvert -benchmem -count=5 ./...
go test -bench=BenchmarkConvert_Small -memprofile=mem.pprof -cpuprofile=cpu.pprof -run=^$
go tool pprof -top mem.pprof
go tool pprof -http=:0 mem.pprof
```

## Baseline (2026-07-22, linux/amd64, Intel Xeon, GOMAXPROCS=4)

| Bench | ns/op | B/op | allocs/op |
|-------|-------|------|-----------|
| Convert_Small | ~47k | ~28.7k | 329 |
| Convert_Medium (500 rows) | ~1.40M | ~1.07M | 18310 |

`Convert_Small` uses `fixtureTwoSheets` (two-sheet workbook, one shared-string cell on sheet 1).  
`Convert_Medium` uses `fixtureMediumSheet` (500 rows × one shared-string cell).

## Alloc notes

Profile (`BenchmarkConvert_Small`, `go tool pprof -top mem.pprof`):

- **io.ReadAll** (~26%): full XLSX buffered once per `Convert` — expected for `io.Reader` API.
- **encoding/xml** (`Decoder.Token`, `rawToken`, `NewDecoder`): parsing workbook, rels, shared strings, and sheet — dominant after zip; no duplicate shared-string parse (once per `Convert`).
- **archive/zip** (`NewReader`, directory read): one-time package open per convert.
- **ooxml.ParseWorkbook** / rels + sheets parsers: proportional to workbook XML size; small for fixtures.
- **ooxml.DensifyPlaced** (~0.8%): sparse-to-dense matrix; scales with cell count (Medium bench ~36 allocs/row largely from streaming + densify maps).

No hot-path code changes in this task: pprof did not show obvious fixable waste (extra copies, re-parsing shared strings, or redundant `ReadAll` on the same part). Future wins would likely need API/shape changes (e.g. mmap/streaming zip, row-major densify without per-row maps).

## Regression watch

`TestConvert_AllocsSmoke` asserts allocs/op on the small fixture via `testing.AllocsPerRun(50)`. Ceiling **400** = measured **329** + ~22% margin (documented here). Raise only after intentional perf work and re-profile.
