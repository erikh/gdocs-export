package converters

import (
	"github.com/erikh/gdocs-export/pkg/downloader"
	"google.golang.org/api/docs/v1"
)

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
