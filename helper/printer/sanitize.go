package printer

import (
	"fmt"
	"slices"
	"strings"
)

// Sanitize escapes control characters in s, except the ones in keep, so that
// text received from a guest cannot drive the terminal.
func Sanitize(s string, keep ...rune) string {
	var b strings.Builder
	for _, r := range s {
		if isControl(r) && !slices.Contains(keep, r) {
			fmt.Fprintf(&b, "\\x%02x", r)
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func isControl(r rune) bool {
	return r < 0x20 || r == 0x7f || (r >= 0x80 && r <= 0x9f)
}
