package tests

import (
	"bytes"
	"encoding/json"
	"github.com/cj3636/gobabel/internal/alphabet"
	"github.com/cj3636/gobabel/internal/api"
	"github.com/cj3636/gobabel/internal/config"
	"github.com/cj3636/gobabel/internal/engine"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func router(t *testing.T) http.Handler {
	c, e := config.Load()
	if e != nil {
		t.Fatal(e)
	}
	return api.Routes(engine.Engine{Alphabet: alphabet.CodeASCII, Sealer: c.Sealer}, slog.New(slog.NewTextHandler(io.Discard, nil)))
}
func TestAPI(t *testing.T) {
	h := router(t)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/v1/health", nil))
	if w.Code != 200 {
		t.Fatal(w.Code)
	}
	body := bytes.NewBufferString(`{"text":"hello world","placement":"hash"}`)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("POST", "/v1/locate", body))
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	var loc map[string]any
	json.Unmarshal(w.Body.Bytes(), &loc)
	rangeURL := loc["range_url"].(string)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", rangeURL, nil))
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	var rg map[string]any
	json.Unmarshal(w.Body.Bytes(), &rg)
	if rg["text"] != "hello world" {
		t.Fatalf("got %q", rg["text"])
	}
}

func TestAPIRejectsInvalidBookPaths(t *testing.T) {
	h := router(t)
	body := bytes.NewBufferString(`{"text":"hello world","placement":"hash"}`)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("POST", "/v1/locate", body))
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	var loc map[string]any
	json.Unmarshal(w.Body.Bytes(), &loc)
	pageURL := loc["page_url"].(string)
	rangeURL := loc["range_url"].(string)

	invalid := []string{
		"/v1/book/bf1",
		strings.Replace(pageURL, "/bf1/", "/bf1/+/", 1),
		strings.Replace(pageURL, "/bf1/", "/bf1/abc=/", 1),
		rangeURL + "/1:2",
		pageURL + "/1:",
		pageURL + "/-1:2",
		pageURL + "/3:2",
		pageURL + "/0:5001",
		"/v1/book/bf1/" + strings.Repeat("A", 2050),
	}
	for _, path := range invalid {
		t.Run(path, func(t *testing.T) {
			w := httptest.NewRecorder()
			h.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
			if w.Code < 400 {
				t.Fatalf("status = %d, want error; body = %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestAPIRejectsInvalidCorpusPaths(t *testing.T) {
	h := router(t)
	invalid := []string{
		"/v1/corpus/hex/0/notwall/1/shelf/2/book/3",
		"/v1/corpus/hex/0/wall/1/shelf/2/book/3/extra/segment",
		"/v1/corpus/hex/0/wall/1/shelf/2/book/3/1:",
		"/v1/corpus/hex/0/wall/1/shelf/2/book/3/0:5001",
	}
	for _, path := range invalid {
		t.Run(path, func(t *testing.T) {
			w := httptest.NewRecorder()
			h.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
			if w.Code < 400 {
				t.Fatalf("status = %d, want error; body = %s", w.Code, w.Body.String())
			}
		})
	}
}
