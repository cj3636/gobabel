package engine

import (
	"crypto/sha256"
	"encoding/binary"
	"github.com/cj3636/gobabel/internal/alphabet"
)

// Generate expands seed material into deterministic filler bytes for v1 pages.
// v1 uses SHA-256 counter expansion (label || seed || uint64 counter) and
// rejection sampling to map digest bytes uniformly onto the 98-character
// alphabet; it does not use ChaCha20 for stream generation.
func Generate(seed []byte, label string, a alphabet.Alphabet, n int) ([]byte, error) {
	out := make([]byte, 0, n)
	var ctr uint64
	limit := byte(196)
	for len(out) < n {
		h := sha256.New()
		h.Write([]byte(label))
		h.Write(seed)
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], ctr)
		h.Write(b[:])
		sum := h.Sum(nil)
		ctr++
		for _, x := range sum {
			if x < limit {
				out = append(out, a.Chars[int(x)%98])
				if len(out) == n {
					break
				}
			}
		}
	}
	return out, nil
}
