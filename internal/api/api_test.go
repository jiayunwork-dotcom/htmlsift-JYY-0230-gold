package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestServer() *Server {
	cfg := DefaultConfig()
	cfg.WebDir = "" // no static files in tests
	return New(cfg)
}

func TestHealthEndpoint(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("status = %d", w.Code)
	}
	var resp map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["status"] != "ok" {
		t.Fatalf("status = %q", resp["status"])
	}
}

func TestSanitizeEndpoint(t *testing.T) {
	srv := newTestServer()
	body := `{"html":"<p onclick=\"x\">hello</p><script>bad</script>"}`
	req := httptest.NewRequest(http.MethodPost, "/api/sanitize", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var resp SanitizeResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.RemovedElements < 1 {
		t.Fatalf("expected removed elements, got %d", resp.RemovedElements)
	}
	if bytes.Contains([]byte(resp.Output), []byte("<script")) {
		t.Fatal("script survived")
	}
}

func TestSanitizeEmptyBody(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest(http.MethodPost, "/api/sanitize", bytes.NewBufferString(`{"html":""}`))
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestParseEndpoint(t *testing.T) {
	srv := newTestServer()
	body := `{"html":"<html><body><p>hi</p><a href=\"/x\">link</a><img src=\"a.png\"></body></html>"}`
	req := httptest.NewRequest(http.MethodPost, "/api/parse", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("status = %d", w.Code)
	}
	var resp ParseResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Links != 1 {
		t.Fatalf("links = %d, want 1", resp.Links)
	}
	if resp.Images != 1 {
		t.Fatalf("images = %d, want 1", resp.Images)
	}
}

func TestLinksEndpoint(t *testing.T) {
	srv := newTestServer()
	body := `{"html":"<html><body><a href=\"/page\">go</a></body></html>","base_url":"https://example.com/"}`
	req := httptest.NewRequest(http.MethodPost, "/api/links", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("status = %d", w.Code)
	}
	var resp LinksResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Count != 1 {
		t.Fatalf("count = %d, want 1", resp.Count)
	}
	if resp.Links[0].Resolved != "https://example.com/page" {
		t.Fatalf("resolved = %q", resp.Links[0].Resolved)
	}
}

func TestTextEndpoint(t *testing.T) {
	srv := newTestServer()
	body := `{"html":"<html><body><p>Hello <b>world</b></p><script>x</script></body></html>"}`
	req := httptest.NewRequest(http.MethodPost, "/api/text", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("status = %d", w.Code)
	}
	var resp TextResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Text != "Hello world" {
		t.Fatalf("text = %q", resp.Text)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest(http.MethodGet, "/api/sanitize", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != 405 {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}
