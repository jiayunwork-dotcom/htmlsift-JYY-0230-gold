// Example: sanitize a snippet, extract links and print visible text.
package main

import (
	"fmt"

	"htmlsift/internal/htmllink"
	"htmlsift/internal/htmlparse"
	"htmlsift/internal/htmlsanitize"
)

const snippet = `<div class="post">
  <h2>Hello</h2>
  <p>Read <a href="https://example.com/docs" onclick="steal()">the docs</a>
     or <a href="javascript:alert(1)">run this</a>.</p>
  <script>track()</script>
  <img src="https://example.com/logo.png" alt="logo">
</div>`

func main() {
	out, rep, err := htmlsanitize.DefaultPolicy().SanitizeReport(snippet)
	if err != nil {
		fmt.Println("sanitize error:", err)
		return
	}
	fmt.Println("--- sanitized ---")
	fmt.Println(out)
	fmt.Printf("removed: %d elements, %d attrs, %d URLs\n",
		rep.RemovedElements, rep.RemovedAttrs, rep.RemovedURLs)

	d, _ := htmlparse.Parse(out)
	fmt.Println("--- visible text ---")
	fmt.Println(htmlparse.VisibleText(d))

	fmt.Println("--- links ---")
	links, _ := htmllink.Extract(d, "https://example.com/blog/")
	for _, l := range links {
		fmt.Printf("  %s -> %s (%s)\n", l.Text, l.Resolved, htmllink.Classify(l.Href))
	}
}
