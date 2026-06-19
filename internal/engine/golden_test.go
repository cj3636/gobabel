package engine

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"testing"

	"github.com/cj3636/gobabel/internal/alphabet"
	"github.com/cj3636/gobabel/internal/codec"
)

const goldenSealedBlob = "AAECAwQFBgcICQoLd59KojSrbRJ9CKnRcWnZ9jfe44r3qqaRGeQ0NFMOI8GAOUbLK_i2H8klF1JxMOoSq3uSgnwlUeTP83g_qXiKJgVQm4nPgb-6_7apnRLtOKccQp4taGEy3HQmnwkApWwSZidSlffd4A"

var goldenPlaintext = []byte("The quick brown fox jumps over 13 lazy dogs.")

type goldenFields struct {
	magic         string
	flags         byte
	alphabetID    byte
	codecID       byte
	placement     byte
	pageSize      uint16
	start         uint16
	textLength    uint16
	fillerSeedLen int
	packedTextLen uint32
}

func TestGoldenSealedBlobFields(t *testing.T) {
	blob := mustGoldenBlob(t)
	pt := openGoldenPlaintext(t, blob)
	got := decodeGoldenFields(t, pt)
	want := goldenFields{
		magic:         "BF01",
		flags:         0,
		alphabetID:    1,
		codecID:       1,
		placement:     3,
		pageSize:      PageSize,
		start:         2478,
		textLength:    uint16(len(goldenPlaintext)),
		fillerSeedLen: 32,
		packedTextLen: 37,
	}
	if got != want {
		t.Fatalf("golden fields mismatch\n got: %+v\nwant: %+v", got, want)
	}
}

func TestGoldenSealedBlobReconstructsRange(t *testing.T) {
	e := Engine{Alphabet: alphabet.CodeASCII, Sealer: goldenDeterministicSealer{}}
	blob := mustGoldenBlob(t)
	page, d, err := e.Page(blob)
	if err != nil {
		t.Fatal(err)
	}
	if d.Start != 2478 || d.Length != len(goldenPlaintext) || d.Placement != "center" {
		t.Fatalf("decoded range metadata mismatch: start=%d length=%d placement=%q", d.Start, d.Length, d.Placement)
	}
	if got := page[d.Start : d.Start+d.Length]; !bytes.Equal(got, goldenPlaintext) {
		t.Fatalf("reconstructed range mismatch\n got: %q\nwant: %q", got, goldenPlaintext)
	}
}

func TestGoldenAddressStillReadable(t *testing.T) {
	e := Engine{Alphabet: alphabet.CodeASCII, Sealer: goldenDeterministicSealer{}}
	blob := mustGoldenBlob(t)
	d, err := e.Decode(blob)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(d.Text, goldenPlaintext) {
		t.Fatalf("decoded golden text mismatch\n got: %q\nwant: %q", d.Text, goldenPlaintext)
	}
}

func TestGoldenDeterministicSealingHelper(t *testing.T) {
	blob := sealGoldenBlob(t, goldenPlaintext)
	want := mustGoldenBlob(t)
	if !bytes.Equal(blob, want) {
		t.Fatalf("deterministic golden blob changed\n got: %s\nwant: %s", base64.RawURLEncoding.EncodeToString(blob), goldenSealedBlob)
	}
}

func mustGoldenBlob(t *testing.T) []byte {
	t.Helper()
	blob, err := base64.RawURLEncoding.DecodeString(goldenSealedBlob)
	if err != nil {
		t.Fatal(err)
	}
	return blob
}

func openGoldenPlaintext(t *testing.T, blob []byte) []byte {
	t.Helper()
	pt, err := (goldenDeterministicSealer{}).Open(blob, AAD())
	if err != nil {
		t.Fatal(err)
	}
	return pt
}

func decodeGoldenFields(t *testing.T, pt []byte) goldenFields {
	t.Helper()
	if len(pt) < 50 {
		t.Fatalf("golden plaintext too short: %d", len(pt))
	}
	return goldenFields{
		magic:         string(pt[:4]),
		flags:         pt[4],
		alphabetID:    pt[5],
		codecID:       pt[6],
		placement:     pt[7],
		pageSize:      binary.BigEndian.Uint16(pt[8:10]),
		start:         binary.BigEndian.Uint16(pt[10:12]),
		textLength:    binary.BigEndian.Uint16(pt[12:14]),
		fillerSeedLen: 32,
		packedTextLen: binary.BigEndian.Uint32(pt[46:50]),
	}
}

func sealGoldenBlob(t *testing.T, text []byte) []byte {
	t.Helper()
	packed, err := codec.Pack(text, alphabet.CodeASCII)
	if err != nil {
		t.Fatal(err)
	}
	seed := goldenSeed()
	var pt bytes.Buffer
	pt.Write(magic)
	pt.WriteByte(0)
	pt.WriteByte(1)
	pt.WriteByte(1)
	pt.WriteByte(3)
	_ = binary.Write(&pt, binary.BigEndian, uint16(PageSize))
	_ = binary.Write(&pt, binary.BigEndian, uint16((PageSize-len(text))/2))
	_ = binary.Write(&pt, binary.BigEndian, uint16(len(text)))
	pt.Write(seed)
	_ = binary.Write(&pt, binary.BigEndian, uint32(len(packed)))
	pt.Write(packed)
	blob, err := (goldenDeterministicSealer{}).Seal(pt.Bytes(), AAD())
	if err != nil {
		t.Fatal(err)
	}
	return blob
}

type goldenDeterministicSealer struct{}

func (goldenDeterministicSealer) Seal(pt, aad []byte) ([]byte, error) {
	gcm, err := goldenGCM()
	if err != nil {
		return nil, err
	}
	nonce := goldenNonce()
	ct := gcm.Seal(nil, nonce, pt, aad)
	return append(append([]byte(nil), nonce...), ct...), nil
}

func (goldenDeterministicSealer) Open(blob, aad []byte) ([]byte, error) {
	gcm, err := goldenGCM()
	if err != nil {
		return nil, err
	}
	return gcm.Open(nil, blob[:gcm.NonceSize()], blob[gcm.NonceSize():], aad)
}

func goldenGCM() (cipher.AEAD, error) {
	key := sha256.Sum256([]byte("gobabel golden test deterministic key v1"))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func goldenNonce() []byte { return []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11} }

func goldenSeed() []byte {
	seed := sha256.Sum256([]byte("gobabel golden test filler seed v1"))
	return seed[:]
}
