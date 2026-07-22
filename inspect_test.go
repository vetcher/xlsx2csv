package xlsx2csv_test

import (
	"bytes"
	"testing"

	"github.com/vetcher/xlsx2csv"
	"golang.org/x/text/encoding/charmap"
)

func TestInspect_SheetList(t *testing.T) {
	meta, err := xlsx2csv.Inspect(bytes.NewReader(fixtureTwoSheets(t)))
	if err != nil {
		t.Fatal(err)
	}
	if len(meta.Sheets) != 2 {
		t.Fatalf("%#v", meta)
	}
	if meta.Sheets[0].Name != "one" || meta.Sheets[1].Index != 1 {
		t.Fatalf("%#v", meta.Sheets)
	}
}

func TestConvert_EncodingOverrideWindows1251(t *testing.T) {
	raw := fixtureSheetWindows1251(t)
	rows, err := xlsx2csv.Convert(bytes.NewReader(raw), xlsx2csv.WithEncoding(charmap.Windows1251))
	if err != nil {
		t.Fatal(err)
	}
	if rows[0][0] != "Привет" {
		t.Fatalf("%#v", rows)
	}
}

func TestInspect_ReportsEncodingName(t *testing.T) {
	meta, err := xlsx2csv.Inspect(bytes.NewReader(fixtureSheetWindows1251(t)), xlsx2csv.WithEncoding(charmap.Windows1251))
	if err != nil {
		t.Fatal(err)
	}
	if meta.EncodingName == "" {
		t.Fatal("empty encoding name")
	}
}
