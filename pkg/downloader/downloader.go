package downloader

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"google.golang.org/api/docs/v1"
)

// ManifestFile is a file with size data.
type ManifestFile struct {
	Filename string
	Height   int64
	Width    int64
}

// Manifest is just an object ID -> filename mapping.
type Manifest map[string]ManifestFile

// Agent is a new downloader agent. It will ingest the document's assets and
// download them.
type Agent struct {
	client    *http.Client
	idFileMap Manifest
}

// New creates a new agent for use. The HTTP client provided must have oauth2
// capabilities.
func New(client *http.Client) (*Agent, error) {
	return &Agent{
		client:    client,
		idFileMap: Manifest{},
	}, nil
}

// Download downloads all the assets to the directory that are contained the
// doc.
func (a *Agent) Download(dir string, doc *docs.Document) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("While making directory %q: %w", dir, err)
	}

	for id, obj := range doc.InlineObjects {
		if err := a.fetch(obj, dir); err != nil {
			return fmt.Errorf("%q: %w", id, err)
		}
	}

	return nil
}

// Manifest returns the manifest
func (a *Agent) Manifest() Manifest {
	return a.idFileMap
}

// ManifestJSON returns the manifest is JSON form
func (a *Agent) ManifestJSON() ([]byte, error) {
	return json.Marshal(a.idFileMap)
}

func (a *Agent) fetch(obj docs.InlineObject, dir string) error {
	if obj.InlineObjectProperties == nil || obj.InlineObjectProperties.EmbeddedObject == nil {
		return nil
	}

	resp, err := a.client.Get(obj.InlineObjectProperties.EmbeddedObject.ImageProperties.ContentUri)
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

	fn := obj.ObjectId
	if len(exts) > 0 {
		fn += exts[0]
	}

	fn = filepath.Join(dir, fn)

	f, err := os.Create(fn)
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

	a.idFileMap[obj.ObjectId] = ManifestFile{
		Filename: fn,
		Height:   int64(obj.InlineObjectProperties.EmbeddedObject.Size.Height.Magnitude),
		Width:    int64(obj.InlineObjectProperties.EmbeddedObject.Size.Width.Magnitude),
	}

	return nil
}
