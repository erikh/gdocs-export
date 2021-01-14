package converters

import "google.golang.org/api/docs/v1"

type sortedElems []*docs.ParagraphElement

func (s sortedElems) Less(i, j int) bool {
	return s[i].StartIndex < s[j].StartIndex
}

func (s sortedElems) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s sortedElems) Len() int {
	return len(s)
}
