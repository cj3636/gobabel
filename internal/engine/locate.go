package engine

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/cj3636/gobabel/internal/address"
	"github.com/cj3636/gobabel/internal/alphabet"
	"github.com/cj3636/gobabel/internal/codec"
	gbcrypto "github.com/cj3636/gobabel/internal/crypto"
)

const AddressType = "sealed-anchor-v1"

var magic = []byte("BF01")

var (
	ErrInvalidJSON         = errors.New("invalid_json")
	ErrInvalidCharacter    = errors.New("invalid_character")
	ErrTextTooLong         = errors.New("text_too_long")
	ErrEmptyText           = errors.New("empty_text")
	ErrInvalidRange        = errors.New("invalid_range")
	ErrInvalidAddress      = errors.New("invalid_address")
	ErrUnsupportedVersion  = errors.New("unsupported_version")
	ErrUnsupportedAlphabet = errors.New("unsupported_alphabet")
	ErrUnsupportedCodec    = errors.New("unsupported_codec")
	ErrCryptoOpenFailed    = errors.New("crypto_open_failed")
)

type InvalidCharacterError struct {
	Invalid []alphabet.InvalidByte
}

func (e InvalidCharacterError) Error() string     { return ErrInvalidCharacter.Error() }
func (e InvalidCharacterError) Unwrap() error     { return ErrInvalidCharacter }
func (e InvalidCharacterError) ErrorCode() string { return ErrInvalidCharacter.Error() }
func (e InvalidCharacterError) ErrorDetails() any {
	if len(e.Invalid) == 0 {
		return nil
	}
	return map[string]any{
		"position": e.Invalid[0].Position,
		"byte":     e.Invalid[0].Byte,
		"invalid":  e.Invalid,
	}
}

type Engine struct {
	Alphabet alphabet.Alphabet
	Sealer   gbcrypto.Sealer
}
type LocateResult struct {
	RangeURL    string `json:"range_url"`
	PageURL     string `json:"page_url"`
	Start       int    `json:"start"`
	End         int    `json:"end"`
	Length      int    `json:"length"`
	PageSize    int    `json:"page_size"`
	Alphabet    string `json:"alphabet"`
	Codec       string `json:"codec"`
	Placement   string `json:"placement"`
	AddressType string `json:"address_type"`
}

func AAD() []byte {
	return []byte("gobabel-sealed-anchor-v1" + address.Version + alphabet.CodeASCIIV1ID + codec.ID)
}
func (e Engine) Locate(text []byte, placement string) (LocateResult, error) {
	if placement == "" {
		placement = "hash"
	}
	if len(text) == 0 {
		return LocateResult{}, ErrEmptyText
	}
	if len(text) > PageSize {
		return LocateResult{}, ErrTextTooLong
	}
	if inv := e.Alphabet.ValidateBytes(text); len(inv) > 0 {
		return LocateResult{}, InvalidCharacterError{Invalid: inv}
	}
	seed := make([]byte, 32)
	rand.Read(seed)
	start := 0
	pm := byte(1)
	switch placement {
	case "start":
		start = 0
		pm = 1
	case "hash":
		h := sha256.New()
		h.Write(text)
		h.Write(seed)
		h.Write([]byte(e.Alphabet.ID))
		sum := h.Sum(nil)
		start = int(binary.BigEndian.Uint64(sum[:8]) % uint64(PageSize-len(text)+1))
		pm = 2
	case "center":
		start = (PageSize - len(text)) / 2
		pm = 3
	default:
		return LocateResult{}, fmt.Errorf("invalid placement")
	}
	packed, err := codec.Pack(text, e.Alphabet)
	if err != nil {
		return LocateResult{}, err
	}
	var b bytes.Buffer
	b.Write(magic)
	b.WriteByte(0)
	b.WriteByte(1)
	b.WriteByte(1)
	b.WriteByte(pm)
	binary.Write(&b, binary.BigEndian, uint16(PageSize))
	binary.Write(&b, binary.BigEndian, uint16(start))
	binary.Write(&b, binary.BigEndian, uint16(len(text)))
	b.Write(seed)
	binary.Write(&b, binary.BigEndian, uint32(len(packed)))
	b.Write(packed)
	blob, err := e.Sealer.Seal(b.Bytes(), AAD())
	if err != nil {
		return LocateResult{}, err
	}
	page := address.Encode(address.Version, blob)
	return LocateResult{RangeURL: address.WithRange(page, start, start+len(text)), PageURL: page, Start: start, End: start + len(text), Length: len(text), PageSize: PageSize, Alphabet: e.Alphabet.ID, Codec: codec.ID, Placement: placement, AddressType: AddressType}, nil
}

type Decoded struct {
	Text          []byte
	Start, Length int
	Placement     string
	Seed          []byte
}

func (e Engine) Decode(blob []byte) (Decoded, error) {
	pt, err := e.Sealer.Open(blob, AAD())
	if err != nil {
		return Decoded{}, ErrCryptoOpenFailed
	}
	if len(pt) < 50 || string(pt[:4]) != "BF01" {
		return Decoded{}, ErrInvalidAddress
	}
	if pt[5] != 1 {
		return Decoded{}, ErrUnsupportedAlphabet
	}
	if pt[6] != 1 {
		return Decoded{}, ErrUnsupportedCodec
	}
	ps := binary.BigEndian.Uint16(pt[8:10])
	if ps != PageSize {
		return Decoded{}, ErrInvalidAddress
	}
	start := int(binary.BigEndian.Uint16(pt[10:12]))
	l := int(binary.BigEndian.Uint16(pt[12:14]))
	seed := append([]byte(nil), pt[14:46]...)
	plen := int(binary.BigEndian.Uint32(pt[46:50]))
	if 50+plen != len(pt) || start+l > PageSize {
		return Decoded{}, ErrInvalidAddress
	}
	txt, err := codec.Unpack(pt[50:], l, e.Alphabet)
	if err != nil {
		return Decoded{}, err
	}
	return Decoded{Text: txt, Start: start, Length: l, Placement: map[byte]string{1: "start", 2: "hash", 3: "center"}[pt[7]], Seed: seed}, nil
}
func (e Engine) Page(blob []byte) ([]byte, Decoded, error) {
	d, err := e.Decode(blob)
	if err != nil {
		return nil, d, err
	}
	p, err := Generate(d.Seed, "sealed-anchor-filler-v1", e.Alphabet, PageSize)
	if err != nil {
		return nil, d, err
	}
	copy(p[d.Start:d.Start+d.Length], d.Text)
	return p, d, nil
}
