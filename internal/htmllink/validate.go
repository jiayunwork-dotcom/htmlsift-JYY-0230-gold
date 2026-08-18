// validate.go extends htmllink with URL validation utilities.
package htmllink

import (
	"net/url"
	"strings"
)

// ValidationResult holds the outcome of link validation.
type ValidationResult struct {
	Valid   bool
	URL     string
	Issues  []string
}

// ValidateURL checks a URL string for common issues.
func ValidateURL(rawURL string) ValidationResult {
	res := ValidationResult{URL: rawURL, Valid: true}
	if strings.TrimSpace(rawURL) == "" {
		res.Valid = false
		res.Issues = append(res.Issues, "empty URL")
		return res
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		res.Valid = false
		res.Issues = append(res.Issues, "parse error: "+err.Error())
		return res
	}
	// Check for suspicious patterns.
	if strings.Contains(rawURL, " ") {
		res.Issues = append(res.Issues, "contains spaces")
	}
	if u.Scheme != "" && u.Host == "" && u.Scheme != "mailto" && u.Scheme != "tel" && u.Scheme != "data" && u.Scheme != "javascript" {
		res.Issues = append(res.Issues, "scheme without host")
	}
	if strings.HasSuffix(u.Host, ".") {
		res.Issues = append(res.Issues, "trailing dot in host")
	}
	if len(res.Issues) > 0 {
		res.Valid = false
	}
	return res
}

// ValidateLinks validates all links and returns results.
func ValidateLinks(links []Link) []ValidationResult {
	results := make([]ValidationResult, len(links))
	for i, l := range links {
		target := l.Resolved
		if target == "" {
			target = l.Href
		}
		results[i] = ValidateURL(target)
	}
	return results
}

// NormalizeURL normalizes a URL by lowercasing the scheme and host,
// removing default ports, and sorting query parameters.
func NormalizeURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)
	// Remove default ports.
	if (u.Scheme == "http" && u.Port() == "80") || (u.Scheme == "https" && u.Port() == "443") {
		u.Host = u.Hostname()
	}
	// Remove trailing slash from path if it's just "/".
	if u.Path == "/" && u.RawQuery == "" && u.Fragment == "" {
		u.Path = ""
	}
	return u.String()
}

// ExtractDomain returns just the domain/host portion of a URL.
func ExtractDomain(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return u.Hostname()
}

// IsSecure reports whether the URL uses HTTPS.
func IsSecure(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return strings.ToLower(u.Scheme) == "https"
}

// CountSecureLinks returns the number of links using HTTPS.
func CountSecureLinks(links []Link) int {
	n := 0
	for _, l := range links {
		target := l.Resolved
		if target == "" {
			target = l.Href
		}
		if IsSecure(target) {
			n++
		}
	}
	return n
}
