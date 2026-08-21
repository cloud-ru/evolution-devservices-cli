// Package output handles printing human-readable tables and JSON output.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// Format selects the output format.
type Format int

const (
	// FormatAuto picks a format based on whether the output is a TTY.
	// When non-TTY (e.g. pipe to another command), JSON is used.
	FormatAuto Format = iota
	// FormatTable forces a human-readable table.
	FormatTable
	// FormatJSON forces JSON output.
	FormatJSON
)

// Printer renders data to w.
type Printer struct {
	W      io.Writer
	Format Format
}

// New creates a Printer using stdout and the requested format.
func New(format Format) *Printer {
	if format == FormatAuto {
		if isTerminal(os.Stdout) {
			format = FormatTable
		} else {
			format = FormatJSON
		}
	}
	return &Printer{W: os.Stdout, Format: format}
}

// PrintJSON marshals v as indented JSON to the writer.
func (p *Printer) PrintJSON(v any) error {
	enc := json.NewEncoder(p.W)
	enc.SetIndent("", "  ")

	err := enc.Encode(v)
	if err != nil {
		return fmt.Errorf("json encode %w", err)
	}

	return nil
}

// Table renders a human-readable table. Headers is the header row,
// rows are arbitrary string cells. A nil or empty rows slice prints
// a friendly placeholder.
func (p *Printer) Table(headers []string, rows [][]string) {
	if p.Format == FormatJSON {
		out := make([]map[string]string, 0, len(rows))
		for _, r := range rows {
			row := map[string]string{}
			for i, h := range headers {
				if i < len(r) {
					row[h] = r[i]
				}
			}
			out = append(out, row)
		}
		_ = p.PrintJSON(out)
		return
	}

	if len(rows) == 0 {
		fmt.Fprintln(p.W, "No results.")
		return
	}

	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, r := range rows {
		for i, cell := range r {
			if i >= len(widths) {
				continue
			}
			if w := displayWidth(cell); w > widths[i] {
				widths[i] = w
			}
		}
	}

	fmt.Fprintln(p.W, formatRow(headers, widths))
	fmt.Fprintln(p.W, formatSeparator(widths))
	for _, r := range rows {
		fmt.Fprintln(p.W, formatRow(r, widths))
	}
}

// KeyValue renders a small two-column key/value listing.
// Useful for `eds repo show` and `eds config`.
func (p *Printer) KeyValue(pairs [][2]string) {
	if p.Format == FormatJSON {
		obj := make(map[string]string, len(pairs))
		for _, kv := range pairs {
			obj[kv[0]] = kv[1]
		}
		_ = p.PrintJSON(obj)
		return
	}

	if len(pairs) == 0 {
		fmt.Fprintln(p.W, "(empty)")
		return
	}

	maxLen := 0
	for _, kv := range pairs {
		if len(kv[0]) > maxLen {
			maxLen = len(kv[0])
		}
	}
	for _, kv := range pairs {
		fmt.Fprintf(p.W, "%-*s  %s\n", maxLen, kv[0]+":", kv[1])
	}
}

// HumanTime returns a friendly short-form representation.
// Falls back to the raw string when parsing fails.
func HumanTime(s string) string {
	if s == "" {
		return "-"
	}
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC().Format("2006-01-02 15:04 UTC")
		}
	}
	return s
}

// HumanSize returns a human-readable byte count.
func HumanSize(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for n2 := n / unit; n2 >= unit; n2 /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}

func formatRow(cells []string, widths []int) string {
	parts := make([]string, len(cells))
	for i, c := range cells {
		w := 0
		if i < len(widths) {
			w = widths[i]
		}
		parts[i] = padRight(c, w)
	}
	return strings.Join(parts, "  ")
}

func formatSeparator(widths []int) string {
	parts := make([]string, len(widths))
	for i, w := range widths {
		parts[i] = strings.Repeat("-", w)
	}
	return strings.Join(parts, "  ")
}

func padRight(s string, width int) string {
	w := displayWidth(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}

// displayWidth approximates the visible width of s. It treats runes
// as single-width which is sufficient for typical repo names.
func displayWidth(s string) int {
	return len([]rune(s))
}
