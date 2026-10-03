package ui

import (
	"fmt"
	"io"
	"strings"
)

// Table renders aligned columns in ASCII format
type Table struct {
	Headers []string
	Rows    [][]string
}

// NewTable initializes a Table with given headers
func NewTable(headers ...string) *Table {
	return &Table{
		Headers: headers,
		Rows:    make([][]string, 0),
	}
}

// AddRow adds a row of values to the table
func (t *Table) AddRow(cols ...string) {
	t.Rows = append(t.Rows, cols)
}

// Render writes the formatted table to w
func (t *Table) Render(w io.Writer) {
	if len(t.Headers) == 0 && len(t.Rows) == 0 {
		return
	}

	numCols := len(t.Headers)
	for _, r := range t.Rows {
		if len(r) > numCols {
			numCols = len(r)
		}
	}

	colWidths := make([]int, numCols)
	for i, h := range t.Headers {
		if len(h) > colWidths[i] {
			colWidths[i] = len(h)
		}
	}

	for _, row := range t.Rows {
		for i, cell := range row {
			if i < numCols && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// Format header
	if len(t.Headers) > 0 {
		headerParts := make([]string, numCols)
		for i := 0; i < numCols; i++ {
			h := ""
			if i < len(t.Headers) {
				h = t.Headers[i]
			}
			headerParts[i] = padRight(h, colWidths[i])
		}
		fmt.Fprintln(w, Bold(strings.Join(headerParts, "   ")))
	}

	// Format rows
	for _, row := range t.Rows {
		rowParts := make([]string, numCols)
		for i := 0; i < numCols; i++ {
			val := ""
			if i < len(row) {
				val = row[i]
			}
			rowParts[i] = padRight(val, colWidths[i])
		}
		fmt.Fprintln(w, strings.Join(rowParts, "   "))
	}
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}
