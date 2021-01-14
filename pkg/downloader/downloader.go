package downloader

import (
	"errors"
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
func New(client *http.Client) (*Agent, error) {
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
		if err := a.fetch(id, dir, obj.InlineObjectProperties.EmbeddedObject.ImageProperties.ContentUri); err != nil {
			return fmt.Errorf("%q: %w", id, err)
		}
	}

	return nil
}

func (a *Agent) fetch(id, dir, url string) error {
	resp, err := a.client.Get(url)
	if err != nil {
		return fmt.Errorf("while downloading: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("Status code was not 200, was %d: %v", resp.StatusCode, resp.Status)
	}

	ct := resp.Header.Get("content-type")
	if ct == "" {
		return errors.New("No content-type")
	}

	exts, err := mime.ExtensionsByType(ct)
	if err != nil {
		return fmt.Errorf("gathering extensions for content-type: %w", err)
	}

	fn := id
	if len(exts) > 0 {
		fn += exts[0]
	}

	f, err := os.Create(filepath.Join(dir, fn))
	if err != nil {
		return fmt.Errorf("could not create file %q: %w", fn, err)
	}
	defer f.Close()

	n, err := io.Copy(f, resp.Body)
	if err != nil {
		return fmt.Errorf("while copying content: %w", err)
	}

	if n != resp.ContentLength {
		return fmt.Errorf("Short read copying file")
	}

	return nil
}
