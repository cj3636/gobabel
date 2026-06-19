package codec

type BitWriter struct {
	buf []byte
	cur byte
	n   uint8
}

func NewBitWriter(capacity int) *BitWriter { return &BitWriter{buf: make([]byte, 0, capacity)} }
func (w *BitWriter) WriteBits(v uint64, bits int) {
	for i := bits - 1; i >= 0; i-- {
		w.cur = (w.cur << 1) | byte((v>>i)&1)
		w.n++
		if w.n == 8 {
			w.buf = append(w.buf, w.cur)
			w.cur = 0
			w.n = 0
		}
	}
}
func (w *BitWriter) Bytes() []byte {
	if w.n > 0 {
		w.buf = append(w.buf, w.cur<<(8-w.n))
		w.n = 0
		w.cur = 0
	}
	return w.buf
}
