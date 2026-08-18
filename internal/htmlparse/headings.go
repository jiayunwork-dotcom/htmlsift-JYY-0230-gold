// headings.go provides heading-level analysis: outline extraction, heading
// hierarchy validation, and table-of-contents generation.
package htmlparse

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

// Heading represents an extracted heading from the document.
type Heading struct {
	Level int    // 1-6
	Text  string // visible text content
	ID    string // id attribute if present
}

// ExtractHeadings returns all h1-h6 elements in document order.
func ExtractHeadings(d *Doc) []Heading {
	if d == nil || d.Root == nil {
		return nil
	}
	var headings []Heading
	_ = Walk(d.Root, func(n *html.Node) error {
		if n.Type != html.ElementNode {
			return nil
		}
		level := headingLevel(n.Data)
		if level == 0 {
			return nil
		}
		headings = append(headings, Heading{
			Level: level,
			Text:  InnerText(n),
			ID:    Attr(n, "id"),
		})
		return nil
	})
	return headings
}

// headingLevel returns the heading level (1-6) for a tag name, or 0 if not.
func headingLevel(tag string) int {
	switch tag {
	case "h1":
		return 1
	case "h2":
		return 2
	case "h3":
		return 3
	case "h4":
		return 4
	case "h5":
		return 5
	case "h6":
		return 6
	}
	return 0
}

// TOCEntry is a single entry in a table of contents.
type TOCEntry struct {
	Level    int
	Text     string
	Anchor   string // #id link target
	Children []TOCEntry
}

// GenerateTOC builds a nested table of contents from headings.
func GenerateTOC(headings []Heading) []TOCEntry {
	var toc []TOCEntry
	var stack []*[]TOCEntry
	stack = append(stack, &toc)
	prevLevel := 0

	for _, h := range headings {
		entry := TOCEntry{
			Level:  h.Level,
			Text:   h.Text,
			Anchor: h.ID,
		}
		if h.Level > prevLevel && prevLevel > 0 && len(*stack[len(stack)-1]) > 0 {
			parent := &(*stack[len(stack)-1])[len(*stack[len(stack)-1])-1].Children
			stack = append(stack, parent)
		} else if h.Level < prevLevel {
			for len(stack) > 1 && h.Level <= prevLevel {
				stack = stack[:len(stack)-1]
				prevLevel--
			}
		}
		*stack[len(stack)-1] = append(*stack[len(stack)-1], entry)
		prevLevel = h.Level
	}
	return toc
}

// ValidateHeadingHierarchy checks that headings don't skip levels (e.g. h1→h3
// without h2). Returns a list of warnings.
func ValidateHeadingHierarchy(headings []Heading) []string {
	if len(headings) == 0 {
		return nil
	}
	var warnings []string
	if headings[0].Level != 1 {
		warnings = append(warnings, fmt.Sprintf("first heading is h%d, expected h1", headings[0].Level))
	}
	for i := 1; i < len(headings); i++ {
		diff := headings[i].Level - headings[i-1].Level
		if diff > 1 {
			warnings = append(warnings, fmt.Sprintf("heading level skipped: h%d → h%d at %q",
				headings[i-1].Level, headings[i].Level, headings[i].Text))
		}
	}
	return warnings
}

// HeadingCount returns the count of headings at each level.
func HeadingCount(headings []Heading) map[int]int {
	counts := make(map[int]int)
	for _, h := range headings {
		counts[h.Level]++
	}
	return counts
}

// HeadingsAsText formats headings as indented text outline.
func HeadingsAsText(headings []Heading) string {
	var sb strings.Builder
	for _, h := range headings {
		indent := strings.Repeat("  ", h.Level-1)
		sb.WriteString(indent)
		sb.WriteString(h.Text)
		sb.WriteByte('\n')
	}
	return sb.String()
}
