package kslformat_test

import (
	"fmt"
	"os"
	"testing"
	"text/tabwriter"

	"ksl/kslformat"

	"github.com/stretchr/testify/require"
)

func TestFormat(t *testing.T) {
	wr := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintln(wr, "    id:\t[]int\t@pk @default(autoincrement())")
	fmt.Fprintln(wr, "    name:\tstring?\t")
	fmt.Fprintln(wr, "    age:\tint(8)\t@default(21)")
	fmt.Fprintln(wr)
	fmt.Fprintln(wr, "    email_address:\tstring?\t@unique")
	wr.Flush()
}

func TestFormatFile(t *testing.T) {
	file, err := os.ReadFile("/Users/bryan/Desktop/ksl/data/test.kwil")
	require.Nil(t, err)
	data, err := kslformat.Format(file)
	require.Nil(t, err)
	_ = data
}
