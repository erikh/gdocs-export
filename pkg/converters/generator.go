package converters

import (
	"errors"
	"fmt"
	"strings"

	"github.com/erikh/gdocs-export/pkg/downloader"
)

func Generate(typ string, node *Node, manifest downloader.Manifest) (string, error) {
	converter, ok := ConvertMap[typ]
	if !ok {
		return "", fmt.Errorf("%q is an invalid format. Try `-c help`", typ)
	}

	var res string
	tag, ok := converter[node.Token]
	if !ok {
		return "", fmt.Errorf("Parser is broken: missing handler for token %q", node.Token)
	}

	if node.ObjectId != "" {
		if tag.MapFile != nil {
			filename, ok := manifest[node.ObjectId]
			if !ok {
				return "", nil
			}

			return tag.MapFile(filename), nil
		}
		return "", errors.New("filename was yielded yet no handler could be found for the token")
	}

	res = node.Content

	var (
		lastSib *Node
	)

	for _, sib := range node.Children {
		tmp, err := Generate(typ, sib, manifest)
		if err != nil {
			return "", err
		}

		sibConv := converter[sib.Token]

		if lastSib != nil && lastSib.Token != sib.Token && sibConv.LeftPad && (res == "" || res[len(res)-1] != ' ' || res[len(res)-1] != '\n') &&
			tmp != "" && tmp[0] != ' ' && tmp[0] != '\n' {
			res += " "
		}

		if sibConv.Escape != nil {
			tmp = sibConv.Escape(tmp)
		}

		res += tmp
		lastSib = sib
	}

	if tag.TrimInside {
		res = strings.TrimSpace(res)
	}

	if tag.RequiresContent && res == "" {
		// do not add before/after tags to empty content
		return "", nil
	}

	if tag.Repeat != nil {
		if node.Repeat > 0 {
			res = tag.Repeat(node.Repeat, res)
		} else if node.BulletNesting > 0 {
			res = tag.Repeat(int(node.BulletNesting), res)
		}
	}

	var lastChild *Node

	if len(node.Children) > 1 {
		lastChild = node.Children[len(node.Children)-2]
	}

	if tag.Before != nil {
		switch {
		case !tag.SkipFirst && !tag.Collapse:
			res = tag.Before(res)
		case tag.Collapse && (lastChild == nil || lastChild.Token != node.Token):
			res = tag.Before(res)
		case !tag.SkipFirst && lastChild == nil:
			res = tag.Before(res)
		}
	}

	if tag.After != nil && (!tag.Collapse || (node.parent != nil && node.parent.Token != node.Token)) {
		switch {
		case !tag.SkipFirst && !tag.Collapse:
			res = tag.After(res)
		case tag.Collapse && (lastChild == nil || lastChild.Token != node.Token):
			res = tag.After(res)
		case !tag.SkipFirst && (lastChild == nil || lastChild.Token == node.Token):
			res = tag.After(res)
		}
	}

	if tag.Link != nil && node.Url != "" {
		res = tag.Link(node.Url, res)
	}

	return res, nil
}