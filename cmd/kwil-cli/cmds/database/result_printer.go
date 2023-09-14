package database

import (
	"fmt"
	"strings"
)

func PrintTableAny(data []map[string]any) {
	conv := make([]map[string]string, len(data))
	for i, row := range data {
		conv[i] = make(map[string]string)
		for k, v := range row {
			conv[i][k] = fmt.Sprintf("%v", v)
		}
	}

	printTable(conv)
}

func printTable(data []map[string]string) {
	if len(data) == 0 {
		fmt.Println("No data to display.")
		return
	}

	// Find the column names and their maximum widths
	headers := make([]string, 0)
	colWidths := make(map[string]int)
	for _, row := range data {
		for k, v := range row {
			if _, exists := colWidths[k]; !exists {
				headers = append(headers, k)
				colWidths[k] = len(k)
			}
			if len(v) > colWidths[k] {
				colWidths[k] = len(v)
			}
		}
	}

	// Print the header
	for _, h := range headers {
		fmt.Printf("| %-*s ", colWidths[h], h)
	}
	fmt.Println("|")

	// Print the separator line
	for _, h := range headers {
		fmt.Printf("|-%s-",
			strings.Repeat("-", colWidths[h]))
	}
	fmt.Println("|")

	// Print the rows
	for _, row := range data {
		for _, h := range headers {
			fmt.Printf("| %-*s ", colWidths[h], row[h])
		}
		fmt.Println("|")
	}
}

// printTableTrunc prints a table, truncating values that are too long.
func printTableTrunc(data []map[string]string) {
	if len(data) == 0 {
		fmt.Println("No data to display.")
		return
	}

	headers := make([]string, 0)
	colWidths := make(map[string]int)
	for _, row := range data {
		for k, v := range row {
			if _, exists := colWidths[k]; !exists {
				headers = append(headers, k)
				colWidths[k] = len(k)
			}
			if len(v) > colWidths[k] {
				colWidths[k] = len(v)
			}
			if colWidths[k] > 32 {
				colWidths[k] = 32
			}
		}
	}

	for _, h := range headers {
		fmt.Printf("| %-*s ", colWidths[h], h)
	}
	fmt.Println("|")

	for _, h := range headers {
		fmt.Printf("|-%s-", strings.Repeat("-", colWidths[h]))
	}
	fmt.Println("|")

	for _, row := range data {
		for _, h := range headers {
			val := row[h]
			if len(val) > 32 {
				val = val[:29] + "..."
			}
			fmt.Printf("| %-*s ", colWidths[h], val)
		}
		fmt.Println("|")
	}
}
