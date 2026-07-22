package ooxml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
)

// CellKind classifies a worksheet cell value.
type CellKind int

const (
	KindString CellKind = iota
	KindNumber
	KindBool
	KindBlank
	KindFormula
	KindError
	KindUnsupported
)

// CellValue is one parsed cell with coordinates and classification.
type CellValue struct {
	Row  int
	Col  int
	Text string
	Kind CellKind
	Err  error
}

// SheetCellFault records a cell that could not be converted to a plain string.
type SheetCellFault struct {
	Row  int
	Col  int
	Kind CellKind
}

// CellHandler is called for each cell while streaming a sheet.
type CellHandler func(row, col int, text string, kind CellKind) error

// StreamSheet walks sheet XML and invokes fn for each cell.
func StreamSheet(utf8 []byte, shared []string, fn CellHandler) error {
	dec := xml.NewDecoder(bytes.NewReader(utf8))
	var inSheetData bool
	var inCell bool
	var cellRef, cellType string
	var hasFormula bool
	var vText, inlineText strings.Builder
	var inV, inInline, inInlineT bool

	flushCell := func() error {
		if !inCell {
			return nil
		}
		row, col, err := parseCellRef(cellRef)
		if err != nil {
			return err
		}
		if hasFormula {
			return fn(row, col, "", KindFormula)
		}
		text, kind, err := cellText(cellType, vText.String(), inlineText.String(), shared)
		if err != nil {
			return err
		}
		return fn(row, col, text, kind)
	}

	for {
		tok, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("ooxml: parse sheet: %w", err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			local := t.Name.Local
			nsOK := t.Name.Space == "" || t.Name.Space == spreadsheetMLNS
			if !nsOK {
				continue
			}
			switch local {
			case "sheetData":
				inSheetData = true
			case "c":
				if inSheetData {
					if inCell {
						if err := flushCell(); err != nil {
							return err
						}
					}
					inCell = true
					cellRef, cellType = "", ""
					hasFormula = false
					vText.Reset()
					inlineText.Reset()
					for _, a := range t.Attr {
						if a.Name.Space != "" {
							continue
						}
						switch a.Name.Local {
						case "r":
							cellRef = a.Value
						case "t":
							cellType = a.Value
						}
					}
				}
			case "f":
				if inCell {
					hasFormula = true
				}
			case "v":
				if inCell {
					inV = true
				}
			case "is":
				if inCell {
					inInline = true
				}
			case "t":
				if inCell && inInline {
					inInlineT = true
				}
			}
		case xml.EndElement:
			local := t.Name.Local
			nsOK := t.Name.Space == "" || t.Name.Space == spreadsheetMLNS
			if !nsOK {
				continue
			}
			switch local {
			case "sheetData":
				inSheetData = false
			case "c":
				if inSheetData && inCell {
					if err := flushCell(); err != nil {
						return err
					}
					inCell = false
				}
			case "v":
				inV = false
			case "is":
				inInline = false
			case "t":
				inInlineT = false
			}
		case xml.CharData:
			if inCell && inV {
				vText.Write(t)
			}
			if inCell && inInlineT {
				inlineText.Write(t)
			}
		}
	}
	if inCell {
		if err := flushCell(); err != nil {
			return err
		}
	}
	return nil
}

func cellText(cellType, v, inline string, shared []string) (text string, kind CellKind, err error) {
	switch cellType {
	case "s":
		idx, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return "", KindString, fmt.Errorf("ooxml: shared string index %q: %w", v, err)
		}
		if idx < 0 || idx >= len(shared) {
			return "", KindString, fmt.Errorf("ooxml: shared string index %d out of range", idx)
		}
		return shared[idx], KindString, nil
	case "b":
		if strings.TrimSpace(v) == "1" {
			return "TRUE", KindBool, nil
		}
		return "FALSE", KindBool, nil
	case "inlineStr":
		return inline, KindString, nil
	case "str":
		return "", KindFormula, nil
	case "e":
		return "", KindError, nil
	case "n":
		if v == "" {
			return "", KindBlank, nil
		}
		return v, KindNumber, nil
	case "":
		if v == "" {
			return "", KindBlank, nil
		}
		return v, KindNumber, nil
	default:
		return "", KindUnsupported, nil
	}
}

// PlacedCell is a cell value at 1-based row and column coordinates.
type PlacedCell struct {
	Row, Col int
	Text     string
}

// DensifyPlaced builds a dense row matrix from sparse cell placements.
func DensifyPlaced(placedCells []PlacedCell) [][]string {
	if len(placedCells) == 0 {
		return nil
	}

	maxRow := 0
	rowCols := make(map[int]map[int]string)
	for _, c := range placedCells {
		if c.Row > maxRow {
			maxRow = c.Row
		}
		if rowCols[c.Row] == nil {
			rowCols[c.Row] = make(map[int]string)
		}
		rowCols[c.Row][c.Col] = c.Text
	}

	rows := make([][]string, maxRow)
	for r := 1; r <= maxRow; r++ {
		cols := rowCols[r]
		maxCol := 0
		for col := range cols {
			if col > maxCol {
				maxCol = col
			}
		}
		if maxCol == 0 {
			rows[r-1] = nil
			continue
		}
		row := make([]string, maxCol)
		for col := 1; col <= maxCol; col++ {
			if v, ok := cols[col]; ok {
				row[col-1] = v
			}
		}
		rows[r-1] = row
	}
	return rows
}

// SheetToRows parses sheet XML into dense row slices and cell faults.
func SheetToRows(utf8 []byte, shared []string) ([][]string, []SheetCellFault, error) {
	var placedCells []PlacedCell
	var faults []SheetCellFault

	err := StreamSheet(utf8, shared, func(row, col int, text string, kind CellKind) error {
		switch kind {
		case KindFormula, KindError, KindUnsupported:
			faults = append(faults, SheetCellFault{Row: row, Col: col, Kind: kind})
			placedCells = append(placedCells, PlacedCell{Row: row, Col: col, Text: ""})
			return nil
		}
		placedCells = append(placedCells, PlacedCell{Row: row, Col: col, Text: text})
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	rows := DensifyPlaced(placedCells)
	return rows, faults, nil
}

// parseCellRef parses an A1-style reference into 1-based row and column.
func parseCellRef(ref string) (row, col int, err error) {
	if ref == "" {
		return 0, 0, fmt.Errorf("ooxml: empty cell reference")
	}
	i := 0
	for i < len(ref) && isColLetter(ref[i]) {
		i++
	}
	if i == 0 || i == len(ref) {
		return 0, 0, fmt.Errorf("ooxml: invalid cell reference %q", ref)
	}
	col = colLettersToIndex(ref[:i])
	row, err = strconv.Atoi(ref[i:])
	if err != nil || row < 1 {
		return 0, 0, fmt.Errorf("ooxml: invalid cell reference %q", ref)
	}
	return row, col, nil
}

func isColLetter(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}

func colLettersToIndex(s string) int {
	col := 0
	for _, r := range s {
		col = col*26 + int(unicode.ToUpper(r)-'A') + 1
	}
	return col
}
