package htmlparse

import (
	"strings"

	"golang.org/x/net/html"
)

// RemoveByTag removes all elements with the given tag from the document tree.
// Their children are discarded.
func RemoveByTag(d *Doc, tag string) int {
	if d == nil || d.Root == nil {
		return 0
	}
	removed := 0
	removeByTagRec(d.Root, tag, &removed)
	return removed
}

func removeByTagRec(n *html.Node, tag string, count *int) {
	for c := n.FirstChild; c != nil; {
		next := c.NextSibling
		if c.Type == html.ElementNode && c.Data == tag {
			n.RemoveChild(c)
			*count++
		} else {
			removeByTagRec(c, tag, count)
		}
		c = next
	}
}

// RemoveByID removes the element with the given id from the tree.
func RemoveByID(d *Doc, id string) bool {
	if d == nil || d.Root == nil {
		return false
	}
	var found *html.Node
	_ = Walk(d.Root, func(n *html.Node) error {
		if n.Type == html.ElementNode && Attr(n, "id") == id {
			found = n
		}
		return nil
	})
	if found != nil && found.Parent != nil {
		found.Parent.RemoveChild(found)
		return true
	}
	return false
}

// AddClass adds a class to an element node.
func AddClass(n *html.Node, class string) {
	if n == nil || n.Type != html.ElementNode {
		return
	}
	existing := Attr(n, "class")
	if existing == "" {
		SetAttr(n, "class", class)
		return
	}
	if strings.Contains(" "+existing+" ", " "+class+" ") {
		return
	}
	SetAttr(n, "class", existing+" "+class)
}

// RemoveClass removes a class from an element node.
func RemoveClass(n *html.Node, class string) {
	if n == nil || n.Type != html.ElementNode {
		return
	}
	existing := Attr(n, "class")
	if existing == "" {
		return
	}
	fields := strings.Fields(existing)
	var kept []string
	for _, f := range fields {
		if f != class {
			kept = append(kept, f)
		}
	}
	if len(kept) == 0 {
		RemoveAttr(n, "class")
	} else {
		SetAttr(n, "class", strings.Join(kept, " "))
	}
}

// SetAttr sets (or adds) an attribute on an element node.
func SetAttr(n *html.Node, key, val string) {
	if n == nil {
		return
	}
	for i := range n.Attr {
		if n.Attr[i].Key == key {
			n.Attr[i].Val = val
			return
		}
	}
	n.Attr = append(n.Attr, html.Attribute{Key: key, Val: val})
}

// RemoveAttr removes an attribute from an element node.
func RemoveAttr(n *html.Node, key string) {
	if n == nil {
		return
	}
	kept := n.Attr[:0]
	for _, a := range n.Attr {
		if a.Key != key {
			kept = append(kept, a)
		}
	}
	n.Attr = kept
}

// HasClass reports whether the node has the given class.
func HasClass(n *html.Node, class string) bool {
	if n == nil || n.Type != html.ElementNode {
		return false
	}
	existing := Attr(n, "class")
	for _, f := range strings.Fields(existing) {
		if f == class {
			return true
		}
	}
	return false
}

// Children returns the direct child element nodes of n.
func Children(n *html.Node) []*html.Node {
	var out []*html.Node
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode {
			out = append(out, c)
		}
	}
	return out
}

// InnerText returns the concatenated text content of a node (excluding
// children's structure), with whitespace collapsed.
func InnerText(n *html.Node) string {
	var sb strings.Builder
	collectInnerText(n, &sb)
	return CollapseSpace(sb.String())
}

func collectInnerText(n *html.Node, sb *strings.Builder) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			sb.WriteString(c.Data)
			sb.WriteByte(' ')
		} else if c.Type == html.ElementNode {
			collectInnerText(c, sb)
		}
	}
}
