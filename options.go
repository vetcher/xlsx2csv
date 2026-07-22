package xlsx2csv

import "golang.org/x/text/encoding"

// Option configures Convert and Inspect.
type Option func(*config)

type config struct {
	sheetIndex  int
	sheetName   string
	useName     bool
	enc         encoding.Encoding
	errorLimit  int
	placeholder string
}

func apply(opts []Option) config {
	cfg := config{errorLimit: 1, placeholder: ""}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.errorLimit == 0 {
		cfg.errorLimit = 1
	}
	return cfg
}

// WithSheetIndex selects a worksheet by zero-based index.
func WithSheetIndex(i int) Option {
	return func(c *config) { c.sheetIndex = i; c.useName = false }
}

// WithSheetName selects a worksheet by its display name.
func WithSheetName(name string) Option {
	return func(c *config) { c.sheetName = name; c.useName = true }
}

// WithEncoding forces decoding of workbook and sheet XML with enc.
func WithEncoding(enc encoding.Encoding) Option {
	return func(c *config) { c.enc = enc }
}

// WithErrorLimit sets how many cell errors are collected before stopping (-1 for no limit).
func WithErrorLimit(n int) Option {
	return func(c *config) { c.errorLimit = n }
}

// WithErrorPlaceholder is the string written for cells that cannot be converted.
func WithErrorPlaceholder(s string) Option {
	return func(c *config) { c.placeholder = s }
}
