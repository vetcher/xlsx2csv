package xlsx2csv

import (
	"bytes"
	"testing"

	"golang.org/x/text/encoding/charmap"
)

func TestDecodeXML_OverrideWindows1251(t *testing.T) {
	// "Привет" in windows-1251 bytes inside fake xml
	payload, _ := charmap.Windows1251.NewEncoder().Bytes([]byte("Привет"))
	raw := append([]byte(`<?xml version="1.0"?><t>`), payload...)
	raw = append(raw, []byte(`</t>`)...)
	out, name, err := decodeXML(raw, charmap.Windows1251)
	if err != nil {
		t.Fatal(err)
	}
	if name == "" {
		t.Fatal("empty name")
	}
	if !bytes.Contains(out, []byte("Привет")) {
		t.Fatalf("out=%q", out)
	}
}

func TestDecodeXML_RespectsDecl(t *testing.T) {
	payload, _ := charmap.Windows1251.NewEncoder().Bytes([]byte("Я"))
	raw := append([]byte(`<?xml version="1.0" encoding="windows-1251"?><t>`), payload...)
	raw = append(raw, []byte(`</t>`)...)
	out, _, err := decodeXML(raw, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(out, []byte("Я")) {
		t.Fatalf("out=%q", out)
	}
}

func TestDecodeXML_UTF8Default(t *testing.T) {
	raw := []byte(`<?xml version="1.0"?><t>ok</t>`)
	out, name, err := decodeXML(raw, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, raw) && !bytes.Contains(out, []byte("ok")) {
		t.Fatalf("out=%q name=%s", out, name)
	}
}
