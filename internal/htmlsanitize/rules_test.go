package htmlsanitize

import (
	"strings"
	"testing"
)

func TestStrictPolicyDropsLinks(t *testing.T) {
	out, _ := StrictPolicy().Sanitize(`<html><body><a href="https://x.com">link</a><p>keep</p></body></html>`)
	if strings.Contains(out, "<a") {
		t.Fatalf("link survived strict policy: %q", out)
	}
	if !strings.Contains(out, "<p>") {
		t.Fatalf("p dropped: %q", out)
	}
}

func TestPermissivePolicyAllowsVideo(t *testing.T) {
	out, _ := PermissivePolicy().Sanitize(`<html><body><video src="v.mp4" controls></video></body></html>`)
	if !strings.Contains(out, "<video") {
		t.Fatalf("video dropped: %q", out)
	}
}

func TestTextOnlyPolicy(t *testing.T) {
	out, _ := TextOnlyPolicy().Sanitize(`<html><body><p>hello</p><b>world</b></body></html>`)
	// Structural wrappers survive; <p> is block so dropped with content;
	// <b> is phrasing so its text survives, but the <b> tag itself is removed.
	if strings.Contains(out, "<b>") {
		t.Fatalf("<b> tag survived text-only: %q", out)
	}
	if !strings.Contains(out, "world") {
		t.Fatalf("phrasing text lost: %q", out)
	}
}

func TestPolicyBuilder(t *testing.T) {
	p := NewPolicyBuilder().
		AllowElements("p", "a").
		AllowAttrs("a", "href").
		AllowSchemes("https").
		StripComments(true).
		RequireNofollow(true).
		Build()
	out, _ := p.Sanitize(`<html><body><p>hi</p><a href="https://x.com">x</a><div>gone</div></body></html>`)
	if !strings.Contains(out, "<p>") {
		t.Fatalf("p dropped: %q", out)
	}
	if !strings.Contains(out, "href=\"https://x.com\"") {
		t.Fatalf("link dropped: %q", out)
	}
	if strings.Contains(out, "<div") {
		t.Fatalf("div survived: %q", out)
	}
}
