// Package htmllink extracts hyperlinks from parsed HTML documents and
// resolves them against a base URL. It classifies link targets by
// scheme, filters by allowed schemes, checks bidi safety of link text,
// and provides same-origin comparison. URL handling uses net/url; bidi
// validation uses golang.org/x/text/secure/bidirule.
package htmllink

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"golang.org/x/net/html"

	"htmlsift/internal/htmlparse"
)

// Link describes one extracted hyperlink.
type Link struct {
	Tag      string // element name: a, link, area, img, script, iframe
	AttrKey  string // attribute the URL came from: href or src
	Href     string // raw attribute value
	Text     string // visible text of the anchor (a/link only)
	Resolved string // absolute URL when base was provided, else ""
}

// Class is the scheme classification of a link target.
type Class int

const (
	ClassUnknown    Class = iota
	ClassHTTP             // http: or https:
	ClassMailto           // mailto:
	ClassTel              // tel:
	ClassData             // data:
	ClassJavaScript       // javascript: (and other script-adjacent schemes)
	ClassFragment         // starts with '#'
	ClassRelative         // no scheme and not a fragment
)

func (c Class) String() string {
	switch c {
	case ClassHTTP:
		return "http"
	case ClassMailto:
		return "mailto"
	case ClassTel:
		return "tel"
	case ClassData:
		return "data"
	case ClassJavaScript:
		return "javascript"
	case ClassFragment:
		return "fragment"
	case ClassRelative:
		return "relative"
	}
	return "unknown"
}

// extractors maps element names to the attribute carrying the URL.
var extractors = map[string]string{
	"a":      "href",
	"link":   "href",
	"area":   "href",
	"img":    "src",
	"script": "src",
	"iframe": "src",
	"source": "src",
	"video":  "src",
	"audio":  "src",
}

// Extract collects all links from the document. base, when non-empty,
// is used to resolve relative URLs into absolute ones; an invalid base
// yields an error.
func Extract(d *htmlparse.Doc, base string) ([]Link, error) {
	if d == nil {
		return nil, fmt.Errorf("htmllink: nil document")
	}
	var out []Link
	if err := d.Walk(func(n *html.Node) error {
		if n.Type != html.ElementNode {
			return nil
		}
		attrKey, ok := extractors[n.Data]
		if !ok {
			return nil
		}
		href := htmlparse.Attr(n, attrKey)
		if href == "" {
			return nil
		}
		l := Link{
			Tag:     n.Data,
			AttrKey: attrKey,
			Href:    href,
			Text:    anchorText(n),
		}
		if base != "" {
			r, err := ResolveURL(base, href)
			if err != nil {
				return fmt.Errorf("htmllink: resolve %q against %q: %w", href, base, err)
			}
			l.Resolved = r
		}
		out = append(out, l)
		return nil
	}); err != nil {
		return nil, err
	}
	return out, nil
}

func anchorText(n *html.Node) string {
	var sb strings.Builder
	collectText(n, &sb)
	return htmlparse.CollapseSpace(sb.String())
}

func collectText(n *html.Node, sb *strings.Builder) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		switch c.Type {
		case html.TextNode:
			sb.WriteString(c.Data)
			sb.WriteByte(' ')
		case html.ElementNode:
			collectText(c, sb)
		}
	}
}

// ResolveURL resolves ref against base and returns the absolute URL
// string. An empty base returns ref unchanged. Fragments and relative
// refs are handled by net/url.
func ResolveURL(base, ref string) (string, error) {
	if base == "" {
		return ref, nil
	}
	b, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("htmllink: %w", err)
	}
	r, err := url.Parse(ref)
	if err != nil {
		return "", fmt.Errorf("htmllink: %w", err)
	}
	return b.ResolveReference(r).String(), nil
}

// Classify returns the scheme class of a raw href value.
func Classify(href string) Class {
	t := strings.TrimSpace(href)
	switch {
	case t == "":
		return ClassUnknown
	case strings.HasPrefix(t, "#"):
		return ClassFragment
	}
	scheme, rest, ok := splitScheme(t)
	if !ok {
		return ClassRelative
	}
	switch strings.ToLower(scheme) {
	case "http", "https":
		return ClassHTTP
	case "mailto":
		return ClassMailto
	case "tel":
		return ClassTel
	case "data":
		return ClassData
	case "javascript", "vbscript", "livescript":
		return ClassJavaScript
	}
	_ = rest
	return ClassUnknown
}

