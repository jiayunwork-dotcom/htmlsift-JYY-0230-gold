// escape.go provides HTML entity encoding/decoding utilities and additional
// text sanitization functions for safe output rendering.
package htmlsanitize

import (
	"strings"
	"unicode/utf8"
)

// EscapeHTML escapes special HTML characters to their entity equivalents.
func EscapeHTML(s string) string {
	var sb strings.Builder
	sb.Grow(len(s))
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '&':
			sb.WriteString("&amp;")
		case '<':
			sb.WriteString("&lt;")
		case '>':
			sb.WriteString("&gt;")
		case '"':
			sb.WriteString("&quot;")
		case '\'':
			sb.WriteString("&#39;")
		default:
			sb.WriteByte(s[i])
		}
	}
	return sb.String()
}

// UnescapeHTML reverses basic HTML entity escaping.
func UnescapeHTML(s string) string {
	r := strings.NewReplacer(
		"&amp;", "&",
		"&lt;", "<",
		"&gt;", ">",
		"&quot;", `"`,
		"&#39;", "'",
		"&#x27;", "'",
		"&#34;", `"`,
	)
	return r.Replace(s)
}

// StripTags removes all HTML tags from a string, leaving only text content.
func StripTags(s string) string {
	var sb strings.Builder
	inTag := false
	for i := 0; i < len(s); i++ {
		if s[i] == '<' {
			inTag = true
			continue
		}
		if s[i] == '>' {
			inTag = false
			continue
		}
		if !inTag {
			sb.WriteByte(s[i])
		}
	}
	return sb.String()
}

// TruncateText truncates s to maxLen runes, appending "..." if truncated.
func TruncateText(s string, maxLen int) string {
	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}
	runes := []rune(s)
	if maxLen < 3 {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-3]) + "..."
}

// NormalizeWhitespace collapses all whitespace runs to single spaces and trims.
func NormalizeWhitespace(s string) string {
	fields := strings.Fields(s)
	return strings.Join(fields, " ")
}

// RemoveControlChars strips ASCII control characters (0x00-0x1F except
// tab/newline/carriage return) from a string.
func RemoveControlChars(s string) string {
	var sb strings.Builder
	for _, r := range s {
		if r < 0x20 && r != '\t' && r != '\n' && r != '\r' {
			continue
		}
		sb.WriteRune(r)
	}
	return sb.String()
}

// SafeAttrValue ensures an attribute value doesn't contain characters that
// could break out of a quoted attribute context.
func SafeAttrValue(val string) string {
	val = strings.ReplaceAll(val, `"`, "&quot;")
	val = strings.ReplaceAll(val, `'`, "&#39;")
	val = strings.ReplaceAll(val, "<", "&lt;")
	val = strings.ReplaceAll(val, ">", "&gt;")
	return val
}

// IsASCII reports whether all characters in s are ASCII (< 128).
func IsASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= 128 {
			return false
		}
	}
	return true
}

// CountNonASCII returns the number of non-ASCII characters.
func CountNonASCII(s string) int {
	n := 0
	for _, r := range s {
		if r >= 128 {
			n++
		}
	}
	return n
}
