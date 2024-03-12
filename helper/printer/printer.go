package printer

import (
	"fmt"
	"io"
	"strings"

	"github.com/olekukonko/tablewriter"
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
	tw := tablewriter.NewWriter(out)
	tw.SetAutoWrapText(false)
	tw.SetAutoFormatHeaders(true)
	tw.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	tw.SetAlignment(tablewriter.ALIGN_LEFT)
	tw.SetCenterSeparator("")
	tw.SetColumnSeparator("")
	tw.SetRowSeparator("")
	tw.SetHeaderLine(false)
	tw.SetBorder(false)
	tw.SetTablePadding("  ")
	tw.SetNoWhiteSpace(true)

	cols := item.Cols()
	if len(includeCols) > 0 && includeCols[0] != "" {
		cols = includeCols
	}

	for _, c := range includeCols {
		if _, ok := item.ColsMap()[c]; !ok {
			fmt.Fprintf(out, "Column doesn't exist: %s. Available columns: %v\n", c, strings.Join(item.Cols(), ", "))
			return
		}
	}

	tw.SetHeader(cols)

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

	tw.AppendBulk(values)
	tw.Render()
}
