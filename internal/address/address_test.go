package address

import (
	"strings"
	"testing"
)

func TestSplitBookPath(t *testing.T) {
	valid := []struct {
		name      string
		path      string
		wantRange bool
	}{
		{"page", "/v1/book/bf1/QUJD/RA", false},
		{"range", "/v1/book/bf1/QUJD/RA/0:4", true},
		{"max range", "/v1/book/bf1/QUJD/RA/0:5000", true},
	}
	for _, tc := range valid {
		t.Run(tc.name, func(t *testing.T) {
			segs, rg, err := SplitBookPath(tc.path)
			if err != nil {
				t.Fatalf("SplitBookPath() error = %v", err)
			}
			if len(segs) == 0 {
				t.Fatal("missing sealed segments")
			}
			if (rg != nil) != tc.wantRange {
				t.Fatalf("range present = %v, want %v", rg != nil, tc.wantRange)
			}
		})
	}

	invalid := []struct {
		name string
		path string
	}{
		{"empty sealed segment", "/v1/book/bf1/QUJD//RA"},
		{"missing sealed payload", "/v1/book/bf1"},
		{"multiple range suffixes", "/v1/book/bf1/QUJD/0:1/1:2"},
		{"malformed range suffix", "/v1/book/bf1/QUJD/1:"},
		{"negative range", "/v1/book/bf1/QUJD/-1:2"},
		{"range start after end", "/v1/book/bf1/QUJD/3:2"},
		{"range past page size", "/v1/book/bf1/QUJD/0:5001"},
		{"too long", "/v1/book/bf1/" + strings.Repeat("A", MaxAddressLength)},
	}
	for _, tc := range invalid {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := SplitBookPath(tc.path); err == nil {
				t.Fatal("SplitBookPath() error = nil, want error")
			}
		})
	}
}

func TestDecodeSegments(t *testing.T) {
	valid := [][]string{{"QUJD", "RA"}, {"abc-DEF_012"}}
	for _, segs := range valid {
		if _, err := DecodeSegments(segs); err != nil {
			t.Fatalf("DecodeSegments(%v) error = %v", segs, err)
		}
	}

	invalid := []struct {
		name string
		segs []string
	}{
		{"missing", nil},
		{"empty", []string{"QUJD", ""}},
		{"plus", []string{"QU+JD"}},
		{"slash", []string{"QU/JD"}},
		{"padding", []string{"QUJD="}},
		{"other non base64url", []string{"QUJD!"}},
	}
	for _, tc := range invalid {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := DecodeSegments(tc.segs); err == nil {
				t.Fatal("DecodeSegments() error = nil, want error")
			}
		})
	}
}
