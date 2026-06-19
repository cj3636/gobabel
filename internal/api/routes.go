package api

import (
	"github.com/cj3636/gobabel/internal/engine"
	"log/slog"
	"net/http"
	"time"
)

func Routes(e engine.Engine, logger *slog.Logger) http.Handler {
	h := Handler{Engine: e, Log: logger}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/health", h.Health)
	mux.HandleFunc("GET /v1/alphabet", h.Alphabet)
	mux.HandleFunc("POST /v1/validate", h.Validate)
	mux.HandleFunc("POST /v1/locate", h.Locate)
	mux.HandleFunc("GET /v1/book/", h.Book)
	mux.HandleFunc("GET /v1/corpus/hex/", h.Corpus)
	return logmw(logger, mux)
}
func logmw(l *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		st := time.Now()
		rr := &rw{ResponseWriter: w, status: 200}
		next.ServeHTTP(rr, r)
		l.Info("request", "method", r.Method, "path", r.URL.Path, "status", rr.status, "duration", time.Since(st).String())
	})
}

type rw struct {
	http.ResponseWriter
	status int
}

func (r *rw) WriteHeader(s int) { r.status = s; r.ResponseWriter.WriteHeader(s) }
