package xlsx2csv_test

import (
	"bytes"
	"errors"
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

func TestConvert_SheetNotFound(t *testing.T) {
	raw := fixtureTwoSheets(t)

	t.Run("ByName", func(t *testing.T) {
		rows, err := xlsx2csv.Convert(bytes.NewReader(raw), xlsx2csv.WithSheetName("missing"))
		assertConvertSheetNotFound(t, rows, err)
	})

	t.Run("ByIndex", func(t *testing.T) {
		rows, err := xlsx2csv.Convert(bytes.NewReader(raw), xlsx2csv.WithSheetIndex(99))
		assertConvertSheetNotFound(t, rows, err)
	})
}

func assertConvertSheetNotFound(t *testing.T, rows [][]string, err error) {
	t.Helper()
	if rows != nil {
		t.Fatalf("rows=%v want nil", rows)
	}
	if err == nil {
		t.Fatal("err is nil")
	}
	var list *xlsx2csv.ErrorList
	if !errors.As(err, &list) {
		t.Fatalf("errors.As ErrorList: %v", err)
	}
	var ce *xlsx2csv.CellError
	if !errors.As(err, &ce) {
		t.Fatalf("errors.As CellError: %v", err)
	}
	if ce.Code != xlsx2csv.ErrSheetNotFound {
		t.Fatalf("Code=%v want ErrSheetNotFound", ce.Code)
	}
}
