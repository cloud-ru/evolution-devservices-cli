package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"go.yaml.in/yaml/v3"
)

type Format int

const (
	FormatJSON Format = iota
	FormatYAML
)

type Printer struct {
	writer io.Writer
	format Format
	quiet  bool
}

func New(format Format, quiet bool) *Printer {
	return &Printer{writer: os.Stdout, format: format, quiet: quiet}
}

func (p *Printer) Print(v any) error {
	if p.quiet {
		return nil
	}

	if p.format == FormatJSON {
		enc := json.NewEncoder(p.writer)
		enc.SetIndent("", "  ")

		err := enc.Encode(v)
		if err != nil {
			return fmt.Errorf("json encode %w", err)
		}
	}
	if p.format == FormatYAML {
		enc := yaml.NewEncoder(p.writer)
		defer enc.Close()

		err := enc.Encode(v)
		if err != nil {
			return fmt.Errorf("yaml encode: %w", err)
		}
	}

	return nil
}
