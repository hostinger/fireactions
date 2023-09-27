package printer

import (
	"io"
	"os"

	"github.com/olekukonko/tablewriter"
)

var printer *Printer

type Printable interface {
	Headers() []string
	Rows() [][]string
}

type Printer struct {
	tablewriter *tablewriter.Table
}

func New(w io.Writer) *Printer {
	tw := tablewriter.NewWriter(w)
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

	p := &Printer{
		tablewriter: tw,
	}

	return p
}

func (p *Printer) Print(printable Printable) {
	p.tablewriter.SetHeader(printable.Headers())
	p.tablewriter.AppendBulk(printable.Rows())
	p.tablewriter.Render()
}

func Get() *Printer {
	if printer == nil {
		printer = New(os.Stdout)
	}

	return printer
}
