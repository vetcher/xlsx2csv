package ooxml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// ParseSharedStrings parses sharedStrings.xml (UTF-8) into a dense string table.
func ParseSharedStrings(utf8 []byte) ([]string, error) {
	dec := xml.NewDecoder(bytes.NewReader(utf8))
	var out []string
	var inSI bool
	var siText strings.Builder
	var inT bool
	for {
		tok, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("ooxml: parse shared strings: %w", err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "si":
				if t.Name.Space == "" || t.Name.Space == spreadsheetMLNS {
					inSI = true
					siText.Reset()
				}
			case "t":
				if inSI && (t.Name.Space == "" || t.Name.Space == spreadsheetMLNS) {
					inT = true
				}
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "si":
				if inSI && (t.Name.Space == "" || t.Name.Space == spreadsheetMLNS) {
					out = append(out, siText.String())
					inSI = false
				}
			case "t":
				inT = false
			}
		case xml.CharData:
			if inT {
				siText.Write(t)
			}
		}
	}
	return out, nil
}
