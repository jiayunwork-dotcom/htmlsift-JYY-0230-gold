// social.go extends htmlmeta with social media sharing metadata analysis.
package htmlmeta

import "strings"

// SocialPreview represents how a page would appear when shared on social media.
type SocialPreview struct {
	Title       string
	Description string
	Image       string
	URL         string
	SiteName    string
	Type        string
}

// SocialPreviewFromMeta builds a social sharing preview from extracted metadata.
func SocialPreviewFromMeta(m *Meta) *SocialPreview {
	sp := &SocialPreview{}
	// Prefer OG, fall back to Twitter, then to basic meta.
	sp.Title = coalesce(m.OpenGraph["og:title"], m.Twitter["twitter:title"], m.Title)
	sp.Description = coalesce(m.OpenGraph["og:description"], m.Twitter["twitter:description"], m.Description)
	sp.Image = coalesce(m.OpenGraph["og:image"], m.Twitter["twitter:image"])
	sp.URL = coalesce(m.OpenGraph["og:url"], m.Canonical)
	sp.SiteName = m.OpenGraph["og:site_name"]
	sp.Type = m.OpenGraph["og:type"]
	return sp
}

// IsComplete reports whether the preview has all fields for a rich card.
func (sp *SocialPreview) IsComplete() bool {
	return sp.Title != "" && sp.Description != "" && sp.Image != ""
}

// coalesce returns the first non-empty string from the arguments.
func coalesce(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// Recommendations returns a list of missing metadata fields for optimal
// social sharing.
func Recommendations(m *Meta) []string {
	var recs []string
	if m.Title == "" {
		recs = append(recs, "missing <title>")
	}
	if m.Description == "" {
		recs = append(recs, "missing meta description")
	}
	if !m.HasOpenGraph() {
		recs = append(recs, "no Open Graph tags (og:title, og:description, og:image)")
	} else {
		if m.OpenGraph["og:image"] == "" {
			recs = append(recs, "missing og:image")
		}
	}
	if m.Canonical == "" {
		recs = append(recs, "missing canonical URL")
	}
	if m.Language == "" {
		recs = append(recs, "missing html lang attribute")
	}
	return recs
}

// SEOScore returns a simple 0-100 score for basic SEO readiness.
func SEOScore(m *Meta) int {
	score := 0
	if m.Title != "" {
		score += 20
	}
	if m.Description != "" {
		score += 20
	}
	if m.Canonical != "" {
		score += 15
	}
	if m.HasOpenGraph() {
		score += 15
	}
	if m.Language != "" {
		score += 10
	}
	if len(m.Keywords) > 0 {
		score += 10
	}
	if m.Robots == "" || !strings.Contains(m.Robots, "noindex") {
		score += 10
	}
	if score > 100 {
		score = 100
	}
	return score
}
