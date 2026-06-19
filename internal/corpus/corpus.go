package corpus

import (
	"crypto/sha256"
	"github.com/cj3636/gobabel/internal/alphabet"
	"github.com/cj3636/gobabel/internal/engine"
)

const AddressType = "canonical-corpus-v1"

func Page(c Coordinate, a alphabet.Alphabet) ([]byte, error) {
	h := sha256.Sum256([]byte("corpus-v1" + a.ID + c.Normalize()))
	return engine.Generate(h[:], "corpus-page-v1", a, engine.PageSize)
}
