package codec

import (
	"errors"
	"fmt"
	"github.com/cj3636/gobabel/internal/alphabet"
)

const ID = "pack98-v1"

var GroupBits = [10]int{0, 7, 14, 20, 27, 34, 40, 47, 53, 60}

func Pack(text []byte, a alphabet.Alphabet) ([]byte, error) {
	if inv := a.ValidateBytes(text); len(inv) > 0 {
		return nil, fmt.Errorf("invalid byte at %d", inv[0].Position)
	}
	w := NewBitWriter((len(text)*60 + 71) / 72)
	for i := 0; i < len(text); {
		n := len(text) - i
		if n > 9 {
			n = 9
		}
		var v uint64
		for j := 0; j < n; j++ {
			v = v*98 + uint64(a.Reverse[text[i+j]])
		}
		w.WriteBits(v, GroupBits[n])
		i += n
	}
	return w.Bytes(), nil
}
func Unpack(packed []byte, length int, a alphabet.Alphabet) ([]byte, error) {
	if length < 0 {
		return nil, errors.New("negative length")
	}
	out := make([]byte, 0, length)
	r := NewBitReader(packed)
	for i := 0; i < length; {
		n := length - i
		if n > 9 {
			n = 9
		}
		v, err := r.ReadBits(GroupBits[n])
		if err != nil {
			return nil, err
		}
		digits := make([]byte, n)
		for j := n - 1; j >= 0; j-- {
			d := v % 98
			v /= 98
			digits[j] = a.Chars[d]
		}
		if v != 0 {
			return nil, errors.New("packed value out of range")
		}
		out = append(out, digits...)
		i += n
	}
	return out, nil
}
