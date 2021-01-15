package converters

import (
	"fmt"
	"sort"
	"strings"

	"github.com/erikh/gdocs-export/pkg/downloader"
	"google.golang.org/api/docs/v1"
)

// Markdown generates a markdown representation from the gdocs document.
func Markdown(doc *docs.Document, manifest downloader.Manifest) (string, error) {
	var res string

	for _, node := range doc.Body.Content {
		nodeRes, err := generateNode(node, manifest)
		if err != nil {
			return "", err
		}

		res += nodeRes
	}

	return res, nil
}

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

	return strings.TrimLeft(b.String(), " ")
}

func generateNode(node *docs.StructuralElement, manifest downloader.Manifest) (string, error) {
	var res string

	if node.Paragraph != nil {
		paraRes, err := generateParagraph(node.Paragraph, manifest)
		if err != nil {
			return res, err
		}

		res += paraRes
	}

	if node.Table != nil {
		tableRes, err := generateTable(node.Table, manifest)
		if err != nil {
			return res, err
		}

		res += tableRes
	}

	return res, nil
}

func generateTable(table *docs.Table, manifest downloader.Manifest) (string, error) {
	res := "<table>\n"

	for _, row := range table.TableRows {
		res += "\t<tr>\n"
		for _, cell := range row.TableCells {
			for _, node := range cell.Content {
				cellRes, err := generateNode(node, manifest)
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

func generateParagraph(para *docs.Paragraph, manifest downloader.Manifest) (string, error) {
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

	first := true

	for _, elem := range para.Elements {
		if elem.InlineObjectElement == nil && (elem.TextRun == nil || strings.TrimSpace(elem.TextRun.Content) == "") {
			continue
		}

		elemRes, err := generateParagraphElement(elem, first, manifest)
		if err != nil {
			return res, err
		}

		if !first && strings.TrimSpace(elemRes) != "" {
			res += " "
		}

		res += elemRes
		first = false
	}

	res += "\n"

	return res, nil
}

func generateParagraphElement(elem *docs.ParagraphElement, first bool, manifest downloader.Manifest) (string, error) {
	var res string

	if elem.InlineObjectElement != nil {
		obj := elem.InlineObjectElement
		if filename, ok := manifest[obj.InlineObjectId]; ok {
			res += fmt.Sprintf("<img src=%q />", filename)
		}
	}

	if elem.TextRun == nil {
		return res, nil
	}

	if elem.TextRun != nil {
		ts := elem.TextRun.TextStyle
		if ts != nil {
			if ts.Italic {
				res += "_"
			}

			if ts.Bold {
				res += "**"
			}

		}

		if elem.TextRun.TextStyle.Link != nil {
			res += "[" + markdownEscape(elem.TextRun.Content) + "](" + elem.TextRun.TextStyle.Link.Url + ")"
		} else {
			res += markdownEscape(strings.Replace(elem.TextRun.Content, "\u000b", "\n\n", -1))
		}

		if ts != nil {
			if ts.Bold {
				res = strings.TrimRight(res, " ")
				res += "**"
			}

			if ts.Italic {
				res = strings.TrimRight(res, " ")
				res += "_"
			}
		}
	}

	return res, nil
}
