package crypto

type Sealer interface {
	Seal([]byte, []byte) ([]byte, error)
	Open([]byte, []byte) ([]byte, error)
}
