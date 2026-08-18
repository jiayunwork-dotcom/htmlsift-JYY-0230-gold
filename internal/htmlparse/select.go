package htmlparse

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

// Selector is a parsed, CSS-lite selector: an optional tag name
// combined with optional #id and .class constraints. Supported forms:
//
//	div            any <div>
//	#main          any element with id="main"
//	.note          any element with class containing "note"
//	div#main       a <div> with id="main"
//	div.note.a     a <div> whose class list contains both note and a
//	*              any element
//
// Attribute-level selectors and combinators are not supported.
type Selector struct {
	Tag     string // "" or "*" means any tag
	ID      string
	Classes []string
}

// ParseSelector parses a selector expression. An empty or "*" tag is
// kept as "". Any error wraps ErrBadSelector.
var ErrBadSelector = fmt.Errorf("htmlparse: bad selector")

func ParseSelector(s string) (Selector, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Selector{}, fmt.Errorf("%w: empty selector", ErrBadSelector)
	}
	if s == "*" {
		return Selector{}, nil
	}
	sel := Selector{}
	// Tokenize on # and . boundaries.
	i := 0
	for i < len(s) {
		c := s[i]
		switch {
		case c == '#':
			j := scanIdent(s, i+1)
			if j == i+1 {
				return Selector{}, fmt.Errorf("%w: missing id in %q", ErrBadSelector, s)
			}
			if sel.ID != "" {
				return Selector{}, fmt.Errorf("%w: multiple ids in %q", ErrBadSelector, s)
			}
			sel.ID = s[i+1 : j]
			i = j
		case c == '.':
			j := scanIdent(s, i+1)
			if j == i+1 {
				return Selector{}, fmt.Errorf("%w: missing class in %q", ErrBadSelector, s)
			}
			sel.Classes = append(sel.Classes, s[i+1:j])
			i = j
		default:
			j := scanIdent(s, i)
			if j == i {
				return Selector{}, fmt.Errorf("%w: unexpected %q in %q", ErrBadSelector, string(c), s)
			}
			if sel.Tag != "" {
				return Selector{}, fmt.Errorf("%w: multiple tags in %q", ErrBadSelector, s)
			}
			tag := strings.ToLower(s[i:j])
			if tag != "*" {
				sel.Tag = tag
			}
			i = j
		}
	}
	return sel, nil
}

// scanIdent consumes [a-zA-Z0-9_-] and returns the end index.
func scanIdent(s string, start int) int {
	i := start
	for i < len(s) {
		c := s[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '-' || c == '_' {
			i++
			continue
		}
		break
	}
	return i
}

// Match reports whether n (an element) satisfies the selector.
func (s Selector) Match(n *html.Node) bool {
	if n.Type != html.ElementNode {
		return false
	}
	if s.Tag != "" && n.Data != s.Tag {
		return false
	}
	if s.ID != "" && Attr(n, "id") != s.ID {
		return false
	}
	if len(s.Classes) > 0 {
		got := classSet(n)
		for _, c := range s.Classes {
			if !got[c] {
				return false
			}
		}
	}
	return true
}

func classSet(n *html.Node) map[string]bool {
	out := map[string]bool{}
	for _, a := range n.Attr {
		if a.Key == "class" {
			for _, c := range strings.Fields(a.Val) {
				out[c] = true
			}
		}
	}
	return out
}

// Select returns all elements in the document matching the selector,
// in document order.
func Select(d *Doc, expr string) ([]*html.Node, error) {
	sel, err := ParseSelector(expr)
	if err != nil {
		return nil, err
	}
	var out []*html.Node
	if d == nil || d.Root == nil {
		return out, nil
	}
	_ = Walk(d.Root, func(n *html.Node) error {
		if sel.Match(n) {
			out = append(out, n)
		}
		return nil
	})
	return out, nil
}

// ByID is a convenience wrapper for Select(d, "#"+id).
func ByID(d *Doc, id string) (*html.Node, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: empty id", ErrBadSelector)
	}
	nodes, err := Select(d, "#"+id)
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, nil
	}
	return nodes[0], nil
}
