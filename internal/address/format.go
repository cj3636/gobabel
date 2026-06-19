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
	return base64.RawURLEncoding.DecodeString(strings.Join(segs, ""))
}
func WithRange(page string, start, end int) string { return fmt.Sprintf("%s/%d:%d", page, start, end) }
