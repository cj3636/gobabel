package corpus

import (
	"bytes"
	"github.com/cj3636/gobabel/internal/alphabet"
	"testing"
)

func TestCorpus(t *testing.T) {
	c := Coordinate{"zk4n2", "3", "1", "k9Lm2", "412"}
	p1, _ := Page(c, alphabet.CodeASCII)
	p2, _ := Page(c, alphabet.CodeASCII)
	if !bytes.Equal(p1, p2) || len(p1) != 5000 {
		t.Fatal("bad corpus")
	}
	c.Page = "413"
	p3, _ := Page(c, alphabet.CodeASCII)
	if bytes.Equal(p1, p3) {
		t.Fatal("same")
	}
	if inv := alphabet.CodeASCII.ValidateBytes(p1); len(inv) > 0 {
		t.Fatal(inv[0])
	}
}
func BenchmarkCorpusPage(b *testing.B) {
	c := Coordinate{"zk4n2", "3", "1", "k9Lm2", "412"}
	for i := 0; i < b.N; i++ {
		Page(c, alphabet.CodeASCII)
	}
}
