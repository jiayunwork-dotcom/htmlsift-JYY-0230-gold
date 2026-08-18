package htmlmeta

import (
	"testing"

	"htmlsift/internal/htmlparse"
)

func mustParse(t *testing.T, s string) *htmlparse.Doc {
	t.Helper()
	d, err := htmlparse.Parse(s)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return d
}

func TestExtractTitle(t *testing.T) {
	d := mustParse(t, `<html><head><title>Hello World</title></head><body></body></html>`)
	m := Extract(d)
	if m.Title != "Hello World" {
		t.Fatalf("title = %q", m.Title)
	}
}

func TestExtractDescription(t *testing.T) {
	d := mustParse(t, `<html><head><meta name="description" content="A test page"></head><body></body></html>`)
	m := Extract(d)
	if m.Description != "A test page" {
		t.Fatalf("description = %q", m.Description)
	}
}

func TestExtractKeywords(t *testing.T) {
	d := mustParse(t, `<html><head><meta name="keywords" content="go, html, parser"></head><body></body></html>`)
	m := Extract(d)
	if len(m.Keywords) != 3 || m.Keywords[0] != "go" {
		t.Fatalf("keywords = %v", m.Keywords)
	}
}

func TestExtractCharset(t *testing.T) {
	d := mustParse(t, `<html><head><meta charset="utf-8"></head><body></body></html>`)
	m := Extract(d)
	if m.Charset != "utf-8" {
		t.Fatalf("charset = %q", m.Charset)
	}
}

func TestExtractCanonical(t *testing.T) {
	d := mustParse(t, `<html><head><link rel="canonical" href="https://example.com/page"></head><body></body></html>`)
	m := Extract(d)
	if m.Canonical != "https://example.com/page" {
		t.Fatalf("canonical = %q", m.Canonical)
	}
}

func TestExtractOpenGraph(t *testing.T) {
	d := mustParse(t, `<html><head>
		<meta property="og:title" content="OG Title">
		<meta property="og:description" content="OG Desc">
	</head><body></body></html>`)
	m := Extract(d)
	if !m.HasOpenGraph() {
		t.Fatal("should have OG")
	}
	if m.OGTitle() != "OG Title" {
		t.Fatalf("og title = %q", m.OGTitle())
	}
	if m.OGDescription() != "OG Desc" {
		t.Fatalf("og desc = %q", m.OGDescription())
	}
}

func TestExtractLanguage(t *testing.T) {
	d := mustParse(t, `<html lang="en"><head></head><body></body></html>`)
	m := Extract(d)
	if m.Language != "en" {
		t.Fatalf("language = %q", m.Language)
	}
}

func TestExtractAuthorRobots(t *testing.T) {
	d := mustParse(t, `<html><head>
		<meta name="author" content="John">
		<meta name="robots" content="noindex,nofollow">
	</head><body></body></html>`)
	m := Extract(d)
	if m.Author != "John" {
		t.Fatalf("author = %q", m.Author)
	}
	if m.Robots != "noindex,nofollow" {
		t.Fatalf("robots = %q", m.Robots)
	}
}
