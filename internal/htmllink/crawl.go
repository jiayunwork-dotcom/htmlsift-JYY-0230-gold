// crawl.go extends htmllink with link analysis utilities: broken link detection
// helpers, link graph construction, and outbound link counting.
package htmllink

import (
	"net/url"
	"strings"
)

// LinkGraph represents a directed graph of links between pages.
type LinkGraph struct {
	Nodes map[string]bool         // set of unique URLs
	Edges map[string][]string     // source -> destinations
}

// NewLinkGraph creates an empty link graph.
func NewLinkGraph() *LinkGraph {
	return &LinkGraph{
		Nodes: make(map[string]bool),
		Edges: make(map[string][]string),
	}
}

// AddPage adds a page and its outbound links to the graph.
func (g *LinkGraph) AddPage(pageURL string, links []Link) {
	g.Nodes[pageURL] = true
	for _, l := range links {
		target := l.Resolved
		if target == "" {
			target = l.Href
		}
		g.Nodes[target] = true
		g.Edges[pageURL] = append(g.Edges[pageURL], target)
	}
}

// OutDegree returns the number of outbound links from a page.
func (g *LinkGraph) OutDegree(pageURL string) int {
	return len(g.Edges[pageURL])
}

// InDegree returns the number of pages linking to target.
func (g *LinkGraph) InDegree(target string) int {
	count := 0
	for _, dests := range g.Edges {
		for _, d := range dests {
			if d == target {
				count++
			}
		}
	}
	return count
}

// NumNodes returns the total number of unique URLs.
func (g *LinkGraph) NumNodes() int { return len(g.Nodes) }

// NumEdges returns the total number of edges.
func (g *LinkGraph) NumEdges() int {
	n := 0
	for _, dests := range g.Edges {
		n += len(dests)
	}
	return n
}

// ExternalLinks returns links that point to a different host than base.
func ExternalLinks(links []Link, baseHost string) []Link {
	lowerBase := strings.ToLower(baseHost)
	var out []Link
	for _, l := range links {
		target := l.Resolved
		if target == "" {
			target = l.Href
		}
		u, err := url.Parse(target)
		if err != nil || u.Host == "" {
			continue
		}
		if strings.ToLower(u.Host) != lowerBase {
			out = append(out, l)
		}
	}
	return out
}

// InternalLinks returns links that point to the same host as base.
func InternalLinks(links []Link, baseHost string) []Link {
	lowerBase := strings.ToLower(baseHost)
	var out []Link
	for _, l := range links {
		target := l.Resolved
		if target == "" {
			target = l.Href
		}
		u, err := url.Parse(target)
		if err != nil {
			continue
		}
		host := strings.ToLower(u.Host)
		if host == lowerBase || host == "" {
			out = append(out, l)
		}
	}
	return out
}

// BrokenLinkCandidate identifies links that look potentially broken:
// empty href, javascript: scheme, or data: that's not an image.
func BrokenLinkCandidates(links []Link) []Link {
	var out []Link
	for _, l := range links {
		c := Classify(l.Href)
		if c == ClassJavaScript || c == ClassData || c == ClassUnknown {
			out = append(out, l)
		}
	}
	return out
}

// CountByTag groups links by their source tag and returns counts.
func CountByTag(links []Link) map[string]int {
	counts := make(map[string]int)
	for _, l := range links {
		counts[l.Tag]++
	}
	return counts
}

// HasDuplicates reports whether any two links point to the same resolved URL.
func HasDuplicates(links []Link) bool {
	seen := make(map[string]bool)
	for _, l := range links {
		key := l.Resolved
		if key == "" {
			key = l.Href
		}
		if seen[key] {
			return true
		}
		seen[key] = true
	}
	return false
}
