package converters

import (
	"fmt"
	"html"
	"sort"
	"strings"

	"google.golang.org/api/docs/v1"
)

// Markdown generates a markdown representation from the gdocs document.
type Markdown struct {
	payload Payload
}

// NewMarkdown creates a new *Markdown for use.
func NewMarkdown() *Markdown {
	return &Markdown{}
}

// Convert performs the conversion
func (md *Markdown) Convert(p Payload) (string, error) {
	md.payload = p
	return md.convert()
}

func (md *Markdown) convert() (string, error) {
	var res string

	for _, node := range md.payload.doc.Body.Content {
		nodeRes, err := md.generateNode(node)
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

	return strings.TrimLeft(html.EscapeString(b.String()), " ")
}

func (md *Markdown) generateNode(node *docs.StructuralElement) (string, error) {
	var res string

	if node.Paragraph != nil {
		paraRes, err := md.generateParagraph(node.Paragraph)
		if err != nil {
			return res, err
		}

		res += paraRes
	}

	if node.Table != nil {
		tableRes, err := md.generateTable(node.Table)
		if err != nil {
			return res, err
		}

		res += tableRes
	}

	return res, nil
}

func (md *Markdown) generateTable(table *docs.Table) (string, error) {
	res := "<table>\n"

	for _, row := range table.TableRows {
		res += "\t<tr>\n"
		for _, cell := range row.TableCells {
			for _, node := range cell.Content {
				cellRes, err := md.generateNode(node)
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

func (md *Markdown) generateParagraph(para *docs.Paragraph) (string, error) {
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

		elemRes, err := md.generateParagraphElement(elem, first)
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

func (md *Markdown) generateParagraphElement(elem *docs.ParagraphElement, first bool) (string, error) {
	var res string

	if elem.InlineObjectElement != nil {
		obj := elem.InlineObjectElement
		if filename, ok := md.payload.manifest[obj.InlineObjectId]; ok {
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
