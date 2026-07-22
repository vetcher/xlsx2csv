package xlsx2csv

import (
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

	wbRaw, err := pkg.ReadPart(workbookPath)
	if err != nil {
		return nil, &ErrorList{Errs: []error{err}}
	}
	wbUTF8, _, err := decodeXML(wbRaw, cfg.enc)
	if err != nil {
		return nil, decodeFileError(err)
	}

	relsRaw, err := pkg.ReadPart(workbookRelsPath)
	if err != nil {
		return nil, &ErrorList{Errs: []error{err}}
	}
	relsUTF8, _, err := decodeXML(relsRaw, cfg.enc)
	if err != nil {
		return nil, decodeFileError(err)
	}

	sheets, err := ooxml.ParseWorkbook(wbUTF8, relsUTF8)
	if err != nil {
		return nil, &ErrorList{Errs: []error{err}}
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

	rows, _, err := ooxml.SheetToRows(sheetUTF8, shared)
	if err != nil {
		return nil, &ErrorList{Errs: []error{err}}
	}
	return rows, nil
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
