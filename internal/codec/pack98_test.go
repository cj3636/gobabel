package codec

import (
	"bytes"
	"github.com/cj3636/gobabel/internal/alphabet"
	"math/rand"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	a := alphabet.CodeASCII
	cases := [][]byte{[]byte(""), []byte("a"), []byte("ab"), []byte("12345678"), []byte("123456789"), []byte("1234567890"), bytes.Repeat([]byte("x"), 100), bytes.Repeat(a.Chars, 60)[:5000], bytes.Repeat(a.Chars, 2)}
	for _, c := range cases {
		p, e := Pack(c, a)
		if e != nil {
			t.Fatal(e)
		}
		got, e := Unpack(p, len(c), a)
		if e != nil {
			t.Fatal(e)
		}
		if !bytes.Equal(got, c) {
			t.Fatalf("mismatch len %d", len(c))
		}
	}
}
func TestRandom(t *testing.T) {
	a := alphabet.CodeASCII
	r := rand.New(rand.NewSource(1))
	for n := 1; n < 200; n++ {
		b := make([]byte, n)
		for i := range b {
			b[i] = a.Chars[r.Intn(len(a.Chars))]
		}
		p, _ := Pack(b, a)
		g, e := Unpack(p, len(b), a)
		if e != nil || !bytes.Equal(b, g) {
			t.Fatal(n, e)
		}
	}
}
func BenchmarkPack100(b *testing.B) {
	a := alphabet.CodeASCII
	s := bytes.Repeat([]byte("a"), 100)
	for i := 0; i < b.N; i++ {
		Pack(s, a)
	}
}
func BenchmarkPack5000(b *testing.B) {
	a := alphabet.CodeASCII
	s := bytes.Repeat([]byte("a"), 5000)
	for i := 0; i < b.N; i++ {
		Pack(s, a)
	}
}
func BenchmarkUnpack100(b *testing.B) {
	a := alphabet.CodeASCII
	s := bytes.Repeat([]byte("a"), 100)
	p, _ := Pack(s, a)
	for i := 0; i < b.N; i++ {
		Unpack(p, 100, a)
	}
}
func BenchmarkUnpack5000(b *testing.B) {
	a := alphabet.CodeASCII
	s := bytes.Repeat([]byte("a"), 5000)
	p, _ := Pack(s, a)
	for i := 0; i < b.N; i++ {
		Unpack(p, 5000, a)
	}
}
