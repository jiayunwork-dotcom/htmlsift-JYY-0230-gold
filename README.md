# htmlsift

htmlsift is an HTML analysis and sanitization toolkit for Go. It wraps
`golang.org/x/net/html` behind a stable API for parsing documents and
fragments, walking and selecting nodes (CSS-lite selectors), and
extracting visible text; it extracts and resolves hyperlinks with URL
scheme classification and bidi-safety checks; and it rewrites untrusted
HTML against an allowlist policy that strips scripts, event handlers
and unsafe URL schemes. The sanitizer is deterministic and idempotent.

Dependencies: `golang.org/x/net` (HTML5 tokenizer/tree builder) and
`golang.org/x/text` (NFC normalization). No cgo, no network at runtime.

## Packages

- `internal/htmlparse` — parse/render, node walking, stats, selectors, visible text
- `internal/htmllink` — link extraction, URL resolution and classification, host grouping
- `internal/htmlsanitize` — allowlist sanitizer with removal reporting

## CLI

```
htmlsift parse <file|->                 print document statistics
htmlsift text <file|->                  extract visible text
htmlsift links <file|-> [-base url]     list extracted hyperlinks
htmlsift sanitize <file|->              print policy-sanitized HTML
```

`-` reads from stdin. Example:

```
$ echo '<p>hi <a href="/x">x</a></p>' | go run . text -
hi x

$ echo '<p onclick="x()">hi</p><script>bad()</script>' | go run . sanitize -
<html><head></head><body><p>hi</p></body></html>
```

## Library usage

```go
d, err := htmlparse.Parse(input)
links, err := htmllink.Extract(d, "https://example.com/")
out, rep, err := htmlsanitize.DefaultPolicy().SanitizeReport(input)
// rep.RemovedElements / rep.RemovedURLs tell you what was stripped.
```

## Build and test

```
go mod download
go build ./...
go test ./...
go vet ./...
```

The module targets Go 1.21 and builds offline after `go mod download`.
