package converters

import (
	"fmt"
	"html"
	"strings"

	"github.com/erikh/gdocs-export/pkg/downloader"
	"google.golang.org/api/docs/v1"
)

// Convert converts google docs json types to string format documents in the format provided.
// Formats available: md (markdown), html
func Convert(typ string, doc *docs.Document, manifest downloader.Manifest) (string, error) {
	node, err := Parse(doc, manifest)
	if err != nil {
		return "", err
	}

	return Generate(typ, node, manifest)
}

type TagSet map[Token]Tag

var ConvertMap = map[string]TagSet{
	"md": {
		TokenPlain: Tag{
			Collapse: true,
			LeftPad:  true,
			Escape:   func(s string) string { return markdownEscape(s) },
		},
		TokenBold: Tag{
			Collapse:        true,
			LeftPad:         true,
			TrimInside:      true,
			RequiresContent: true,
			Before:          func(s string) string { return "**" + s },
			After:           func(s string) string { return s + "**" },
		},
		TokenItalic: Tag{
			TrimInside:      true,
			LeftPad:         true,
			Collapse:        true,
			RequiresContent: true,
			Before:          func(s string) string { return "_" + s },
			After:           func(s string) string { return s + "_" },
		},
		TokenParagraph: Tag{
			TrimInside: true,
			Before:     func(s string) string { return "\n" + s },
			After:      func(s string) string { return s + "\n" },
		},
		TokenUnorderedList: Tag{
			Repeat: func(times int, s string) string { return strings.Repeat("  ", times-1) + s },
		},
		TokenBullet: Tag{
			TrimInside:      true,
			RequiresContent: true,
			Before:          func(s string) string { return "* " + s },
			After:           func(s string) string { return s + "\n" },
		},
		TokenHeading: Tag{
			TrimInside:      true,
			RequiresContent: true,
			Repeat:          func(times int, s string) string { return strings.Repeat("#", times) + " " + s },
			After:           func(s string) string { return s + "\n" },
		},
		TokenTable: Tag{
			Before: func(s string) string { return "<table>" + s },
			After:  func(s string) string { return s + "</table>" },
		},
		TokenTableCell: Tag{
			TrimInside: true,
			Before:     func(s string) string { return "<td>" + s },
			After:      func(s string) string { return s + "</td>" },
		},
		TokenTableRow: Tag{
			Before: func(s string) string { return "<tr>" + s },
			After:  func(s string) string { return s + "</tr>" },
		},
		TokenImage: Tag{
			MapFile: func(file downloader.ManifestFile) string {
				return fmt.Sprintf("<img src=%q height=%d width=%d />", file.Filename, file.Height, file.Width)
			},
		},
		TokenCode: Tag{
			RequiresContent: true,
			Before: func(s string) string {
				if strings.Contains(s, "\n") {
					return "```\n" + s
				}
				return "`" + s
			},
			After: func(s string) string {
				if strings.Contains(s, "\n") {
					return s + "\n```\n"
				}
				return s + "`"
			},
		},
		TokenLink: Tag{
			LeftPad: true,
			Link: func(href, s string) string {
				return fmt.Sprintf("[%s](%s)", s, href)
			},
		},
	},
	"html": {
		TokenPlain: Tag{
			Collapse: true,
			Escape: func(s string) string {
				if len(s) > 0 {
					s = html.EscapeString(s)
					if s[0] == ' ' {
						s = "&nbsp;" + s[1:]
					}
					if s[len(s)-1] == ' ' {
						s = s[:len(s)-1] + "&nbsp;"
					}
				}

				return s
			},
		},
		TokenBold: Tag{
			Collapse:   true,
			TrimInside: true,
			Before:     func(s string) string { return "<b>" + s },
			After:      func(s string) string { return s + "</b>" },
		},
		TokenItalic: Tag{
			Collapse:        true,
			RequiresContent: true,
			TrimInside:      true,
			Before:          func(s string) string { return "<i>" + s },
			After:           func(s string) string { return s + "</i>" },
		},
		TokenParagraph: Tag{
			Collapse:        true,
			LeftPad:         true,
			TrimInside:      true,
			RequiresContent: true,
			Before:          func(s string) string { return "<p>" + s },
			After:           func(s string) string { return s + "</p>\n" },
		},
		TokenUnorderedList: Tag{
			Before: func(s string) string { return "<ul>" + s },
			After:  func(s string) string { return s + "</ul>" },
		},
		TokenBullet: Tag{
			Before: func(s string) string { return "<li>" + s },
			After:  func(s string) string { return s + "</li>" },
		},
		TokenHeading: Tag{
			TrimInside:      true,
			RequiresContent: true,
			Repeat:          func(times int, s string) string { return fmt.Sprintf("<h%d>%s</h%d>", times, s, times) },
		},
		TokenTable: Tag{
			Before: func(s string) string { return "<table>" + s },
			After:  func(s string) string { return s + "</table>" },
		},
		TokenTableCell: Tag{
			TrimInside: true,
			Before:     func(s string) string { return "<td>" + s },
			After:      func(s string) string { return s + "</td>" },
		},
		TokenTableRow: Tag{
			Before: func(s string) string { return "<tr>" + s },
			After:  func(s string) string { return s + "</tr>" },
		},
		TokenImage: Tag{
			MapFile: func(file downloader.ManifestFile) string {
				return fmt.Sprintf("<img src=%q height=%d width=%d />", file.Filename, file.Height, file.Width)
			},
		},
		TokenCode: Tag{
			TrimInside:      true,
			RequiresContent: true,
			LeftPad:         true,
			Before: func(s string) string {
				if strings.Contains(s, "\n") {
					return "\n<pre><code>" + s
				}
				return "<code>" + s
			},
			After: func(s string) string {
				if strings.Contains(s, "\n") {
					return s + "</code></pre>\n"
				}

				return s + "</code>"
			},
		},
		TokenLink: Tag{
			Link: func(href, s string) string {
				return fmt.Sprintf("<a href=%q>%s</a>", href, s)
			},
		},
	},
}
