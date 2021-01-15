package converters

import (
	"fmt"
	"html"
	"sort"
	"strings"

	"google.golang.org/api/docs/v1"
)

// HTML generates a markdown recodesentation from the gdocs document.
type HTML struct {
	payload   Payload
	nestLevel int64
	nesting   bool
	code      bool
	inPara    bool
}

// NewHTML creates a new *HTML for use.
func NewHTML() *HTML {
	return &HTML{}
}

// Convert performs the conversion
func (h *HTML) Convert(p Payload) (string, error) {
	h.payload = p
	return h.convert()
}

func (h *HTML) convert() (string, error) {
	var res string

	for _, node := range h.payload.doc.Body.Content {
		nodeRes, err := h.generateNode(node)
		if err != nil {
			return "", err
		}

		if nodeRes != "" {
			res += nodeRes
		}
	}

	return res, nil
}

func (h *HTML) generateNode(node *docs.StructuralElement) (string, error) {
	var res string

	if node.Paragraph != nil {
		paraRes, err := h.generateParagraph(node.Paragraph)
		if err != nil {
			return res, err
		}

		if !h.inPara {
			res += "<p>"
		}

		res += paraRes
	}

	if node.Table != nil {
		tableRes, err := h.generateTable(node.Table)
		if err != nil {
			return res, err
		}

		res += tableRes
	}

	return res, nil
}

func (h *HTML) generateTable(table *docs.Table) (string, error) {
	res := "<table>\n"

	for _, row := range table.TableRows {
		res += "\t<tr>\n"
		for _, cell := range row.TableCells {
			for _, node := range cell.Content {
				cellRes, err := h.generateNode(node)
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

func (h *HTML) generateParagraph(para *docs.Paragraph) (string, error) {
	var res string

	if para.Bullet == nil && h.nesting {
		res += strings.Repeat("</ul>", int(h.nestLevel)+1)
		h.nestLevel = 0
		h.nesting = false
	}

	switch para.ParagraphStyle.NamedStyleType {
	case "HEADING_1":
		res += "<h1>"
	case "HEADING_2":
		res += "<h2>"
	case "HEADING_3":
		res += "<h3>"
	case "HEADING_4":
		res += "<h4>"
	case "HEADING_5":
		res += "<h5>"
	case "HEADING_6":
		res += "<h6>"
	}

	if para.Bullet != nil {
		if para.Bullet.NestingLevel > h.nestLevel {
			res += strings.Repeat("<ul>", int(para.Bullet.NestingLevel-h.nestLevel))
			h.nestLevel = para.Bullet.NestingLevel
		} else if para.Bullet.NestingLevel < h.nestLevel {
			res += strings.Repeat("</ul>", int(h.nestLevel-para.Bullet.NestingLevel))
			h.nestLevel = para.Bullet.NestingLevel
		} else if para.Bullet.NestingLevel == 0 && !h.nesting {
			res += "<ul>"
		}

		h.nesting = true

		res += "<li>"
	}
	elems := sortedElems(para.Elements)
	sort.Sort(elems)

	for _, elem := range para.Elements {
		if elem.InlineObjectElement == nil && (elem.TextRun == nil || strings.TrimSpace(elem.TextRun.Content) == "") {
			continue
		}

		elemRes, err := h.generateParagraphElement(elem)
		if err != nil {
			return res, err
		}

		elemRes = strings.TrimRight(elemRes, " ")
		if elemRes != "" && elemRes[0] == '<' {
			res += " "
		}

		res += elemRes
	}

	if h.code {
		res += "</code>"
		h.code = false
	}

	if para.Bullet != nil {
		res += "</li>"
	}

	switch para.ParagraphStyle.NamedStyleType {
	case "HEADING_1":
		res += "</h1>"
	case "HEADING_2":
		res += "</h2>"
	case "HEADING_3":
		res += "</h3>"
	case "HEADING_4":
		res += "</h4>"
	case "HEADING_5":
		res += "</h5>"
	case "HEADING_6":
		res += "</h6>"
	}

	res = strings.Replace(res, "\u000b", "<br />", -1)
	if res != "" && res[len(res)-1] == '\n' {
		res += "</p>"
		h.inPara = false
	}

	return res, nil
}

func (h *HTML) generateParagraphElement(elem *docs.ParagraphElement) (string, error) {
	var res string

	if elem.InlineObjectElement != nil {
		obj := elem.InlineObjectElement
		if filename, ok := h.payload.manifest[obj.InlineObjectId]; ok {
			res += fmt.Sprintf("<img src=%q />", filename)
		}
	}

	if elem.TextRun == nil {
		return res, nil
	}

	ts := elem.TextRun.TextStyle
	if ts != nil {
		if ts.Italic {
			res += "<i>"
		}

		if ts.Bold {
			res += "<b>"
		}

		if ts.WeightedFontFamily != nil && ts.WeightedFontFamily.FontFamily == "Consolas" && !h.code {
			res += "<code>"
			h.code = true
		}
	}

	if elem.TextRun.TextStyle.Link != nil {
		res += fmt.Sprintf("<a href=%q>%s</a>", elem.TextRun.TextStyle.Link.Url, html.EscapeString(elem.TextRun.Content))
	} else {
		res += html.EscapeString(elem.TextRun.Content)
	}

	if ts != nil {
		if (ts.WeightedFontFamily == nil || ts.WeightedFontFamily.FontFamily != "Consolas") && h.code {
			h.code = false
			res += "</code>"
		}

		if ts.Bold {
			res += "</b>"
		}

		if ts.Italic {
			res += "</i>"
		}
	}

	return res, nil
}
