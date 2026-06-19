package config

import (
	"encoding/base64"
	"fmt"
	gbcrypto "github.com/cj3636/gobabel/internal/crypto"
	"os"
)

type Config struct {
	Addr, SealMode, LogLevel string
	Sealer                   gbcrypto.Sealer
}

func Load() (Config, error) {
	c := Config{Addr: get("gobabel_ADDR", ":3000"), SealMode: get("gobabel_SEAL_MODE", "public"), LogLevel: get("gobabel_LOG_LEVEL", "info")}
	switch c.SealMode {
	case "public":
		s, e := gbcrypto.PublicSealer()
		c.Sealer = s
		return c, e
	case "private":
		raw := os.Getenv("gobabel_SEAL_KEY")
		if raw == "" {
			return c, fmt.Errorf("gobabel_SEAL_KEY required")
		}
		k, e := base64.RawURLEncoding.DecodeString(raw)
		if e != nil {
			return c, e
		}
		s, e := gbcrypto.NewSealerFromKey(k)
		c.Sealer = s
		return c, e
	default:
		return c, fmt.Errorf("invalid seal mode")
	}
}
func get(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
