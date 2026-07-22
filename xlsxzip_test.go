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
