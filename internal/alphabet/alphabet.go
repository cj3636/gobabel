package alphabet

import "fmt"

const CodeASCIIV1ID = "code-ascii-v1"

type Alphabet struct {
	ID      string
	Chars   []byte
	Reverse [256]int16
}
type InvalidByte struct {
	Position int    `json:"position"`
	Byte     byte   `json:"byte"`
	Display  string `json:"display"`
}

func New(id string, chars []byte) Alphabet {
	a := Alphabet{ID: id, Chars: append([]byte(nil), chars...)}
	for i := range a.Reverse {
		a.Reverse[i] = -1
	}
	for i, b := range a.Chars {
		a.Reverse[b] = int16(i)
	}
	return a
}
func (a Alphabet) ValidateBytes(b []byte) []InvalidByte {
	out := []InvalidByte{}
	for i, c := range b {
		if a.Reverse[c] < 0 {
			out = append(out, InvalidByte{Position: i, Byte: c, Display: fmt.Sprintf("\\x%02X", c)})
		}
	}
	return out
}
func (a Alphabet) Valid(b []byte) bool { return len(a.ValidateBytes(b)) == 0 }
func EscapeByte(b byte) string {
	switch b {
	case '\t':
		return `\t`
	case '\n':
		return `\n`
	case '\r':
		return `\r`
	case ' ':
		return " "
	default:
		if b >= 33 && b <= 126 {
			return string([]byte{b})
		}
		return fmt.Sprintf("\\x%02X", b)
	}
}
func (a Alphabet) EscapedChars() []string {
	r := make([]string, len(a.Chars))
	for i, b := range a.Chars {
		r[i] = EscapeByte(b)
	}
	return r
}
