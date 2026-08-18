// Command htmlsift is an HTML analysis and sanitization toolkit. It
// wraps the internal htmlparse/htmllink/htmlsanitize packages behind a
// small CLI.
//
// Subcommands:
//
//	htmlsift parse <file|->                 print document statistics
//	htmlsift text <file|->                  extract visible text
//	htmlsift links <file|-> [-base url]     list extracted hyperlinks
//	htmlsift sanitize <file|->              print policy-sanitized HTML
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"htmlsift/internal/api"
	"htmlsift/internal/htmllink"
	"htmlsift/internal/htmlparse"
	"htmlsift/internal/htmlsanitize"
)

func main() {
	if len(os.Args) < 2 {
		usage(os.Stderr)
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "parse":
		err = runParse(os.Args[2:])
	case "text":
		err = runText(os.Args[2:])
	case "links":
		err = runLinks(os.Args[2:])
	case "sanitize":
		err = runSanitize(os.Args[2:])
	case "serve":
		err = runServe(os.Args[2:])
	case "help", "-h", "--help":
		usage(os.Stdout)
		return
	default:
		fmt.Fprintf(os.Stderr, "htmlsift: unknown subcommand %q\n\n", os.Args[1])
		usage(os.Stderr)
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "htmlsift:", err)
		os.Exit(1)
	}
}

func usage(w *os.File) {
	fmt.Fprintln(w, `usage: htmlsift <subcommand> [flags] <file|->   (- reads stdin)

subcommands:
  parse <file|->              print document statistics
  text <file|->               extract visible text
  links <file|-> [-base url]  list extracted hyperlinks
  sanitize <file|->           print policy-sanitized HTML
  serve [-addr :8080]         start web UI and API server`)
}

// reorderFlags moves flags (and their values) ahead of positional
// arguments so flag.FlagSet.Parse sees them first. Go's flag package
// stops at the first non-flag argument.
func reorderFlags(args []string, boolFlags map[string]bool) []string {
	flags := make([]string, 0, len(args))
	rest := make([]string, 0, len(args))
	i := 0
	for i < len(args) {
		a := args[i]
		if !strings.HasPrefix(a, "-") || a == "-" {
			rest = append(rest, a)
			i++
			continue
		}
		name := strings.TrimLeft(a, "-")
		key := name
		if eq := strings.IndexByte(name, '='); eq >= 0 {
			key = name[:eq]
		}
		flags = append(flags, a)
		i++
		if !boolFlags[key] && i < len(args) && !strings.HasPrefix(args[i], "-") {
			flags = append(flags, args[i])
			i++
		}
	}
	return append(flags, rest...)
}

func readInput(path string) (string, error) {
	if path == "-" {
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("read stdin: %w", err)
		}
		return string(b), nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read %q: %w", path, err)
	}
	return string(b), nil
}

func runParse(args []string) error {
	fs := flag.NewFlagSet("parse", flag.ContinueOnError)
	if err := fs.Parse(reorderFlags(args, nil)); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("parse needs exactly one file argument")
	}
	input, err := readInput(fs.Arg(0))
	if err != nil {
		return err
	}
	d, err := htmlparse.Parse(input)
	if err != nil {
		return err
	}
	st := d.Stats()
	fmt.Printf("elements:   %d\n", st.Elements)
	fmt.Printf("text nodes: %d\n", st.TextNodes)
	fmt.Printf("comments:   %d\n", st.Comments)
	fmt.Printf("max depth:  %d\n", st.MaxDepth)
	fmt.Printf("links:      %d\n", st.Links)
	fmt.Printf("images:     %d\n", st.Images)
	fmt.Printf("text bytes: %d\n", st.TotalBytes)
	return nil
}

func runText(args []string) error {
	fs := flag.NewFlagSet("text", flag.ContinueOnError)
	if err := fs.Parse(reorderFlags(args, nil)); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("text needs exactly one file argument")
	}
	input, err := readInput(fs.Arg(0))
	if err != nil {
		return err
	}
	d, err := htmlparse.Parse(input)
	if err != nil {
		return err
	}
	fmt.Println(htmlparse.VisibleText(d))
	return nil
}

func runLinks(args []string) error {
	fs := flag.NewFlagSet("links", flag.ContinueOnError)
	base := fs.String("base", "", "base URL for resolution")
	if err := fs.Parse(reorderFlags(args, nil)); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("links needs exactly one file argument")
	}
	input, err := readInput(fs.Arg(0))
	if err != nil {
		return err
	}
	d, err := htmlparse.Parse(input)
	if err != nil {
		return err
	}
	links, err := htmllink.Extract(d, *base)
	if err != nil {
		return err
	}
	for _, l := range links {
		target := l.Href
		if l.Resolved != "" {
			target = l.Resolved
		}
		fmt.Printf("%-6s %-5s %s  (%s)\n", l.Tag, htmllink.Classify(l.Href), target, l.Text)
	}
	return nil
}

func runSanitize(args []string) error {
	fs := flag.NewFlagSet("sanitize", flag.ContinueOnError)
	fragment := fs.Bool("fragment", false, "treat input as a body fragment")
	if err := fs.Parse(reorderFlags(args, map[string]bool{"fragment": true})); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("sanitize needs exactly one file argument")
	}
	input, err := readInput(fs.Arg(0))
	if err != nil {
		return err
	}
	p := htmlsanitize.DefaultPolicy()
	var out string
	if *fragment {
		out, err = p.SanitizeFragment(input)
	} else {
		out, err = p.Sanitize(input)
	}
	if err != nil {
		return err
	}
	fmt.Println(out)
	return nil
}

func runServe(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	addr := fs.String("addr", ":8080", "listen address")
	webDir := fs.String("web", "web", "path to web static files")
	if err := fs.Parse(reorderFlags(args, nil)); err != nil {
		return err
	}
	cfg := api.Config{Addr: *addr, WebDir: *webDir}
	srv := api.New(cfg)
	fmt.Fprintf(os.Stderr, "htmlsift: listening on %s\n", *addr)
	return srv.ListenAndServe()
}
