// rules.go extends the sanitize package with predefined security rule sets
// and additional policy builders for common use cases.
package htmlsanitize

import "strings"

// StrictPolicy returns a very restrictive policy: only basic text formatting
// elements, no links, no images.
func StrictPolicy() Policy {
	el := map[string]bool{
		"p": true, "br": true, "hr": true,
		"b": true, "i": true, "u": true, "em": true, "strong": true,
		"ul": true, "ol": true, "li": true,
		"blockquote": true, "pre": true, "code": true,
	}
	return Policy{
		Elements:      el,
		Attrs:         map[string]map[string]bool{},
		Schemes:       map[string]bool{},
		StripComments: true,
	}
}

// PermissivePolicy returns a more permissive policy: all default elements
// plus media elements, without rel=nofollow enforcement.
func PermissivePolicy() Policy {
	p := DefaultPolicy()
	p.RequireRelNofollow = false
	p.Elements["video"] = true
	p.Elements["audio"] = true
	p.Elements["source"] = true
	p.Elements["iframe"] = true
	p.Attrs["iframe"] = map[string]bool{
		"src": true, "width": true, "height": true,
		"frameborder": true, "allowfullscreen": true,
	}
	p.Attrs["video"] = map[string]bool{"src": true, "controls": true, "width": true, "height": true}
	p.Attrs["audio"] = map[string]bool{"src": true, "controls": true}
	return p
}

// TextOnlyPolicy strips everything except text content. No elements survive.
func TextOnlyPolicy() Policy {
	return Policy{
		Elements:      map[string]bool{},
		Attrs:         map[string]map[string]bool{},
		Schemes:       map[string]bool{},
		StripComments: true,
	}
}

// PolicyBuilder helps construct custom policies fluently.
type PolicyBuilder struct {
	p Policy
}

// NewPolicyBuilder starts building a policy from scratch.
func NewPolicyBuilder() *PolicyBuilder {
	return &PolicyBuilder{p: Policy{
		Elements: make(map[string]bool),
		Attrs:    make(map[string]map[string]bool),
		Schemes:  make(map[string]bool),
	}}
}

// AllowElements adds elements to the allowlist.
func (b *PolicyBuilder) AllowElements(tags ...string) *PolicyBuilder {
	for _, t := range tags {
		b.p.Elements[strings.ToLower(t)] = true
	}
	return b
}

// AllowAttrs adds attributes for a specific element.
func (b *PolicyBuilder) AllowAttrs(element string, attrs ...string) *PolicyBuilder {
	el := strings.ToLower(element)
	if b.p.Attrs[el] == nil {
		b.p.Attrs[el] = make(map[string]bool)
	}
	for _, a := range attrs {
		b.p.Attrs[el][a] = true
	}
	return b
}

// AllowSchemes adds URL schemes to the allowlist.
func (b *PolicyBuilder) AllowSchemes(schemes ...string) *PolicyBuilder {
	for _, s := range schemes {
		b.p.Schemes[strings.ToLower(s)] = true
	}
	return b
}

// StripComments configures whether comments are removed.
func (b *PolicyBuilder) StripComments(strip bool) *PolicyBuilder {
	b.p.StripComments = strip
	return b
}

// RequireNofollow configures automatic rel=nofollow on links.
func (b *PolicyBuilder) RequireNofollow(req bool) *PolicyBuilder {
	b.p.RequireRelNofollow = req
	return b
}

// AllowDataImages permits data: image URLs.
func (b *PolicyBuilder) AllowDataImages(allow bool) *PolicyBuilder {
	b.p.AllowDataImages = allow
	return b
}

// Build returns the constructed policy.
func (b *PolicyBuilder) Build() Policy { return b.p }
