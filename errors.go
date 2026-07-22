package xlsx2csv

import (
	"fmt"
	"strings"
)

// CellRef identifies a cell by sheet name and 1-based row/column.
type CellRef struct {
	Sheet string
	Row   int
	Col   int
}

// String returns an A1-style reference, prefixed with "Sheet!" when Sheet is set.
func (c CellRef) String() string {
	addr := colLetters(c.Col) + fmt.Sprintf("%d", c.Row)
	if c.Sheet != "" {
		return c.Sheet + "!" + addr
	}
	return addr
}

func colLetters(col int) string {
	if col <= 0 {
		return ""
	}
	var b []byte
	for col > 0 {
		col--
		b = append([]byte{byte('A' + col%26)}, b...)
		col /= 26
	}
	return string(b)
}

// ErrCode classifies a cell or conversion error.
type ErrCode int

const (
	// ErrFormulaNotSupported is returned for formula cells.
	ErrFormulaNotSupported ErrCode = iota
	// ErrDecode indicates XML decoding failed.
	ErrDecode
	// ErrUnsupportedType is returned for unsupported cell types or error values.
	ErrUnsupportedType
	// ErrCorrupt indicates malformed workbook or sheet data.
	ErrCorrupt
	// ErrSheetNotFound means the requested sheet does not exist.
	ErrSheetNotFound
	// ErrInvalidPackage means the input is not a valid XLSX zip package.
	ErrInvalidPackage
)

// CellError is a typed error tied to a cell reference.
type CellError struct {
	Ref   CellRef
	Code  ErrCode
	Msg   string
	Cause error
}

func (e *CellError) Error() string {
	return fmt.Sprintf("%s: %s", e.Ref.String(), e.Msg)
}

func (e *CellError) Unwrap() error {
	return e.Cause
}

// ErrorList aggregates multiple errors.
type ErrorList struct {
	Errs []error
}

func (e *ErrorList) Error() string {
	parts := make([]string, len(e.Errs))
	for i, err := range e.Errs {
		if err != nil {
			parts[i] = err.Error()
		}
	}
	return strings.Join(parts, "; ")
}

func (e *ErrorList) Unwrap() []error {
	return e.Errs
}
