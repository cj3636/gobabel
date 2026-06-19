package address

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseRange(s string) (int, int, bool, error) {
	if !strings.Contains(s, ":") {
		return 0, 0, false, nil
	}
	p := strings.Split(s, ":")
	if len(p) != 2 {
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
