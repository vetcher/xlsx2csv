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