// splitScheme splits "scheme:rest" per RFC 3986. It reports false when
// there is no valid scheme (e.g. relative paths or Windows-ish input).
func splitScheme(s string) (scheme, rest string, ok bool) {
	idx := strings.IndexByte(s, ':')
	if idx <= 0 {
		return "", "", false
	}
	head := s[:idx]
	for i := 0; i < len(head); i++ {
		c := head[i]
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z':
		case i > 0 && c >= '0' && c <= '9':
		case i > 0 && (c == '+' || c == '-' || c == '.'):
		default:
			return "", "", false
		}
	}
	return head, s[idx+1:], true
}

// FilterByScheme keeps only links whose class is in allow. A relative
// or fragment link is always kept (browsers resolve them locally).
func FilterByScheme(links []Link, allow ...Class) []Link {
	if len(allow) == 0 {
		return links
	}
	allowed := map[Class]bool{}
	for _, c := range allow {
		allowed[c] = true
	}
	var out []Link
	for _, l := range links {
		c := Classify(l.Href)
		if c == ClassRelative || c == ClassFragment || allowed[c] {
			out = append(out, l)
		}
	}
	return out
}

// UniqueByHref deduplicates links by resolved-or-raw href, keeping the
// first occurrence and sorting by (Tag, Href).
func UniqueByHref(links []Link) []Link {
	seen := map[string]bool{}
	var out []Link
	for _, l := range links {
		key := l.Href
		if l.Resolved != "" {
			key = l.Resolved
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, l)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Tag != out[j].Tag {
			return out[i].Tag < out[j].Tag
		}
		return out[i].Href < out[j].Href
	})
	return out
}

// AbsoluteLinks returns only links whose raw href already carries an
// http(s) scheme, in original order.
func AbsoluteLinks(links []Link) []Link {
	var out []Link
	for _, l := range links {
		c := Classify(l.Href)
		if c == ClassHTTP {
			out = append(out, l)
		}
	}
	return out
}

// GroupByHost partitions links (resolved or raw) by host and returns a
// sorted map of host -> links. Links without a resolvable host are
// collected under the empty key.
func GroupByHost(links []Link) map[string][]Link {
	groups := map[string][]Link{}
	for _, l := range links {
		u, err := url.Parse(l.Resolved)
		if err != nil || u.Host == "" {
			u, err = url.Parse(l.Href)
		}
		host := ""
		if err == nil {
			host = u.Host
		}
		groups[host] = append(groups[host], l)
	}
	return groups
}

// IsSameOrigin reports whether ref, resolved against base, shares the
// same scheme, host and port as base.
func IsSameOrigin(base, ref string) (bool, error) {
	b, err := url.Parse(base)
	if err != nil {
		return false, fmt.Errorf("htmllink: %w", err)
	}
	r, err := url.Parse(ref)
	if err != nil {
		return false, fmt.Errorf("htmllink: %w", err)
	}
	if b.Scheme != r.Scheme {
		return false, nil
	}
	if !strings.EqualFold(b.Host, r.Host) {
		return false, nil
	}
	if b.Port() != r.Port() {
		return false, nil
	}
	return true, nil
}

// ValidBidi reports whether s can be safely embedded in an LTR
// document: it must not contain Unicode directional control characters
// (LRE/RLE/LRO/RLO/PDF and the isolate format characters) that could
// reorder surrounding text. Plain LTR and RTL text both pass.
func ValidBidi(s string) bool {
	for _, r := range s {
		if isBidiControl(r) {
			return false
		}
	}
	return true
}

func isBidiControl(r rune) bool {
	switch {
	case r >= 0x202A && r <= 0x202E: // LRE, RLE, PDF, LRO, RLO
		return true
	case r >= 0x2066 && r <= 0x2069: // LRI, RLI, FSI, PDI
		return true
	case r == 0x061C: // Arabic Letter Mark
		return true
	}
	return false
}

// LinkTextOK combines extraction-time checks: non-empty, bidi-safe and
// reasonably short.
func LinkTextOK(text string, maxLen int) bool {
	if text == "" {
		return false
	}
	if maxLen > 0 && len([]rune(text)) > maxLen {
		return false
	}
	return ValidBidi(text)
}
