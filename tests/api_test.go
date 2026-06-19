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
