package htmltable

import (
	"strings"
	"testing"

	"htmlsift/internal/htmlparse"
)

func mustDoc(t *testing.T, s string) *htmlparse.Doc {
	t.Helper()
	d, err := htmlparse.Parse(s)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return d
}

func TestExtractFirstBasic(t *testing.T) {
	d := mustDoc(t, `<html><body>
		<table>
			<thead><tr><th>Name</th><th>Age</th></tr></thead>
			<tbody>
				<tr><td>Alice</td><td>30</td></tr>
				<tr><td>Bob</td><td>25</td></tr>
			</tbody>
		</table>
	</body></html>`)
	tbl, err := ExtractFirst(d)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if len(tbl.Headers) != 2 || tbl.Headers[0] != "Name" {
		t.Fatalf("headers = %v", tbl.Headers)
	}
	if tbl.NumRows() != 2 {
		t.Fatalf("rows = %d", tbl.NumRows())
	}
	if tbl.Cell(0, 0) != "Alice" {
		t.Fatalf("cell(0,0) = %q", tbl.Cell(0, 0))
	}
}

func TestExtractNoThead(t *testing.T) {
	d := mustDoc(t, `<html><body>
		<table>
			<tr><td>H1</td><td>H2</td></tr>
			<tr><td>D1</td><td>D2</td></tr>
		</table>
	</body></html>`)
	tbl, _ := ExtractFirst(d)
	if len(tbl.Headers) != 2 || tbl.Headers[0] != "H1" {
		t.Fatalf("headers = %v", tbl.Headers)
	}
	if tbl.NumRows() != 1 {
		t.Fatalf("rows = %d", tbl.NumRows())
	}
}

func TestExtractCaption(t *testing.T) {
	d := mustDoc(t, `<html><body><table><caption>My Table</caption><tr><td>x</td></tr></table></body></html>`)
	tbl, _ := ExtractFirst(d)
	if tbl.Caption != "My Table" {
		t.Fatalf("caption = %q", tbl.Caption)
	}
}

func TestExtractAllMultiple(t *testing.T) {
	d := mustDoc(t, `<html><body>
		<table><tr><td>A</td></tr></table>
		<table><tr><td>B</td></tr></table>
	</body></html>`)
	tables, err := ExtractAll(d)
	if err != nil {
		t.Fatalf("extract all: %v", err)
	}
	if len(tables) != 2 {
		t.Fatalf("tables = %d, want 2", len(tables))
	}
}

func TestExtractNoTable(t *testing.T) {
	d := mustDoc(t, `<html><body><p>no table</p></body></html>`)
	_, err := ExtractFirst(d)
	if err != ErrNoTable {
		t.Fatalf("expected ErrNoTable, got %v", err)
	}
}

func TestColumn(t *testing.T) {
	d := mustDoc(t, `<html><body><table>
		<thead><tr><th>X</th><th>Y</th></tr></thead>
		<tbody><tr><td>1</td><td>2</td></tr><tr><td>3</td><td>4</td></tr></tbody>
	</table></body></html>`)
	tbl, _ := ExtractFirst(d)
	col := tbl.Column(1)
	if len(col) != 2 || col[0] != "2" || col[1] != "4" {
		t.Fatalf("column = %v", col)
	}
}

func TestToCSV(t *testing.T) {
	d := mustDoc(t, `<html><body><table>
		<thead><tr><th>A</th><th>B</th></tr></thead>
		<tbody><tr><td>1</td><td>2</td></tr></tbody>
	</table></body></html>`)
	tbl, _ := ExtractFirst(d)
	csv, err := tbl.ToCSVString()
	if err != nil {
		t.Fatalf("csv: %v", err)
	}
	if !strings.Contains(csv, "A,B") {
		t.Fatalf("csv headers: %q", csv)
	}
	if !strings.Contains(csv, "1,2") {
		t.Fatalf("csv data: %q", csv)
	}
}
