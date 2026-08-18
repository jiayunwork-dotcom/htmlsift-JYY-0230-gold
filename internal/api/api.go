// Package api provides an HTTP server exposing the htmlsift functionality as
// a REST API. It serves endpoints for parsing, sanitizing, link extraction,
// and text extraction. The server also serves a static frontend from the
// embedded web directory.
package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"htmlsift/internal/htmllink"
	"htmlsift/internal/htmlparse"
	"htmlsift/internal/htmlsanitize"
)

// Server is the HTTP API server for htmlsift.
type Server struct {
	mux      *http.ServeMux
	policy   htmlsanitize.Policy
	addr     string
	webDir   string
}

// Config holds server configuration.
type Config struct {
	Addr   string // listen address (e.g. ":8080")
	WebDir string // path to static web files
}

// DefaultConfig returns sensible defaults for local development.
func DefaultConfig() Config {
	return Config{Addr: ":8080", WebDir: "web"}
}

// New creates a new API server.
func New(cfg Config) *Server {
	s := &Server{
		mux:    http.NewServeMux(),
		policy: htmlsanitize.DefaultPolicy(),
		addr:   cfg.Addr,
		webDir: cfg.WebDir,
	}
	s.routes()
	return s
}

// Handler returns the HTTP handler for use in testing or custom setups.
func (s *Server) Handler() http.Handler { return s.mux }

// Addr returns the configured listen address.
func (s *Server) Addr() string { return s.addr }

// ListenAndServe starts the HTTP server.
func (s *Server) ListenAndServe() error {
	return http.ListenAndServe(s.addr, s.mux)
}

// routes registers all API endpoints and static file serving.
func (s *Server) routes() {
	s.mux.HandleFunc("/api/sanitize", s.handleSanitize)
	s.mux.HandleFunc("/api/parse", s.handleParse)
	s.mux.HandleFunc("/api/links", s.handleLinks)
	s.mux.HandleFunc("/api/text", s.handleText)
	s.mux.HandleFunc("/api/health", s.handleHealth)
	// Serve static frontend files.
	if s.webDir != "" {
		fs := http.FileServer(http.Dir(s.webDir))
		s.mux.Handle("/", fs)
	}
}

// SanitizeRequest is the JSON body for /api/sanitize.
type SanitizeRequest struct {
	HTML     string `json:"html"`
	Fragment bool   `json:"fragment"`
}

// SanitizeResponse is the JSON response for /api/sanitize.
type SanitizeResponse struct {
	Output          string `json:"output"`
	RemovedElements int    `json:"removed_elements"`
	RemovedAttrs    int    `json:"removed_attrs"`
	RemovedURLs     int    `json:"removed_urls"`
}

// ParseRequest is the JSON body for /api/parse.
type ParseRequest struct {
	HTML string `json:"html"`
}

// ParseResponse is the JSON response for /api/parse.
type ParseResponse struct {
	Elements  int            `json:"elements"`
	TextNodes int            `json:"text_nodes"`
	Comments  int            `json:"comments"`
	MaxDepth  int            `json:"max_depth"`
	Links     int            `json:"links"`
	Images    int            `json:"images"`
	TextBytes int            `json:"text_bytes"`
	Tags      map[string]int `json:"tags"`
}

// LinksRequest is the JSON body for /api/links.
type LinksRequest struct {
	HTML    string `json:"html"`
	BaseURL string `json:"base_url"`
}

// LinkItem is a single link in the response.
type LinkItem struct {
	Tag      string `json:"tag"`
	Href     string `json:"href"`
	Text     string `json:"text"`
	Resolved string `json:"resolved,omitempty"`
	Class    string `json:"class"`
}

// LinksResponse is the JSON response for /api/links.
type LinksResponse struct {
	Links []LinkItem `json:"links"`
	Count int        `json:"count"`
}

// TextRequest is the JSON body for /api/text.
type TextRequest struct {
	HTML string `json:"html"`
}

// TextResponse is the JSON response for /api/text.
type TextResponse struct {
	Text string `json:"text"`
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (s *Server) handleSanitize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpError(w, http.StatusMethodNotAllowed, "POST only")
		return
	}
	var req SanitizeRequest
	if err := readJSON(r, &req); err != nil {
		httpError(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.HTML) == "" {
		httpError(w, http.StatusBadRequest, "html field is required")
		return
	}
	var out string
	var rep htmlsanitize.Report
	var err error
	if req.Fragment {
		out, err = s.policy.SanitizeFragment(req.HTML)
		// Fragment doesn't return a report; use zero report.
	} else {
		out, rep, err = s.policy.SanitizeReport(req.HTML)
	}
	if err != nil {
		httpError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, SanitizeResponse{
		Output:          out,
		RemovedElements: rep.RemovedElements,
		RemovedAttrs:    rep.RemovedAttrs,
		RemovedURLs:     rep.RemovedURLs,
	})
}

func (s *Server) handleParse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpError(w, http.StatusMethodNotAllowed, "POST only")
		return
	}
	var req ParseRequest
	if err := readJSON(r, &req); err != nil {
		httpError(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.HTML) == "" {
		httpError(w, http.StatusBadRequest, "html field is required")
		return
	}
	d, err := htmlparse.Parse(req.HTML)
	if err != nil {
		httpError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	st := d.Stats()
	writeJSON(w, ParseResponse{
		Elements:  st.Elements,
		TextNodes: st.TextNodes,
		Comments:  st.Comments,
		MaxDepth:  st.MaxDepth,
		Links:     st.Links,
		Images:    st.Images,
		TextBytes: st.TotalBytes,
		Tags:      st.Tags,
	})
}

func (s *Server) handleLinks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpError(w, http.StatusMethodNotAllowed, "POST only")
		return
	}
	var req LinksRequest
	if err := readJSON(r, &req); err != nil {
		httpError(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.HTML) == "" {
		httpError(w, http.StatusBadRequest, "html field is required")
		return
	}
	d, err := htmlparse.Parse(req.HTML)
	if err != nil {
		httpError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	links, err := htmllink.Extract(d, req.BaseURL)
	if err != nil {
		httpError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	items := make([]LinkItem, len(links))
	for i, l := range links {
		items[i] = LinkItem{
			Tag:      l.Tag,
			Href:     l.Href,
			Text:     l.Text,
			Resolved: l.Resolved,
			Class:    htmllink.Classify(l.Href).String(),
		}
	}
	writeJSON(w, LinksResponse{Links: items, Count: len(items)})
}

func (s *Server) handleText(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpError(w, http.StatusMethodNotAllowed, "POST only")
		return
	}
	var req TextRequest
	if err := readJSON(r, &req); err != nil {
		httpError(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.HTML) == "" {
		httpError(w, http.StatusBadRequest, "html field is required")
		return
	}
	d, err := htmlparse.Parse(req.HTML)
	if err != nil {
		httpError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	text := htmlparse.VisibleText(d)
	writeJSON(w, TextResponse{Text: text})
}

func readJSON(r *http.Request, v interface{}) error {
	body, err := io.ReadAll(io.LimitReader(r.Body, 10*1024*1024))
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}
	if len(body) == 0 {
		return fmt.Errorf("empty request body")
	}
	return json.Unmarshal(body, v)
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(v)
}

func httpError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
