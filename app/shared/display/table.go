package display

import (
	"bytes"
	"fmt"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// FormatTable prints the table with the given columns and rows.
// It is meant to be used in a MarshalText method.
func FormatTable(cmd *cobra.Command, columns []string, rows [][]string) ([]byte, error) {
	outputFmt, err := getOutputFormat(cmd)
	if err != nil {
		return nil, err
	}

	if outputFmt == outputFormatJSON {
		// this means that we called FormatTable in MarshalJSON
		return nil, fmt.Errorf("json output is not supported for tables. this is a CLI bug")
	}

	c, err := getTableConfig(cmd)
	if err != nil {
		return nil, err
	}

	return recordsToTable(columns, rows, c), nil
}

// BindTableFlags binds the flags to the table config
func BindTableFlags(cmd *cobra.Command) {
	cmd.Flags().IntP("width", "w", 0, "Set the width of the table columns. Text beyond this width will be wrapped.")
	cmd.Flags().Bool("row-border", false, "Show border lines between rows.")
	cmd.Flags().Int("max-row-width", 0, "Set the maximum width of the row. Text beyond this width will be truncated.")
}

func getTableConfig(cmd *cobra.Command) (*tableConfig, error) {
	width, err := cmd.Flags().GetInt("width")
	if err != nil {
		return nil, err
	}
	rowBorder, err := cmd.Flags().GetBool("row-border")
	if err != nil {
		return nil, err
	}
	maxRowWidth, err := cmd.Flags().GetInt("max-row-width")
	if err != nil {
		return nil, err
	}

	return &tableConfig{
		width:              width,
		topAndBottomBorder: rowBorder,
		maxRowWidth:        maxRowWidth,
	}, nil
}

type tableConfig struct {
	// width is the width of the column.
	// text will be wrapped if it exceeds the width
	width int
	// topAndBottomBorder is the flag to show top and bottom border
	topAndBottomBorder bool
	// maxRowWidth is the maximum width of the row
	maxRowWidth int
}

func (t *tableConfig) apply(table *tablewriter.Table) {
	if t.width > 0 {
		table.SetColWidth(t.width)
	}
	if t.topAndBottomBorder {
		table.SetRowLine(true)
	}
}

// recordsToTable converts records to a formatted table structure
// that can be printed
func recordsToTable(columns []string, rows [][]string, c *tableConfig) []byte {
	if c == nil {
		c = &tableConfig{}
	}
	if len(rows) == 0 {
		return []byte("No data to display.")
	}

	// collect headers

	var buf bytes.Buffer
	table := tablewriter.NewWriter(&buf)
	table.SetHeader(columns)
	table.SetAutoFormatHeaders(false)
	table.SetBorders(
		tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	c.apply(table)

	for _, row := range rows {
		rs := make([]string, 0, len(columns))
		for _, col := range row {
			if c.maxRowWidth > 0 && len(col) > c.maxRowWidth {
				col = col[:c.maxRowWidth] + "..."
			}
			rs = append(rs, col)
		}
		table.Append(rs)
	}

	table.Render()
	return buf.Bytes()
}
