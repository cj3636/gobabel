package address

import (
	"fmt"
	"strconv"
	"strings"
)

const MaxAddressLength = 2048

func ParseRange(s string) (int, int, bool, error) {
	if !strings.Contains(s, ":") {
		return 0, 0, false, nil
	}
	if strings.Count(s, ":") != 1 {
		return 0, 0, true, fmt.Errorf("bad range")
	}
	p := strings.Split(s, ":")
	if len(p) != 2 || p[0] == "" || p[1] == "" {
		return 0, 0, true, fmt.Errorf("bad range")
	}
	a, e := strconv.Atoi(p[0])
	if e != nil {
		return 0, 0, true, e
	}
	b, e := strconv.Atoi(p[1])
	if e != nil {
		return 0, 0, true, e
	}
	if a < 0 || b < a || b > 5000 {
		return 0, 0, true, fmt.Errorf("invalid range")
	}
	return a, b, true, nil
}

func SplitBookPath(path string) ([]string, *[2]int, error) {
	if len(path) > MaxAddressLength {
		return nil, nil, fmt.Errorf("address_too_long")
	}
	trimmed := strings.TrimPrefix(path, "/v1/book/")
	parts := strings.Split(trimmed, "/")
	for _, p := range parts {
		if p == "" {
			return nil, nil, fmt.Errorf("empty_segment")
		}
	}
	if len(parts) < 2 || parts[0] != Version {
		return nil, nil, fmt.Errorf("unsupported_version")
	}
	for i, p := range parts[1:] {
		if strings.Contains(p, ":") && i != len(parts[1:])-1 {
			return nil, nil, fmt.Errorf("invalid_range")
		}
	}
	var rg *[2]int
	if s, e, ok, err := ParseRange(parts[len(parts)-1]); ok {
		if err != nil {
			return nil, nil, fmt.Errorf("invalid_range")
		}
		rg = &[2]int{s, e}
		parts = parts[:len(parts)-1]
	}
	if len(parts) < 2 {
		return nil, nil, fmt.Errorf("missing_payload")
	}
	return parts[1:], rg, nil
}
