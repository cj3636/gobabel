package engine

import (
	"bytes"
	"testing"

	"github.com/cj3636/gobabel/internal/address"
	"github.com/cj3636/gobabel/internal/alphabet"
	gbcrypto "github.com/cj3636/gobabel/internal/crypto"
)

func testEngine(t testing.TB) Engine {
	s, err := gbcrypto.PublicSealer()
	if err != nil {
		t.Fatal(err)
	}
	return Engine{Alphabet: alphabet.CodeASCII, Sealer: s}
}
func TestLocatePlacements(t *testing.T) {
	e := testEngine(t)
	for _, p := range []string{"start", "hash", "center"} {
		r, err := e.Locate([]byte("hello"), p)
		if err != nil {
			t.Fatal(err)
		}
		blob, err := address.DecodeSegments(byteSlicesToStrings(bytes.Split([]byte(r.PageURL), []byte("/"))[4:]))
		if err != nil {
			t.Fatal(err)
		}
		page, _, err := e.Page(blob)
		if err != nil {
			t.Fatal(err)
		}
		if string(page[r.Start:r.End]) != "hello" {
			t.Fatal("range mismatch")
		}
	}
}
func BenchmarkLocate100(b *testing.B) {
	e := testEngine(b)
	s := bytes.Repeat([]byte("a"), 100)
	for i := 0; i < b.N; i++ {
		e.Locate(s, "hash")
	}
}
func BenchmarkLocate5000(b *testing.B) {
	e := testEngine(b)
	s := bytes.Repeat([]byte("a"), 5000)
	for i := 0; i < b.N; i++ {
		e.Locate(s, "hash")
	}
}
func BenchmarkDecodeSealedPage(b *testing.B) {
	e := testEngine(b)
	r, _ := e.Locate(bytes.Repeat([]byte("a"), 100), "hash")
	seg := bytes.Split([]byte(r.PageURL), []byte("/"))[4:]
	blob, _ := address.DecodeSegments(byteSlicesToStrings(seg))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Page(blob)
	}
}
func byteSlicesToStrings(in [][]byte) []string {
	out := make([]string, len(in))
	for i := range in {
		out[i] = string(in[i])
	}
	return out
}
