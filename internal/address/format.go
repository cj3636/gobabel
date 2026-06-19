package address

import (
	"encoding/base64"
	"fmt"
	"strings"
)

func Encode(version string, blob []byte) string {
	s := base64.RawURLEncoding.EncodeToString(blob)
	parts := []string{"", "v1", "book", version}
	for len(s) > 0 {
		n := 10
		if len(s) < n {
			n = len(s)
		}
		parts = append(parts, s[:n])
		s = s[n:]
	}
	return strings.Join(parts, "/")
}
func DecodeSegments(segs []string) ([]byte, error) {
	if len(segs) == 0 {
		return nil, fmt.Errorf("missing_payload")
	}
	for _, s := range segs {
		if s == "" {
			return nil, fmt.Errorf("empty_segment")
		}
		for _, r := range s {
			if !(r >= 'A' && r <= 'Z') && !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9') && r != '-' && r != '_' {
				return nil, fmt.Errorf("invalid_base64url")
			}
		}
	}
	return base64.RawURLEncoding.DecodeString(strings.Join(segs, ""))
}
func WithRange(page string, start, end int) string { return fmt.Sprintf("%s/%d:%d", page, start, end) }
