package crypto_test

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/cj3636/gobabel/internal/config"
	gbcrypto "github.com/cj3636/gobabel/internal/crypto"
)

func TestPublicSealerRoundTrip(t *testing.T) {
	sealer, err := gbcrypto.PublicSealer()
	if err != nil {
		t.Fatalf("PublicSealer() error = %v", err)
	}

	plaintext := []byte("public mode plaintext")
	aad := []byte("public aad")

	sealed, err := sealer.Seal(plaintext, aad)
	if err != nil {
		t.Fatalf("Seal() error = %v", err)
	}
	if bytes.Contains(sealed, plaintext) {
		t.Fatalf("sealed blob exposes plaintext")
	}

	opened, err := sealer.Open(sealed, aad)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if !bytes.Equal(opened, plaintext) {
		t.Fatalf("Open() = %q, want %q", opened, plaintext)
	}
}

func TestNewSealerFromKeyRoundTrip(t *testing.T) {
	key := fixedKey(0x10)
	sealer, err := gbcrypto.NewSealerFromKey(key)
	if err != nil {
		t.Fatalf("NewSealerFromKey() error = %v", err)
	}

	plaintext := []byte("private mode plaintext")
	aad := []byte("private aad")

	sealed, err := sealer.Seal(plaintext, aad)
	if err != nil {
		t.Fatalf("Seal() error = %v", err)
	}
	if bytes.Contains(sealed, plaintext) {
		t.Fatalf("sealed blob exposes plaintext")
	}

	opened, err := sealer.Open(sealed, aad)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if !bytes.Equal(opened, plaintext) {
		t.Fatalf("Open() = %q, want %q", opened, plaintext)
	}
}

func TestConfigLoadPrivateModeKeyValidation(t *testing.T) {
	validKey := fixedKey(0x20)

	t.Run("missing key", func(t *testing.T) {
		setPrivateModeEnv(t, "")

		_, err := config.Load()
		if err == nil || !strings.Contains(err.Error(), "gobabel_SEAL_KEY required") {
			t.Fatalf("Load() error = %v, want required key error", err)
		}
	})

	t.Run("invalid base64 key", func(t *testing.T) {
		setPrivateModeEnv(t, "not-valid!!!")

		_, err := config.Load()
		if err == nil {
			t.Fatalf("Load() error = nil, want base64 decode error")
		}
	})

	t.Run("invalid decoded key length", func(t *testing.T) {
		setPrivateModeEnv(t, base64.RawURLEncoding.EncodeToString(validKey[:31]))

		_, err := config.Load()
		if err == nil || !strings.Contains(err.Error(), "32 bytes") {
			t.Fatalf("Load() error = %v, want clear 32-byte key error", err)
		}
	})

	t.Run("valid key", func(t *testing.T) {
		setPrivateModeEnv(t, base64.RawURLEncoding.EncodeToString(validKey))

		cfg, err := config.Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg.Sealer == nil {
			t.Fatalf("Load().Sealer = nil, want configured sealer")
		}
	})
}

func TestOpenWithWrongKeyFailsWithoutPlaintext(t *testing.T) {
	plaintext := []byte("secret plaintext")
	aad := []byte("aad")

	sealer, err := gbcrypto.NewSealerFromKey(fixedKey(0x30))
	if err != nil {
		t.Fatalf("NewSealerFromKey() error = %v", err)
	}
	wrongSealer, err := gbcrypto.NewSealerFromKey(fixedKey(0x40))
	if err != nil {
		t.Fatalf("NewSealerFromKey(wrong) error = %v", err)
	}

	sealed, err := sealer.Seal(plaintext, aad)
	if err != nil {
		t.Fatalf("Seal() error = %v", err)
	}

	opened, err := wrongSealer.Open(sealed, aad)
	assertOpenFailedWithoutPlaintext(t, opened, err, plaintext)
}

func TestOpenWithFlippedSealedBlobByteFailsWithoutPlaintext(t *testing.T) {
	plaintext := []byte("secret plaintext")
	aad := []byte("aad")
	sealer := mustSealerFromKey(t, fixedKey(0x50))

	sealed, err := sealer.Seal(plaintext, aad)
	if err != nil {
		t.Fatalf("Seal() error = %v", err)
	}
	sealed[len(sealed)-1] ^= 0x01

	opened, err := sealer.Open(sealed, aad)
	assertOpenFailedWithoutPlaintext(t, opened, err, plaintext)
}

func TestOpenWithModifiedAADFailsWithoutPlaintext(t *testing.T) {
	plaintext := []byte("secret plaintext")
	sealer := mustSealerFromKey(t, fixedKey(0x60))

	sealed, err := sealer.Seal(plaintext, []byte("original aad"))
	if err != nil {
		t.Fatalf("Seal() error = %v", err)
	}

	opened, err := sealer.Open(sealed, []byte("modified aad"))
	assertOpenFailedWithoutPlaintext(t, opened, err, plaintext)
}

func TestInvalidKeyLengthsFailClearly(t *testing.T) {
	for _, length := range []int{0, 1, 16, 31, 33, 64} {
		key := make([]byte, length)
		_, err := gbcrypto.NewSealerFromKey(key)
		if err == nil {
			t.Fatalf("NewSealerFromKey(make([]byte, %d)) error = nil, want error", length)
		}
		if !strings.Contains(err.Error(), "32 bytes") {
			t.Fatalf("NewSealerFromKey(make([]byte, %d)) error = %v, want clear 32-byte key error", length, err)
		}
	}
}

func fixedKey(start byte) []byte {
	key := make([]byte, 32)
	for i := range key {
		key[i] = start + byte(i)
	}
	return key
}

func mustSealerFromKey(t *testing.T, key []byte) gbcrypto.Sealer {
	t.Helper()

	sealer, err := gbcrypto.NewSealerFromKey(key)
	if err != nil {
		t.Fatalf("NewSealerFromKey() error = %v", err)
	}
	return sealer
}

func setPrivateModeEnv(t *testing.T, key string) {
	t.Helper()

	t.Setenv("gobabel_SEAL_MODE", "private")
	t.Setenv("gobabel_SEAL_KEY", key)
}

func assertOpenFailedWithoutPlaintext(t *testing.T, opened []byte, err error, plaintext []byte) {
	t.Helper()

	if err == nil {
		t.Fatalf("Open() error = nil, want authentication failure")
	}
	if bytes.Equal(opened, plaintext) || bytes.Contains(opened, plaintext) {
		t.Fatalf("Open() returned plaintext despite error")
	}
}
