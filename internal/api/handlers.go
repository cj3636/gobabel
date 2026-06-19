package api

import (
	"encoding/json"
	"github.com/cj3636/gobabel/internal/address"
	"github.com/cj3636/gobabel/internal/alphabet"
	"github.com/cj3636/gobabel/internal/codec"
	"github.com/cj3636/gobabel/internal/corpus"
	"github.com/cj3636/gobabel/internal/engine"
	"github.com/cj3636/gobabel/internal/version"
	"log/slog"
	"net/http"
	"strings"
)

type Handler struct {
	Engine engine.Engine
	Log    *slog.Logger
}

func jsonw(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
func (h Handler) Health(w http.ResponseWriter, r *http.Request) {
	jsonw(w, map[string]any{"ok": true, "service": "gobabel", "version": version.Version})
}
func (h Handler) Alphabet(w http.ResponseWriter, r *http.Request) {
	a := h.Engine.Alphabet
	jsonw(w, map[string]any{"id": a.ID, "page_size": engine.PageSize, "length": len(a.Chars), "chars_escaped": a.EscapedChars()})
}
func (h Handler) Validate(w http.ResponseWriter, r *http.Request) {
	var req ValidateRequest
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		writeErr(w, 400, "invalid_json", "Invalid JSON.", nil)
		return
	}
	inv := h.Engine.Alphabet.ValidateBytes([]byte(req.Text))
	jsonw(w, map[string]any{"valid": len(inv) == 0 && len(req.Text) <= engine.PageSize, "length": len(req.Text), "max_length": engine.PageSize, "alphabet": h.Engine.Alphabet.ID, "invalid": inv})
}
func (h Handler) Locate(w http.ResponseWriter, r *http.Request) {
	var req LocateRequest
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		writeErr(w, 400, "invalid_json", "Invalid JSON.", nil)
		return
	}
	res, err := h.Engine.Locate([]byte(req.Text), req.Placement)
	if err != nil {
		code := strings.Split(err.Error(), ":")[0]
		status := 400
		writeErr(w, status, code, "Locate request failed.", nil)
		return
	}
	jsonw(w, res)
}
func (h Handler) Book(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/v1/book/"), "/")
	if len(parts) < 2 || parts[0] != address.Version {
		writeErr(w, 400, "unsupported_version", "Unsupported book version.", nil)
		return
	}
	if s, e, ok, err := address.ParseRange(parts[len(parts)-1]); ok {
		if err != nil {
			writeErr(w, 400, "invalid_range", "Invalid range.", nil)
			return
		}
		parts = parts[:len(parts)-1]
		blob, err := address.DecodeSegments(parts[1:])
		if err != nil {
			writeErr(w, 400, "invalid_address", "Invalid address.", nil)
			return
		}
		page, _, err := h.Engine.Page(blob)
		if err != nil {
			writeErr(w, 400, err.Error(), "Could not open address.", nil)
			return
		}
		jsonw(w, map[string]any{"address_type": engine.AddressType, "start": s, "end": e, "text": string(page[s:e])})
		return
	}
	blob, err := address.DecodeSegments(parts[1:])
	if err != nil {
		writeErr(w, 400, "invalid_address", "Invalid address.", nil)
		return
	}
	page, _, err := h.Engine.Page(blob)
	if err != nil {
		writeErr(w, 400, err.Error(), "Could not open address.", nil)
		return
	}
	jsonw(w, map[string]any{"address_type": engine.AddressType, "text": string(page), "page_size": engine.PageSize, "alphabet": alphabet.CodeASCIIV1ID, "codec": codec.ID})
}
func (h Handler) Corpus(w http.ResponseWriter, r *http.Request) {
	p := strings.Split(strings.TrimPrefix(r.URL.Path, "/v1/corpus/hex/"), "/")
	if len(p) != 8 && len(p) != 9 {
		writeErr(w, 404, "invalid_address", "Invalid corpus path.", nil)
		return
	}
	c := corpus.Coordinate{Hexagon: p[0], Wall: p[2], Shelf: p[4], Book: p[6], Page: p[7]}
	if p[1] != "wall" || p[3] != "shelf" || p[5] != "book" {
		writeErr(w, 404, "invalid_address", "Invalid corpus path.", nil)
		return
	}
	page, err := corpus.Page(c, h.Engine.Alphabet)
	if err != nil {
		writeErr(w, 500, "internal_error", "Internal error.", nil)
		return
	}
	if len(p) == 9 {
		s, e, ok, err := address.ParseRange(p[8])
		if !ok || err != nil {
			writeErr(w, 400, "invalid_range", "Invalid range.", nil)
			return
		}
		jsonw(w, map[string]any{"address_type": corpus.AddressType, "start": s, "end": e, "text": string(page[s:e])})
		return
	}
	jsonw(w, map[string]any{"address_type": corpus.AddressType, "coordinate": c, "text": string(page), "page_size": engine.PageSize, "alphabet": h.Engine.Alphabet.ID})
}
