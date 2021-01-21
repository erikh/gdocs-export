package converters

import "github.com/erikh/gdocs-export/pkg/downloader"

type Tag struct {
	NoEscape        bool
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
	TokenPlain           Token = 0
	TokenBold            Token = iota
	TokenItalic          Token = iota
	TokenCode            Token = iota
	TokenParagraph       Token = iota
	TokenImage           Token = iota
	TokenUnorderedBullet Token = iota
	TokenOrderedBullet   Token = iota
	TokenHeading         Token = iota
	TokenTable           Token = iota
	TokenTableRow        Token = iota
	TokenTableCell       Token = iota
	TokenUnorderedList   Token = iota
	TokenOrderedList     Token = iota
	TokenLink            Token = iota
)
