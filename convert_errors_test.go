package xlsx2csv_test

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/vetcher/xlsx2csv"
)

func workbookWithSheet(sheetXML string) []byte {
	const workbook = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"
 xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <sheets>
    <sheet name="one" sheetId="1" r:id="rId1"/>
  </sheets>
</workbook>`
	const rels = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>
</Relationships>`
	return buildXLSX(map[string][]byte{
		"[Content_Types].xml":        []byte(`<?xml version="1.0"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"></Types>`),
		"xl/workbook.xml":            []byte(workbook),
		"xl/_rels/workbook.xml.rels": []byte(rels),
		"xl/worksheets/sheet1.xml":   []byte(sheetXML),
	})
}

func fixtureWithFormulaAndValue(t *testing.T) []byte {
	t.Helper()
	const sheet = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <sheetData>
    <row r="1">
      <c r="A1"><v>1</v></c>
      <c r="B1"><f>SUM(A1:C1)</f><v>6</v></c>
      <c r="C1"><v>3</v></c>
    </row>
  </sheetData>
</worksheet>`
	return workbookWithSheet(sheet)
}

func fixtureTwoFormulas(t *testing.T) []byte {
	t.Helper()
	const sheet = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <sheetData>
    <row r="1">
      <c r="A1"><f>1+1</f><v>2</v></c>
      <c r="B1"><f>2+2</f><v>4</v></c>
    </row>
  </sheetData>
</worksheet>`
	return workbookWithSheet(sheet)
}

func fixtureThreeFormulas(t *testing.T) []byte {
	t.Helper()
	const sheet = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <sheetData>
    <row r="1">
      <c r="A1"><f>1</f><v>1</v></c>
      <c r="B1"><f>2</f><v>2</v></c>
      <c r="C1"><f>3</f><v>3</v></c>
    </row>
  </sheetData>
</worksheet>`
	return workbookWithSheet(sheet)
}

func TestConvert_FormulaCell_FailFast(t *testing.T) {
	raw := fixtureWithFormulaAndValue(t)
	rows, err := xlsx2csv.Convert(bytes.NewReader(raw))
	if err == nil {
		t.Fatal("expected err")
	}
	var list *xlsx2csv.ErrorList
	if !errors.As(err, &list) {
		t.Fatalf("%T", err)
	}
	var ce *xlsx2csv.CellError
	if !errors.As(err, &ce) {
		t.Fatal("no CellError")
	}
	if ce.Code != xlsx2csv.ErrFormulaNotSupported {
		t.Fatalf("code %v", ce.Code)
	}
	if ce.Ref.Col != 2 {
		t.Fatalf("ref %#v", ce.Ref)
	}
	if len(rows) == 0 || rows[0][0] != "1" {
		t.Fatalf("%#v", rows)
	}
	if rows[0][1] != "" {
		t.Fatalf("placeholder %#v", rows[0])
	}
}

func TestConvert_CollectAllAndPlaceholder(t *testing.T) {
	raw := fixtureTwoFormulas(t)
	rows, err := xlsx2csv.Convert(bytes.NewReader(raw),
		xlsx2csv.WithErrorLimit(-1),
		xlsx2csv.WithErrorPlaceholder("#ERR"),
	)
	if err == nil {
		t.Fatal("expected err")
	}
	var list *xlsx2csv.ErrorList
	errors.As(err, &list)
	if len(list.Errs) < 2 {
		t.Fatalf("errs=%d", len(list.Errs))
	}
	joined := fmt.Sprint(rows)
	if !strings.Contains(joined, "#ERR") {
		t.Fatalf("%v", rows)
	}
}

func TestConvert_ErrorLimitN(t *testing.T) {
	raw := fixtureThreeFormulas(t)
	_, err := xlsx2csv.Convert(bytes.NewReader(raw), xlsx2csv.WithErrorLimit(2))
	var list *xlsx2csv.ErrorList
	errors.As(err, &list)
	if len(list.Errs) != 2 {
		t.Fatalf("got %d", len(list.Errs))
	}
}
