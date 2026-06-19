package main

import (
	"encoding/json"
	"fmt"
	"github.com/cj3636/gobabel/internal/address"
	"github.com/cj3636/gobabel/internal/alphabet"
	"github.com/cj3636/gobabel/internal/api"
	"github.com/cj3636/gobabel/internal/config"
	"github.com/cj3636/gobabel/internal/engine"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	e := engine.Engine{Alphabet: alphabet.CodeASCII, Sealer: cfg.Sealer}
	if len(os.Args) < 2 {
		usage()
		return
	}
	switch os.Args[1] {
	case "serve":
		l := slog.New(slog.NewTextHandler(os.Stdout, nil))
		l.Info("server start", "addr", cfg.Addr, "seal_mode", cfg.SealMode)
		log.Fatal(http.ListenAndServe(cfg.Addr, api.Routes(e, l)))
	case "locate":
		need(3)
		r, err := e.Locate([]byte(os.Args[2]), "hash")
		fatal(err)
		enc(r)
	case "validate":
		need(3)
		inv := e.Alphabet.ValidateBytes([]byte(os.Args[2]))
		enc(map[string]any{"valid": len(inv) == 0, "invalid": inv})
	case "page":
		need(3)
		path := os.Args[2]
		parts := strings.Split(strings.TrimPrefix(path, "/v1/book/"), "/")
		if len(parts) < 2 {
			fatal(fmt.Errorf("bad path"))
		}
		if _, _, ok, _ := address.ParseRange(parts[len(parts)-1]); ok {
			parts = parts[:len(parts)-1]
		}
		blob, err := address.DecodeSegments(parts[1:])
		fatal(err)
		page, _, err := e.Page(blob)
		fatal(err)
		fmt.Print(string(page))
	default:
		usage()
	}
}
func usage() { fmt.Println("gobabel serve|locate|page|validate") }
func need(n int) {
	if len(os.Args) < n {
		usage()
		os.Exit(2)
	}
}
func fatal(e error) {
	if e != nil {
		log.Fatal(e)
	}
}
func enc(v any) { json.NewEncoder(os.Stdout).Encode(v) }
