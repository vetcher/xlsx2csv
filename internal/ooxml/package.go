package ooxml

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
)

// Package is an opened XLSX (OOXML) zip package.
type Package struct {
	data []byte
	zr   *zip.Reader
}

// Open reads an XLSX file from raw zip bytes.
func Open(data []byte) (*Package, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("ooxml: empty package")
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("ooxml: open zip: %w", err)
	}
	return &Package{data: data, zr: zr}, nil
}

// ReadPart returns the uncompressed bytes of a part by zip path (e.g. "xl/workbook.xml").
func (p *Package) ReadPart(path string) ([]byte, error) {
	for _, f := range p.zr.File {
		if f.Name != path {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("ooxml: open part %q: %w", path, err)
		}
		body, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, fmt.Errorf("ooxml: read part %q: %w", path, err)
		}
		return body, nil
	}
	return nil, fmt.Errorf("ooxml: part not found: %q", path)
}
