package xlsx2csv_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/vetcher/xlsx2csv"
)

func TestConvert_FirstSheetDefault(t *testing.T) {
	r := bytes.NewReader(fixtureTwoSheets(t))
	rows, err := xlsx2csv.Convert(r)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rows, [][]string{{"a"}}) {
		t.Fatalf("%#v", rows)
	}
}

func TestConvert_SheetByNameAndIndex(t *testing.T) {
	raw := fixtureTwoSheets(t)
	rows, err := xlsx2csv.Convert(bytes.NewReader(raw), xlsx2csv.WithSheetName("two"))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rows, [][]string{{"b"}}) {
		t.Fatalf("%#v", rows)
	}

	rows, err = xlsx2csv.Convert(bytes.NewReader(raw), xlsx2csv.WithSheetIndex(1))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rows, [][]string{{"b"}}) {
		t.Fatalf("%#v", rows)
	}
}

func TestConvert_InvalidZip(t *testing.T) {
	rows, err := xlsx2csv.Convert(bytes.NewReader([]byte("not-a-zip")))
	if err == nil || rows != nil {
		t.Fatalf("rows=%v err=%v", rows, err)
	}
}
