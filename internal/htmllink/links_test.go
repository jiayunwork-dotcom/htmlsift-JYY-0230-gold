package htmllink

import (
	"strings"
	"testing"

	"htmlsift/internal/htmlparse"
)

func mustDoc(t *testing.T, s string) *htmlparse.Doc {
	t.Helper()
	d, err := htmlparse.Parse(s)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	return d
}

func TestExtractLinks(t *testing.T) {
	d := mustDoc(t, `<html><body>
		<a href="/one">one</a>
		<a href="/two" class="x">two</a>
		<a>no href</a>
		<img src="pic.png">
		<link href="/css/main.css">
	</body></html>`)
	links, err := Extract(d, "")
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}
	if len(links) != 4 {
		t.Fatalf("links = %d, want 4", len(links))
	}
	if links[0].Tag != "a" || links[0].Href != "/one" || strings.TrimSpace(links[0].Text) != "one" {
		t.Errorf("first link wrong: %+v", links[0])
	}
}

func TestExtractResolveBase(t *testing.T) {
	d := mustDoc(t, `<html><body><a href="page.html">p</a></body></html>`)
	links, err := Extract(d, "https://example.com/sub/")
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}
	if len(links) != 1 || links[0].Resolved != "https://example.com/sub/page.html" {
		t.Errorf("resolved = %q, want https://example.com/sub/page.html", links[0].Resolved)
	}
}

func TestExtractInvalidBase(t *testing.T) {
	d := mustDoc(t, `<html><body><a href="x">x</a></body></html>`)
	if _, err := Extract(d, "%zz"); err == nil {
		t.Error("invalid base should error")
	}
}

func TestResolveURL(t *testing.T) {
	cases := []struct {
		base, ref, want string
	}{
		{"https://ex.com/a/b", "c", "https://ex.com/a/c"},
		{"https://ex.com/a/", "../d", "https://ex.com/d"},
		{"https://ex.com", "/root", "https://ex.com/root"},
		{"https://ex.com/a", "#frag", "https://ex.com/a#frag"},
		{"", "rel/path", "rel/path"},
	}
	for _, c := range cases {
		got, err := ResolveURL(c.base, c.ref)
		if err != nil {
			t.Errorf("ResolveURL(%q,%q): %v", c.base, c.ref, err)
			continue
		}
		if got != c.want {
			t.Errorf("ResolveURL(%q,%q) = %q, want %q", c.base, c.ref, got, c.want)
		}
	}
}

