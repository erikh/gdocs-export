package converters

import (
	"encoding/json"
	"os"

	"github.com/erikh/gdocs-export/pkg/downloader"
	"google.golang.org/api/docs/v1"
)

type parser struct {
	doc       *docs.Document
	manifest  downloader.Manifest
	bulletMap map[string]*Node
}

func Parse(doc *docs.Document, manifest downloader.Manifest) (*Node, error) {
	origNode := &Node{}
	node := origNode

	parser := &parser{
		doc:       doc,
		manifest:  manifest,
		bulletMap: map[string]*Node{},
	}

	for _, elem := range doc.Body.Content {
		var err error
		if node, err = parser.parseElement(elem, node); err != nil {
			return node, err
		}
	}

	if os.Getenv("DUMP_PARSE_TREE") != "" {
		enc := json.NewEncoder(os.Stderr)
		enc.SetIndent("", "  ")
		enc.Encode(origNode)
	}

	return origNode, nil
}

func (p *parser) parseElement(elem *docs.StructuralElement, origNode *Node) (*Node, error) {
	node := origNode

	if elem.Paragraph != nil {
		if elem.Paragraph.Bullet != nil {
			listID := elem.Paragraph.Bullet.ListId
			glyphType := p.doc.Lists[listID].ListProperties.NestingLevels[elem.Paragraph.Bullet.NestingLevel].GlyphType

			var listToken Token
			var bulletToken Token

			switch glyphType {
			case "DECIMAL":
				listToken = TokenOrderedList
				bulletToken = TokenOrderedBullet
			default:
				listToken = TokenUnorderedList
				bulletToken = TokenUnorderedBullet
			}

			if thisNode, ok := p.bulletMap[listID]; ok {
				node = thisNode.append(&Node{Token: bulletToken, ListNumber: len(thisNode.Children) + 1, BulletNesting: elem.Paragraph.Bullet.NestingLevel + 1})
			} else {
				node = node.append(&Node{Token: listToken, BulletNesting: elem.Paragraph.Bullet.NestingLevel + 1})
				p.bulletMap[listID] = node
				node = node.append(&Node{Token: bulletToken, ListNumber: len(node.Children) + 1, BulletNesting: elem.Paragraph.Bullet.NestingLevel + 1})
			}
		}

		node = node.append(&Node{Token: TokenParagraph})

		var headingLevel int
		switch elem.Paragraph.ParagraphStyle.NamedStyleType {
		case "HEADING_1":
			headingLevel = 1
		case "HEADING_2":
			headingLevel = 2
		case "HEADING_3":
			headingLevel = 3
		case "HEADING_4":
			headingLevel = 4
		case "HEADING_5":
			headingLevel = 5
		case "HEADING_6":
			headingLevel = 6
		}

		if headingLevel > 0 {
			node = node.append(&Node{Token: TokenHeading, Repeat: headingLevel})
		}

		code := true
		for _, pelem := range elem.Paragraph.Elements {
			if !(pelem.TextRun != nil && pelem.TextRun.TextStyle.WeightedFontFamily != nil && pelem.TextRun.TextStyle.WeightedFontFamily.FontFamily == "Consolas") {
				code = false
				break
			}
		}
		if code {
			if node.Children[len(node.Children)-1].Token == TokenCode {
				node = node.Children[len(node.Children)-1]
			} else {
				node = node.append(&Node{Token: TokenCode})
			}
		}

		for _, pelem := range elem.Paragraph.Elements {
			paraNode := node

			if pelem.TextRun != nil {
				tr := pelem.TextRun
				ts := tr.TextStyle
				if ts != nil {
					if ts.Bold {
						paraNode = paraNode.append(&Node{Token: TokenBold})
					}
					if ts.Italic {
						paraNode = paraNode.append(&Node{Token: TokenItalic})
					}
					if ts.Link != nil {
						paraNode = paraNode.append(&Node{Token: TokenLink, Url: ts.Link.Url})
					}
				}

				paraNode.append(&Node{Token: TokenPlain, Content: tr.Content})
			}

			if pelem.InlineObjectElement != nil {
				paraNode.append(&Node{Token: TokenImage, ObjectId: pelem.InlineObjectElement.InlineObjectId})
			}
		}
	}

	if elem.Table != nil {
		if err := p.parseTable(elem.Table, node); err != nil {
			return node, err
		}
	}

	return node, nil
}

func (p *parser) parseTable(table *docs.Table, node *Node) error {
	tableNode := node.append(&Node{Token: TokenTable})

	for _, row := range table.TableRows {
		rowNode := tableNode.append(&Node{Token: TokenTableRow})
		for _, cell := range row.TableCells {
			for _, elem := range cell.Content {
				cellNode := rowNode.append(&Node{Token: TokenTableCell})
				var err error
				if cellNode, err = p.parseElement(elem, cellNode); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
