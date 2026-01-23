package printer

import (
	"fmt"
	"io"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

// OutputType is the type of output
type OutputType string

const (
	// Text prints the output in text (table) format
	Text OutputType = "text"
	// JSON prints the output in JSON format
	JSON OutputType = "json"
)

// Printable is the interface for any printable object
type Printable interface {
	Cols() []string
	ColsMap() map[string]string
	KV() []map[string]interface{}
}

// PrintText prints the output in text (table) format
func PrintText(item Printable, out io.Writer, includeCols []string) {
	// Create table with new v1.1.3 API
	table := tablewriter.NewTable(out,
		tablewriter.WithRenderer(renderer.NewBlueprint()),
		tablewriter.WithRendition(tw.Rendition{
			Borders: tw.BorderNone,
			Symbols: tw.NewSymbols(tw.StyleNone),
			Settings: tw.Settings{
				Separators: tw.SeparatorsNone,
				Lines:      tw.LinesNone,
			},
		}),
		tablewriter.WithHeaderAlignment(tw.AlignLeft),
		tablewriter.WithRowAlignment(tw.AlignLeft),
		tablewriter.WithHeaderAutoFormat(tw.On),
		tablewriter.WithRowAutoWrap(tw.WrapNone),
		tablewriter.WithConfig(tablewriter.Config{
			Header: tw.CellConfig{
				Padding: tw.CellPadding{
					Global: tw.Padding{Left: "", Right: "  "},
				},
			},
			Row: tw.CellConfig{
				Padding: tw.CellPadding{
					Global: tw.Padding{Left: "", Right: "  "},
				},
			},
			Behavior: tw.Behavior{
				TrimSpace: tw.Off,
			},
		}),
	)

	cols := item.Cols()
	if len(includeCols) > 0 && includeCols[0] != "" {
		cols = includeCols
	}

	for _, c := range includeCols {
		if _, ok := item.ColsMap()[c]; !ok {
			_, _ = fmt.Fprintf(out, "Column doesn't exist: %s. Available columns: %v\n", c, strings.Join(item.Cols(), ", "))
			return
		}
	}

	// Convert []string to []any for Header method
	headerCols := make([]any, len(cols))
	for i, col := range cols {
		headerCols[i] = col
	}
	table.Header(headerCols...)

	values := make([][]string, 0, len(item.KV()))
	for _, r := range item.KV() {
		row := make([]string, 0, len(cols))
		for _, c := range cols {
			v := r[c]
			if v == nil {
				v = "N/A"
			}

			var format string
			switch v.(type) {
			case bool:
				format = "%t"
			case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
				format = "%d"
			case float32, float64:
				format = "%f"
			default:
				format = "%s"
			}

			row = append(row, fmt.Sprintf(format, v))
		}

		values = append(values, row)
	}

	// Convert [][]string to []any for Bulk method
	bulkData := make([]any, len(values))
	for i, v := range values {
		bulkData[i] = v
	}

	if err := table.Bulk(bulkData); err != nil {
		_, _ = fmt.Fprintf(out, "Error rendering table: %v\n", err)
		return
	}

	if err := table.Render(); err != nil {
		_, _ = fmt.Fprintf(out, "Error rendering table: %v\n", err)
	}
}
