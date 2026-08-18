// Package htmlparse is a thin, typed wrapper over golang.org/x/net/html.
// It parses full documents or fragments, walks the resulting node tree,
// collects statistics and extracts visible text. It exists so callers
// work with a stable API instead of raw html.Node plumbing.
package htmlparse

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Doc wraps a parsed HTML tree.
type Doc struct {
	Root *html.Node
}

// Stats holds aggregate facts about a parsed document.
type Stats struct {
	Elements   int            // total element nodes
	TextNodes  int            // total text nodes
	Comments   int            // total comment nodes
	MaxDepth   int            // deepest nesting level (document root = 0)
	Tags       map[string]int // element name -> occurrence count
	Links      int            // <a href> elements
	Images     int            // <img> elements
	TotalBytes int            // bytes of raw text content
}

// Parse parses a full HTML document. Input is read as UTF-8. A trailing
// parse error after partial content is reported, matching
// html.Parse's behavior for well-formed fragments.
func Parse(input string) (*Doc, error) {
	if strings.TrimSpace(input) == "" {
		return nil, fmt.Errorf("htmlparse: empty input")
	}
	root, err := html.Parse(strings.NewReader(input))
	if err != nil {
		return nil, fmt.Errorf("htmlparse: %w", err)
	}
	return &Doc{Root: root}, nil
}

// ParseFragment parses a fragment of HTML in the context of a given
// context element name (e.g. "body", "table"). The context element
// influences how implied tags are inserted.
func ParseFragment(input, context string) ([]*html.Node, error) {
	if context == "" {
		context = "body"
	}
	ctx := &html.Node{
		Type:     html.ElementNode,
		Data:     context,
		DataAtom: atom.Lookup([]byte(context)),
	}
	nodes, err := html.ParseFragment(strings.NewReader(input), ctx)
	if err != nil {
		return nil, fmt.Errorf("htmlparse: %w", err)
	}
	return nodes, nil
}

// Render serializes the document back to HTML.
func (d *Doc) Render() (string, error) {
	if d == nil || d.Root == nil {
		return "", fmt.Errorf("htmlparse: nil document")
	}
	var buf bytes.Buffer
	if err := html.Render(&buf, d.Root); err != nil {
		return "", fmt.Errorf("htmlparse: render: %w", err)
	}
	return buf.String(), nil
}

// RenderNode serializes a single node (and its subtree).
func RenderNode(n *html.Node) (string, error) {
	if n == nil {
		return "", fmt.Errorf("htmlparse: nil node")
	}
	var buf bytes.Buffer
	if err := html.Render(&buf, n); err != nil {
		return "", fmt.Errorf("htmlparse: render node: %w", err)
	}
	return buf.String(), nil
}

// Walk performs a pre-order traversal of the tree rooted at n, calling
// fn for every node. The traversal stops early if fn returns an error.
func Walk(n *html.Node, fn func(*html.Node) error) error {
	if n == nil {
		return nil
	}
	if err := fn(n); err != nil {
		return err
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if err := Walk(c, fn); err != nil {
			return err
		}
	}
	return nil
}

// (d *Doc) Walk walks the whole document tree.
func (d *Doc) Walk(fn func(*html.Node) error) error {
	if d == nil || d.Root == nil {
		return fmt.Errorf("htmlparse: nil document")
	}
	return Walk(d.Root, fn)
}

// Stats computes aggregate statistics for the document.
func (d *Doc) Stats() Stats {
	st := Stats{Tags: map[string]int{}}
	if d == nil || d.Root == nil {
		return st
	}
	depth := 0
	_ = Walk(d.Root, func(n *html.Node) error {
		switch n.Type {
		case html.ElementNode:
			st.Elements++
			st.Tags[n.Data]++
			if n.Data == "a" && hasAttr(n, "href") {
				st.Links++
			}
			if n.Data == "img" {
				st.Images++
			}
		case html.TextNode:
			st.TextNodes++
			st.TotalBytes += len(n.Data)
		case html.CommentNode:
			st.Comments++
		}
		if depth < nodeDepth(n) {
			depth = nodeDepth(n)
		}
		return nil
	})
	st.MaxDepth = depth
	return st
}

func nodeDepth(n *html.Node) int {
	d := 0
	for p := n.Parent; p != nil; p = p.Parent {
		d++
	}
	return d
}

func hasAttr(n *html.Node, key string) bool {
	for _, a := range n.Attr {
		if a.Key == key {
			return true
		}
	}
	return false
}

// Attr returns the value of an attribute, or "" when absent.
func Attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

// FindByTag returns all element nodes with the given tag name (in
// document order).
func FindByTag(d *Doc, tag string) []*html.Node {
	var out []*html.Node
	if d == nil || d.Root == nil {
		return out
	}
	_ = Walk(d.Root, func(n *html.Node) error {
		if n.Type == html.ElementNode && n.Data == tag {
			out = append(out, n)
		}
		return nil
	})
	return out
}

// HiddenTags are element names whose text is not rendered visually.
var HiddenTags = map[string]bool{
	"script":   true,
	"style":    true,
	"noscript": true,
	"template": true,
	"head":     true,
}

// VisibleText extracts the text that a browser would visually render:
// it concatenates text nodes while skipping script/style/head content
// and then normalizes whitespace runs to single spaces.
func VisibleText(d *Doc) string {
	var sb strings.Builder
	if d != nil && d.Root != nil {
		_ = Walk(d.Root, func(n *html.Node) error {
			if n.Type == html.TextNode {
				sb.WriteString(n.Data)
				sb.WriteByte(' ')
			}
			return nil
		})
	}
	return collapseSpace(sb.String())
}

// CollapseSpace is exposed for callers that want the same whitespace
// normalization on their own strings.
func CollapseSpace(s string) string {
	return collapseSpace(s)
}

func collapseSpace(s string) string {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return ""
	}
	return strings.Join(fields, " ")
}

// ReadAll is a convenience for tests and the CLI: it reads an io.Reader
// fully into a string.
func ReadAll(r io.Reader) (string, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("htmlparse: read: %w", err)
	}
	return string(b), nil
}
