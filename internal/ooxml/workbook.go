package ooxml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"path"
	"strings"
)

const (
	spreadsheetMLNS = "http://schemas.openxmlformats.org/spreadsheetml/2006/main"
	relationshipNS  = "http://schemas.openxmlformats.org/officeDocument/2006/relationships"
	packageRelNS    = "http://schemas.openxmlformats.org/package/2006/relationships"
)

// SheetInfo describes one worksheet in the workbook.
type SheetInfo struct {
	Name   string
	Index  int
	Path   string
	Hidden bool
}

// ParseWorkbook parses workbook.xml and workbook.xml.rels (both UTF-8) into sheet metadata.
func ParseWorkbook(workbookUTF8 []byte, relsUTF8 []byte) ([]SheetInfo, error) {
	targets, err := parseWorkbookRels(relsUTF8)
	if err != nil {
		return nil, err
	}
	sheets, err := parseWorkbookSheets(workbookUTF8)
	if err != nil {
		return nil, err
	}
	out := make([]SheetInfo, 0, len(sheets))
	for i, sh := range sheets {
		target, ok := targets[sh.relID]
		if !ok {
			return nil, fmt.Errorf("ooxml: workbook rel %q not found", sh.relID)
		}
		out = append(out, SheetInfo{
			Name:   sh.name,
			Index:  i,
			Path:   resolveXLPath(target),
			Hidden: sh.hidden,
		})
	}
	return out, nil
}

type workbookSheet struct {
	name   string
	relID  string
	hidden bool
}

func parseWorkbookSheets(workbookUTF8 []byte) ([]workbookSheet, error) {
	dec := xml.NewDecoder(bytes.NewReader(workbookUTF8))
	var sheets []workbookSheet
	for {
		tok, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("ooxml: parse workbook: %w", err)
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "sheet" {
			continue
		}
		if se.Name.Space != "" && se.Name.Space != spreadsheetMLNS {
			continue
		}
		var name, relID, state string
		for _, a := range se.Attr {
			switch {
			case a.Name.Local == "name" && a.Name.Space == "":
				name = a.Value
			case a.Name.Local == "id" && (a.Name.Space == relationshipNS || strings.HasSuffix(a.Name.Space, "/relationships")):
				relID = a.Value
			case a.Name.Local == "state" && a.Name.Space == "":
				state = a.Value
			}
		}
		hidden := state == "hidden" || state == "veryHidden"
		sheets = append(sheets, workbookSheet{name: name, relID: relID, hidden: hidden})
	}
	if len(sheets) == 0 {
		return nil, fmt.Errorf("ooxml: no sheets in workbook")
	}
	return sheets, nil
}

func parseWorkbookRels(relsUTF8 []byte) (map[string]string, error) {
	dec := xml.NewDecoder(bytes.NewReader(relsUTF8))
	targets := make(map[string]string)
	for {
		tok, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("ooxml: parse workbook rels: %w", err)
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "Relationship" {
			continue
		}
		if se.Name.Space != "" && se.Name.Space != packageRelNS {
			continue
		}
		var id, target string
		for _, a := range se.Attr {
			switch a.Name.Local {
			case "Id":
				id = a.Value
			case "Target":
				target = a.Value
			}
		}
		if id != "" && target != "" {
			targets[id] = target
		}
	}
	return targets, nil
}

// resolveXLPath resolves a relationship Target relative to xl/.
func resolveXLPath(target string) string {
	target = strings.ReplaceAll(target, "\\", "/")
	if strings.HasPrefix(target, "/") {
		return strings.TrimPrefix(target, "/")
	}
	return path.Join("xl", target)
}
