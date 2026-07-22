package xlsx2csv_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/vetcher/xlsx2csv"
)

func TestCellRef_String_A1(t *testing.T) {
	ref := xlsx2csv.CellRef{Sheet: "Data", Row: 1, Col: 1}
	if got := ref.String(); got != "Data!A1" {
		t.Fatalf("got %q", got)
	}
	ref2 := xlsx2csv.CellRef{Row: 2, Col: 27} // AA2
	if got := ref2.String(); got != "AA2" {
		t.Fatalf("got %q", got)
	}
}

func TestCellError_ErrorsAsAndMessage(t *testing.T) {
	cause := fmt.Errorf("boom")
	err := &xlsx2csv.CellError{
		Ref:   xlsx2csv.CellRef{Sheet: "S", Row: 3, Col: 2},
		Code:  xlsx2csv.ErrFormulaNotSupported,
		Msg:   "formula not supported",
		Cause: cause,
	}
	if !errors.Is(err, cause) {
		t.Fatal("Unwrap/Is failed")
	}
	var ce *xlsx2csv.CellError
	if !errors.As(err, &ce) {
		t.Fatal("As CellError failed")
	}
	if want := "S!B3: formula not supported"; err.Error() != want {
		t.Fatalf("Error()=%q want %q", err.Error(), want)
	}
}

func TestErrorList_UnwrapAll(t *testing.T) {
	e1 := &xlsx2csv.CellError{Ref: xlsx2csv.CellRef{Row: 1, Col: 1}, Code: xlsx2csv.ErrCorrupt, Msg: "bad"}
	e2 := errors.New("file")
	list := &xlsx2csv.ErrorList{Errs: []error{e1, e2}}
	if list.Error() == "" {
		t.Fatal("empty Error()")
	}
	unwrapped := list.Unwrap()
	if len(unwrapped) != 2 {
		t.Fatalf("len=%d", len(unwrapped))
	}
	var ce *xlsx2csv.CellError
	if !errors.As(list, &ce) {
		// errors.As on ErrorList walks Unwrap() []error in Go 1.20+
		t.Fatal("As through ErrorList failed")
	}
}
