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
		writeErr(w, 400, engine.ErrInvalidJSON.Error(), "Invalid JSON.", nil)
		return
	}
	inv := h.Engine.Alphabet.ValidateBytes([]byte(req.Text))
	jsonw(w, map[string]any{"valid": len(inv) == 0 && len(req.Text) <= engine.PageSize, "length": len(req.Text), "max_length": engine.PageSize, "alphabet": h.Engine.Alphabet.ID, "invalid": inv})
}
func (h Handler) Locate(w http.ResponseWriter, r *http.Request) {
	var req LocateRequest
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		writeErr(w, 400, engine.ErrInvalidJSON.Error(), "Invalid JSON.", nil)
		return
	}
	res, err := h.Engine.Locate([]byte(req.Text), req.Placement)
	if err != nil {
		writeAPIError(w, 400, err, "invalid_address", "Locate request failed.")
		return
	}
	jsonw(w, res)
}
func (h Handler) Book(w http.ResponseWriter, r *http.Request) {
	segs, rg, err := address.SplitBookPath(r.URL.Path)
	if err != nil {
		switch err.Error() {
		case engine.ErrUnsupportedVersion.Error():
			writeErr(w, 400, engine.ErrUnsupportedVersion.Error(), "Unsupported book version.", nil)
		case engine.ErrInvalidRange.Error():
			writeErr(w, 400, engine.ErrInvalidRange.Error(), "Invalid range.", nil)
		default:
			writeErr(w, 400, err.Error(), "Invalid address.", nil)
		}
		return
	}
	blob, err := address.DecodeSegments(segs)
	if err != nil {
		writeErr(w, 400, engine.ErrInvalidAddress.Error(), "Invalid address.", nil)
		return
	}
	page, _, err := h.Engine.Page(blob)
	if err != nil {
		writeAPIError(w, 400, err, "invalid_address", "Could not open address.")
		return
	}
	if rg != nil {
		jsonw(w, map[string]any{"address_type": engine.AddressType, "start": rg[0], "end": rg[1], "text": string(page[rg[0]:rg[1]])})
		return
	}
	jsonw(w, map[string]any{"address_type": engine.AddressType, "text": string(page), "page_size": engine.PageSize, "alphabet": alphabet.CodeASCIIV1ID, "codec": codec.ID})
}
func (h Handler) Corpus(w http.ResponseWriter, r *http.Request) {
	p := strings.Split(strings.TrimPrefix(r.URL.Path, "/v1/corpus/hex/"), "/")
	if len(p) != 8 && len(p) != 9 {
		writeErr(w, 404, engine.ErrInvalidAddress.Error(), "Invalid corpus path.", nil)
		return
	}
	if p[0] == "" || p[2] == "" || p[4] == "" || p[6] == "" || p[7] == "" || strings.Contains(p[7], ":") || p[1] != "wall" || p[3] != "shelf" || p[5] != "book" {
		writeErr(w, 404, engine.ErrInvalidAddress.Error(), "Invalid corpus path.", nil)
		return
	}
	c := corpus.Coordinate{Hexagon: p[0], Wall: p[2], Shelf: p[4], Book: p[6], Page: p[7]}
	page, err := corpus.Page(c, h.Engine.Alphabet)
	if err != nil {
		writeErr(w, 500, "internal_error", "Internal error.", nil)
		return
	}
	if len(p) == 9 {
		s, e, ok, err := address.ParseRange(p[8])
		if !ok || err != nil {
			writeErr(w, 400, engine.ErrInvalidRange.Error(), "Invalid range.", nil)
			return
		}
		jsonw(w, map[string]any{"address_type": corpus.AddressType, "start": s, "end": e, "text": string(page[s:e])})
		return
	}
	jsonw(w, map[string]any{"address_type": corpus.AddressType, "coordinate": c, "text": string(page), "page_size": engine.PageSize, "alphabet": h.Engine.Alphabet.ID})
}
