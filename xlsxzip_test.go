package xlsx2csv_test

import (
	"archive/zip"
	"bytes"
	"testing"
)

func buildXLSX(parts map[string][]byte) []byte {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, body := range parts {
		w, err := zw.Create(name)
		if err != nil {
			panic(err)
		}
		if _, err := w.Write(body); err != nil {
			panic(err)
		}
	}
	if err := zw.Close(); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func fixtureTwoSheets(t testing.TB) []byte {
	t.Helper()
	return minimalDataXLSX()
}

func minimalDataXLSX() []byte {
	const workbook = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"
 xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <sheets>
    <sheet name="one" sheetId="1" r:id="rId1"/>
    <sheet name="two" sheetId="2" r:id="rId2" state="hidden"/>
  </sheets>
</workbook>`
	const rels = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet2.xml"/>
</Relationships>`
	const sharedStrings = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" count="2" uniqueCount="2">
  <si><t>a</t></si>
  <si><t>b</t></si>
</sst>`
	const sheet1 = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <sheetData>
    <row r="1"><c r="A1" t="s"><v>0</v></c></row>
  </sheetData>
</worksheet>`
	const sheet2 = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <sheetData>
    <row r="1"><c r="A1" t="s"><v>1</v></c></row>
  </sheetData>
</worksheet>`
	return buildXLSX(map[string][]byte{
		"[Content_Types].xml":        []byte(`<?xml version="1.0"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"></Types>`),
		"xl/workbook.xml":            []byte(workbook),
		"xl/_rels/workbook.xml.rels": []byte(rels),
		"xl/sharedStrings.xml":       []byte(sharedStrings),
		"xl/worksheets/sheet1.xml":   []byte(sheet1),
		"xl/worksheets/sheet2.xml":   []byte(sheet2),
	})
}

func TestBuildXLSX_IsValidZipWithParts(t *testing.T) {
	raw := buildXLSX(map[string][]byte{
		"[Content_Types].xml": []byte(`<?xml version="1.0"?><Types></Types>`),
		"xl/workbook.xml":     []byte(`<?xml version="1.0"?><workbook></workbook>`),
	})
	zr, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		t.Fatalf("zip open: %v", err)
	}
	if len(zr.File) != 2 {
		t.Fatalf("files=%d want 2", len(zr.File))
	}
}
