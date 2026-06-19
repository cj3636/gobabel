package codec

import "io"

type BitReader struct {
	data []byte
	bit  int
}

func NewBitReader(b []byte) *BitReader { return &BitReader{data: b} }
func (r *BitReader) ReadBits(bits int) (uint64, error) {
	if r.bit+bits > len(r.data)*8 {
		return 0, io.ErrUnexpectedEOF
	}
	var v uint64
	for i := 0; i < bits; i++ {
		idx := r.bit / 8
		off := 7 - (r.bit % 8)
		v = (v << 1) | uint64((r.data[idx]>>off)&1)
		r.bit++
	}
	return v, nil
}
