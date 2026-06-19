package alphabet

import (
	"bytes"
	"testing"
)

func TestCodeASCII(t *testing.T) {
	a := CodeASCII
	if len(a.Chars) != 98 {
		t.Fatalf("len=%d", len(a.Chars))
	}
	valid := append([]byte{'\t', '\n', '\r', ' '}, []byte("!~AZaz09[]{}")...)
	if inv := a.ValidateBytes(valid); len(inv) != 0 {
		t.Fatalf("valid rejected: %+v", inv)
	}
	if inv := a.ValidateBytes([]byte{0, 0xc3}); len(inv) != 2 {
		t.Fatalf("invalid not rejected: %+v", inv)
	}
	for i, b := range a.Chars {
		if a.Reverse[b] != int16(i) {
			t.Fatalf("reverse mismatch")
		}
	}
}

func BenchmarkValidate100(b *testing.B) {
	s := bytes.Repeat([]byte("a"), 100)
	for i := 0; i < b.N; i++ {
		CodeASCII.ValidateBytes(s)
	}
}
func BenchmarkValidate5000(b *testing.B) {
	s := bytes.Repeat([]byte("a"), 5000)
	for i := 0; i < b.N; i++ {
		CodeASCII.ValidateBytes(s)
	}
}
