package converters

import (
	"html"
	"strings"
)

// taken from golang.org/x/tools/cmd/present2md and hacked up
func markdownEscape(s string) string {
	var b strings.Builder
	for i, r := range s {
		switch {
		case r == '#' && (i == 0 || s[i-1] == '\n'),
			r == '*',
			r == '_',
			r == '`',
			r == '<' && (i == 0 || s[i-1] != ' ') && i+1 < len(s) && s[i+1] != ' ',
			r == '[' && strings.Contains(s[i:], "]("):
			b.WriteRune('\\')
		}

		b.WriteRune(r)
	}

	return strings.TrimLeft(html.EscapeString(b.String()), " ")
}
