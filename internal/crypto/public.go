package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
)

type gcmSealer struct{ key [32]byte }

func PublicSealer() (Sealer, error) {
	k := sha256.Sum256([]byte("gobabel public deterministic sealed-address key v1"))
	return gcmSealer{key: k}, nil
}
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
