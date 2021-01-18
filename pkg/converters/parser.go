package converters

import (
	"encoding/json"
	"os"

	"github.com/erikh/gdocs-export/pkg/downloader"
	"google.golang.org/api/docs/v1"
)

func Parse(doc *docs.Document, manifest downloader.Manifest) (*Node, error) {
	node := &Node{}

	for _, elem := range doc.Body.Content {
		if err := parseElement(elem, node); err != nil {
			return node, err
		}
	}

	if os.Getenv("DUMP_PARSE_TREE") != "" {
		enc := json.NewEncoder(os.Stderr)
		enc.SetIndent("", "  ")
		enc.Encode(node)
	}

	return node, nil
}

func parseElement(elem *docs.StructuralElement, origNode *Node) error {
	node := origNode

	if elem.Paragraph != nil {
		if elem.Paragraph.Bullet != nil {
			if len(node.Children) != 0 {
				var lastChild *Node

				if elem.Paragraph.Bullet.NestingLevel > 0 {
					if len(node.Children) > 0 && node.Children[len(node.Children)-1].Token == TokenUnorderedList {
						lastChild = node.Children[len(node.Children)-1]
					}
				}

				if lastChild == nil {
					node = node.append(&Node{Token: TokenUnorderedList, BulletNesting: elem.Paragraph.Bullet.NestingLevel + 1})
				} else {
					node = lastChild
					if lastChild.BulletNesting < elem.Paragraph.Bullet.NestingLevel+1 {
						for i := (elem.Paragraph.Bullet.NestingLevel + 1) - lastChild.BulletNesting; i >= 1; i-- {
							node = node.append(&Node{Token: TokenUnorderedList, BulletNesting: i})
						}
					}
				}

				node = node.append(&Node{Token: TokenBullet, BulletNesting: elem.Paragraph.Bullet.NestingLevel + 1})
			} else {
				node = node.append(&Node{Token: TokenUnorderedList, BulletNesting: elem.Paragraph.Bullet.NestingLevel + 1})
				node = node.append(&Node{Token: TokenBullet, BulletNesting: elem.Paragraph.Bullet.NestingLevel + 1})
			}
		} else {
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
			} else {
				node = node.append(&Node{Token: TokenParagraph})
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
			} else if pelem.InlineObjectElement != nil {
				paraNode.append(&Node{Token: TokenImage, ObjectId: pelem.InlineObjectElement.InlineObjectId})
			}
		}
	}

	if elem.Table != nil {
		if err := parseTable(elem.Table, origNode); err != nil {
			return err
		}
	}

	return nil
}

func parseTable(table *docs.Table, node *Node) error {
	tableNode := node.append(&Node{Token: TokenTable})

	for _, row := range table.TableRows {
		rowNode := tableNode.append(&Node{Token: TokenTableRow})
		for _, cell := range row.TableCells {
			for _, elem := range cell.Content {
				cellNode := rowNode.append(&Node{Token: TokenTableCell})

				if err := parseElement(elem, cellNode); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
