// images.go provides image extraction and analysis from HTML documents.
package htmlparse

import (
	"strings"

	"golang.org/x/net/html"
)

// Image represents an extracted <img> element.
type Image struct {
	Src     string
	Alt     string
	Width   string
	Height  string
	Loading string // lazy, eager, or empty
	Classes []string
}

// ExtractImages returns all <img> elements from the document.
func ExtractImages(d *Doc) []Image {
	if d == nil || d.Root == nil {
		return nil
	}
	var images []Image
	_ = Walk(d.Root, func(n *html.Node) error {
		if n.Type != html.ElementNode || n.Data != "img" {
			return nil
		}
		img := Image{
			Src:     Attr(n, "src"),
			Alt:     Attr(n, "alt"),
			Width:   Attr(n, "width"),
			Height:  Attr(n, "height"),
			Loading: Attr(n, "loading"),
		}
		if cls := Attr(n, "class"); cls != "" {
			img.Classes = strings.Fields(cls)
		}
		images = append(images, img)
		return nil
	})
	return images
}

// MissingAlt returns images that lack alt text (accessibility issue).
func MissingAlt(images []Image) []Image {
	var out []Image
	for _, img := range images {
		if strings.TrimSpace(img.Alt) == "" {
			out = append(out, img)
		}
	}
	return out
}

// ImagesBySrc deduplicates images by their src attribute.
func ImagesBySrc(images []Image) map[string]Image {
	m := make(map[string]Image)
	for _, img := range images {
		if img.Src != "" {
			m[img.Src] = img
		}
	}
	return m
}

// LazyLoadedCount returns the number of images with loading="lazy".
func LazyLoadedCount(images []Image) int {
	n := 0
	for _, img := range images {
		if strings.ToLower(img.Loading) == "lazy" {
			n++
		}
	}
	return n
}

// ExternalImages returns images whose src starts with http:// or https://.
func ExternalImages(images []Image) []Image {
	var out []Image
	for _, img := range images {
		lower := strings.ToLower(img.Src)
		if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
			out = append(out, img)
		}
	}
	return out
}

// DataImages returns images whose src is a data: URI.
func DataImages(images []Image) []Image {
	var out []Image
	for _, img := range images {
		if strings.HasPrefix(strings.ToLower(img.Src), "data:") {
			out = append(out, img)
		}
	}
	return out
}
