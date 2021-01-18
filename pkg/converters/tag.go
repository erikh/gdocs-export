package converters

import "github.com/erikh/gdocs-export/pkg/downloader"

type Tag struct {
	SkipFirst       bool
	Collapse        bool
	RequiresContent bool
	LeftPad         bool
	TrimInside      bool
	Escape          func(string) string
	Link            func(string, string) string
	Repeat          func(int, string) string
	Before          func(string) string
	After           func(string) string
	MapFile         func(downloader.ManifestFile) string
}

type Token int

const (
	TokenPlain         = 0
	TokenBold          = iota
	TokenItalic        = iota
	TokenCode          = iota
	TokenParagraph     = iota
	TokenImage         = iota
	TokenBullet        = iota
	TokenHeading       = iota
	TokenTable         = iota
	TokenTableRow      = iota
	TokenTableCell     = iota
	TokenUnorderedList = iota
	TokenLink          = iota
)
