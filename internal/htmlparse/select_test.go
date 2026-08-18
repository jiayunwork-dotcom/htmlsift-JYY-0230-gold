package htmlparse

import (
	"testing"
)

func TestParseSelector(t *testing.T) {
	cases := []struct {
		in  string
		tag string
		id  string
		cls []string
		ok  bool
	}{
		{"div", "div", "", nil, true},
		{"#main", "", "main", nil, true},
		{".note", "", "", []string{"note"}, true},
		{"div#main", "div", "main", nil, true},
		{"div.note.warn", "div", "", []string{"note", "warn"}, true},
		{"*", "", "", nil, true},
		{"DIV", "div", "", nil, true},
		{"", "", "", nil, false},
		{"#", "", "", nil, false},
		{".", "", "", nil, false},
		{"div#a#b", "", "", nil, false},
		{"div span", "", "", nil, false},
	}
	for _, c := range cases {
		sel, err := ParseSelector(c.in)
		if c.ok {
			if err != nil {
				t.Errorf("ParseSelector(%q) error: %v", c.in, err)
				continue
			}
			if sel.Tag != c.tag || sel.ID != c.id || !sameStrings(sel.Classes, c.cls) {
				t.Errorf("ParseSelector(%q) = %+v, want tag=%q id=%q cls=%v", c.in, sel, c.tag, c.id, c.cls)
			}
		} else if err == nil {
			t.Errorf("ParseSelector(%q) should fail", c.in)
		}
	}
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestSelectByTag(t *testing.T) {
	d, _ := Parse(`<html><body><p>1</p><div><p>2</p></div><span>3</span></body></html>`)
	nodes, err := Select(d, "p")
	if err != nil {
		t.Fatalf("Select: %v", err)
	}
	if len(nodes) != 2 {
		t.Errorf("p count = %d, want 2", len(nodes))
	}
}

func TestSelectById(t *testing.T) {
	d, _ := Parse(`<html><body><div id="a">1</div><div id="b">2</div></body></html>`)
	n, err := ByID(d, "b")
	if err != nil {
		t.Fatalf("ByID: %v", err)
	}
	if n == nil {
		t.Fatal("ByID(b) = nil")
	}
	if Attr(n, "id") != "b" {
		t.Errorf("ByID returned wrong node")
	}
	if n, _ := ByID(d, "missing"); n != nil {
		t.Errorf("ByID(missing) should be nil")
	}
}

func TestSelectByClass(t *testing.T) {
	d, _ := Parse(`<html><body><p class="note">1</p><p class="warn note">2</p><p>3</p></body></html>`)
	nodes, _ := Select(d, "p.note")
	if len(nodes) != 2 {
		t.Errorf("p.note count = %d, want 2", len(nodes))
	}
	both, _ := Select(d, ".note.warn")
	if len(both) != 1 {
		t.Errorf(".note.warn count = %d, want 1", len(both))
	}
}

func TestSelectComposed(t *testing.T) {
	d, _ := Parse(`<html><body><div id="post" class="entry">x</div></body></html>`)
	nodes, _ := Select(d, "div#post.entry")
	if len(nodes) != 1 {
		t.Errorf("div#post.entry count = %d", len(nodes))
	}
	nodes, _ = Select(d, "div#post.other")
	if len(nodes) != 0 {
		t.Errorf("div#post.other should match nothing")
	}
}

func TestSelectWildcard(t *testing.T) {
	d, _ := Parse(`<html><body><p>a</p><div>b</div></body></html>`)
	nodes, _ := Select(d, "*")
	// html, head (implied), body, p, div.
	if len(nodes) != 5 {
		t.Errorf("* count = %d, want 5", len(nodes))
	}
}
