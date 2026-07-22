package xlsx2csv_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/vetcher/xlsx2csv"
)

func BenchmarkConvert_Small(b *testing.B) {
	raw := fixtureTwoSheets(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := xlsx2csv.Convert(bytes.NewReader(raw))
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkConvert_Medium(b *testing.B) {
	raw := fixtureMediumSheet(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := xlsx2csv.Convert(bytes.NewReader(raw))
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestConvert_AllocsSmoke(t *testing.T) {
	raw := fixtureTwoSheets(t)
	var n float64
	n = testing.AllocsPerRun(50, func() {
		_, _ = xlsx2csv.Convert(bytes.NewReader(raw))
	})
	t.Logf("allocs/op=%.0f", n)
	if n > 400 {
		t.Fatalf("allocs/op=%.0f exceeds ceiling (measured ~329 + margin)", n)
	}
}

// fixtureMediumSheet builds an XLSX with one sheet and many string cells (shared strings).
func fixtureMediumSheet(tb testing.TB) []byte {
	tb.Helper()
	const numRows = 500
	var sst strings.Builder
	sst.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" count="`)
	fmt.Fprintf(&sst, "%d", numRows)
	sst.WriteString(`" uniqueCount="`)
	fmt.Fprintf(&sst, "%d", numRows)
	sst.WriteString(`">`)
	for i := 0; i < numRows; i++ {
		fmt.Fprintf(&sst, `<si><t>cell-%d</t></si>`, i)
	}
	sst.WriteString(`</sst>`)

	var sheet strings.Builder
	sheet.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <sheetData>`)
	for i := 1; i <= numRows; i++ {
		fmt.Fprintf(&sheet, `<row r="%d"><c r="A%d" t="s"><v>%d</v></c></row>`, i, i, i-1)
	}
	sheet.WriteString(`
  </sheetData>
</worksheet>`)

	const workbook = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"
 xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <sheets>
    <sheet name="data" sheetId="1" r:id="rId1"/>
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
		"xl/sharedStrings.xml":       []byte(sst.String()),
		"xl/worksheets/sheet1.xml":   []byte(sheet.String()),
	})
}
