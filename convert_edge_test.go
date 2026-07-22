package xlsx2csv_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/vetcher/xlsx2csv"
)

func fixtureMergedCells(t *testing.T) []byte {
	t.Helper()
	const sheet = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <sheetData>
    <row r="1">
      <c r="A1" t="inlineStr"><is><t>a</t></is></c>
      <c r="B1" t="inlineStr"><is><t>b</t></is></c>
    </row>
  </sheetData>
  <mergeCells count="1">
    <mergeCell ref="A1:B1"/>
  </mergeCells>
</worksheet>`
	return workbookWithSheet(sheet)
}

func fixtureSparseWhitespace(t *testing.T) []byte {
	t.Helper()
	// Real-ish: 3x4 table in A1:D3, accidental space typed in Z20.
	const sheet = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <sheetData>
    <row r="1">
      <c r="A1"><v>11</v></c>
      <c r="B1"><v>12</v></c>
      <c r="C1"><v>13</v></c>
      <c r="D1"><v>14</v></c>
    </row>
    <row r="2">
      <c r="A2"><v>21</v></c>
      <c r="B2"><v>22</v></c>
      <c r="C2"><v>23</v></c>
      <c r="D2"><v>24</v></c>
    </row>
    <row r="3">
      <c r="A3"><v>31</v></c>
      <c r="B3"><v>32</v></c>
      <c r="C3"><v>33</v></c>
      <c r="D3"><v>34</v></c>
    </row>
    <row r="20">
      <c r="Z20" t="inlineStr"><is><t> </t></is></c>
    </row>
  </sheetData>
</worksheet>`
	return workbookWithSheet(sheet)
}

func TestConvert_MergedCells_Rejected(t *testing.T) {
	rows, err := xlsx2csv.Convert(bytes.NewReader(fixtureMergedCells(t)))
	if err == nil {
		t.Fatalf("expected error for merged cells, got rows=%#v", rows)
	}
	var ce *xlsx2csv.CellError
	if !errors.As(err, &ce) {
		t.Fatalf("want *CellError, got %T %v", err, err)
	}
	if ce.Code != xlsx2csv.ErrUnsupportedType {
		t.Fatalf("Code=%v want ErrUnsupportedType", ce.Code)
	}
	if rows != nil {
		t.Fatalf("rows=%#v want nil before sheet grid is trusted", rows)
	}
}

func TestConvert_SparseWhitespaceFarFromTable(t *testing.T) {
	rows, err := xlsx2csv.Convert(bytes.NewReader(fixtureSparseWhitespace(t)))
	if err != nil {
		t.Fatal(err)
	}
	// Current densify rule: max row = farthest used cell row; empty middle rows are nil;
	// far whitespace cell widens that row to column Z (26).
	if len(rows) != 20 {
		t.Fatalf("len(rows)=%d want 20 (Z20 pulls grid to row 20)", len(rows))
	}
	wantTop := [][]string{
		{"11", "12", "13", "14"},
		{"21", "22", "23", "24"},
		{"31", "32", "33", "34"},
	}
	for i, want := range wantTop {
		if len(rows[i]) != len(want) {
			t.Fatalf("row %d len=%d want %d (%#v)", i+1, len(rows[i]), len(want), rows[i])
		}
		for j := range want {
			if rows[i][j] != want[j] {
				t.Fatalf("row %d col %d = %q want %q", i+1, j+1, rows[i][j], want[j])
			}
		}
	}
	for i := 3; i < 19; i++ {
		if rows[i] != nil {
			t.Fatalf("row %d = %#v want nil (no cells on that row)", i+1, rows[i])
		}
	}
	if len(rows[19]) != 26 {
		t.Fatalf("row 20 len=%d want 26 (A..Z)", len(rows[19]))
	}
	for i := 0; i < 25; i++ {
		if rows[19][i] != "" {
			t.Fatalf("row 20 col %d = %q want empty", i+1, rows[19][i])
		}
	}
	if rows[19][25] != " " {
		t.Fatalf("Z20=%q want single space", rows[19][25])
	}
}
