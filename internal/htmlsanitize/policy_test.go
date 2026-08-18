package htmlsanitize

import (
	"strings"
	"testing"
)

func TestSanitizeDropsScript(t *testing.T) {
	out, err := DefaultPolicy().Sanitize(`<html><body><p>ok</p><script>alert(1)</script></body></html>`)
	if err != nil {
		t.Fatalf("Sanitize: %v", err)
	}
	if strings.Contains(out, "<script") || strings.Contains(out, "alert") {
		t.Errorf("script survived: %q", out)
	}
	if !strings.Contains(out, "ok") {
		t.Errorf("prose lost: %q", out)
	}
}

func TestSanitizeStripsEventHandlers(t *testing.T) {
	out, _ := DefaultPolicy().Sanitize(`<html><body><p onclick="x()" title="t">hi</p></body></html>`)
	if strings.Contains(out, "onclick") {
		t.Errorf("event handler survived: %q", out)
	}
	if !strings.Contains(out, "title=\"t\"") {
		t.Errorf("allowed attr lost: %q", out)
	}
}

func TestSanitizeDropsJavascriptHref(t *testing.T) {
	out, _ := DefaultPolicy().Sanitize(`<html><body><a href="javascript:alert(1)">x</a></body></html>`)
	if strings.Contains(out, "javascript:") {
		t.Errorf("javascript: href survived: %q", out)
	}
}

func TestSanitizeKeepsSafeMarkup(t *testing.T) {
	in := `<html><body><p>Hello <a href="https://ex.com" rel="ugc">link</a></p><ul><li>a</li></ul></body></html>`
	out, err := DefaultPolicy().Sanitize(in)
	if err != nil {
		t.Fatalf("Sanitize: %v", err)
	}
	for _, want := range []string{"<p>", "<a href=\"https://ex.com\"", "<ul>", "<li>"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %q", want, out)
		}
	}
}

func TestSanitizeAddsNofollow(t *testing.T) {
	out, _ := DefaultPolicy().Sanitize(`<html><body><a href="https://ex.com">x</a></body></html>`)
	if !strings.Contains(out, "rel=\"nofollow\"") {
		t.Errorf("nofollow not added: %q", out)
	}
	// An existing rel without nofollow is extended, not duplicated.
	out2, _ := DefaultPolicy().Sanitize(`<html><body><a href="https://ex.com" rel="ugc">x</a></body></html>`)
	if !strings.Contains(out2, "ugc nofollow") {
		t.Errorf("existing rel not extended: %q", out2)
	}
	if strings.Count(out2, "nofollow") != 1 {
		t.Errorf("nofollow duplicated: %q", out2)
	}
}

func TestSanitizeDataImageAllowed(t *testing.T) {
	in := `<html><body><img src="data:image/png;base64,AAAA"></body></html>`
	out, _ := DefaultPolicy().Sanitize(in)
	if !strings.Contains(out, "data:image/png") {
		t.Errorf("safe data image dropped: %q", out)
	}
}

func TestSanitizeDataNonImageDenied(t *testing.T) {
	in := `<html><body><img src="data:text/html;base64,PHNjcmlwdD4="></body></html>`
	out, _ := DefaultPolicy().Sanitize(in)
	if strings.Contains(out, "data:text/html") {
		t.Errorf("non-image data URL survived: %q", out)
	}
}

func TestSanitizeStripsComments(t *testing.T) {
	out, _ := DefaultPolicy().Sanitize(`<html><body><!-- secret --><p>x</p></body></html>`)
	if strings.Contains(out, "secret") {
		t.Errorf("comment survived: %q", out)
	}
}

func TestSanitizeKeepsCommentsWhenConfigured(t *testing.T) {
	p := DefaultPolicy()
	p.StripComments = false
	out, _ := p.Sanitize(`<html><body><!-- keep --><p>x</p></body></html>`)
	if !strings.Contains(out, "keep") {
		t.Errorf("comment dropped despite config: %q", out)
	}
}

func TestSanitizeIdempotent(t *testing.T) {
	inputs := []string{
		`<html><body><p onclick="x">a <b>b</b></p><script>1</script><a href="javascript:x">c</a></body></html>`,
		`<html><body><img src="data:text/html;base64,AAAA"><img src="data:image/gif;base64,AAAA"></body></html>`,
		`<html><body><table><tr><td colspan="2">x</td></tr></table></body></html>`,
		`<html><body><p>café e\u0301</p></body></html>`,
	}
	for _, in := range inputs {
		once, err := DefaultPolicy().Sanitize(in)
		if err != nil {
			t.Fatalf("Sanitize: %v", err)
		}
		twice, err := DefaultPolicy().Sanitize(once)
		if err != nil {
			t.Fatalf("Sanitize(once): %v", err)
		}
		if once != twice {
			t.Errorf("not idempotent:\n once=%q\ntwice=%q", once, twice)
		}
	}
}

