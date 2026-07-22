package xlsx2csv

import (
	"errors"
	"fmt"
	"io"

	"golang.org/x/text/encoding"

	"github.com/vetcher/xlsx2csv/internal/ooxml"
)

const (
	workbookPath      = "xl/workbook.xml"
	workbookRelsPath  = "xl/_rels/workbook.xml.rels"
	sharedStringsPath = "xl/sharedStrings.xml"
)

// Convert reads an XLSX file and returns the selected worksheet as dense rows.
func Convert(r io.Reader, opts ...Option) ([][]string, error) {
	cfg := apply(opts)
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	pkg, err := ooxml.Open(data)
	if err != nil {
		return nil, &ErrorList{Errs: []error{err}}
	}

	sheets, _, err := loadWorkbookSheets(pkg, cfg.enc)
	if err != nil {
		return nil, err
	}

	sheet, err := selectSheet(sheets, cfg)
	if err != nil {
		return nil, err
	}

	shared, err := loadSharedStrings(pkg, cfg.enc)
	if err != nil {
		return nil, err
	}

	sheetRaw, err := pkg.ReadPart(sheet.Path)
	if err != nil {
		return nil, &ErrorList{Errs: []error{err}}
	}
	sheetUTF8, _, err := decodeXML(sheetRaw, cfg.enc)
	if err != nil {
		return nil, decodeFileError(err)
	}

	rows, cellErrs, err := sheetToRowsWithErrors(sheetUTF8, shared, sheet.Name, cfg)
	if err != nil {
		return nil, &ErrorList{Errs: []error{err}}
	}
	if len(cellErrs) > 0 {
		return rows, &ErrorList{Errs: cellErrs}
	}
	return rows, nil
}

func loadWorkbookSheets(pkg *ooxml.Package, enc encoding.Encoding) ([]ooxml.SheetInfo, string, error) {
	wbRaw, err := pkg.ReadPart(workbookPath)
	if err != nil {
		return nil, "", &ErrorList{Errs: []error{err}}
	}
	wbUTF8, encName, err := decodeXML(wbRaw, enc)
	if err != nil {
		return nil, "", decodeFileError(err)
	}

	relsRaw, err := pkg.ReadPart(workbookRelsPath)
	if err != nil {
		return nil, "", &ErrorList{Errs: []error{err}}
	}
	relsUTF8, _, err := decodeXML(relsRaw, enc)
	if err != nil {
		return nil, "", decodeFileError(err)
	}

	sheets, err := ooxml.ParseWorkbook(wbUTF8, relsUTF8)
	if err != nil {
		return nil, "", &ErrorList{Errs: []error{err}}
	}
	return sheets, encName, nil
}

func selectSheet(sheets []ooxml.SheetInfo, cfg config) (*ooxml.SheetInfo, error) {
	if cfg.useName {
		for i := range sheets {
			if sheets[i].Name == cfg.sheetName {
				return &sheets[i], nil
			}
		}
		return nil, sheetNotFound(cfg.sheetName)
	}
	if cfg.sheetIndex < 0 || cfg.sheetIndex >= len(sheets) {
		return nil, sheetNotFound("")
	}
	return &sheets[cfg.sheetIndex], nil
}

func sheetNotFound(name string) error {
	msg := "sheet not found"
	if name != "" {
		msg = fmt.Sprintf("sheet %q not found", name)
	}
	return &ErrorList{Errs: []error{&CellError{
		Ref:  CellRef{Sheet: name},
		Code: ErrSheetNotFound,
		Msg:  msg,
	}}}
}

func loadSharedStrings(pkg *ooxml.Package, enc encoding.Encoding) ([]string, error) {
	raw, err := pkg.ReadPart(sharedStringsPath)
	if err != nil {
		return nil, nil
	}
	utf8, _, err := decodeXML(raw, enc)
	if err != nil {
		return nil, decodeFileError(err)
	}
	shared, err := ooxml.ParseSharedStrings(utf8)
	if err != nil {
		return nil, &ErrorList{Errs: []error{err}}
	}
	return shared, nil
}

func decodeFileError(cause error) error {
	return &ErrorList{Errs: []error{&CellError{
		Code:  ErrDecode,
		Msg:   "decode XML",
		Cause: cause,
	}}}
}

var errErrorLimitReached = errors.New("xlsx2csv: error limit reached")

func sheetToRowsWithErrors(utf8 []byte, shared []string, sheetName string, cfg config) ([][]string, []error, error) {
	var placed []ooxml.PlacedCell
	var errs []error

	streamErr := ooxml.StreamSheet(utf8, shared, func(row, col int, text string, kind ooxml.CellKind) error {
		switch kind {
		case ooxml.KindFormula, ooxml.KindError, ooxml.KindUnsupported:
			placed = append(placed, ooxml.PlacedCell{Row: row, Col: col, Text: cfg.placeholder})
			errs = append(errs, cellErrorFromKind(sheetName, row, col, kind))
			if cfg.errorLimit != -1 && len(errs) >= cfg.errorLimit {
				return errErrorLimitReached
			}
			return nil
		}
		placed = append(placed, ooxml.PlacedCell{Row: row, Col: col, Text: text})
		return nil
	})
	if streamErr != nil && !errors.Is(streamErr, errErrorLimitReached) {
		return nil, nil, streamErr
	}

	return ooxml.DensifyPlaced(placed), errs, nil
}

func cellErrorFromKind(sheetName string, row, col int, kind ooxml.CellKind) *CellError {
	ref := CellRef{Sheet: sheetName, Row: row, Col: col}
	switch kind {
	case ooxml.KindFormula:
		return &CellError{Ref: ref, Code: ErrFormulaNotSupported, Msg: "formula not supported"}
	case ooxml.KindError:
		return &CellError{Ref: ref, Code: ErrUnsupportedType, Msg: "cell error value"}
	default:
		return &CellError{Ref: ref, Code: ErrUnsupportedType, Msg: "unsupported cell type"}
	}
}
