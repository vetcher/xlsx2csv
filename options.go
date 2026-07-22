package xlsx2csv

import "golang.org/x/text/encoding"

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

func WithSheetIndex(i int) Option {
	return func(c *config) { c.sheetIndex = i; c.useName = false }
}

func WithSheetName(name string) Option {
	return func(c *config) { c.sheetName = name; c.useName = true }
}

func WithEncoding(enc encoding.Encoding) Option {
	return func(c *config) { c.enc = enc }
}

func WithErrorLimit(n int) Option {
	return func(c *config) { c.errorLimit = n }
}

func WithErrorPlaceholder(s string) Option {
	return func(c *config) { c.placeholder = s }
}
