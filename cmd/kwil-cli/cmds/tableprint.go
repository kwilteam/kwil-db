package cmds

import (
	"bytes"
	"sort"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func bindTableOutputFlags(cmd *cobra.Command) {
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
func recordsToTable(data []map[string]string, c *tableConfig) []byte {
	if c == nil {
		c = &tableConfig{}
	}
	if len(data) == 0 {
		return []byte("No data to display.")
	}

	// collect headers
	headers := make([]string, 0, len(data[0]))
	for k := range data[0] {
		headers = append(headers, k)
	}

	// keep the headers in a sorted order
	sort.Strings(headers)

	var buf bytes.Buffer
	table := tablewriter.NewWriter(&buf)
	table.SetHeader(headers)
	table.SetAutoFormatHeaders(false)
	table.SetBorders(
		tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	c.apply(table)

	for _, row := range data {
		rs := make([]string, 0, len(headers))
		for _, h := range headers {
			v := row[h]
			if c.maxRowWidth > 0 && len(v) > c.maxRowWidth {
				v = v[:c.maxRowWidth] + "..."
			}
			rs = append(rs, v)
		}
		table.Append(rs)
	}

	table.Render()
	return buf.Bytes()
}
