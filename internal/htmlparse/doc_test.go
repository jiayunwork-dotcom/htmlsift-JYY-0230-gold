package htmlparse

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestParseBasic(t *testing.T) {
	d, err := Parse(`<html><body><p>hello</p></body></html>`)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if d.Root == nil {
		t.Fatal("nil root")
	}
	st := d.Stats()
	if st.Elements < 3 {
		t.Errorf("elements = %d, want >= 3 (html/body/p)", st.Elements)
	}
	if st.TextNodes < 1 {
		t.Errorf("text nodes = %d, want >= 1", st.TextNodes)
	}
}

func TestParseEmptyInput(t *testing.T) {
	if _, err := Parse("   "); err == nil {
		t.Error("Parse(blank) should fail")
	}
}

func TestStatsCountsLinksAndImages(t *testing.T) {
	d, _ := Parse(`<html><body>
		<a href="/x">x</a><a>no href</a>
		<img src="a.png"><img src="b.png">
	</body></html>`)
	st := d.Stats()
	if st.Links != 1 {
		t.Errorf("links = %d, want 1 (a with href only)", st.Links)
	}
	if st.Images != 2 {
		t.Errorf("images = %d, want 2", st.Images)
	}
	if st.Tags["a"] != 2 {
		t.Errorf("a tags = %d, want 2", st.Tags["a"])
	}
}

func TestStatsComments(t *testing.T) {
	d, _ := Parse(`<html><body><!-- note --><p>x</p></body></html>`)
	if st := d.Stats(); st.Comments != 1 {
		t.Errorf("comments = %d, want 1", st.Comments)
	}
}

func TestVisibleTextSkipsScriptAndStyle(t *testing.T) {
	d, _ := Parse(`<html><head><title>ignored</title><style>body{}</style></head>
		<body><p>Hello <b>world</b></p><script>alert(1)</script><p>again</p></body></html>`)
	got := VisibleText(d)
	if !strings.Contains(got, "Hello world") {
		t.Errorf("visible text missing prose: %q", got)
	}
	if strings.Contains(got, "alert") || strings.Contains(got, "body{}") {
		t.Errorf("script/style leaked into visible text: %q", got)
	}
	if !strings.Contains(got, "again") {
		t.Errorf("visible text missing trailing prose: %q", got)
	}
}

func TestCollapseSpace(t *testing.T) {
	if got := CollapseSpace("a   b\n\t c "); got != "a b c" {
		t.Errorf("CollapseSpace = %q, want %q", got, "a b c")
	}
	if got := CollapseSpace("   "); got != "" {
		t.Errorf("CollapseSpace(blank) = %q, want empty", got)
	}
}

func TestFindByTag(t *testing.T) {
	d, _ := Parse(`<html><body><p>1</p><p>2</p><span>3</span></body></html>`)
	ps := FindByTag(d, "p")
	if len(ps) != 2 {
		t.Fatalf("p count = %d, want 2", len(ps))
	}
	if Attr(ps[0], "id") != "" {
		t.Errorf("missing attr should be empty")
	}
}

func TestWalkPreOrder(t *testing.T) {
	d, _ := Parse(`<html><body><p>a</p></body></html>`)
	var order []string
	_ = d.Walk(func(n *html.Node) error {
		if n.Type == html.ElementNode {
			order = append(order, n.Data)
		}
		return nil
	})
	// html comes first (document root is a DocumentNode).
	if len(order) == 0 || order[0] != "html" {
		t.Errorf("pre-order should start with html, got %v", order)
	}
}

func TestParseFragmentTable(t *testing.T) {
	nodes, err := ParseFragment(`<tr><td>1</td></tr>`, "table")
	if err != nil {
		t.Fatalf("ParseFragment error: %v", err)
	}
	if len(nodes) == 0 {
		t.Fatal("no fragment nodes")
	}
	s, err := RenderNode(nodes[0])
	if err != nil {
		t.Fatalf("RenderNode error: %v", err)
	}
	if !strings.Contains(s, "<tr>") {
		t.Errorf("fragment render = %q, want a <tr>", s)
	}
}

func TestRenderRoundTrip(t *testing.T) {
	in := `<html><body><p title="a&amp;b">x &lt; y</p></body></html>`
	d, err := Parse(in)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	out, err := d.Render()
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if !strings.Contains(out, "title=\"a&amp;b\"") || !strings.Contains(out, "x &lt; y") {
		t.Errorf("render lost escaping: %q", out)
	}
}

func TestAttr(t *testing.T) {
	d, _ := Parse(`<html><body><img src="x.png" alt="y"></body></html>`)
	imgs := FindByTag(d, "img")
	if len(imgs) != 1 {
		t.Fatalf("img count = %d", len(imgs))
	}
	if got := Attr(imgs[0], "src"); got != "x.png" {
		t.Errorf("src = %q", got)
	}
	if got := Attr(imgs[0], "missing"); got != "" {
		t.Errorf("missing attr = %q, want empty", got)
	}
}

func TestHiddenTagsSet(t *testing.T) {
	if !HiddenTags["script"] || !HiddenTags["style"] || !HiddenTags["head"] {
		t.Errorf("HiddenTags missing core entries: %v", HiddenTags)
	}
}
