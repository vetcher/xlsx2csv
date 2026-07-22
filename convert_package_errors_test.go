package xlsx2csv_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/vetcher/xlsx2csv"
)

func TestConvert_InvalidZip_ErrInvalidPackage(t *testing.T) {
	_, err := xlsx2csv.Convert(bytes.NewReader([]byte("not a zip file")))
	if err == nil {
		t.Fatal("expected error")
	}
	var ce *xlsx2csv.CellError
	if !errors.As(err, &ce) {
		t.Fatalf("errors.As CellError: %v", err)
	}
	if ce.Code != xlsx2csv.ErrInvalidPackage {
		t.Fatalf("code %v want ErrInvalidPackage", ce.Code)
	}
}

func TestConvert_ErrorCellTypeE(t *testing.T) {
	const sheet = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <sheetData>
    <row r="1">
      <c r="A1" t="e"><v>#DIV/0!</v></c>
    </row>
  </sheetData>
</worksheet>`
	raw := workbookWithSheet(sheet)
	rows, err := xlsx2csv.Convert(bytes.NewReader(raw), xlsx2csv.WithErrorPlaceholder("#ERR"))
	if err == nil {
		t.Fatal("expected error")
	}
	var ce *xlsx2csv.CellError
	if !errors.As(err, &ce) {
		t.Fatalf("errors.As CellError: %v", err)
	}
	if ce.Code != xlsx2csv.ErrUnsupportedType {
		t.Fatalf("code %v", ce.Code)
	}
	if len(rows) != 1 || rows[0][0] != "#ERR" {
		t.Fatalf("rows %#v", rows)
	}
}
