// format.go extends htmltable with additional output formats and utilities.
package htmltable

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ToJSON converts the table to a JSON array of objects using headers as keys.
func (t *Table) ToJSON() (string, error) {
	if len(t.Headers) == 0 {
		return "[]", nil
	}
	var records []map[string]string
	for _, row := range t.Rows {
		record := make(map[string]string)
		for i, h := range t.Headers {
			if i < len(row) {
				record[h] = row[i]
			} else {
				record[h] = ""
			}
		}
		records = append(records, record)
	}
	b, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ToMarkdown converts the table to a Markdown formatted table string.
func (t *Table) ToMarkdown() string {
	var sb strings.Builder
	// Header row.
	if len(t.Headers) > 0 {
		sb.WriteString("| ")
		sb.WriteString(strings.Join(t.Headers, " | "))
		sb.WriteString(" |\n")
		// Separator.
		sb.WriteString("|")
		for range t.Headers {
			sb.WriteString(" --- |")
		}
		sb.WriteString("\n")
	}
	// Data rows.
	for _, row := range t.Rows {
		sb.WriteString("| ")
		sb.WriteString(strings.Join(row, " | "))
		sb.WriteString(" |\n")
	}
	return sb.String()
}

// Search returns rows where the cell at column colIdx contains the search term.
func (t *Table) Search(colIdx int, term string) [][]string {
	var out [][]string
	lower := strings.ToLower(term)
	for _, row := range t.Rows {
		if colIdx < len(row) && strings.Contains(strings.ToLower(row[colIdx]), lower) {
			out = append(out, row)
		}
	}
	return out
}

// SortBy returns a new table sorted by the given column index (ascending).
func (t *Table) SortBy(colIdx int) *Table {
	if colIdx < 0 || colIdx >= t.NumCols() {
		return t
	}
	rows := make([][]string, len(t.Rows))
	copy(rows, t.Rows)
	// Simple insertion sort for stability.
	for i := 1; i < len(rows); i++ {
		key := cellAt(rows[i], colIdx)
		j := i - 1
		for j >= 0 && cellAt(rows[j], colIdx) > key {
			rows[j+1] = rows[j]
			j--
		}
		rows[j+1] = rows[i]
	}
	return &Table{Caption: t.Caption, Headers: t.Headers, Rows: rows}
}

func cellAt(row []string, idx int) string {
	if idx < len(row) {
		return row[idx]
	}
	return ""
}

// Summary returns a human-readable summary of the table dimensions.
func (t *Table) Summary() string {
	return fmt.Sprintf("%d columns x %d rows", t.NumCols(), t.NumRows())
}

// IsEmpty reports whether the table has no data rows.
func (t *Table) IsEmpty() bool { return len(t.Rows) == 0 }

// Transpose returns a new table with rows and columns swapped.
func (t *Table) Transpose() *Table {
	cols := t.NumCols()
	rows := t.NumRows()
	if cols == 0 {
		return &Table{}
	}
	newRows := make([][]string, cols)
	for c := 0; c < cols; c++ {
		var row []string
		if c < len(t.Headers) {
			row = append(row, t.Headers[c])
		} else {
			row = append(row, "")
		}
		for r := 0; r < rows; r++ {
			row = append(row, t.Cell(r, c))
		}
		newRows[c] = row
	}
	return &Table{Rows: newRows}
}
