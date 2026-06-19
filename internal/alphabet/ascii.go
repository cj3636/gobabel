package alphabet

var CodeASCII = func() Alphabet {
	chars := []byte{'\t', '\n', '\r', ' '}
	for b := byte('!'); b <= '~'; b++ {
		chars = append(chars, b)
	}
	return New(CodeASCIIV1ID, chars)
}()
