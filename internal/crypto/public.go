package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
)

type gcmSealer struct{ key [32]byte }

// PublicSealer returns the v1 public-mode sealer. v1 sealed addresses
// intentionally use AES-256-GCM with a 32-byte key; the address path version
// and authenticated data bind ciphertexts to bf1 so future crypto suites can
// move to a new address version without reinterpreting existing addresses.
func PublicSealer() (Sealer, error) {
	k := sha256.Sum256([]byte("gobabel public deterministic sealed-address key v1"))
	return gcmSealer{key: k}, nil
}

// NewSealerFromKey returns the v1 private-mode sealer. The supplied 32-byte
// key is used directly as an AES-256-GCM key for bf1 sealed addresses.
func NewSealerFromKey(k []byte) (Sealer, error) {
	if len(k) != 32 {
		return nil, errors.New("seal key must be 32 bytes")
	}
	var a [32]byte
	copy(a[:], k)
	return gcmSealer{key: a}, nil
}
func (s gcmSealer) Seal(pt, aad []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.key[:])
	if err != nil {
		return nil, err
	}
	a, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, a.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	ct := a.Seal(nil, nonce, pt, aad)
	return append(nonce, ct...), nil
}
func (s gcmSealer) Open(blob, aad []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.key[:])
	if err != nil {
		return nil, err
	}
	a, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(blob) < a.NonceSize() {
		return nil, errors.New("short blob")
	}
	return a.Open(nil, blob[:a.NonceSize()], blob[a.NonceSize():], aad)
}