func TestSanitizeNFCNormalizes(t *testing.T) {
	// "é" as e + combining acute (U+0065 U+0301) normalizes to U+00E9.
	out, _ := DefaultPolicy().Sanitize("<html><body><p>cafe\u0301</p></body></html>")
	if !strings.Contains(out, "café") {
		t.Errorf("text not NFC normalized: %q", out)
	}
}

func TestSanitizeCustomPolicyDropsWholeSubtree(t *testing.T) {
	p := DefaultPolicy()
	delete(p.Elements, "table")
	out, _ := p.Sanitize(`<html><body><p>keep</p><table><tr><td>drop me</td></tr></table></body></html>`)
	if strings.Contains(out, "drop me") {
		t.Errorf("disallowed block subtree survived: %q", out)
	}
	if !strings.Contains(out, "keep") {
		t.Errorf("allowed prose lost: %q", out)
	}
}

func TestSanitizeCustomPolicyPreservesPhrasingText(t *testing.T) {
	p := DefaultPolicy()
	delete(p.Elements, "code")
	out, _ := p.Sanitize(`<html><body><p>run <code>go test</code> now</p></body></html>`)
	if !strings.Contains(out, "go test") {
		t.Errorf("phrasing text lost when element dropped: %q", out)
	}
	if strings.Contains(out, "<code>") {
		t.Errorf("disallowed phrasing element survived: %q", out)
	}
}

func TestSanitizeFragment(t *testing.T) {
	out, err := DefaultPolicy().SanitizeFragment(`<p>a</p><script>x</script>`)
	if err != nil {
		t.Fatalf("SanitizeFragment: %v", err)
	}
	if strings.Contains(out, "<script") {
		t.Errorf("script survived in fragment: %q", out)
	}
	if !strings.Contains(out, "<p>a</p>") {
		t.Errorf("fragment prose lost: %q", out)
	}
}

func TestSanitizeReportCounts(t *testing.T) {
	in := `<html><body>
		<script>bad</script>
		<p onclick="x" data-a="1" title="t">keep</p>
		<a href="javascript:x">bad link</a>
		<a href="https://ok.example/">good</a>
		<table><tr><td>x</td></tr></table>
	</body></html>`
	out, rep, err := DefaultPolicy().SanitizeReport(in)
	if err != nil {
		t.Fatalf("SanitizeReport: %v", err)
	}
	if strings.Contains(out, "<script") || strings.Contains(out, "javascript:") {
		t.Errorf("dangerous content survived: %q", out)
	}
	if rep.RemovedElements != 1 {
		t.Errorf("RemovedElements = %d, want 1 (script)", rep.RemovedElements)
	}
	if rep.RemovedURLs != 1 {
		t.Errorf("RemovedURLs = %d, want 1 (javascript href)", rep.RemovedURLs)
	}
	if rep.RemovedAttrs < 2 {
		t.Errorf("RemovedAttrs = %d, want >= 2 (onclick, data-a)", rep.RemovedAttrs)
	}
	if rep.KeptElements == 0 || rep.TextBytes == 0 {
		t.Errorf("KeptElements=%d TextBytes=%d, want > 0", rep.KeptElements, rep.TextBytes)
	}
}

func TestSanitizeReportCleanDoc(t *testing.T) {
	in := `<html><body><p>hi</p></body></html>`
	_, rep, err := DefaultPolicy().SanitizeReport(in)
	if err != nil {
		t.Fatalf("SanitizeReport: %v", err)
	}
	if rep.RemovedElements != 0 || rep.RemovedAttrs != 0 || rep.RemovedURLs != 0 {
		t.Errorf("clean doc should remove nothing, got %+v", rep)
	}
}

func TestSanitizeEmpty(t *testing.T) {
	out, err := DefaultPolicy().Sanitize(`<html></html>`)
	if err != nil {
		t.Fatalf("Sanitize(empty): %v", err)
	}
	if out == "" {
		t.Error("sanitize of empty doc should keep the wrappers")
	}
}

func TestSanitizeDisallowedGlobalAttr(t *testing.T) {
	// id/class are globally allowed; custom data-* and style must go.
	out, _ := DefaultPolicy().Sanitize(`<html><body><p id="x" class="y" data-z="1" style="color:red">t</p></body></html>`)
	if !strings.Contains(out, "id=\"x\"") || !strings.Contains(out, "class=\"y\"") {
		t.Errorf("global attrs lost: %q", out)
	}
	if strings.Contains(out, "data-z") || strings.Contains(out, "style=") {
		t.Errorf("non-allowed attrs survived: %q", out)
	}
}
