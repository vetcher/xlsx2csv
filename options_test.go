package xlsx2csv

import (
	"testing"

	"golang.org/x/text/encoding/charmap"
)

func TestApplyOptions(t *testing.T) {
	cfg := apply(nil)
	if cfg.errorLimit != 1 || cfg.placeholder != "" || cfg.sheetIndex != 0 {
		t.Fatalf("defaults %#v", cfg)
	}

	cfg = apply([]Option{WithErrorLimit(0)})
	if cfg.errorLimit != 1 {
		t.Fatal(cfg.errorLimit)
	}

	cfg = apply([]Option{
		WithSheetIndex(2),
		WithSheetName("X"),
		WithEncoding(charmap.Windows1251),
		WithErrorLimit(0),
		WithErrorPlaceholder("?"),
	})
	if cfg.errorLimit != 1 || cfg.placeholder != "?" || cfg.sheetIndex != 2 ||
		cfg.sheetName != "X" || cfg.enc == nil || !cfg.useName {
		t.Fatalf("overrides %#v", cfg)
	}

	cfg = apply([]Option{WithErrorLimit(-1), WithSheetName("A")})
	if cfg.errorLimit != -1 || !cfg.useName || cfg.sheetName != "A" {
		t.Fatalf("%#v", cfg)
	}

	cfg = apply([]Option{WithSheetIndex(3)})
	if cfg.sheetIndex != 3 || cfg.useName {
		t.Fatalf("sheet index %#v", cfg)
	}
}
