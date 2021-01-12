package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"google.golang.org/api/docs/v1"
)

func main() {
	var doc docs.Document

	if err := json.NewDecoder(os.Stdin).Decode(&doc); err != nil {
		panic(err)
	}

	res, err := generateMD(&doc)
	if err != nil {
		panic(err)
	}

	fmt.Println(res)
}

func generateMD(doc *docs.Document) (string, error) {
	var res string

	for _, node := range doc.Body.Content {
		nodeRes, err := generateNode(node, node.StartIndex, node.EndIndex)
		if err != nil {
			return "", err
		}

		res += nodeRes
	}

	return res, nil
}

func generateNode(node *docs.StructuralElement, start, end int64) (string, error) {
	var res string

	if node.Paragraph != nil {
		switch node.Paragraph.ParagraphStyle.NamedStyleType {
		case "HEADING_1":
			res += "# "
		case "HEADING_2":
			res += "## "
		case "HEADING_3":
			res += "### "
		case "HEADING_4":
			res += "#### "
		case "HEADING_5":
			res += "##### "
		case "HEADING_6":
			res += "###### "
		}

		if node.Paragraph.Bullet != nil {
			res += strings.Repeat("  ", int(node.Paragraph.Bullet.NestingLevel))
			res += "* "
		}

		elems := sortedElems(node.Paragraph.Elements)
		sort.Sort(elems)

		for _, elem := range node.Paragraph.Elements {
			elemRes, err := generateParagraphElement(elem)
			if err != nil {
				return res, err
			}

			res += elemRes
		}

		if node.Paragraph.Bullet == nil {
			res += "\n"
		}
	}

	return res, nil
}

func generateParagraphElement(elem *docs.ParagraphElement) (string, error) {
	var res string

	if elem.TextRun != nil {
		if elem.TextRun.TextStyle != nil {
			if elem.TextRun.TextStyle.Italic {
				res += "_"
			}
			if elem.TextRun.TextStyle.Bold {
				res += "**"
			}
		}

		res += elem.TextRun.Content

		if elem.TextRun.TextStyle != nil {
			if elem.TextRun.TextStyle.Bold {
				res += "**"
			}
			if elem.TextRun.TextStyle.Italic {
				res += "_"
			}
		}
	}

	return res, nil
}

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
