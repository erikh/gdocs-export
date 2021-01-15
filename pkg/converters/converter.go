package converters

import (
	"fmt"

	"github.com/erikh/gdocs-export/pkg/downloader"
	"google.golang.org/api/docs/v1"
)

// Convert converts google docs json types to string format documents in the format provided.
// Formats available: md (markdown), html (coming soon)
func Convert(typ string, doc *docs.Document, manifest downloader.Manifest) (string, error) {
	pl := NewPayload(doc, manifest)

	switch typ {
	case "md":
		res, err := NewMarkdown().Convert(pl)
		if err != nil {
			return "", fmt.Errorf("Unable to produce markdown: %v", err)
		}

		return res, nil
	default:
		return "", fmt.Errorf("%q is an invalid format. Try `-c help`", typ)
	}
}

// Payload is the document structure + any downloaded assets
type Payload struct {
	doc      *docs.Document
	manifest downloader.Manifest
}

// NewPayload creates a new payload for use.
func NewPayload(doc *docs.Document, manifest downloader.Manifest) Payload {
	return Payload{doc, manifest}
}

// Converter is the interface to different types of converters
type Converter interface {
	Convert(Payload) (string, error)
}
