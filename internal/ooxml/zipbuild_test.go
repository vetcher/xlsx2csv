package ooxml_test

import (
	"archive/zip"
	"bytes"
	"testing"
)

func buildZip(parts map[string][]byte) []byte {
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

func mustBuildMinimalWorkbook(t *testing.T) []byte {
	t.Helper()
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
	const emptySheet = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData/></worksheet>`
	return buildZip(map[string][]byte{
		"[Content_Types].xml":              []byte(`<?xml version="1.0"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"></Types>`),
		"xl/workbook.xml":                  []byte(workbook),
		"xl/_rels/workbook.xml.rels":       []byte(rels),
		"xl/worksheets/sheet1.xml":         []byte(emptySheet),
		"xl/worksheets/sheet2.xml":         []byte(emptySheet),
	})
}
