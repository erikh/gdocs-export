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
		paraRes, err := generateParagraph(node.Paragraph)
		if err != nil {
			return res, err
		}

		res += paraRes
	}

	if node.Table != nil {
		tableRes, err := generateTable(node.Table)
		if err != nil {
			return res, err
		}

		res += tableRes
	}

	return res, nil
}

func generateTable(table *docs.Table) (string, error) {
	res := "<table>\n"

	for _, row := range table.TableRows {
		res += "\t<tr>\n"
		for _, cell := range row.TableCells {
			for _, node := range cell.Content {
				cellRes, err := generateNode(node, cell.StartIndex, cell.EndIndex)
				if err != nil {
					return res, err
				}

				res += "\t\t<td>" + strings.TrimSpace(cellRes) + "</td>\n"
			}
		}

		res += "\t</tr>\n"
	}

	res += "</table>\n"

	return res, nil
}

func generateParagraph(para *docs.Paragraph) (string, error) {
	var res string
	switch para.ParagraphStyle.NamedStyleType {
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

	if para.Bullet != nil {
		res += strings.Repeat("  ", int(para.Bullet.NestingLevel))
		res += "* "
	}

	elems := sortedElems(para.Elements)
	sort.Sort(elems)

	for _, elem := range para.Elements {
		elemRes, err := generateParagraphElement(elem)
		if err != nil {
			return res, err
		}

		res += elemRes
	}

	if para.Bullet == nil {
		res += "\n"
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
