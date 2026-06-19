package tests

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cj3636/gobabel/internal/address"
	"github.com/cj3636/gobabel/internal/alphabet"
	"github.com/cj3636/gobabel/internal/api"
	"github.com/cj3636/gobabel/internal/config"
	"github.com/cj3636/gobabel/internal/engine"
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

func TestAPIErrorCodes(t *testing.T) {
	h := router(t)
	c, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}

	unsupportedAlphabetURL := sealedTestBookURL(t, c.Sealer, func(pt []byte) {
		pt[5] = 2
	})
	unsupportedCodecURL := sealedTestBookURL(t, c.Sealer, func(pt []byte) {
		pt[6] = 2
	})

	cases := []struct {
		name         string
		method       string
		path         string
		body         string
		wantStatus   int
		wantCode     string
		checkDetails bool
	}{
		{name: "invalid_json", method: "POST", path: "/v1/locate", body: `{"text":`, wantStatus: 400, wantCode: "invalid_json"},
		{name: "invalid_character", method: "POST", path: "/v1/locate", body: `{"text":"\u0000"}`, wantStatus: 400, wantCode: "invalid_character", checkDetails: true},
		{name: "text_too_long", method: "POST", path: "/v1/locate", body: fmt.Sprintf(`{"text":%q}`, string(bytes.Repeat([]byte("a"), engine.PageSize+1))), wantStatus: 400, wantCode: "text_too_long"},
		{name: "empty_text", method: "POST", path: "/v1/locate", body: `{"text":""}`, wantStatus: 400, wantCode: "empty_text"},
		{name: "invalid_range", method: "GET", path: "/v1/corpus/hex/a/wall/b/shelf/c/book/d/e/5:1", wantStatus: 400, wantCode: "invalid_range"},
		{name: "invalid_address", method: "GET", path: "/v1/book/bf1/!!!!", wantStatus: 400, wantCode: "invalid_address"},
		{name: "unsupported_version", method: "GET", path: "/v1/book/bad/address", wantStatus: 400, wantCode: "unsupported_version"},
		{name: "unsupported_alphabet", method: "GET", path: unsupportedAlphabetURL, wantStatus: 400, wantCode: "unsupported_alphabet"},
		{name: "unsupported_codec", method: "GET", path: unsupportedCodecURL, wantStatus: 400, wantCode: "unsupported_codec"},
		{name: "crypto_open_failed", method: "GET", path: "/v1/book/bf1/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", wantStatus: 400, wantCode: "crypto_open_failed"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			h.ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, bytes.NewBufferString(tc.body)))
			if w.Code != tc.wantStatus {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}
			var got api.ErrorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if got.Error.Code != tc.wantCode {
				t.Fatalf("code=%q body=%s", got.Error.Code, w.Body.String())
			}
			if tc.checkDetails {
				details, ok := got.Error.Details.(map[string]any)
				if !ok {
					t.Fatalf("details=%T", got.Error.Details)
				}
				if details["position"] != float64(0) || details["byte"] != float64(0) {
					t.Fatalf("details=%v", details)
				}
				if _, ok := details["invalid"].([]any); !ok {
					t.Fatalf("invalid details=%v", details)
				}
			}
		})
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

func sealedTestBookURL(t *testing.T, sealer interface {
	Seal([]byte, []byte) ([]byte, error)
}, mutate func([]byte)) string {
	t.Helper()
	pt := make([]byte, 50)
	copy(pt[:4], []byte("BF01"))
	pt[5] = 1
	pt[6] = 1
	pt[7] = 1
	binary.BigEndian.PutUint16(pt[8:10], uint16(engine.PageSize))
	mutate(pt)
	blob, err := sealer.Seal(pt, engine.AAD())
	if err != nil {
		t.Fatal(err)
	}
	return address.Encode(address.Version, blob)
}
