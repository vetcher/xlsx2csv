package xlsx2csv

import (
	"bytes"
	"regexp"
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"
	"golang.org/x/text/encoding/unicode"
)

var xmlEncodingRE = regexp.MustCompile(`(?i)encoding\s*=\s*["']([^"']+)["']`)

func decodeXML(raw []byte, override encoding.Encoding) ([]byte, string, error) {
	enc := override
	name := ""
	if enc == nil {
		enc, name = detectEncoding(raw)
	} else {
		name = encodingName(enc)
	}
	if enc == nil {
		return raw, "utf-8", nil
	}
	norm := strings.ToLower(strings.TrimSpace(name))
	if norm == "utf-8" || norm == "utf8" || norm == "" {
		return raw, "utf-8", nil
	}
	decoded, err := enc.NewDecoder().Bytes(raw)
	if err != nil {
		return nil, name, err
	}
	decoded = rewriteXMLDeclEncoding(decoded, "UTF-8")
	return decoded, name, nil
}

func detectEncoding(raw []byte) (encoding.Encoding, string) {
	if enc, name := bomEncoding(raw); enc != nil {
		return enc, name
	}
	if decl := parseXMLEncodingDecl(raw); decl != "" {
		if enc, err := htmlindex.Get(decl); err == nil {
			return enc, decl
		}
	}
	return nil, "utf-8"
}

func bomEncoding(raw []byte) (encoding.Encoding, string) {
	if len(raw) >= 3 && raw[0] == 0xEF && raw[1] == 0xBB && raw[2] == 0xBF {
		return unicode.UTF8, "utf-8"
	}
	if len(raw) >= 2 {
		if raw[0] == 0xFE && raw[1] == 0xFF {
			return unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM), "utf-16be"
		}
		if raw[0] == 0xFF && raw[1] == 0xFE {
			return unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM), "utf-16le"
		}
	}
	return nil, ""
}

func parseXMLEncodingDecl(raw []byte) string {
	limit := len(raw)
	if limit > 256 {
		limit = 256
	}
	head := raw[:limit]
	idx := bytes.Index(head, []byte("<?xml"))
	if idx < 0 {
		return ""
	}
	head = head[idx:]
	end := bytes.Index(head, []byte("?>"))
	if end < 0 {
		return ""
	}
	m := xmlEncodingRE.FindSubmatch(head[:end])
	if len(m) < 2 {
		return ""
	}
	return string(m[1])
}

func encodingName(enc encoding.Encoding) string {
	for _, label := range []string{
		"utf-8", "utf-16", "utf-16le", "utf-16be",
		"windows-1251", "windows-1252", "iso-8859-1", "iso-8859-15",
	} {
		e, err := htmlindex.Get(label)
		if err == nil && enc == e {
			return label
		}
	}
	return "unknown"
}

func rewriteXMLDeclEncoding(b []byte, newEnc string) []byte {
	idx := bytes.Index(b, []byte("<?xml"))
	if idx < 0 {
		return b
	}
	endRel := bytes.Index(b[idx:], []byte("?>"))
	if endRel < 0 {
		return b
	}
	end := idx + endRel
	decl := b[idx:end]
	if !xmlEncodingRE.Match(decl) {
		return b
	}
	newDecl := xmlEncodingRE.ReplaceAll(decl, []byte(`encoding="`+newEnc+`"`))
	out := make([]byte, 0, len(b)-(len(decl)-len(newDecl)))
	out = append(out, b[:idx]...)
	out = append(out, newDecl...)
	out = append(out, b[end:]...)
	return out
}