func TestClassify(t *testing.T) {
	cases := []struct {
		in   string
		want Class
	}{
		{"https://x.com", ClassHTTP},
		{"http://x.com", ClassHTTP},
		{"mailto:a@b.c", ClassMailto},
		{"tel:+8613800000000", ClassTel},
		{"data:image/png;base64,AAAA", ClassData},
		{"javascript:alert(1)", ClassJavaScript},
		{"JAVASCRIPT:alert(1)", ClassJavaScript},
		{"#local", ClassFragment},
		{"../rel", ClassRelative},
		{"relative/path", ClassRelative},
		{"", ClassUnknown},
		{"ftp://x", ClassUnknown},
	}
	for _, c := range cases {
		if got := Classify(c.in); got != c.want {
			t.Errorf("Classify(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestFilterByScheme(t *testing.T) {
	d := mustDoc(t, `<html><body>
		<a href="https://ok">https</a>
		<a href="javascript:x">js</a>
		<a href="/rel">rel</a>
		<a href="#f">frag</a>
		<a href="mailto:a@b">mail</a>
	</body></html>`)
	links, _ := Extract(d, "")
	kept := FilterByScheme(links, ClassHTTP)
	if len(kept) != 3 {
		t.Fatalf("kept = %d, want 3 (https + rel + frag)", len(kept))
	}
	for _, l := range kept {
		if Classify(l.Href) == ClassJavaScript || Classify(l.Href) == ClassMailto {
			t.Errorf("unwanted link kept: %+v", l)
		}
	}
}

func TestUniqueByHref(t *testing.T) {
	d := mustDoc(t, `<html><body>
		<a href="/dup">1</a><a href="/dup">2</a><a href="/other">3</a>
	</body></html>`)
	links, _ := Extract(d, "")
	uniq := UniqueByHref(links)
	if len(uniq) != 2 {
		t.Fatalf("unique = %d, want 2", len(uniq))
	}
	if uniq[0].Href != "/dup" || uniq[1].Href != "/other" {
		t.Errorf("unique order wrong: %+v", uniq)
	}
}

func TestIsSameOrigin(t *testing.T) {
	cases := []struct {
		base, ref string
		want      bool
	}{
		{"https://ex.com/a", "https://ex.com/b", true},
		{"https://ex.com", "https://EX.com/x", true},
		{"https://ex.com", "http://ex.com/x", false},
		{"https://ex.com", "https://other.com/x", false},
		{"https://ex.com:8443", "https://ex.com:8443/x", true},
		{"https://ex.com:8443", "https://ex.com/x", false},
	}
	for _, c := range cases {
		got, err := IsSameOrigin(c.base, c.ref)
		if err != nil {
			t.Errorf("IsSameOrigin(%q,%q): %v", c.base, c.ref, err)
			continue
		}
		if got != c.want {
			t.Errorf("IsSameOrigin(%q,%q) = %v, want %v", c.base, c.ref, got, c.want)
		}
	}
}

func TestIsSameOriginInvalid(t *testing.T) {
	if _, err := IsSameOrigin("://bad", "https://x.com"); err == nil {
		t.Error("invalid base should error")
	}
}

func TestValidBidi(t *testing.T) {
	// Plain LTR and plain RTL text are both bidi-safe.
	if !ValidBidi("hello world") {
		t.Error("LTR text should be bidi-safe")
	}
	if !ValidBidi("مرحبا بالعالم") {
		t.Error("RTL text should be bidi-safe")
	}
	// An embedded right-to-left override character is not safe to render
	// into an LTR document without escaping.
	if ValidBidi("evil\u202Estuff") {
		t.Error("RLO embedding should be flagged")
	}
}

func TestLinkTextOK(t *testing.T) {
	if !LinkTextOK("click here", 100) {
		t.Error("normal text should pass")
	}
	if LinkTextOK("", 100) {
		t.Error("empty text should fail")
	}
	if LinkTextOK("very long text exceeding the limit", 10) {
		t.Error("over-length text should fail")
	}
	if LinkTextOK("bad\u202Etext", 100) {
		t.Error("bidi-unsafe text should fail")
	}
}

func TestClassifyString(t *testing.T) {
	if got := ClassHTTP.String(); got != "http" {
		t.Errorf("ClassHTTP.String = %q", got)
	}
	if got := ClassUnknown.String(); got != "unknown" {
		t.Errorf("ClassUnknown.String = %q", got)
	}
}

func TestAbsoluteLinks(t *testing.T) {
	d := mustDoc(t, `<html><body>
		<a href="https://a.example/">a</a>
		<a href="/rel">rel</a>
		<a href="http://b.example/">b</a>
		<a href="#f">f</a>
	</body></html>`)
	links, _ := Extract(d, "")
	abs := AbsoluteLinks(links)
	if len(abs) != 2 {
		t.Fatalf("absolute = %d, want 2", len(abs))
	}
	for _, l := range abs {
		if Classify(l.Href) != ClassHTTP {
			t.Errorf("non-http link in absolute: %+v", l)
		}
	}
}

func TestGroupByHost(t *testing.T) {
	d := mustDoc(t, `<html><body>
		<a href="https://a.example/x">1</a>
		<a href="https://a.example/y">2</a>
		<a href="https://b.example/z">3</a>
		<a href="/rel">4</a>
	</body></html>`)
	links, _ := Extract(d, "https://a.example/")
	groups := GroupByHost(links)
	// x, y and the resolved /rel all land under a.example.
	if len(groups["a.example"]) != 3 {
		t.Errorf("a.example count = %d, want 3", len(groups["a.example"]))
	}
	if len(groups["b.example"]) != 1 {
		t.Errorf("b.example count = %d, want 1", len(groups["b.example"]))
	}
	if len(groups[""]) != 0 {
		t.Errorf("empty-host count = %d, want 0", len(groups[""]))
	}
}
