package util

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"github.com/erikh/gdocs-export/pkg/downloader"
	"google.golang.org/api/docs/v1"
)

// DownloadDoc just downloads the JSON representation of the document.
func DownloadDoc(client *http.Client, docID string) (*docs.Document, error) {
	srv, err := docs.New(client)
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve Docs client: %v", err)
	}

	doc, err := srv.Documents.Get(docID).Do()
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve data from document: %v", err)
	}

	return doc, nil
}

// DownloadAssets fetches the assets and places them in a directory, with a
// mapping file that is used in generating the final document if not done at
// the same time as the fetch.
func DownloadAssets(client *http.Client, doc *docs.Document, assetsDir string, createJSON bool) (downloader.Manifest, error) {
	var manifest downloader.Manifest

	a, err := downloader.New(client)
	if err != nil {
		return manifest, fmt.Errorf("%v", err)
	}

	if err := a.Download(assetsDir, doc); err != nil {
		return manifest, fmt.Errorf("trouble downloading to %q: %v", assetsDir, err)
	}

	if createJSON {
		json, err := a.ManifestJSON()
		if err != nil {
			return manifest, fmt.Errorf("Error marshalling manifest: %v", err)
		}

		if err := ioutil.WriteFile(filepath.Join(assetsDir, "manifest.json"), json, 0600); err != nil {
			return manifest, fmt.Errorf("Error writing manifest to %q: %v", assetsDir, err)
		}
	}

	return a.Manifest(), nil
}
