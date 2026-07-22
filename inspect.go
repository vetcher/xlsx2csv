package xlsx2csv

import (
	"io"

	"github.com/vetcher/xlsx2csv/internal/ooxml"
)

// SheetMeta describes one worksheet without loading cell data.
type SheetMeta struct {
	Name   string
	Index  int
	Path   string
	Hidden bool
}

// Meta is workbook-level metadata from Inspect.
type Meta struct {
	Sheets       []SheetMeta
	EncodingName string
}

// Inspect reads an XLSX file and returns sheet metadata without loading grids.
func Inspect(r io.Reader, opts ...Option) (Meta, error) {
	cfg := apply(opts)
	data, err := io.ReadAll(r)
	if err != nil {
		return Meta{}, err
	}
	pkg, err := ooxml.Open(data)
	if err != nil {
		return Meta{}, &ErrorList{Errs: []error{err}}
	}

	sheets, encName, err := loadWorkbookSheets(pkg, cfg.enc)
	if err != nil {
		return Meta{}, err
	}

	meta := Meta{EncodingName: encName}
	meta.Sheets = make([]SheetMeta, len(sheets))
	for i, sh := range sheets {
		meta.Sheets[i] = SheetMeta{
			Name:   sh.Name,
			Index:  sh.Index,
			Path:   sh.Path,
			Hidden: sh.Hidden,
		}
	}
	return meta, nil
}
