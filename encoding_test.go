package xlsx2csv_test

import (
	"testing"

	"golang.org/x/text/encoding/charmap"
)

func fixtureSheetWindows1251(t *testing.T) []byte {
	t.Helper()
	const workbook = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"
 xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <sheets>
    <sheet name="Sheet1" sheetId="1" r:id="rId1"/>
  </sheets>
</workbook>`
	const rels = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>
</Relationships>`
	prefix := []byte(`<?xml version="1.0"?><worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData><row r="1"><c r="A1" t="inlineStr"><is><t>`)
	suffix := []byte(`</t></is></c></row></sheetData></worksheet>`)
	hello, err := charmap.Windows1251.NewEncoder().Bytes([]byte("Привет"))
	if err != nil {
		t.Fatal(err)
	}
	sheet1 := append(append(append([]byte{}, prefix...), hello...), suffix...)
	return buildXLSX(map[string][]byte{
		"[Content_Types].xml":        []byte(`<?xml version="1.0"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"></Types>`),
		"xl/workbook.xml":            []byte(workbook),
		"xl/_rels/workbook.xml.rels": []byte(rels),
		"xl/worksheets/sheet1.xml":   sheet1,
	})
}
