package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
)

// Printer writes CLI results in table, json, or jsonl form.
type Printer struct {
	Format string
	Out    io.Writer
	Err    io.Writer
}

func New(format string) *Printer {
	return &Printer{Format: strings.ToLower(format), Out: os.Stdout, Err: os.Stderr}
}

// PrintJSON encodes v as pretty JSON (or compact for jsonl).
func (p *Printer) PrintJSON(v any) error {
	enc := json.NewEncoder(p.Out)
	if p.Format != "jsonl" {
		enc.SetIndent("", "  ")
	}
	return enc.Encode(v)
}

// Print encodes v according to format. For table, rows/headers must be provided via PrintTable.
func (p *Printer) Print(v any) error {
	switch p.Format {
	case "json", "jsonl":
		return p.PrintJSON(v)
	default:
		// Fall back to JSON if caller didn't use PrintTable.
		return p.PrintJSON(v)
	}
}

// PrintTable writes a tab-separated table, or JSON if format is json/jsonl.
func (p *Printer) PrintTable(headers []string, rows [][]string, raw any) error {
	if p.Format == "json" || p.Format == "jsonl" {
		if raw != nil {
			return p.PrintJSON(raw)
		}
		// synthesize objects from headers/rows
		objs := make([]map[string]string, 0, len(rows))
		for _, row := range rows {
			m := make(map[string]string, len(headers))
			for i, h := range headers {
				if i < len(row) {
					m[h] = row[i]
				}
			}
			objs = append(objs, m)
		}
		return p.PrintJSON(objs)
	}

	w := tabwriter.NewWriter(p.Out, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, strings.Join(headers, "\t"))
	for _, row := range rows {
		// pad row to header length
		cells := make([]string, len(headers))
		for i := range headers {
			if i < len(row) {
				cells[i] = row[i]
			}
		}
		fmt.Fprintln(w, strings.Join(cells, "\t"))
	}
	return w.Flush()
}

// PrintLine writes a single JSONL object (for stream).
func (p *Printer) PrintLine(v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(p.Out, string(b))
	return err
}
