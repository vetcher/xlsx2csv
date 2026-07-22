package ooxml_test

import (
	"reflect"
	"testing"

	"github.com/vetcher/xlsx2csv/internal/ooxml"
)

func TestSheetToRows_LiteralsAndShared(t *testing.T) {
	sharedXML := []byte(`<?xml version="1.0"?>
<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" count="1" uniqueCount="1">
  <si><t>hello</t></si>
</sst>`)
	sheetXML := []byte(`<?xml version="1.0"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <sheetData>
    <row r="1">
      <c r="A1" t="s"><v>0</v></c>
      <c r="B1"><v>1.50</v></c>
      <c r="D1" t="b"><v>1</v></c>
    </row>
  </sheetData>
</worksheet>`)
	shared, err := ooxml.ParseSharedStrings(sharedXML)
	if err != nil {
		t.Fatal(err)
	}
	rows, faults, err := ooxml.SheetToRows(sheetXML, shared)
	if err != nil {
		t.Fatal(err)
	}
	if len(faults) != 0 {
		t.Fatalf("faults %#v", faults)
	}
	want := [][]string{{"hello", "1.50", "", "TRUE"}}
	if !reflect.DeepEqual(rows, want) {
		t.Fatalf("got %#v", rows)
	}
}

func TestSheetToRows_ExplicitNumericType(t *testing.T) {
	sheetXML := []byte(`<?xml version="1.0"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <sheetData>
    <row r="1">
      <c r="A1" t="n"><v>1.50</v></c>
    </row>
  </sheetData>
</worksheet>`)
	rows, faults, err := ooxml.SheetToRows(sheetXML, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(faults) != 0 {
		t.Fatalf("faults %#v", faults)
	}
	want := [][]string{{"1.50"}}
	if !reflect.DeepEqual(rows, want) {
		t.Fatalf("got %#v want %#v", rows, want)
	}
}

func TestSheetToRows_FormulaFault(t *testing.T) {
	sheetXML := []byte(`<?xml version="1.0"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <sheetData>
    <row r="1">
      <c r="B2"><f>SUM(A1:A9)</f><v>42</v></c>
    </row>
  </sheetData>
</worksheet>`)
	rows, faults, err := ooxml.SheetToRows(sheetXML, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("rows: got %#v", rows)
	}
	wantRows := [][]string{nil, {"", ""}}
	if !reflect.DeepEqual(rows, wantRows) {
		t.Fatalf("got %#v, want %#v", rows, wantRows)
	}
	if len(faults) != 1 {
		t.Fatalf("faults: got %d, want 1: %#v", len(faults), faults)
	}
	wantFault := ooxml.SheetCellFault{Row: 2, Col: 2, Kind: ooxml.KindFormula}
	if faults[0] != wantFault {
		t.Fatalf("fault: got %#v, want %#v", faults[0], wantFault)
	}
}
