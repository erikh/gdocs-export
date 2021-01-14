package downloader

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"google.golang.org/api/docs/v1"
)

// Agent is a new downloader agent. It will ingest the document's assets and
// download them.
type Agent struct {
	client *http.Client
}

// New creates a new agent for use. The HTTP client provided must have oauth2
// capabilities.
func New(log bool, client *http.Client) (*Agent, error) {
	return &Agent{
		client: client,
	}, nil
}

// Download downloads all the assets to the directory that are contained the
// doc.
func (a *Agent) Download(dir string, doc *docs.Document) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("While making directory %q: %w", dir, err)
	}

	for id, obj := range doc.InlineObjects {
		resp, err := a.client.Get(obj.InlineObjectProperties.EmbeddedObject.ImageProperties.ContentUri)
		if err != nil {
			return fmt.Errorf("while downloading %q: %w", id, err)
		}

		if resp.StatusCode != 200 {
			return fmt.Errorf("Status code for %q was not 200, was %d: %v", id, resp.StatusCode, resp.Status)
		}

		ct := resp.Header.Get("content-type")
		if ct == "" {
			return fmt.Errorf("No content-type for id %q", id)
		}

		exts, err := mime.ExtensionsByType(ct)
		if err != nil {
			return fmt.Errorf("Content-type for id %q: %w", id, err)
		}

		fn := id
		if len(exts) > 0 {
			fn += exts[0]
		}

		f, err := os.Create(filepath.Join(dir, fn))
		if err != nil {
			return fmt.Errorf("could not create file %q: %w", fn, err)
		}

		n, err := io.Copy(f, resp.Body)
		if err != nil {
			f.Close()
			return fmt.Errorf("%q: while copying content: %w", id, err)
		}

		if n != resp.ContentLength {
			f.Close()
			return fmt.Errorf("Short read copying %q: %w", id, err)
		}

		f.Close()
		resp.Body.Close()
	}

	return nil
}
