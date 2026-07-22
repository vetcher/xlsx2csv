package ooxml_test

import (
	"testing"

	"github.com/vetcher/xlsx2csv/internal/ooxml"
)

func TestOpenAndParseWorkbookSheets(t *testing.T) {
	data := mustBuildMinimalWorkbook(t)
	pkg, err := ooxml.Open(data)
	if err != nil {
		t.Fatal(err)
	}
	wb, err := pkg.ReadPart("xl/workbook.xml")
	if err != nil {
		t.Fatal(err)
	}
	rels, err := pkg.ReadPart("xl/_rels/workbook.xml.rels")
	if err != nil {
		t.Fatal(err)
	}
	sheets, err := ooxml.ParseWorkbook(wb, rels)
	if err != nil {
		t.Fatal(err)
	}
	if len(sheets) != 2 || sheets[0].Name != "one" || sheets[1].Name != "two" {
		t.Fatalf("%#v", sheets)
	}
	if sheets[0].Path != "xl/worksheets/sheet1.xml" {
		t.Fatalf("path %s", sheets[0].Path)
	}
	if !sheets[1].Hidden {
		t.Fatalf("expected sheet two hidden, %#v", sheets[1])
	}
}
